package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketUpgrader handles upgrading HTTP connections to WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// WebSocketServer handles WebSocket connections
type WebSocketServer struct {
	hub    *Hub
	config HubConfig
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(config HubConfig) *WebSocketServer {
	return &WebSocketServer{
		hub:    NewHub(config),
		config: config,
	}
}

// Start starts the WebSocket hub
func (ws *WebSocketServer) Start() {
	go ws.hub.Run()
}

// Stop stops the WebSocket hub
func (ws *WebSocketServer) Stop() {
	ws.hub.Stop()
}

// HandleWebSocket handles WebSocket connection requests
func (ws *WebSocketServer) HandleWebSocket(c *gin.Context) {
	// Extract authenticated user information from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Get username from context or user lookup - for now, use userID as fallback
	username := c.GetString("username")
	if username == "" {
		username = userID // Fallback to userID if username not available
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create new client
	client := &Client{
		ID:       generateClientID(),
		UserID:   userID,
		Username: username,
		Conn:     conn,
		Send:     make(chan Message, ws.config.MessageBufferSize),
		Rooms:    make(map[string]bool),
		LastSeen: time.Now(),
	}

	// Register client with hub
	ws.hub.register <- client

	// Start client goroutines
	go client.writePump(ws.config)
	go client.readPump(ws.hub)
}

// GetHub returns the WebSocket hub
func (ws *WebSocketServer) GetHub() *Hub {
	return ws.hub
}

// BroadcastMessage broadcasts a message to all clients
func (ws *WebSocketServer) BroadcastMessage(messageType, content string) {
	message := Message{
		Type:      messageType,
		UserID:    "system",
		Username:  "System",
		Content:   content,
		Timestamp: time.Now(),
	}

	select {
	case ws.hub.broadcast <- message:
	default:
		log.Printf("Broadcast channel is full, dropping message")
	}
}

// SendToRoom sends a message to all clients in a room
func (ws *WebSocketServer) SendToRoom(roomID, messageType, content string) {
	message := Message{
		Type:      messageType,
		RoomID:    roomID,
		UserID:    "system",
		Username:  "System",
		Content:   content,
		Timestamp: time.Now(),
	}

	select {
	case ws.hub.broadcast <- message:
	default:
		log.Printf("Broadcast channel is full, dropping room message")
	}
}

// SendToUser sends a message to a specific user
func (ws *WebSocketServer) SendToUser(userID, messageType, content string) {
	message := Message{
		Type:      messageType,
		UserID:    "system",
		Username:  "System",
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"target_user_id": userID},
	}

	select {
	case ws.hub.broadcast <- message:
	default:
		log.Printf("Broadcast channel is full, dropping user message")
	}
}

// GetRoomStats returns statistics for a room
func (ws *WebSocketServer) GetRoomStats(roomID string) map[string]interface{} {
	return ws.hub.GetRoomInfo(roomID)
}

// GetServerStats returns server statistics
func (ws *WebSocketServer) GetServerStats() map[string]interface{} {
	return ws.hub.GetHubStats()
}

// Client methods

// readPump handles reading messages from WebSocket connection
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	// Set read deadline and pong handler
	c.Conn.SetReadLimit(hub.config.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(hub.config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(hub.config.PongWait))
		return nil
	})

	for {
		var message Message
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Set message metadata using authenticated user identity
		message.UserID = c.UserID
		message.Username = c.Username
		message.Timestamp = time.Now()

		// Send message to hub
		select {
		case hub.broadcast <- message:
		default:
			log.Printf("Broadcast channel is full, dropping message from client %s", c.ID)
		}
	}
}

// writePump handles writing messages to WebSocket connection
func (c *Client) writePump(config HubConfig) {
	ticker := time.NewTicker(config.PingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WriteJSON(message)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Helper functions

// generateClientID generates a unique client ID
func generateClientID() string {
	return "client_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// NotificationService handles sending notifications through WebSocket
type NotificationService struct {
	server *WebSocketServer
}

// NewNotificationService creates a new notification service
func NewNotificationService(server *WebSocketServer) *NotificationService {
	return &NotificationService{
		server: server,
	}
}

// SendNotification sends a notification to a user
func (ns *NotificationService) SendNotification(userID, title, message string) {
	notification := map[string]interface{}{
		"title":   title,
		"message": message,
		"type":    "notification",
	}

	data, _ := json.Marshal(notification)

	msg := Message{
		Type:      MessageTypeNotification,
		UserID:    "system",
		Username:  "System",
		Content:   string(data),
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"target_user_id": userID},
	}

	select {
	case ns.server.hub.broadcast <- msg:
	default:
		log.Printf("Broadcast channel is full, dropping notification")
	}
}

// SendBroadcast sends a broadcast notification to all users
func (ns *NotificationService) SendBroadcast(title, message string) {
	notification := map[string]interface{}{
		"title":   title,
		"message": message,
		"type":    "broadcast",
	}

	data, _ := json.Marshal(notification)

	ns.server.BroadcastMessage(MessageTypeNotification, string(data))
}

// WatchPartyService manages watch party functionality
type WatchPartyService struct {
	server *WebSocketServer
}

// NewWatchPartyService creates a new watch party service
func NewWatchPartyService(server *WebSocketServer) *WatchPartyService {
	return &WatchPartyService{
		server: server,
	}
}

// CreateWatchParty creates a new watch party room
func (wps *WatchPartyService) CreateWatchParty(movieID, creatorID string) string {
	roomID := "watch_party_" + movieID + "_" + randomString(8)

	// Send notification about new watch party
	message := map[string]interface{}{
		"action":   "created",
		"movie_id": movieID,
		"room_id":  roomID,
	}

	data, _ := json.Marshal(message)

	wps.server.SendToRoom(roomID, MessageTypeWatchParty, string(data))

	return roomID
}

// JoinWatchParty allows a user to join a watch party
func (wps *WatchPartyService) JoinWatchParty(roomID, userID string) {
	message := map[string]interface{}{
		"action":  "joined",
		"user_id": userID,
	}

	data, _ := json.Marshal(message)

	wps.server.SendToRoom(roomID, MessageTypeWatchParty, string(data))
}

// SyncPlayback synchronizes playback state across watch party
func (wps *WatchPartyService) SyncPlayback(roomID, userID string, currentTime float64, isPlaying bool) {
	message := map[string]interface{}{
		"action":       "sync",
		"user_id":      userID,
		"current_time": currentTime,
		"is_playing":   isPlaying,
	}

	data, _ := json.Marshal(message)

	wps.server.SendToRoom(roomID, MessageTypeWatchParty, string(data))
}

// ChatService handles real-time chat functionality
type ChatService struct {
	server *WebSocketServer
}

// NewChatService creates a new chat service
func NewChatService(server *WebSocketServer) *ChatService {
	return &ChatService{
		server: server,
	}
}

// SendMessage sends a chat message to a room
func (cs *ChatService) SendMessage(roomID, userID, username, content string) {
	message := Message{
		Type:      MessageTypeText,
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
		Content:   content,
		Timestamp: time.Now(),
	}

	select {
	case cs.server.hub.broadcast <- message:
	default:
		log.Printf("Broadcast channel is full, dropping chat message")
	}
}

// HandleWebSocketConnection handles WebSocket connection using an existing hub
func HandleWebSocketConnection(hub *Hub, c *gin.Context) {
	// Extract authenticated user information from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Get username from context or user lookup - for now, use userID as fallback
	username := c.GetString("username")
	if username == "" {
		username = userID // Fallback to userID if username not available
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create new client
	client := &Client{
		ID:       generateClientID(),
		UserID:   userID,
		Username: username,
		Conn:     conn,
		Send:     make(chan Message, hub.config.MessageBufferSize),
		Rooms:    make(map[string]bool),
		LastSeen: time.Now(),
	}

	// Register client with hub
	hub.register <- client

	// Start client goroutines
	go client.writePump(hub.config)
	go client.readPump(hub)
}

// SendTypingIndicator sends typing indicator to a room
func (cs *ChatService) SendTypingIndicator(roomID, userID, username string, isTyping bool) {
	messageType := MessageTypingStop
	if isTyping {
		messageType = MessageTypingStart
	}

	message := Message{
		Type:      messageType,
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
		Content:   map[string]interface{}{"is_typing": isTyping},
		Timestamp: time.Now(),
	}

	select {
	case cs.server.hub.broadcast <- message:
	default:
		log.Printf("Broadcast channel is full, dropping typing indicator")
	}
}
