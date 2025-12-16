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
	"github.com/emuthianimbithi/GoStack/internal/server/router"
	"github.com/emuthianimbithi/GoStack/internal/tracing"
)

func main() {
	// 1. Load Config
	cfg := config.Load()

	// 2. Observability
	shutdownTrace := tracing.Init(cfg.OTLP)
	defer func() {
		if err := shutdownTrace(context.Background()); err != nil {
			log.Printf("failed to shutdown tracer: %v", err)
		}
	}()

	// 3. DI Container (Handles DB & Redis connection internally)
	c := container.New(cfg)

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
