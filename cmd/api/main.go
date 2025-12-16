package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emuthianimbithi/GoStack/internal/config"
	"github.com/emuthianimbithi/GoStack/internal/container"
	"github.com/emuthianimbithi/GoStack/internal/db"
	"github.com/emuthianimbithi/GoStack/internal/server/router"
)

func main() {
	// 1. Load Config
	cfg := config.Load()

	// 3. DI Container (Handles DB & Redis connection internally)
	c := container.New(cfg)

	// 4. Database Migrations
	// Using GORM AutoMigrate as requested
	if err := db.AutoMigrate(c.DB); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	// 5. Router
	r := router.NewRouter(c)

	// 6. Server
	srv := &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: r,
	}

	// Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Server listening on %s", cfg.HTTP.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("Shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
