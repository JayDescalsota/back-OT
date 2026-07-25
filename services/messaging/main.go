package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/redis/go-redis/v9"

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
	router.HandleFunc("GET /sse/messages", func(w http.ResponseWriter, r *http.Request) {
		sseHandler(w, r, redisClient, encryptor)
	})

	loggedRouter := middleware.LoggingMiddleware(router)
	securedRouter := middleware.RecoveryMiddleware(loggedRouter)

	server := &http.Server{
		Addr:        ":" + port,
		Handler:     securedRouter,
		ReadTimeout: 10 * time.Second,
		IdleTimeout: 30 * time.Second,
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

func sseHandler(w http.ResponseWriter, r *http.Request, redisClient *redis.Client, encryptor *service.Encryptor) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId query parameter required", http.StatusBadRequest)
		return
	}
	lastID := r.URL.Query().Get("lastMessageId")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	writer, ok := w.(io.Writer)
	if !ok {
		return
	}

	streamStart := "$"
	if lastID != "" {
		streamStart = lastID
	}

	for {
		select {
		case <-r.Context().Done():
			return
		default:
		}

		args := &redis.XReadArgs{
			Streams: []string{"messaging:messages", streamStart},
			Count:   20,
			Block:   10000,
		}
		streams, err := redisClient.XRead(r.Context(), args).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return
		}

		for _, stream := range streams {
			for _, m := range stream.Messages {
				threadID, _ := m.Values["thread_id"].(string)
				senderID, _ := m.Values["sender_id"].(string)
				body, _ := m.Values["body"].(string)
				createdAt, _ := m.Values["created_at"].(string)
				participantsRaw, _ := m.Values["participants"].(string)

				var participantIDs []string
				json.Unmarshal([]byte(participantsRaw), &participantIDs)

				found := false
				for _, pid := range participantIDs {
					if pid == userID {
						found = true
						break
					}
				}
				if !found {
					continue
				}

				payload, _ := json.Marshal(map[string]any{
					"id":         m.Values["id"],
					"thread_id":  threadID,
					"sender_id":  senderID,
					"body":       body,
					"created_at": createdAt,
				})
				fmt.Fprintf(writer, "data: %s\n\n", payload)
				flusher.Flush()
				streamStart = m.ID
			}
		}
	}
}
