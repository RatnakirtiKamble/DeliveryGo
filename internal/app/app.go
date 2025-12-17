package app

import (
	"context"
	"net/http"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/order"
	batchsvc "github.com/RatnakirtiKamble/DeliveryGO/internal/service/batch"
	pg "github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
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

	hub := ws.NewHub()

	router := httpt.NewRouter(orderService, batchService, hub)

	return &App{
		Router: router,
	}, nil
}

