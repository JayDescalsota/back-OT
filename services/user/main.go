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

	"github.com/clinicmanager/services/user/graph"
	"github.com/clinicmanager/services/user/graph/generated"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/services/user/service"
	sharedDB "github.com/clinicmanager/shared/db"
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
	})
	if errorEnv != nil {
		log.Fatal("failed to load user service settings: ", errorEnv)
	}
	port := env["USER_PORT"]
	userDbUrl := env["USERDB_URL"]
	jwtSecret := env["JWT_SECRET"]
	baseURL := env["BASE_URL"] + ":" + env["USER_PORT"]

	db, err := sharedDB.NewDB(userDbUrl)
	if err != nil {
		log.Fatal("failed to connect to user database: ", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepo(db)
	userService := service.NewUserService(userRepo, middleware.UserIDFromCtx)
	authService := service.NewAuthService(userRepo, jwtSecret, SMTPMailer{}, baseURL)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{UserService: userService},
	}))

	mux := http.NewServeMux()
	mux.Handle("/graphql", srv)
	mux.Handle("/playground", playground.Handler("User", "/graphql"))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /register", registerHandler(authService))
	mux.HandleFunc("POST /login", loginHandler(authService))
	mux.HandleFunc("GET /verify", verifyHandler(authService))

	server := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		log.Printf("user service listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen error: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("shutting down user service...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
