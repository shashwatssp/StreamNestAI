package middleware

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/security"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/utils"
)

// SecurityConfig holds security configuration
type SecurityConfig struct {
	CSRFEnabled           bool
	CSRFTokenLength       int
	CSRFTokenExpiry       time.Duration
	InputValidation       bool
	RateLimitEnabled      bool
	AuditLogging          bool
	MaxRequestSize        int64
	AllowedOrigins        []string
	AllowedMethods        []string
	AllowedHeaders        []string
	EnableCORS            bool
	EnableSecurityHeaders bool
}

// SecurityMiddleware provides security features
type SecurityMiddleware struct {
	config          SecurityConfig
	validate        *validator.Validate
	securityService *security.SecurityService
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config SecurityConfig, securityService *security.SecurityService) *SecurityMiddleware {
	return &SecurityMiddleware{
		config:          config,
		validate:        validator.New(),
		securityService: securityService,
	}
}

// CSRFProtection provides CSRF protection
func (sm *SecurityMiddleware) CSRFProtection() gin.HandlerFunc {
	if !sm.config.CSRFEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Skip CSRF for GET, HEAD, OPTIONS
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get CSRF token from header
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = c.PostForm("_csrf")
		}

		// Get user ID from context (should be set by auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Validate CSRF token using security service
		if sm.securityService == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Security service not available"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		if !sm.securityService.ValidateCSRFToken(csrfToken, userIDStr) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken generates a new CSRF token
func (sm *SecurityMiddleware) GenerateCSRFToken() gin.HandlerFunc {
	if !sm.config.CSRFEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Only generate CSRF token for authenticated users
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// Generate new CSRF token using security service
		if sm.securityService == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Security service not available"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		csrfToken, err := sm.securityService.GenerateCSRFToken(userIDStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSRF token"})
			c.Abort()
			return
		}

		// Set CSRF token in context for validation
		c.Set("csrf_token", csrfToken)

		// Set CSRF token in header for GET requests
		if c.Request.Method == "GET" {
			c.Header("X-CSRF-Token", csrfToken)
		}

		c.Next()
	}
}

// InputValidation provides input validation
func (sm *SecurityMiddleware) InputValidation() gin.HandlerFunc {
	if !sm.config.InputValidation {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Validate request size
		if c.Request.ContentLength > sm.config.MaxRequestSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request too large"})
			c.Abort()
			return
		}

		// Validate URL parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if !sm.isValidInput(value) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input in query parameter: " + key})
					c.Abort()
					return
				}
			}
		}

		// Validate headers for common injection patterns
		for key, values := range c.Request.Header {
			for _, value := range values {
				if sm.containsInjectionPatterns(value) {
					sm.logSecurityEvent(c, "header_injection_attempt", map[string]interface{}{
						"header": key,
						"value":  value,
					})
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid header value"})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// SecurityHeaders adds security headers
func (sm *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	if !sm.config.EnableSecurityHeaders {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' wss: https:; frame-ancestors 'none';")

		// Strict Transport Security
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		c.Next()
	}
}

// CORS provides CORS protection
func (sm *SecurityMiddleware) CORS() gin.HandlerFunc {
	if !sm.config.EnableCORS {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range sm.config.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(sm.config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(sm.config.AllowedHeaders, ", "))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuditLogging provides audit logging
func (sm *SecurityMiddleware) AuditLogging() gin.HandlerFunc {
	if !sm.config.AuditLogging {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log audit information
		duration := time.Since(start)
		userID, _ := c.Get("user_id")

		auditData := map[string]interface{}{
			"timestamp":     time.Now(),
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"query":         c.Request.URL.RawQuery,
			"status":        c.Writer.Status(),
			"duration":      duration,
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"user_id":       userID,
			"request_size":  c.Request.ContentLength,
			"response_size": c.Writer.Size(),
		}

		// Add security events if any
		if securityEvents, exists := c.Get("security_events"); exists {
			auditData["security_events"] = securityEvents
		}

		sm.logAuditEvent(auditData)
	}
}

// RateLimiting provides rate limiting (simplified version)
func (sm *SecurityMiddleware) RateLimiting() gin.HandlerFunc {
	if !sm.config.RateLimitEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// This is a simplified rate limiting implementation
		// In production, you would use Redis or a dedicated rate limiting service

		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// For now, just log the request
		// In a real implementation, you would check and update rate limits
		sm.logSecurityEvent(c, "rate_limit_check", map[string]interface{}{
			"client_ip": clientIP,
			"key":       key,
		})

		c.Next()
	}
}

// JWTValidation validates JWT tokens
func (sm *SecurityMiddleware) JWTValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = tokenString[7:]
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(utils.SECRET_KEY), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user_id", claims["user_id"])
			c.Set("username", claims["username"])
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper methods

func (sm *SecurityMiddleware) isValidInput(input string) bool {
	// Check for SQL injection patterns
	sqlPatterns := []string{
		`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`,
		`(?i)(or|and)\s+\d+\s*=\s*\d+`,
		`(?i)(or|and)\s+['"]?\w+['"]?\s*=\s*['"]?\w+['"]?`,
		`(?i)(--|;|/\*|\*/)`,
	}

	for _, pattern := range sqlPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return false
		}
	}

	// Check for XSS patterns
	xssPatterns := []string{
		`(?i)(<script|</script|javascript:|vbscript:|onload=|onerror=)`,
		`(?i)(<iframe|<object|<embed|<link)`,
		`(?i)(expression\(|@import|behavior:)`,
	}

	for _, pattern := range xssPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return false
		}
	}

	// Check for path traversal patterns
	pathTraversalPatterns := []string{
		`\.\./`,
		`%2e%2e%2f`,
		`%2e%2e\\`,
		`\.\.\\`,
	}

	for _, pattern := range pathTraversalPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return false
		}
	}

	return true
}

func (sm *SecurityMiddleware) containsInjectionPatterns(input string) bool {
	// Check for common injection patterns in headers
	patterns := []string{
		`(?i)(union|select|insert|update|delete|drop|create|alter)`,
		`(?i)(<script|javascript:|vbscript:)`,
		`(?i)(\.\./|%2e%2e)`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}

	return false
}

func (sm *SecurityMiddleware) logSecurityEvent(c *gin.Context, eventType string, data map[string]interface{}) {
	if !sm.config.AuditLogging {
		return
	}

	// Add to context for later logging
	events, exists := c.Get("security_events")
	if !exists {
		events = []map[string]interface{}{}
	}

	securityEvent := map[string]interface{}{
		"type":      eventType,
		"timestamp": time.Now(),
		"data":      data,
		"client_ip": c.ClientIP(),
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
	}

	events = append(events.([]map[string]interface{}), securityEvent)
	c.Set("security_events", events)
}

func (sm *SecurityMiddleware) logAuditEvent(data map[string]interface{}) {
	// In a real implementation, you would log to a file, database, or logging service
	// For now, we'll just print it (in production, use proper logging)
	fmt.Printf("AUDIT: %+v\n", data)
}

// ValidationMiddleware provides request validation
func (sm *SecurityMiddleware) ValidationMiddleware(validationRules map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate request body based on rules
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
			c.Abort()
			return
		}

		// Apply validation rules
		for field, rule := range validationRules {
			value, exists := body[field]
			if !exists {
				continue
			}

			switch rule {
			case "required":
				if value == nil || value == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": field + " is required"})
					c.Abort()
					return
				}
			case "email":
				if !sm.isValidEmail(fmt.Sprintf("%v", value)) {
					c.JSON(http.StatusBadRequest, gin.H{"error": field + " must be a valid email"})
					c.Abort()
					return
				}
			case "alpha":
				if !sm.isAlpha(fmt.Sprintf("%v", value)) {
					c.JSON(http.StatusBadRequest, gin.H{"error": field + " must contain only letters"})
					c.Abort()
					return
				}
			case "alphanumeric":
				if !sm.isAlphanumeric(fmt.Sprintf("%v", value)) {
					c.JSON(http.StatusBadRequest, gin.H{"error": field + " must contain only letters and numbers"})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// Validation helper methods
func (sm *SecurityMiddleware) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (sm *SecurityMiddleware) isAlpha(s string) bool {
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	return alphaRegex.MatchString(s)
}

func (sm *SecurityMiddleware) isAlphanumeric(s string) bool {
	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return alphanumericRegex.MatchString(s)
}

// IPWhitelist provides IP whitelist functionality
func (sm *SecurityMiddleware) IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check if client IP is in whitelist
		allowed := false
		for _, ip := range allowedIPs {
			if ip == clientIP {
				allowed = true
				break
			}
		}

		if !allowed {
			sm.logSecurityEvent(c, "ip_whitelist_violation", map[string]interface{}{
				"client_ip":   clientIP,
				"allowed_ips": allowedIPs,
			})
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied from this IP"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestSizeLimit limits request size
func (sm *SecurityMiddleware) RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("Request size exceeds limit of %d bytes", maxSize),
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}
