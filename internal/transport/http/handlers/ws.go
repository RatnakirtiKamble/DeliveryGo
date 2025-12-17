package handlers

import (
	"net/http"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/transport/http/ws"
)

func WebSocketHandler(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client, err := ws.NewClient(w, r)

		if err != nil {
			return
		}

		hub.Register(client)
		defer hub.Unregister(client)

		go client.WritePump()

		for {
			if _, _, err := client.Conn().ReadMessage(); err != nil {
				return
			}
		}
	}
}