package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SecurityService handles security features like CSRF protection, audit logs, and input validation
type SecurityService struct {
	client *mongo.Client
	db     *mongo.Database
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	Action    string    `bson:"action" json:"action"`
	Resource  string    `bson:"resource" json:"resource"`
	IPAddress string    `bson:"ip_address" json:"ip_address"`
	UserAgent string    `bson:"user_agent" json:"user_agent"`
	Success   bool      `bson:"success" json:"success"`
	Details   bson.M    `bson:"details" json:"details"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// CSRFToken represents a CSRF token
type CSRFToken struct {
	Token     string    `bson:"token" json:"token"`
	UserID    string    `bson:"user_id" json:"user_id"`
	ExpiresAt time.Time `bson:"expires_at" json:"expires_at"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// SecurityEvent represents a security event
type SecurityEvent struct {
	ID        string    `bson:"_id" json:"id"`
	Type      string    `bson:"type" json:"type"`         // brute_force, suspicious_activity, etc.
	Severity  string    `bson:"severity" json:"severity"` // low, medium, high, critical
	UserID    string    `bson:"user_id" json:"user_id"`
	IPAddress string    `bson:"ip_address" json:"ip_address"`
	Details   bson.M    `bson:"details" json:"details"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// NewSecurityService creates a new security service
func NewSecurityService(client *mongo.Client) *SecurityService {
	return &SecurityService{
		client: client,
		db:     client.Database("streamnestai"),
	}
}

// Initialize sets up the security service
func (ss *SecurityService) Initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create indexes for security collections
	collections := map[string][]bson.D{
		"audit_logs": {
			{{"user_id", 1}, {"created_at", -1}},
			{{"created_at", -1}},
			{{"action", 1}},
		},
		"csrf_tokens": {
			{{"token", 1}},
			{{"user_id", 1}},
			{{"expires_at", 1}},
		},
		"security_events": {
			{{"type", 1}, {"created_at", -1}},
			{{"severity", 1}, {"created_at", -1}},
			{{"ip_address", 1}},
		},
	}

	for collName, indexes := range collections {
		coll := ss.db.Collection(collName)
		for _, index := range indexes {
			_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: index})
			if err != nil {
				log.Printf("Warning: Failed to create index for %s: %v", collName, err)
			}
		}
	}

	log.Println("Security service initialized successfully")
	return nil
}

// LogAudit logs an audit event
func (ss *SecurityService) LogAudit(userID, action, resource, ipAddress, userAgent string, success bool, details bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("audit_logs")

	auditLog := AuditLog{
		ID:        bson.NewObjectID().Hex(),
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
		Details:   details,
		CreatedAt: time.Now(),
	}

	_, err := coll.InsertOne(ctx, auditLog)
	return err
}

// GenerateCSRFToken generates a new CSRF token for a user
func (ss *SecurityService) GenerateCSRFToken(userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Generate random token
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Clean up old tokens for this user
	coll := ss.db.Collection("csrf_tokens")
	_, err = coll.DeleteMany(ctx, bson.M{
		"user_id":    userID,
		"expires_at": bson.M{"$lt": time.Now()},
	})
	if err != nil {
		log.Printf("Warning: Failed to clean up old CSRF tokens: %v", err)
	}

	// Insert new token
	csrfToken := CSRFToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour expiry
		CreatedAt: time.Now(),
	}

	_, err = coll.InsertOne(ctx, csrfToken)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateCSRFToken validates a CSRF token
func (ss *SecurityService) ValidateCSRFToken(token, userID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("csrf_tokens")
	var csrfToken CSRFToken

	err := coll.FindOne(ctx, bson.M{
		"token":      token,
		"user_id":    userID,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&csrfToken)

	return err == nil
}

// RevokeCSRFToken revokes a CSRF token
func (ss *SecurityService) RevokeCSRFToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("csrf_tokens")
	_, err := coll.DeleteOne(ctx, bson.M{"token": token})
	return err
}

// LogSecurityEvent logs a security event
func (ss *SecurityService) LogSecurityEvent(eventType, severity, userID, ipAddress string, details bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("security_events")

	event := SecurityEvent{
		ID:        bson.NewObjectID().Hex(),
		Type:      eventType,
		Severity:  severity,
		UserID:    userID,
		IPAddress: ipAddress,
		Details:   details,
		CreatedAt: time.Now(),
	}

	_, err := coll.InsertOne(ctx, event)
	return err
}

// GetAuditLogs retrieves audit logs for a user or all users
func (ss *SecurityService) GetAuditLogs(userID string, limit int) ([]AuditLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("audit_logs")

	filter := bson.M{}
	if userID != "" {
		filter["user_id"] = userID
	}

	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit))
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []AuditLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// GetSecurityEvents retrieves security events
func (ss *SecurityService) GetSecurityEvents(eventType, severity string, limit int) ([]SecurityEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("security_events")

	filter := bson.M{}
	if eventType != "" {
		filter["type"] = eventType
	}
	if severity != "" {
		filter["severity"] = severity
	}

	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit))
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []SecurityEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// ValidateInput performs basic input validation
func (ss *SecurityService) ValidateInput(input string, maxLength int) bool {
	if len(input) > maxLength {
		return false
	}

	// Check for common SQL injection patterns
	dangerousPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"<script", "</script>", "javascript:", "vbscript:",
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "UNION",
	}

	for _, pattern := range dangerousPatterns {
		if contains(input, pattern) {
			return false
		}
	}

	return true
}

// DetectSuspiciousActivity detects suspicious patterns in user behavior
func (ss *SecurityService) DetectSuspiciousActivity(userID, ipAddress string, actions []string) bool {
	// Check for rapid successive actions
	if len(actions) > 100 { // More than 100 actions in short time
		ss.LogSecurityEvent("rapid_actions", "medium", userID, ipAddress, bson.M{
			"action_count": len(actions),
		})
		return true
	}

	// Check for unusual patterns
	loginAttempts := 0
	for _, action := range actions {
		if action == "login_attempt" {
			loginAttempts++
		}
	}

	if loginAttempts > 10 { // More than 10 login attempts
		ss.LogSecurityEvent("brute_force", "high", userID, ipAddress, bson.M{
			"login_attempts": loginAttempts,
		})
		return true
	}

	return false
}

// CleanupExpiredTokens cleans up expired CSRF tokens
func (ss *SecurityService) CleanupExpiredTokens() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	coll := ss.db.Collection("csrf_tokens")
	result, err := coll.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	if err != nil {
		return err
	}

	log.Printf("Cleaned up %d expired CSRF tokens", result.DeletedCount)
	return nil
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
