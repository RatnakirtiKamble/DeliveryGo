package app

import (
	"net/http"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/order"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/memory"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/ws"
	httpt "github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http"
)

type App struct {
	Router http.Handler
}

func New(cfg Config) (*App, error) {
	orderStore := memory.NewOrderStore()
	
	orderService := order.NewService(orderStore)

	hub := ws.NewHub()

	router := httpt.NewRouter(orderService, hub)

	return &App{
		Router: router,
	}, nil
}

