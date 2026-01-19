package websocket

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// WebSocketService wraps the existing WebSocket server for service integration
type WebSocketService struct {
	client *mongo.Client
	db     *mongo.Database
	server *WebSocketServer
}

// NewWebSocketService creates a new WebSocket service wrapper
func NewWebSocketService(client *mongo.Client, server *WebSocketServer) *WebSocketService {
	return &WebSocketService{
		client: client,
		db:     client.Database("streamnestai"),
		server: server,
	}
}

// Initialize sets up the WebSocket service
func (wss *WebSocketService) Initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create indexes for WebSocket collections
	collections := map[string][]bson.D{
		"websocket_sessions": {
			{{"user_id", 1}},
			{{"created_at", 1}},
		},
		"watch_parties": {
			{{"movie_id", 1}},
			{{"host_id", 1}},
			{{"is_active", 1}},
		},
	}

	for collName, indexes := range collections {
		coll := wss.db.Collection(collName)
		for _, index := range indexes {
			_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: index})
			if err != nil {
				log.Printf("Warning: Failed to create index for %s: %v", collName, err)
			}
		}
	}

	log.Println("WebSocket service initialized successfully")
	return nil
}

// GetServer returns the underlying WebSocket server
func (wss *WebSocketService) GetServer() *WebSocketServer {
	return wss.server
}

// GetHub returns the WebSocket hub
func (wss *WebSocketService) GetHub() *Hub {
	return wss.server.GetHub()
}

// BroadcastMessage broadcasts a message to all connected clients
func (wss *WebSocketService) BroadcastMessage(messageType, content string) {
	wss.server.BroadcastMessage(messageType, content)
}

// SendToUser sends a message to a specific user
func (wss *WebSocketService) SendToUser(userID, messageType, content string) {
	wss.server.SendToUser(userID, messageType, content)
}

// SendToRoom sends a message to all clients in a room
func (wss *WebSocketService) SendToRoom(roomID, messageType, content string) {
	wss.server.SendToRoom(roomID, messageType, content)
}

// GetConnectedUsers returns the count of connected users
func (wss *WebSocketService) GetConnectedUsers() int {
	stats := wss.server.GetServerStats()
	if count, ok := stats["connected_clients"].(int64); ok {
		return int(count)
	}
	return 0
}

// GetRoomUsers returns the count of users in a specific room
func (wss *WebSocketService) GetRoomUsers(roomID string) int {
	stats := wss.server.GetRoomStats(roomID)
	if count, ok := stats["client_count"].(int); ok {
		return count
	}
	return 0
}

// CreateNotificationService creates a new notification service
func (wss *WebSocketService) CreateNotificationService() *NotificationService {
	return NewNotificationService(wss.server)
}

// CreateWatchPartyService creates a new watch party service
func (wss *WebSocketService) CreateWatchPartyService() *WatchPartyService {
	return NewWatchPartyService(wss.server)
}

// CreateChatService creates a new chat service
func (wss *WebSocketService) CreateChatService() *ChatService {
	return NewChatService(wss.server)
}

// logWebSocketSession logs a WebSocket session to the database
func (wss *WebSocketService) logWebSocketSession(userID, ipAddress string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := wss.db.Collection("websocket_sessions")

	session := bson.M{
		"user_id":    userID,
		"ip_address": ipAddress,
		"created_at": time.Now(),
	}

	_, err := coll.InsertOne(ctx, session)
	if err != nil {
		log.Printf("Warning: Failed to log WebSocket session: %v", err)
	}
}
