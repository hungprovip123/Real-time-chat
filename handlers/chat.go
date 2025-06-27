package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"week3-chat/models"
	"week3-chat/redis"
	"week3-chat/websocket"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type ChatHandler struct {
	redisClient *redis.Client
	hub         *websocket.Hub
}

func NewChatHandler(redisClient *redis.Client, hub *websocket.Hub) *ChatHandler {
	return &ChatHandler{
		redisClient: redisClient,
		hub:         hub,
	}
}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	userID := c.Query("user_id")
	username := c.Query("username")
	roomID := c.Query("room_id")

	if userID == "" || username == "" || roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
		return
	}

	// Check if username already exists in room
	existingClients := h.hub.GetRoomClients(roomID)
	for _, client := range existingClients {
		if client.Username == username {
			log.Printf("⚠️ Username '%s' already exists in room '%s'", username, roomID)
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists in this room"})
			return
		}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upgrade connection"})
		return
	}

	client := websocket.NewClient(h.hub, conn, userID, username, roomID)

	// Set user online
	h.redisClient.SetUserOnline(userID, roomID)

	// Register client with hub
	h.hub.Register <- client

	// Start client goroutines
	client.Run()
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	roomID := c.Param("roomId")
	limitStr := c.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	messages, err := h.redisClient.GetMessages(roomID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *ChatHandler) GetOnlineUsers(c *gin.Context) {
	roomID := c.Param("roomId")

	// Get online users from WebSocket connections
	clients := h.hub.GetRoomClients(roomID)
	var users []gin.H

	for _, client := range clients {
		users = append(users, gin.H{
			"id":       client.UserID,
			"username": client.Username,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		Username string `json:"username" binding:"required"`
		RoomID   string `json:"room_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check rate limit
	allowed, err := h.redisClient.CheckRateLimit(req.UserID, 10, 60)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
		return
	}
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		return
	}

	message := &models.Message{
		UserID:    req.UserID,
		Username:  req.Username,
		RoomID:    req.RoomID,
		Content:   req.Content,
		Type:      "message",
		Timestamp: time.Now(),
	}

	// Save to Redis
	if err := h.redisClient.SaveMessage(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save message"})
		return
	}

	// Broadcast via WebSocket
	h.hub.BroadcastMessage(message)

	c.JSON(http.StatusOK, gin.H{"message": "Message sent successfully"})
}
