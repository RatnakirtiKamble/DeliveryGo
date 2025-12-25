package app

import (
	"context"
	"net/http"
	"log"

	"google.golang.org/grpc"
	pb "github.com/RatnakirtiKamble/DeliveryGO/internal/transport/grpc/matchingpb"
	orderSvc "github.com/RatnakirtiKamble/DeliveryGO/internal/service/order"
	batchSvc "github.com/RatnakirtiKamble/DeliveryGO/internal/service/batch"
	matchingSvc "github.com/RatnakirtiKamble/DeliveryGO/internal/service/matching"
	pg "github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
	kafkaq "github.com/RatnakirtiKamble/DeliveryGO/internal/queue/kafka"
	goredis "github.com/redis/go-redis/v9"
	httpt "github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/ws"
)

type App struct {
	Router http.Handler
	MatchingService *matchingSvc.Service
}

func New(cfg Config) (*App, error) {
	ctx := context.Background()

	pool, err := pg.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}

	pathTemplateStore := pg.NewPathTemplateStore(pool)

	paths, err := pathTemplateStore.ListAll(ctx)
	if err != nil {
		return nil, err  
	}

	log.Printf("[api] loaded %d path templates", len(paths))

	estimator := &matchingSvc.SimpleEstimator{}


	matchingService := matchingSvc.NewService(
		paths,
		estimator,
	)


	orderStore := pg.NewOrderStore(pool)
	batchStore := pg.NewBatchStore(pool)
	batchPathStore := pg.NewBatchPathStore(pool)
	riderStore := pg.NewRiderStore(pool)

	orderService := orderSvc.NewService(orderStore)
	batchService := batchSvc.NewService(
		batchStore,
		batchPathStore)

	
	

	redisClient := goredis.NewClient(&goredis.Options{
	Addr: cfg.RedisAddr,
	})

	riderCache := redis.NewRiderCache(redisClient)
	pathIndex := redis.NewPathIndex(redisClient)

	producer := kafkaq.NewProducer(cfg.KafkaBrokers)

	grpcConn, err := grpc.Dial(
		cfg.GRPCDialAddr,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	matchingClient := matchingSvc.NewGRPCClient(
		pb.NewMatchingServiceClient(grpcConn),
	)


	hub := ws.NewHub()

	router := httpt.NewRouter(
		orderService,
		batchService,
		matchingClient,
		pathIndex,
		riderCache,
		batchPathStore,
		pathTemplateStore,
		riderStore,
		producer,
		hub,
	)

	return &App{
		Router: router,
		MatchingService: matchingService,
	}, nil
}
