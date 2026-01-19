package websocket

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message types
const (
	MessageTypeText         = "text"
	MessageTypeJoin         = "join"
	MessageTypeLeave        = "leave"
	MessageTypeWatchParty   = "watch_party"
	MessageTypeNotification = "notification"
	MessageTypingStart      = "typing_start"
	MessageTypingStop       = "typing_stop"
	MessageUserStatus       = "user_status"
	MessageHeartbeat        = "heartbeat"
)

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	RoomID    string                 `json:"room_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Username  string                 `json:"username,omitempty"`
	Content   interface{}            `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Client represents a WebSocket client
type Client struct {
	ID       string
	Username string
	UserID   string
	Conn     *websocket.Conn
	Send     chan Message
	Rooms    map[string]bool
	mu       sync.RWMutex
	LastSeen time.Time
}

// Room represents a chat/watch party room
type Room struct {
	ID      string
	Name    string
	Clients map[string]*Client
	Created time.Time
	mu      sync.RWMutex
}

// Hub manages all WebSocket connections and rooms
type Hub struct {
	// Registered clients
	clients map[string]*Client

	// Rooms
	rooms map[string]*Room

	// Mutex for protecting concurrent access to clients and rooms
	mu sync.RWMutex

	// Inbound messages from the clients
	broadcast chan Message

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Configuration
	config HubConfig

	// Metrics
	metrics *HubMetrics
}

// HubConfig holds configuration for the WebSocket hub
type HubConfig struct {
	MaxClientsPerRoom int
	MessageBufferSize int
	PingInterval      time.Duration
	PongWait          time.Duration
	WriteWait         time.Duration
	MaxMessageSize    int64
	EnableMetrics     bool
}

// HubMetrics tracks hub performance metrics
type HubMetrics struct {
	ConnectedClients   int64
	ActiveRooms        int64
	MessagesSent       int64
	MessagesReceived   int64
	ConnectionsCreated int64
	ConnectionsClosed  int64
	mu                 sync.RWMutex
}

// DefaultHubConfig returns default hub configuration
func DefaultHubConfig() HubConfig {
	return HubConfig{
		MaxClientsPerRoom: 100,
		MessageBufferSize: 256,
		PingInterval:      30 * time.Second,
		PongWait:          60 * time.Second,
		WriteWait:         10 * time.Second,
		MaxMessageSize:    512,
		EnableMetrics:     true,
	}
}

// NewHub creates a new WebSocket hub
func NewHub(config HubConfig) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]*Room),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		ctx:        ctx,
		cancel:     cancel,
		config:     config,
		metrics:    &HubMetrics{},
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	log.Printf("WebSocket hub started with config: %+v", h.config)

	// Start cleanup goroutine
	go h.cleanup()

	// Start metrics collection if enabled
	if h.config.EnableMetrics {
		go h.collectMetrics()
	}

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.handleMessage(message)

		case <-h.ctx.Done():
			h.shutdown()
			return
		}
	}
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	h.clients[client.ID] = client
	h.mu.Unlock()

	client.LastSeen = time.Now()

	if h.config.EnableMetrics {
		h.metrics.incrementConnectionsCreated()
	}

	log.Printf("Client %s (%s) connected", client.ID, client.Username)

	// Send welcome message
	welcome := Message{
		Type:      "welcome",
		UserID:    "system",
		Username:  "System",
		Content:   fmt.Sprintf("Welcome to StreamNestAI, %s!", client.Username),
		Timestamp: time.Now(),
	}

	select {
	case client.Send <- welcome:
	default:
		close(client.Send)
		h.mu.Lock()
		delete(h.clients, client.ID)
		h.mu.Unlock()
	}
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.RLock()
	_, exists := h.clients[client.ID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	// Remove client from all rooms
	client.mu.RLock()
	for roomID := range client.Rooms {
		h.leaveRoom(client, roomID)
	}
	client.mu.RUnlock()

	// Close connection
	close(client.Send)
	h.mu.Lock()
	delete(h.clients, client.ID)
	h.mu.Unlock()

	if h.config.EnableMetrics {
		h.metrics.incrementConnectionsClosed()
	}

	log.Printf("Client %s (%s) disconnected", client.ID, client.Username)
}

// handleMessage processes incoming messages
func (h *Hub) handleMessage(message Message) {
	if h.config.EnableMetrics {
		h.metrics.incrementMessagesReceived()
	}

	switch message.Type {
	case MessageTypeJoin:
		h.handleJoinMessage(message)

	case MessageTypeLeave:
		h.handleLeaveMessage(message)

	case MessageTypeText:
		h.handleTextMessage(message)

	case MessageTypeWatchParty:
		h.handleWatchPartyMessage(message)

	case MessageTypeNotification:
		h.handleNotificationMessage(message)

	case MessageTypingStart, MessageTypingStop:
		h.handleTypingMessage(message)

	case MessageHeartbeat:
		h.handleHeartbeatMessage(message)

	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}

// handleJoinMessage handles join room messages
func (h *Hub) handleJoinMessage(message Message) {
	h.mu.RLock()
	client, ok := h.clients[message.UserID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	roomID := message.RoomID
	if roomID == "" {
		return
	}

	h.joinRoom(client, roomID)
}

// handleLeaveMessage handles leave room messages
func (h *Hub) handleLeaveMessage(message Message) {
	h.mu.RLock()
	client, ok := h.clients[message.UserID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	roomID := message.RoomID
	if roomID == "" {
		return
	}

	h.leaveRoom(client, roomID)
}

// handleTextMessage handles text messages
func (h *Hub) handleTextMessage(message Message) {
	roomID := message.RoomID
	if roomID == "" {
		return
	}

	h.broadcastToRoom(roomID, message)
}

// handleWatchPartyMessage handles watch party messages
func (h *Hub) handleWatchPartyMessage(message Message) {
	roomID := message.RoomID
	if roomID == "" {
		return
	}

	// Add watch party specific metadata
	if message.Metadata == nil {
		message.Metadata = make(map[string]interface{})
	}
	message.Metadata["is_watch_party"] = true

	h.broadcastToRoom(roomID, message)
}

// handleNotificationMessage handles notification messages
func (h *Hub) handleNotificationMessage(message Message) {
	// Send to specific user if specified
	if userID, ok := message.Metadata["target_user_id"].(string); ok {
		h.sendToUser(userID, message)
		return
	}

	// Otherwise broadcast to all user's rooms
	h.mu.RLock()
	client, ok := h.clients[message.UserID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	client.mu.RLock()
	for roomID := range client.Rooms {
		h.broadcastToRoom(roomID, message)
	}
	client.mu.RUnlock()
}

// handleTypingMessage handles typing indicators
func (h *Hub) handleTypingMessage(message Message) {
	roomID := message.RoomID
	if roomID == "" {
		return
	}

	// Send typing indicator to room (excluding sender)
	h.broadcastToRoomExcluding(roomID, message, message.UserID)
}

// handleHeartbeatMessage handles heartbeat messages
func (h *Hub) handleHeartbeatMessage(message Message) {
	h.mu.RLock()
	client, ok := h.clients[message.UserID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	client.LastSeen = time.Now()

	// Send heartbeat response
	response := Message{
		Type:      MessageHeartbeat,
		UserID:    "system",
		Username:  "System",
		Content:   "pong",
		Timestamp: time.Now(),
	}

	select {
	case client.Send <- response:
	default:
		// Client channel is blocked, consider disconnecting
	}
}

// joinRoom adds a client to a room
func (h *Hub) joinRoom(client *Client, roomID string) {
	// Get or create room
	h.mu.Lock()
	room, ok := h.rooms[roomID]
	if !ok {
		room = &Room{
			ID:      roomID,
			Name:    roomID,
			Clients: make(map[string]*Client),
			Created: time.Now(),
		}
		h.rooms[roomID] = room
	}
	h.mu.Unlock()

	// Add client to room
	room.mu.Lock()
	if len(room.Clients) >= h.config.MaxClientsPerRoom {
		room.mu.Unlock()

		// Send room full message
		fullMsg := Message{
			Type:      "error",
			UserID:    "system",
			Username:  "System",
			Content:   "Room is full",
			Timestamp: time.Now(),
		}

		select {
		case client.Send <- fullMsg:
		default:
		}
		return
	}

	room.Clients[client.ID] = client
	room.mu.Unlock()

	// Add room to client
	client.mu.Lock()
	client.Rooms[roomID] = true
	client.mu.Unlock()

	// Broadcast join message
	joinMsg := Message{
		Type:      "user_joined",
		RoomID:    roomID,
		UserID:    client.ID,
		Username:  client.Username,
		Content:   fmt.Sprintf("%s joined the room", client.Username),
		Timestamp: time.Now(),
	}

	h.broadcastToRoom(roomID, joinMsg)

	log.Printf("Client %s joined room %s", client.ID, roomID)
}

// leaveRoom removes a client from a room
func (h *Hub) leaveRoom(client *Client, roomID string) {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	// Remove client from room
	room.mu.Lock()
	delete(room.Clients, client.ID)
	isEmpty := len(room.Clients) == 0
	room.mu.Unlock()

	// Remove room from client
	client.mu.Lock()
	delete(client.Rooms, roomID)
	client.mu.Unlock()

	// Delete room if empty
	if isEmpty {
		h.mu.Lock()
		delete(h.rooms, roomID)
		h.mu.Unlock()
	} else {
		// Broadcast leave message
		leaveMsg := Message{
			Type:      "user_left",
			RoomID:    roomID,
			UserID:    client.ID,
			Username:  client.Username,
			Content:   fmt.Sprintf("%s left the room", client.Username),
			Timestamp: time.Now(),
		}

		h.broadcastToRoom(roomID, leaveMsg)
	}

	log.Printf("Client %s left room %s", client.ID, roomID)
}

// broadcastToRoom sends a message to all clients in a room
func (h *Hub) broadcastToRoom(roomID string, message Message) {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for _, client := range room.Clients {
		select {
		case client.Send <- message:
			if h.config.EnableMetrics {
				h.metrics.incrementMessagesSent()
			}
		default:
			// Client channel is blocked, close it
			close(client.Send)
			h.mu.Lock()
			delete(h.clients, client.ID)
			h.mu.Unlock()
		}
	}
}

// broadcastToRoomExcluding sends a message to all clients in a room except one
func (h *Hub) broadcastToRoomExcluding(roomID string, message Message, excludeUserID string) {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for _, client := range room.Clients {
		if client.ID == excludeUserID {
			continue
		}

		select {
		case client.Send <- message:
			if h.config.EnableMetrics {
				h.metrics.incrementMessagesSent()
			}
		default:
			// Client channel is blocked, close it
			close(client.Send)
			h.mu.Lock()
			delete(h.clients, client.ID)
			h.mu.Unlock()
		}
	}
}

// sendToUser sends a message to a specific user
func (h *Hub) sendToUser(userID string, message Message) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	select {
	case client.Send <- message:
		if h.config.EnableMetrics {
			h.metrics.incrementMessagesSent()
		}
	default:
		// Client channel is blocked, close it
		close(client.Send)
		h.mu.Lock()
		delete(h.clients, client.ID)
		h.mu.Unlock()
	}
}

// GetRoomInfo returns information about a room
func (h *Hub) GetRoomInfo(roomID string) map[string]interface{} {
	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if !ok {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	clients := make([]map[string]interface{}, 0, len(room.Clients))
	for _, client := range room.Clients {
		clients = append(clients, map[string]interface{}{
			"id":        client.ID,
			"username":  client.Username,
			"user_id":   client.UserID,
			"last_seen": client.LastSeen,
		})
	}

	return map[string]interface{}{
		"id":           room.ID,
		"name":         room.Name,
		"created":      room.Created,
		"clients":      clients,
		"client_count": len(clients),
	}
}

// GetHubStats returns hub statistics
func (h *Hub) GetHubStats() map[string]interface{} {
	h.metrics.mu.RLock()
	metrics := *h.metrics
	h.metrics.mu.RUnlock()

	return map[string]interface{}{
		"connected_clients":   metrics.ConnectedClients,
		"active_rooms":        metrics.ActiveRooms,
		"messages_sent":       metrics.MessagesSent,
		"messages_received":   metrics.MessagesReceived,
		"connections_created": metrics.ConnectionsCreated,
		"connections_closed":  metrics.ConnectionsClosed,
	}
}

// cleanup removes inactive clients and empty rooms
func (h *Hub) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.cleanupInactiveClients()
			h.cleanupEmptyRooms()
		case <-h.ctx.Done():
			return
		}
	}
}

// cleanupInactiveClients removes clients that haven't sent a heartbeat recently
func (h *Hub) cleanupInactiveClients() {
	now := time.Now()
	timeout := 10 * time.Minute

	h.mu.RLock()
	clientsToCheck := make([]*Client, 0, len(h.clients))
	for _, client := range h.clients {
		clientsToCheck = append(clientsToCheck, client)
	}
	h.mu.RUnlock()

	for _, client := range clientsToCheck {
		if now.Sub(client.LastSeen) > timeout {
			log.Printf("Removing inactive client: %s", client.ID)
			h.unregisterClient(client)
		}
	}
}

// cleanupEmptyRooms removes empty rooms
func (h *Hub) cleanupEmptyRooms() {
	h.mu.RLock()
	roomsToCheck := make(map[string]*Room)
	for id, room := range h.rooms {
		roomsToCheck[id] = room
	}
	h.mu.RUnlock()

	for id, room := range roomsToCheck {
		room.mu.RLock()
		isEmpty := len(room.Clients) == 0
		room.mu.RUnlock()

		if isEmpty {
			h.mu.Lock()
			delete(h.rooms, id)
			h.mu.Unlock()
			log.Printf("Removed empty room: %s", id)
		}
	}
}

// collectMetrics updates hub metrics
func (h *Hub) collectMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.mu.RLock()
			clientCount := len(h.clients)
			roomCount := len(h.rooms)
			h.mu.RUnlock()

			h.metrics.mu.Lock()
			h.metrics.ConnectedClients = int64(clientCount)
			h.metrics.ActiveRooms = int64(roomCount)
			h.metrics.mu.Unlock()
		case <-h.ctx.Done():
			return
		}
	}
}

// shutdown gracefully shuts down the hub
func (h *Hub) shutdown() {
	log.Println("Shutting down WebSocket hub...")

	// Close all client connections
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		close(client.Send)
		client.Conn.Close()
	}

	// Close channels
	close(h.broadcast)
	close(h.register)
	close(h.unregister)

	log.Println("WebSocket hub shutdown complete")
}

// Metrics methods

func (hm *HubMetrics) incrementConnectionsCreated() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.ConnectionsCreated++
}

func (hm *HubMetrics) incrementConnectionsClosed() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.ConnectionsClosed++
}

func (hm *HubMetrics) incrementMessagesSent() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.MessagesSent++
}

func (hm *HubMetrics) incrementMessagesReceived() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.MessagesReceived++
}

// Stop stops the hub
func (h *Hub) Stop() {
	h.cancel()
}
