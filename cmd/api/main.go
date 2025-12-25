package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/app"
	matchinggrpc "github.com/RatnakirtiKamble/DeliveryGO/internal/transport/grpc/matching"
	pb "github.com/RatnakirtiKamble/DeliveryGO/internal/transport/grpc/matchingpb"
	"google.golang.org/grpc"
)

func main() {
	cfg := app.LoadConfig()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize the application: %v", err)
	}

	// ---------------- HTTP SERVER ----------------

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: application.Router,
	}

	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// ---------------- gRPC SERVER ----------------

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("failed to listen on grpc port: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterMatchingServiceServer(
		grpcServer,
		matchinggrpc.NewServer(application.MatchingService),
	)

	go func() {
		log.Println("gRPC MatchingService listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc server error: %v", err)
		}
	}()

	// ---------------- SHUTDOWN ----------------

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("http server shutdown failed: %v", err)
	}

	log.Println("Shutdown complete")
}
