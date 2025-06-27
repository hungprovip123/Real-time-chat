package models

import "time"

// Message cấu trúc tin nhắn chat
type Message struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	RoomID    string    `json:"room_id"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // message, join, leave
}

// User thông tin người dùng
type User struct {
	ID       string    `json:"id"`
	Username string    `json:"username"`
	RoomID   string    `json:"room_id"`
	IsOnline bool      `json:"is_online"`
	LastSeen time.Time `json:"last_seen"`
}

// Room phòng chat
type Room struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Users        map[string]*User `json:"users"`
	MessageCount int              `json:"message_count"`
	CreatedAt    time.Time        `json:"created_at"`
}
