package handlers

import (
	"net/http"
	"encoding/json"

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
	svc *order.Service,
	hub *ws.Hub,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		order, err := svc.CreateOrder(
			r.Context(), 
			req.UserID, 
			req.Lat, 
			req.Lon,
		)

		hub.Broadcast(map[string]any{
			"type": "order.created",
			"order": map[string]any{
				"id": order.ID,
				"user_id": order.UserID,
				"lat": order.Lat,
				"lon": order.Lon,
			},
		})

		if err != nil {
			http.Error(w, "failed to create order", http.StatusInternalServerError)
		}

		resp := createOrderResponse{
			OrderID: order.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	
	}
}