package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	sharedCtx "github.com/clinicmanager/shared/context"
	sharedDB "github.com/clinicmanager/shared/db"
	"github.com/clinicmanager/shared/logger"
	"github.com/clinicmanager/shared/middleware"
	"github.com/clinicmanager/shared/setting"

	sharedCache "github.com/clinicmanager/shared/cache"

	"github.com/clinicmanager/services/booking/graph"
	"github.com/clinicmanager/services/booking/graph/generated"
	"github.com/clinicmanager/services/booking/repository"
	"github.com/clinicmanager/services/booking/service"
)

func main() {
	env, err := setting.LoadAndValidateEnv([]string{
		"BOOKING_PORT",
		"BOOKINGDB_URL",
		"REDIS_ADDR",
	})
	if err != nil {
		logger.Error(context.Background(), "failed to load booking service settings", "error", err)
		os.Exit(1)
	}
	port := env["BOOKING_PORT"]
	dbURL := env["BOOKINGDB_URL"]

	db, dbErr := sharedDB.NewDB(dbURL)
	if dbErr != nil {
		logger.Error(context.Background(), "failed to connect to booking database", "error", dbErr)
		os.Exit(1)
	}
	defer db.Close()

	redisAddr := env["REDIS_ADDR"]
	redisPassword := env["REDIS_PASSWORD"]
	redisClient, redisErr := sharedCache.NewRedisClient(redisAddr, redisPassword)
	if redisErr != nil {
		logger.Error(context.Background(), "failed to connect to redis", "error", redisErr)
		os.Exit(1)
	}
	defer redisClient.Close()

	dbSet := sharedDB.NewDBSet(db)
	bookingRepo := repository.NewBookingRepository(dbSet)
	bookingService := service.NewBookingService(bookingRepo, redisClient)

	resolver := graph.NewResolver(bookingService)
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
		logger.Info(context.Background(), "starting booking service", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), "listen error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info(context.Background(), "shutting down booking service")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error(context.Background(), "shutdown error", "error", err)
	}
}
