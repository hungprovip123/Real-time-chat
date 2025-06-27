package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	"week3-chat/models"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	Client *redis.Client
}

func NewClient(redisURL string) *Client {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Redis URL parse error: %v", err)
		// Fallback to default
		opt = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	rdb := redis.NewClient(opt)
	return &Client{Client: rdb}
}

func (r *Client) SaveMessage(message *models.Message) error {
	ctx := context.Background()

	// Generate ID if not set
	if message.ID == "" {
		message.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Set timestamp if not set
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Save to room messages list
	key := fmt.Sprintf("room:%s:messages", message.RoomID)
	err = r.Client.LPush(ctx, key, messageJSON).Err()
	if err != nil {
		return err
	}

	// Keep only recent messages (last 100)
	r.Client.LTrim(ctx, key, 0, 99)

	log.Printf("💾 Saved message from %s to Redis", message.Username)
	return nil
}

func (r *Client) GetMessages(roomID string, limit int) ([]*models.Message, error) {
	ctx := context.Background()
	key := fmt.Sprintf("room:%s:messages", roomID)

	// Get messages (newest first)
	messages, err := r.Client.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	var result []*models.Message
	// Reverse to get chronological order (oldest first)
	for i := len(messages) - 1; i >= 0; i-- {
		var message models.Message
		if err := json.Unmarshal([]byte(messages[i]), &message); err == nil {
			result = append(result, &message)
		}
	}

	log.Printf("📚 Retrieved %d messages for room %s", len(result), roomID)
	return result, nil
}

func (r *Client) SetUserOnline(userID, roomID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("room:%s:online", roomID)

	return r.Client.SAdd(ctx, key, userID).Err()
}

func (r *Client) SetUserOffline(userID, roomID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("room:%s:online", roomID)

	return r.Client.SRem(ctx, key, userID).Err()
}

func (r *Client) GetOnlineUsers(roomID string) ([]string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("room:%s:online", roomID)

	return r.Client.SMembers(ctx, key).Result()
}

func (r *Client) CheckRateLimit(userID string, requests int, window int) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%s", userID)

	// Get current count
	current, err := r.Client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	if current == "" {
		// First request
		err = r.Client.Set(ctx, key, 1, time.Duration(window)*time.Second).Err()
		return true, err
	}

	count, err := strconv.Atoi(current)
	if err != nil {
		return false, err
	}

	if count >= requests {
		return false, nil // Rate limit exceeded
	}

	// Increment counter
	err = r.Client.Incr(ctx, key).Err()
	return true, err
}
