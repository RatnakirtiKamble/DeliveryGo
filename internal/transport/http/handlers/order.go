package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/batch"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/order"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/ws"
)

type createOrderRequest struct {
	UserID	string  `json:"user_id"`
	Lat 	float64 `json:"lat"`
	Lon 	float64 `json:"lon"`
}

type createOrderResponse struct {
	OrderID string `json:"order_id"`
}

func CreateOrder(
    orderSvc *order.Service,
	batchSvc *batch.Service,
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
			http.Error(w, "Failed to create batch", http.StatusInternalServerError)
			return
		}

        hub.Broadcast(map[string]any{
            "type": "batch.updated",
            "batch": map[string]any{
				"id": 		batch.ID,
				"status": 	batch.Status,
			},
        })

        json.NewEncoder(w).Encode(createOrderResponse{
            OrderID: order.ID,
        })
    }
}
