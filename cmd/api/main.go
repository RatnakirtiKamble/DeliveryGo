package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/app"
)

func main() {

	cfg := app.LoadConfig()

	application, err := app.New(cfg)

	if err != nil {
		log.Fatalf("failed to initialize the application: %v", err)
	}

	server := &http.Server{
		Addr:	cfg.HTTPAddr,
		Handler:	application.Router,
	}

	go func() {
		log.Printf("Starting server on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
}