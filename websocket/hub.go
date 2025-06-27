package websocket

import (
	"log"
	"week3-chat/models"
	"week3-chat/redis"
)

type Hub struct {
	clients     map[*Client]bool
	broadcast   chan *models.Message
	Register    chan *Client
	unregister  chan *Client
	rooms       map[string]map[*Client]bool // roomID -> clients
	redisClient *redis.Client
}

func NewHub(redisClient *redis.Client) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan *models.Message),
		Register:    make(chan *Client),
		unregister:  make(chan *Client),
		rooms:       make(map[string]map[*Client]bool),
		redisClient: redisClient,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			log.Printf("👤 Client registered: %s in room %s", client.Username, client.RoomID)
			h.clients[client] = true

			// Add to room
			if h.rooms[client.RoomID] == nil {
				h.rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.rooms[client.RoomID][client] = true

			// Send join message to room
			joinMessage := &models.Message{
				UserID:   client.UserID,
				Username: client.Username,
				RoomID:   client.RoomID,
				Content:  client.Username + " joined the room",
				Type:     "join",
			}
			h.broadcastToRoom(joinMessage)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Printf("👤 Client unregistered: %s from room %s", client.Username, client.RoomID)
				delete(h.clients, client)
				close(client.send)

				// Remove from room
				if room, exists := h.rooms[client.RoomID]; exists {
					delete(room, client)
					if len(room) == 0 {
						delete(h.rooms, client.RoomID)
					}
				}

				// Send leave message to room
				leaveMessage := &models.Message{
					UserID:   client.UserID,
					Username: client.Username,
					RoomID:   client.RoomID,
					Content:  client.Username + " left the room",
					Type:     "leave",
				}
				h.broadcastToRoom(leaveMessage)
			}

		case message := <-h.broadcast:
			log.Printf("📤 Broadcasting message from %s to room %s", message.Username, message.RoomID)

			// Save to Redis first
			if h.redisClient != nil {
				if err := h.redisClient.SaveMessage(message); err != nil {
					log.Printf("❌ Failed to save message to Redis: %v", err)
				}
			}

			h.broadcastToRoom(message)
		}
	}
}

func (h *Hub) broadcastToRoom(message *models.Message) {
	if room, exists := h.rooms[message.RoomID]; exists {
		for client := range room {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
				delete(room, client)
			}
		}
	}
}

func (h *Hub) GetRoomClients(roomID string) []*Client {
	var clients []*Client
	if room, exists := h.rooms[roomID]; exists {
		for client := range room {
			clients = append(clients, client)
		}
	}
	return clients
}

func (h *Hub) BroadcastMessage(message *models.Message) {
	h.broadcast <- message
}
