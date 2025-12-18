package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/batch"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/order"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/matching"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
	kafkaq "github.com/RatnakirtiKamble/DeliveryGO/internal/queue/kafka"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/ws"
)

type createOrderRequest struct {
	UserID string  `json:"user_id"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
}

type createOrderResponse struct {
	OrderID string `json:"order_id"`
}

func CreateOrder(
	orderSvc *order.Service,
	batchSvc *batch.Service,
	matchingSvc *matching.Service,
	pathIndex *redis.PathIndex,
	producer *kafkaq.Producer,
	hub *ws.Hub,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if hub == nil {
			http.Error(w, "event hub not initialized", http.StatusInternalServerError)
			return
		}

		var req createOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		order, err := orderSvc.CreateOrder(
			r.Context(),
			req.UserID,
			req.Lat,
			req.Lon,
		)
		if err != nil {
			http.Error(w, "failed to create order", http.StatusInternalServerError)
			return
		}

		batch, err := batchSvc.AssignOrder(
			r.Context(),
			order.ID,
		)
		if err != nil {
			http.Error(w, "failed to assign batch", http.StatusInternalServerError)
			return
		}

		h3Cell := "demo-h3-cell"

		candidatePathIDs, err := pathIndex.GetCandidatePaths(
			r.Context(),
			h3Cell,
		)
		if err != nil {
			http.Error(w, "failed to lookup paths", http.StatusInternalServerError)
			return
		}

		path, cost, err := matchingSvc.SelectBestPath(order, candidatePathIDs)
		if err != nil {
			http.Error(w, "no viable path", http.StatusInternalServerError)
			return
		}

		event := map[string]any{
			"batch_id":       batch.ID,
			"path_id":        path.ID,
			"orders":         []string{order.ID},
			"estimated_cost": cost,
		}

		payload, _ := json.Marshal(event)

		_ = producer.Publish(
			r.Context(),
			"batches.assigned",
			batch.ID,
			payload,
		)

		_ = producer.Publish(
			r.Context(),
			"routes.refine.requested",
			batch.ID,
			payload,
		)

		hub.Broadcast(map[string]any{
			"type": "batch.updated",
			"batch": map[string]any{
				"id":     batch.ID,
				"status": batch.Status,
			},
			"path": map[string]any{
				"id":   path.ID,
				"cost": cost,
			},
		})

		json.NewEncoder(w).Encode(createOrderResponse{
			OrderID: order.ID,
		})
	}
}
