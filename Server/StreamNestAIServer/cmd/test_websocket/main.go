package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket message types
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type ActivityMessage struct {
	UserID  string `json:"user_id"`
	Action  string `json:"action"`
	MovieID string `json:"movie_id"`
	Data    string `json:"data"`
}

type WatchPartyMessage struct {
	PartyID string `json:"party_id"`
	UserID  string `json:"user_id"`
	Action  string `json:"action"`
	Time    int64  `json:"time"`
}

func main() {
	fmt.Println("🔍 Testing WebSocket Connection...")
	fmt.Println("📍 WebSocket Server: ws://localhost:8080/ws")

	// Test WebSocket connection
	testWebSocketConnection()

	// Test activity broadcasting
	testActivityBroadcast()

	// Test watch party functionality
	testWatchPartyFunctionality()

	fmt.Println("\n🎉 All WebSocket tests completed!")
}

func testWebSocketConnection() {
	fmt.Println("\n🔄 Testing WebSocket Connection...")
	
	// Connect to WebSocket server with required query parameters
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	q := u.Query()
	q.Set("user_id", "test_user_123")
	q.Set("username", "testuser")
	u.RawQuery = q.Encode()
	fmt.Printf("🔗 Connecting to: %s\n", u.String())

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("❌ WebSocket connection failed: %v", err)
	}
	defer conn.Close()

	fmt.Println("✅ WebSocket connection established")

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// Test ping/pong
	fmt.Println("📡 Testing ping/pong...")
	err = conn.WriteMessage(websocket.PingMessage, []byte("test ping"))
	if err != nil {
		log.Printf("❌ Ping failed: %v", err)
	} else {
		fmt.Println("✅ Ping sent successfully")
	}

	// Wait for pong
	_, _, err = conn.ReadMessage()
	if err != nil {
		log.Printf("❌ Pong not received: %v", err)
	} else {
		fmt.Println("✅ Pong received successfully")
	}

	// Test JSON message sending
	fmt.Println("📝 Testing JSON message sending...")
	testMessage := WSMessage{
		Type: "test",
		Payload: map[string]string{
			"message":   "Hello WebSocket!",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	messageBytes, err := json.Marshal(testMessage)
	if err != nil {
		log.Fatalf("❌ JSON marshal failed: %v", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		log.Printf("❌ Message send failed: %v", err)
	} else {
		fmt.Println("✅ JSON message sent successfully")
	}

	// Test message receiving
	fmt.Println("📨 Testing message receiving...")
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Printf("❌ Message receive failed: %v", err)
	} else {
		fmt.Printf("✅ Message received: %s\n", string(message))
	}
}

func testActivityBroadcast() {
	fmt.Println("\n📢 Testing Activity Broadcast...")

	// Connect multiple clients to test broadcasting
	clients := make([]*websocket.Conn, 3)

	for i := 0; i < 3; i++ {
		u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
		q := u.Query()
		q.Set("user_id", fmt.Sprintf("test_user_%d", i+1))
		q.Set("username", fmt.Sprintf("testuser%d", i+1))
		u.RawQuery = q.Encode()
		
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("❌ Client %d connection failed: %v", i+1, err)
			return
		}
		clients[i] = conn
		fmt.Printf("✅ Client %d connected\n", i+1)
	}

	// Clean up connections
	defer func() {
		for i, conn := range clients {
			if conn != nil {
				conn.Close()
				fmt.Printf("🔌 Client %d disconnected\n", i+1)
			}
		}
	}()

	// Send activity message from first client
	fmt.Println("📤 Sending activity message...")
	activity := ActivityMessage{
		UserID:  "test_user_123",
		Action:  "watching",
		MovieID: "movie_456",
		Data:    "Started watching The Matrix",
	}

	activityMessage := WSMessage{
		Type:    "activity",
		Payload: activity,
	}

	messageBytes, err := json.Marshal(activityMessage)
	if err != nil {
		log.Printf("❌ Activity message marshal failed: %v", err)
		return
	}

	err = clients[0].WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		log.Printf("❌ Activity message send failed: %v", err)
		return
	}

	fmt.Println("✅ Activity message sent")

	// Check if other clients receive the broadcast
	time.Sleep(1 * time.Second) // Give time for broadcast

	for i := 1; i < 3; i++ {
		clients[i].SetReadDeadline(time.Now().Add(2 * time.Second))
		_, message, err := clients[i].ReadMessage()
		if err != nil {
			log.Printf("❌ Client %d did not receive broadcast: %v", i+1, err)
		} else {
			fmt.Printf("✅ Client %d received broadcast: %s\n", i+1, string(message))
		}
	}
}

func testWatchPartyFunctionality() {
	fmt.Println("\n🎬 Testing Watch Party Functionality...")

	// Connect to WebSocket
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("❌ Watch party connection failed: %v", err)
		return
	}
	defer conn.Close()

	fmt.Println("✅ Watch party client connected")

	// Join watch party
	fmt.Println("🎟️ Joining watch party...")
	joinMessage := WSMessage{
		Type: "join_watch_party",
		Payload: map[string]string{
			"party_id": "party_123",
			"user_id":  "user_456",
		},
	}

	messageBytes, err := json.Marshal(joinMessage)
	if err != nil {
		log.Printf("❌ Join message marshal failed: %v", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		log.Printf("❌ Join message send failed: %v", err)
		return
	}

	fmt.Println("✅ Watch party join request sent")

	// Send watch party action
	fmt.Println("⏯️ Sending watch party action...")
	action := WatchPartyMessage{
		PartyID: "party_123",
		UserID:  "user_456",
		Action:  "play",
		Time:    time.Now().Unix(),
	}

	actionMessage := WSMessage{
		Type:    "watch_party_action",
		Payload: action,
	}

	actionBytes, err := json.Marshal(actionMessage)
	if err != nil {
		log.Printf("❌ Action message marshal failed: %v", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, actionBytes)
	if err != nil {
		log.Printf("❌ Action message send failed: %v", err)
		return
	}

	fmt.Println("✅ Watch party action sent")

	// Listen for responses
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	for i := 0; i < 3; i++ { // Try to read up to 3 messages
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("📭 No more messages (iteration %d)\n", i+1)
			break
		}

		var wsMsg WSMessage
		err = json.Unmarshal(message, &wsMsg)
		if err != nil {
			fmt.Printf("📨 Received raw message: %s\n", string(message))
			continue
		}

		fmt.Printf("📨 Received structured message: Type=%s, Payload=%+v\n", wsMsg.Type, wsMsg.Payload)
	}
}

func testHTTPHealthCheck() {
	fmt.Println("\n🏥 Testing HTTP Health Check...")

	client := &http.Client{Timeout: 5 * time.Second}

	// Test health endpoint
	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		log.Printf("❌ Health check failed: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ Health check status: %s\n", resp.Status)

	// Test metrics endpoint
	resp, err = client.Get("http://localhost:8080/metrics")
	if err != nil {
		log.Printf("❌ Metrics check failed: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ Metrics check status: %s\n", resp.Status)
}
