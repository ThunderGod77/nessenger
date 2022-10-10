package socket

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	Conn *websocket.Conn
	Send chan PM
	H    *Hub
}

type Read struct {
	Action string            `json:"action"`
	UserId string            `json:"userId"`
	Data   map[string]string `json:"data"`
}

func (c Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			c.Conn.WriteJSON(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				c.Conn.WriteJSON(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
func (c Client) readPump() {
	defer func() {
		c.H.UnregisterClient <- &c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	var sr Read
	for {
		err := c.Conn.ReadJSON(&sr)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		switch sr.Action {
		case "register":
			c.H.RegisterUser <- &ClientInfo{
				UserId: sr.UserId,
				Ct:     &c,
			}
		case "message":
			c.H.message(PM{
				Message:    sr.Data["message"],
				ReceiverId: sr.Data["receiverId"],
				SenderId:   sr.UserId,
			})
		default:
			c.H.message(PM{
				Message:    "Error",
				ReceiverId: sr.UserId,
				SenderId:   "System",
			})

		}

	}
}
