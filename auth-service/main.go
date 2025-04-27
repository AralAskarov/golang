package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"authservice/config"
	"authservice/controller"
	"authservice/middleware"
	"authservice/repository"
	"authservice/service"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg := config.New()
	service.InitSecret(cfg.JWTSecretKey)
	db, err := repository.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection faile %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1000)
	db.SetMaxIdleConns(100)
	db.SetConnMaxLifetime(5 * time.Minute)

	tokenRepo := repository.NewPostgresTokenRepository(db)
	// userRepo := repository.NewPostgresUserRepository(db)

	tokenService := service.NewTokenService(tokenRepo)

	tokenController := controller.NewTokenController(tokenService)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      setupRoutes(tokenController),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Printf("Server starting on port %s...\n", cfg.Port)
	log.Fatal(server.ListenAndServe())

}

func setupRoutes(tokenController *controller.TokenController) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/token", middleware.Recover(middleware.Logger(http.HandlerFunc(tokenController.HandleTokenRequest))))
	// mux.Handle("/refresh", middleware.Recover(middleware.Logger(http.HandlerFunc(tokenController.HandleTokenValidation))))

	return mux
}