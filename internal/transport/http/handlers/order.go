package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	kafkaq "github.com/RatnakirtiKamble/DeliveryGO/internal/queue/kafka"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/batch"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/matching"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/order"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/ws"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/util"
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
            log.Println("Failed to create order ", err)
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

		h3Cell := util.LatLonToH3(order.Lat, order.Lon)

        if h3Cell == "" {
            http.Error(w, "Could not determine location.", http.StatusInternalServerError)
            return
        }

		

		candidatePathIDs, err := pathIndex.GetCandidatePaths(
			r.Context(),
			h3Cell,
		)
		if err != nil {
			http.Error(w, "failed to lookup paths", http.StatusInternalServerError)
			return
		}

		if len(candidatePathIDs) == 0 {
			event := map[string]string{
				"store_id": "store-1",
				"h3"	  : h3Cell,
			}

			payload, _ := json.Marshal(event)

			_ = producer.Publish(
				r.Context(),
				"path.provision.requested",
				h3Cell,
				payload,
			)

			http.Error(
				w, 
				"Delivery routes are being prepared for this area",
				http.StatusConflict,
			)

			return
		}

		path, cost, err := matchingSvc.SelectBestPath(order, candidatePathIDs)
		if err != nil {
            log.Println("Path search error: ", err)
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
