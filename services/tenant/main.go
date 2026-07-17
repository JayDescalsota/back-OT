package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"

	"github.com/clinicmanager/services/tenant/graph"
	"github.com/clinicmanager/services/tenant/graph/generated"
	"github.com/clinicmanager/services/tenant/repository"
	"github.com/clinicmanager/services/tenant/service"
	sharedCtx "github.com/clinicmanager/shared/context"
	sharedDB "github.com/clinicmanager/shared/db"
	"github.com/clinicmanager/shared/logger"
	"github.com/clinicmanager/shared/middleware"
	"github.com/clinicmanager/shared/setting"
)

func main() {
	env, err := setting.LoadAndValidateEnv([]string{
		"TENANT_PORT",
		"TENANTDB_URL",
	})
	if err != nil {
		logger.Error(context.Background(), "failed to load tenant service settings", "missing", err)
		os.Exit(1)
	}
	port := env["TENANT_PORT"]
	dbURL := env["TENANTDB_URL"]

	db, dbErr := sharedDB.NewDB(dbURL)
	if dbErr != nil {
		logger.Error(context.Background(), "failed to connect to tenant database", "error", dbErr)
		os.Exit(1)
	}
	defer db.Close()

	tenantRepo := repository.NewTenantRepo(db)
	tenantService := service.NewTenantService(tenantRepo)

	// 1. Initialize the GraphQL Server
	// We create a resolver with our service dependencies, wrap it in gqlgen's schema,
	// and construct the default GraphQL server handler.
	resolver := graph.NewResolver(tenantService)
	schemaConfig := generated.Config{Resolvers: resolver}
	executableSchema := generated.NewExecutableSchema(schemaConfig)
	graphqlHandler := handler.NewDefaultServer(executableSchema)

	// 2. Create the HTTP Router (Mux)
	router := http.NewServeMux()

	// 3. Register HTTP Routes
	// - /health: simple endpoint for health checks (e.g. Docker, Kubernetes, or ping tools)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	// - /graphql: endpoint to process all GraphQL operations (queries, mutations)
	//   We wrap it in Tenant middleware to parse/handle tenant contexts.
	router.Handle("/graphql", sharedCtx.Tenant(graphqlHandler))

	// 4. Chain Middlewares
	// We wrap the router sequentially to add behaviors:
	// LoggingMiddleware: Logs details of incoming HTTP requests.
	// RecoveryMiddleware: Catches any code panics to prevent the service from crashing.
	loggedRouter := middleware.LoggingMiddleware(router)
	securedRouter := middleware.RecoveryMiddleware(loggedRouter)

	// 5. Configure the HTTP Server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      securedRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		logger.Info(context.Background(), "tenant service starting", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), "listen error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info(context.Background(), "shutting down tenant service")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error(context.Background(), "shutdown error", "error", err)
	}
}
