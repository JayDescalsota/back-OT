package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/ot/identity-svc/graph"
	"github.com/ot/identity-svc/middleware"
	"github.com/ot/identity-svc/repository"
	"github.com/ot/identity-svc/service"
	sharedDB "github.com/ot/shared/db"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4001"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-do-not-use-in-production"
	}

	db, err := sharedDB.NewDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepo(db)

	// Services
	authSvc := service.NewAuthService(userRepo, jwtSecret)
	userSvc := service.NewUserService(userRepo)

	// GraphQL
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{
			AuthService: authSvc,
			UserService: userSvc,
		},
	}))

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("Identity Service", "/graphql"))
	mux.Handle("/graphql", middleware.GraphQLAuthMiddleware(jwtSecret)(srv))

	server := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		log.Printf("identity-svc listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	server.Shutdown(context.Background())
}
