package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	kafkaq "github.com/RatnakirtiKamble/DeliveryGO/internal/queue/kafka"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/batch"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/matching"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/order"
	pg "github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
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

const maxAssignRetries = 3

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

		// -------- Parse request --------
		var req createOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// -------- Create order --------
		orderObj, err := orderSvc.CreateOrder(
			r.Context(),
			req.UserID,
			req.Lat,
			req.Lon,
		)
		if err != nil {
			log.Println("failed to create order:", err)
			http.Error(w, "failed to create order", http.StatusInternalServerError)
			return
		}

		// -------- Assign batch --------
		batchObj, err := batchSvc.AssignOrder(
			r.Context(),
			orderObj.ID,
		)
		if err != nil {
			http.Error(w, "failed to assign batch", http.StatusInternalServerError)
			return
		}

		// -------- H3 lookup --------
		h3Cell := util.LatLonToH3(orderObj.Lat, orderObj.Lon)
		if h3Cell == "" {
			http.Error(w, "could not determine location", http.StatusInternalServerError)
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
				"h3":       h3Cell,
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
				"delivery routes are being prepared for this area",
				http.StatusConflict,
			)
			return
		}

		var (
			chosenPathID string
			chosenCost   int
			assignErr    error
		)

		for i := 0; i < maxAssignRetries; i++ {

			path, cost, err := matchingSvc.SelectBestPath(
				orderObj,
				candidatePathIDs,
			)
			if err != nil {
				log.Println("path selection error:", err)
				http.Error(w, "no viable path", http.StatusInternalServerError)
				return
			}

			assignErr = batchSvc.AssignBatchToPath(
				r.Context(),
				batchObj.ID,
				path.ID,
			)

			if assignErr == nil {
				chosenPathID = path.ID
				chosenCost = cost
				break
			}

			if errors.Is(assignErr, pg.ErrOptimisticConflict) {
				continue 
			}

			if errors.Is(assignErr, pg.ErrBatchAtCapacity) {
				http.Error(w, "batch at capacity", http.StatusConflict)
				return
			}

			log.Println("failed to assign batch:", assignErr)
			http.Error(w, "failed to assign batch", http.StatusInternalServerError)
			return
		}

		if assignErr != nil {
			http.Error(w, "assignment contention", http.StatusServiceUnavailable)
			return
		}

		_ = pathIndex.BindBatchToPath(
			r.Context(),
			chosenPathID,
			batchObj.ID,
		)

		event := map[string]any{
			"batch_id":       batchObj.ID,
			"path_id":        chosenPathID,
			"orders":         []string{orderObj.ID},
			"estimated_cost": chosenCost,
		}

		payload, _ := json.Marshal(event)

		_ = producer.Publish(
			r.Context(),
			"batches.assigned",
			batchObj.ID,
			payload,
		)

		_ = producer.Publish(
			r.Context(),
			"routes.refine.requested",
			batchObj.ID,
			payload,
		)

		hub.Broadcast(map[string]any{
			"type": "batch.updated",
			"batch": map[string]any{
				"id":     batchObj.ID,
				"status": batchObj.Status,
			},
			"path": map[string]any{
				"id":   chosenPathID,
				"cost": chosenCost,
			},
		})

		json.NewEncoder(w).Encode(createOrderResponse{
			OrderID: orderObj.ID,
		})
	}
}
