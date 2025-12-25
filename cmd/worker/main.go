package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	redispkg "github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
	goredis "github.com/redis/go-redis/v9"
	kafkaq "github.com/RatnakirtiKamble/DeliveryGO/internal/queue/kafka"
	pg "github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/worker"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/app"
)

func main() {
	log.Println("[worker] starting")

	cfg := app.LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		log.Println("[worker] shutdown signal received")
		cancel()
	}()

	redisClient := goredis.NewClient(&goredis.Options{
		Addr: cfg.RedisAddr,
	})
	pathIndex := redispkg.NewPathIndex(redisClient)


	pool, err := pg.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to init pg pool: %v", err)
	}

	riderCache := redispkg.NewRiderCache(redisClient)
	riderStore := pg.NewRiderStore(pool)

	pathTemplateStore := pg.NewPathTemplateStore(pool)

	producer := kafkaq.NewProducer(cfg.KafkaBrokers)

	matchingWorker := worker.NewMatchingWorker()

	optimizerWorker := worker.NewOptimizerWorker(
		producer,
	)

	regretWorker := worker.NewRegretWorker()

	provisionWorker := worker.NewPathProvisionWorker(
		pathTemplateStore,
		pathIndex,
	)


	riderAssignWorker := worker.NewRiderAssignmentWorker(
		riderStore,
		riderCache,
	)

	matchingConsumer := kafkaq.NewConsumer(
		cfg.KafkaBrokers,
		"batches.assigned",
		"matching-workers",
	)

	optimizerConsumer := kafkaq.NewConsumer(
		cfg.KafkaBrokers,
		"routes.refine.requested",
		"optimizer-workers",
	)

	regretConsumer := kafkaq.NewConsumer(
		cfg.KafkaBrokers,
		"routes.refined.requested",
		"regret-workers",
	)

	provisionConsumer := kafkaq.NewConsumer(
		cfg.KafkaBrokers,
		"path.provision.requested",
		"path-provision-workers",
	)

	riderAssignConsumer := kafkaq.NewConsumer(
		cfg.KafkaBrokers,
		"routes.refined.requested",
		"rider-assignment-workers",
	)

	var wg sync.WaitGroup
	wg.Add(5)

	go func() {
		defer wg.Done()
		for {
			msg, err := matchingConsumer.Read(ctx)
			if err != nil {
				log.Println("[matching-worker]", err)
				return
			}
			_ = matchingWorker.Handle(ctx, msg)
		}
	}()

	go func() {
		defer wg.Done()
		for {
			msg, err := optimizerConsumer.Read(ctx)
			if err != nil {
				log.Println("[optimizer-worker]", err)
				return
			}
			_ = optimizerWorker.Handle(ctx, msg)
		}
	}()

	go func() {
		defer wg.Done()
		for {
			msg, err := regretConsumer.Read(ctx)
			if err != nil {
				log.Println("[regret-worker]", err)
				return
			}
			_ = regretWorker.Handle(ctx, msg)
		}
	}()
	
	go func() {
		defer wg.Done()
		for {
			msg, err := provisionConsumer.Read(ctx)
			if err != nil {
				return 
			}

			_ = provisionWorker.Handle(ctx, msg)
		}
	} ()

	go func() {
		defer wg.Done()
		for {
			msg, err := riderAssignConsumer.Read(ctx)
			if err != nil {
				log.Println("[rider-assign]", err)
				return 
			}

			_ = riderAssignWorker.Handle(ctx, msg)
		}
	} ()

	wg.Wait()
	log.Println("[worker] stopped cleanly")
}
