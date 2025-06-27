package websocket

import (
	"fmt"
	"log"
	"time"
	"week3-chat/models"

	"github.com/gorilla/websocket"
)

const (
	// Thời gian chờ ghi tin nhắn
	writeWait = 10 * time.Second

	// Thời gian chờ đọc pong message
	pongWait = 60 * time.Second

	// Khoảng thời gian gửi ping
	pingPeriod = (pongWait * 9) / 10

	// Kích thước tối đa tin nhắn
	maxMessageSize = 512
)

type Client struct {
	UserID   string
	Username string
	RoomID   string
	hub      *Hub
	conn     *websocket.Conn
	send     chan *models.Message
}

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// NewClient tạo client mới
func NewClient(hub *Hub, conn *websocket.Conn, userID, username, roomID string) *Client {
	return &Client{
		UserID:   userID,
		Username: username,
		RoomID:   roomID,
		hub:      hub,
		conn:     conn,
		send:     make(chan *models.Message, 256),
	}
}

// readPump xử lý đọc tin nhắn từ WebSocket
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		log.Printf("📨 Received message from client %s: %+v", c.Username, msg)

		if msg.Type == "message" && msg.Content != "" {
			message := &models.Message{
				ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
				UserID:    c.UserID,
				Username:  c.Username,
				RoomID:    c.RoomID,
				Content:   msg.Content,
				Type:      "message",
				Timestamp: time.Now(),
			}

			log.Printf("📤 Broadcasting message from %s to room %s", c.Username, c.RoomID)
			c.hub.BroadcastMessage(message)
		}
	}
}

// writePump xử lý gửi tin nhắn tới WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub đã đóng channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Printf("📤 Sending message to client %s: %+v", c.Username, message)
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("Write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Start bắt đầu client (chạy read và write pumps)
func (c *Client) Run() {
	go c.writePump()
	go c.readPump()
}
