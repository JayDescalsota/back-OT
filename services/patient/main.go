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

	"github.com/clinicmanager/services/patient/graph"
	"github.com/clinicmanager/services/patient/graph/generated"
	"github.com/clinicmanager/services/patient/repository"
	"github.com/clinicmanager/services/patient/service"
)

func main() {
	env, err := setting.LoadAndValidateEnv([]string{
		"PATIENT_PORT",
		"PATIENTDB_URL",
		"REDIS_ADDR",
	})
	if err != nil {
		logger.Error(context.Background(), "failed to load patient service settings", "error", err)
		os.Exit(1)
	}
	port := env["PATIENT_PORT"]
	dbURL := env["PATIENTDB_URL"]

	db, dbErr := sharedDB.NewDB(dbURL)
	if dbErr != nil {
		logger.Error(context.Background(), "failed to connect to patient database", "error", dbErr)
		os.Exit(1)
	}
	defer db.Close()

	dbSet := sharedDB.NewDBSet(db)
	patientRepo := repository.NewPatientRepository(dbSet)

	redisClient, redisErr := sharedCache.NewRedisClient(env["REDIS_ADDR"], env["REDIS_PASSWORD"])
	if redisErr != nil {
		logger.Error(context.Background(), "failed to connect to redis", "error", redisErr)
		os.Exit(1)
	}
	defer redisClient.Close()

	patientService := service.NewPatientService(patientRepo, redisClient)

	resolver := graph.NewResolver(patientService)
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
		logger.Info(context.Background(), "starting patient service", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), "listen error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info(context.Background(), "shutting down patient service")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error(context.Background(), "shutdown error", "error", err)
	}
}
