package ws

import (
	"net/http"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

func NewClient(w http.ResponseWriter, r*http.Request) (*Client, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}, nil
}

func (c *Client) WritePump() {
	defer c.conn.Close()

	for msg := range c.send{
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *Client) Conn() *websocket.Conn {
	return c.conn
}