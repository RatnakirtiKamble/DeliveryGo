package app

import (
	"context"
	"net/http"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/order"
	batchsvc "github.com/RatnakirtiKamble/DeliveryGO/internal/service/batch"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/matching"
	pg "github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
	kafkaq "github.com/RatnakirtiKamble/DeliveryGO/internal/queue/kafka"
	goredis "github.com/redis/go-redis/v9"
	httpt "github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/ws"
)

type App struct {
	Router http.Handler
}

func New(cfg Config) (*App, error) {
	ctx := context.Background()

	pool, err := pg.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}

	orderStore := pg.NewOrderStore(pool)
	batchStore := pg.NewBatchStore(pool)

	orderService := order.NewService(orderStore)
	batchService := batchsvc.NewService(batchStore)

	matchingService := matching.NewService(nil, nil)

	redisClient := goredis.NewClient(&goredis.Options{
	Addr: cfg.RedisAddr,
	})

	pathIndex := redis.NewPathIndex(redisClient)

	producer := kafkaq.NewProducer(cfg.KafkaBrokers)

	hub := ws.NewHub()

	router := httpt.NewRouter(
		orderService,
		batchService,
		matchingService,
		pathIndex,
		producer,
		hub,
	)

	return &App{
		Router: router,
	}, nil
}
