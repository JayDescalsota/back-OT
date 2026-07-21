package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"

	"github.com/clinicmanager/services/user/graph"
	"github.com/clinicmanager/services/user/graph/generated"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/services/user/service"
	sharedCache "github.com/clinicmanager/shared/cache"
	sharedCtx "github.com/clinicmanager/shared/context"
	sharedDB "github.com/clinicmanager/shared/db"
	"github.com/clinicmanager/shared/logger"
	"github.com/clinicmanager/shared/middleware"
	"github.com/clinicmanager/shared/setting"
	"github.com/clinicmanager/shared/tools"
)

type SMTPMailer struct{}

func (m SMTPMailer) Send(to, subject, body string) error {
	return tools.SendEmail(to, subject, body)
}

func main() {
	env, errorEnv := setting.LoadAndValidateEnv([]string{
		"USER_PORT",
		"USERDB_URL",
		"JWT_SECRET",
		"BASE_URL",
		"REDIS_ADDR",
	})
	if errorEnv != nil {
		logger.Error(context.Background(), "failed to load user service settings", "missing", errorEnv)
		os.Exit(1)
	}
	port := env["USER_PORT"]
	DbUrl := env["USERDB_URL"]
	jwtSecret := env["JWT_SECRET"]
	baseURL := env["BASE_URL"] + ":" + env["USER_PORT"]

	db, err := sharedDB.NewDB(DbUrl)
	if err != nil {
		logger.Error(context.Background(), "failed to connect to user database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	dbSet := sharedDB.NewDBSet(db)
	userRepo := repository.NewUserRepo(dbSet)

	redisClient, redisErr := sharedCache.NewRedisClient(env["REDIS_ADDR"], env["REDIS_PASSWORD"])
	if redisErr != nil {
		logger.Error(context.Background(), "failed to connect to redis", "error", redisErr)
		os.Exit(1)
	}
	defer redisClient.Close()

	const (
		accessTokenTTL       = 15 * time.Minute
		refreshTokenTTL      = 7 * 24 * time.Hour
		superAdminAccessTTL  = 5 * time.Minute
		superAdminRefreshTTL = 24 * time.Hour
	)
	authService := service.NewAuthService(userRepo, jwtSecret, SMTPMailer{}, baseURL, accessTokenTTL, refreshTokenTTL)

	isSessionRevoked := func(ctx context.Context, sessionID string) (bool, error) {
		id, err := strconv.ParseInt(sessionID, 10, 64)
		if err != nil {
			return true, err
		}
		session, err := userRepo.FindSessionByID(ctx, id)
		if err != nil {
			return true, err
		}
		if session == nil || session.Revoked {
			return true, nil
		}
		return false, nil
	}

	userService := service.NewUserService(userRepo, sharedCtx.UserIDFromCtx, redisClient)
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{UserService: userService},
	}))

	loginLimiter := middleware.NewRateLimiter(10, time.Minute)
	registerLimiter := middleware.NewRateLimiter(5, time.Minute)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.Handle("/graphql", sharedCtx.Tenant(srv))

	mux.Handle("POST /register", registerLimiter.Limit(registerHandler(authService)))
	mux.Handle("POST /login", loginLimiter.Limit(loginHandler(authService)))
	mux.HandleFunc("GET /verify", verifyHandler(authService))
	mux.Handle("POST /change-password", middleware.AuthMiddleware(jwtSecret, isSessionRevoked, changePasswordHandler(authService)))
	mux.HandleFunc("POST /logout", logoutHandler(authService))
	mux.Handle("POST /logout-all", middleware.AuthMiddleware(jwtSecret, isSessionRevoked, logoutAllHandler(authService)))
	mux.HandleFunc("POST /refresh-token", refreshTokenHandler(authService))
	mux.HandleFunc("POST /forgot-password", forgotPasswordHandler(authService))
	mux.HandleFunc("POST /reset-password", resetPasswordHandler(authService))

	recovery := middleware.RecoveryMiddleware(middleware.LoggingMiddleware(mux))
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      recovery,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		logger.Info(context.Background(), "user service starting", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), "listen error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info(context.Background(), "shutting down user service")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error(context.Background(), "shutdown error", "error", err)
	}
}
