package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	kafkaq "github.com/RatnakirtiKamble/DeliveryGO/internal/queue/kafka"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/worker"
)

func main() {
	log.Println("[worker] starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	optimizerWorker := worker.NewOptimizerWorker()

	consumer := kafkaq.NewConsumer(
		[]string{"localhost:9092"},
		"routes.refine.requested",
		"optimizer-workers",
	)

	go func() {
		<-sig
		log.Println("[worker] shutting down")
		cancel()
	}()

	for {
		msg, err := consumer.Read(ctx)
		if err != nil {
			log.Fatal(err)
		}

		if err := optimizerWorker.Handle(ctx, msg); err != nil {
			log.Printf("[worker] error: %v", err)
		}
	}
}
