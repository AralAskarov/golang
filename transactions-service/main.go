package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"transervice/config"
	"transervice/controller"
	"transervice/middleware"
	"transervice/repository"
	"transervice/service"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg := config.New()
	service.InitSecret(cfg.JWTSecretKey)
	db, err := repository.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection fail %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1000)
	db.SetMaxIdleConns(100)
	db.SetConnMaxLifetime(5 * time.Minute)

	balanceRepo := repository.NewPostgresBalanceRepository(db)
	// userRepo := repository.NewPostgresUserRepository(db)

	balanceService := service.NewBalanceService(balanceRepo)

	balanceController := controller.NewBalanceController(balanceService)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      setupRoutes(balanceController),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  220 * time.Second,
	}

	fmt.Printf("Server starting on port %s...\n", cfg.Port)
	log.Fatal(server.ListenAndServe())

}

func setupRoutes(balanceController *controller.BalanceController) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/balance", middleware.Recover(middleware.Logger(http.HandlerFunc(balanceController.ReplenishmentRequest))))
	mux.Handle("/api/withdrawal", middleware.Recover(middleware.Logger(http.HandlerFunc(balanceController.WithdrawalRequest))))

	return mux
}