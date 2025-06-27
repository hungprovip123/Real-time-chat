package main

import (
	"context"
	"log"
	"net/http"
	"week3-chat/config"
	"week3-chat/handlers"
	"week3-chat/redis"
	"week3-chat/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg := config.Load()

	// Initialize Redis
	redisClient := redis.NewClient(cfg.RedisURL)

	// Test Redis connection
	ctx := context.Background()
	_, err := redisClient.Client.Ping(ctx).Result()
	if err != nil {
		log.Printf("⚠️  Redis connection failed: %v", err)
	} else {
		log.Println("✅ Redis connected successfully")
	}

	// Initialize WebSocket hub
	hub := websocket.NewHub(redisClient)
	go hub.Run()

	// Initialize handlers
	chatHandler := handlers.NewChatHandler(redisClient, hub)

	// Setup Gin router
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/ws", chatHandler.HandleWebSocket)
		api.GET("/rooms/:roomId/messages", chatHandler.GetMessages)
		api.GET("/rooms/:roomId/users", chatHandler.GetOnlineUsers)
		api.POST("/messages", chatHandler.SendMessage)
	}

	// Static files
	r.Static("/static", "./static")

	// Home page
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("Chat server khởi động trên port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
