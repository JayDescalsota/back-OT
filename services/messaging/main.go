package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	sharedCache "github.com/clinicmanager/shared/cache"
	sharedCtx "github.com/clinicmanager/shared/context"
	sharedDB "github.com/clinicmanager/shared/db"
	"github.com/clinicmanager/shared/logger"
	"github.com/clinicmanager/shared/middleware"
	"github.com/clinicmanager/shared/setting"

	"github.com/clinicmanager/services/messaging/graph"
	"github.com/clinicmanager/services/messaging/graph/generated"
	"github.com/clinicmanager/services/messaging/repository"
	"github.com/clinicmanager/services/messaging/service"
)

func main() {
	env, err := setting.LoadAndValidateEnv([]string{
		"MESSAGING_PORT",
		"MESSAGINGDB_URL",
		"REDIS_ADDR",
		"MESSAGE_ENCRYPTION_KEY",
	})
	if err != nil {
		logger.Error(context.Background(), "failed to load messaging service settings", "error", err)
		os.Exit(1)
	}
	port := env["MESSAGING_PORT"]
	dbURL := env["MESSAGINGDB_URL"]

	db, dbErr := sharedDB.NewDB(dbURL)
	if dbErr != nil {
		logger.Error(context.Background(), "failed to connect to messaging database", "error", dbErr)
		os.Exit(1)
	}
	defer db.Close()

	dbSet := sharedDB.NewDBSet(db)
	messagingRepo := repository.NewMessagingRepository(dbSet)

	redisClient, redisErr := sharedCache.NewRedisClient(env["REDIS_ADDR"], env["REDIS_PASSWORD"])
	if redisErr != nil {
		logger.Error(context.Background(), "failed to connect to redis", "error", redisErr)
		os.Exit(1)
	}
	defer redisClient.Close()

	encryptor := service.NewEncryptor(env["MESSAGE_ENCRYPTION_KEY"])
	messagingService := service.NewMessagingService(messagingRepo, redisClient, encryptor)

	resolver := graph.NewResolver(messagingService)
	schemaConfig := generated.Config{Resolvers: resolver}
	executableSchema := generated.NewExecutableSchema(schemaConfig)
	graphqlHandler := handler.NewDefaultServer(executableSchema)
	contextedHandler := sharedCtx.Tenant(graphqlHandler)

	router := http.NewServeMux()
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router.Handle("/graphql", contextedHandler)

	loggedRouter := middleware.LoggingMiddleware(router)
	securedRouter := middleware.RecoveryMiddleware(loggedRouter)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      securedRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		logger.Info(context.Background(), "starting messaging service", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), "listen error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info(context.Background(), "shutting down messaging service")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error(context.Background(), "shutdown error", "error", err)
	}
}
