package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/clinicmanager/shared/logger"
	"github.com/clinicmanager/shared/middleware"
	"github.com/clinicmanager/shared/setting"

	gwMiddleware "github.com/clinicmanager/services/gateway/middleware"
	"github.com/golang-jwt/jwt/v5"
)

type authClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

type route struct {
	prefix string
	proxy  *httputil.ReverseProxy
	noAuth bool // skip JWT validation
}

func main() {
	env, err := setting.LoadAndValidateEnv([]string{
		"GATEWAY_PORT",
		"ROUTER_URL",
		"USER_SVC_URL",
		"JWT_SECRET",
	})
	if err != nil {
		logger.Error(context.Background(), "failed to load gateway settings", "error", err)
		os.Exit(1)
	}

	port := env["GATEWAY_PORT"]
	jwtSecret := env["JWT_SECRET"]

	routes := []*route{
		{prefix: "/graphql", proxy: proxyFor(env["ROUTER_URL"]), noAuth: true},

		// Auth-free REST endpoints (login, register, etc.)
		{prefix: "/health", noAuth: true},
		{prefix: "/login", proxy: proxyFor(env["USER_SVC_URL"]), noAuth: true},
		{prefix: "/register", proxy: proxyFor(env["USER_SVC_URL"]), noAuth: true},
		{prefix: "/invite-accept", proxy: proxyFor(env["USER_SVC_URL"]), noAuth: true},
		{prefix: "/verify", proxy: proxyFor(env["USER_SVC_URL"]), noAuth: true},
		{prefix: "/forgot-password", proxy: proxyFor(env["USER_SVC_URL"]), noAuth: true},
		{prefix: "/reset-password", proxy: proxyFor(env["USER_SVC_URL"]), noAuth: true},
		{prefix: "/refresh-token", proxy: proxyFor(env["USER_SVC_URL"]), noAuth: true},

		// Auth-required REST endpoints (JWT validated at gateway)
		{prefix: "/change-password", proxy: proxyFor(env["USER_SVC_URL"]), noAuth: false},
		{prefix: "/logout", proxy: proxyFor(env["USER_SVC_URL"]), noAuth: false},
		{prefix: "/logout-all", proxy: proxyFor(env["USER_SVC_URL"]), noAuth: false},
	}

	mux := http.NewServeMux()
	for _, r := range routes {
		route := r
		if route.prefix == "/health" {
			mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})
			continue
		}
		mux.HandleFunc(route.prefix, routeHandler(route, jwtSecret))
	}

	var handler http.Handler = mux
	handler = gwMiddleware.NewRateLimiter(100, time.Minute).Limit(handler)
	handler = middleware.LoggingMiddleware(handler)
	handler = middleware.RecoveryMiddleware(handler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info(context.Background(), "starting gateway", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), "listen error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info(context.Background(), "shutting down gateway")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error(context.Background(), "shutdown error", "error", err)
	}
}

func routeHandler(rt *route, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		setCORS(w, req)

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		claims, err := validateJWT(req, jwtSecret)
		if err == nil && claims != nil {
			req.Header.Set("x-user-id", claims.UserID)
		} else if !rt.noAuth {
			logger.Warn(req.Context(), "auth failed", "error", err)
			http.Error(w, `{"errors":[{"message":"unauthorized"}]}`, http.StatusUnauthorized)
			return
		}

		rt.proxy.ServeHTTP(w, req)
	}
}

func validateJWT(r *http.Request, jwtSecret string) (*authClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, nil
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader {
		return nil, nil
	}
	token, err := jwt.ParseWithClaims(tokenStr, &authClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*authClaims)
	if !ok || !token.Valid {
		return nil, nil
	}
	return claims, nil
}

func proxyFor(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		logger.Error(context.Background(), "invalid target URL", "target", target, "error", err)
		os.Exit(1)
	}
	p := httputil.NewSingleHostReverseProxy(u)
	// Strip any CORS headers the upstream sends so the gateway's
	// setCORS headers (already written before the proxy call) are not
	// replaced or duplicated by the upstream response.
	p.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Allow-Credentials")
		resp.Header.Del("Access-Control-Max-Age")
		resp.Header.Del("Access-Control-Expose-Headers")
		return nil
	}
	return p
}

var allowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:4200",
	"http://localhost:5173",
	"http://localhost:8080",
	"http://127.0.0.1:4200",
	"https://studio.apollographql.com",
}

func setCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-tenant-id, x-branch-id, x-user-id, x-app-id, apollo-require-preflight, x-apollo-operation-name")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func handleCORS(w http.ResponseWriter, r *http.Request) {
	setCORS(w, r)
	w.WriteHeader(http.StatusOK)
}
