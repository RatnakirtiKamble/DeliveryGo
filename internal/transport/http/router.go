package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/order"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/handlers"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/ws"
)

func NewRouter(
	orderSvc *order.Service,
	hub *ws.Hub,
	) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", handlers.Health)

	r.Get("/ws", handlers.WebSocketHandler(hub))

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", handlers.CreateOrder(orderSvc, hub))
	})
	
	return r
}

