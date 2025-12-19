package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/order"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/batch"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/matching"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
	kafkaq "github.com/RatnakirtiKamble/DeliveryGO/internal/queue/kafka"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/handlers"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/ws"
)

func NewRouter(
	orderSvc *order.Service,
	batchSvc *batch.Service,
	matchingSvc *matching.Service,
	pathIndex *redis.PathIndex,
	batchPathStore *postgres.BatchPathStore,
	producer *kafkaq.Producer,
	hub *ws.Hub,
) http.Handler {

	r := chi.NewRouter()

	r.Get("/health", handlers.Health)

	r.Get("/ws", handlers.WebSocketHandler(hub))

	r.Route("/orders", func(r chi.Router) {
		r.Post(
			"/",
			handlers.CreateOrder(
				orderSvc,
				batchSvc,
				matchingSvc,
				pathIndex,
				producer,
				hub,
			),
		)
	})

	r.Route("/debug", func(r chi.Router){
		r.Get("/batch/{id}/path", handlers.DebugBatchPath(batchPathStore))
		r.Get("/path/{id}/batches", handlers.DebugPathBatches(pathIndex))
	})

	return r
}
