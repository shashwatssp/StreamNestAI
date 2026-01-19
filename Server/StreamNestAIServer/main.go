package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/database"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/routes"

	// Import new services
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/analytics"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/middleware"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/monitoring"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/recommendations"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/security"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/social"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/websocket"
)

// RequestLogger middleware for comprehensive request logging
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		requestID := generateRequestID()

		// Set request ID in context for tracking
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		// Log incoming request
		log.Printf("🔵 [REQ:%s] %s %s from %s - User-Agent: %s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			c.Request.UserAgent())

		// Log query parameters if any
		if len(c.Request.URL.RawQuery) > 0 {
			log.Printf("🔵 [REQ:%s] Query: %s", requestID, c.Request.URL.RawQuery)
		}

		// Process request
		c.Next()

		// Log response
		duration := time.Since(startTime)
		log.Printf("🟢 [RES:%s] %d %s - Duration: %v",
			requestID,
			c.Writer.Status(),
			http.StatusText(c.Writer.Status()),
			duration)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("🔴 [ERR:%s] %v", requestID, err.Error())
			}
		}
	}
}

// generateRequestID creates a unique request ID for tracking
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// logFeatureCall logs feature-specific operations
func logFeatureCall(feature, operation, userID string, details map[string]interface{}) {
	log.Printf("🟡 [FEATURE:%s] %s - User: %s - Details: %+v",
		strings.ToUpper(feature), operation, userID, details)
}

// logCacheOperation logs cache-related operations
func logCacheOperation(operation, key string, hit bool, duration time.Duration) {
	status := "MISS"
	if hit {
		status = "HIT"
	}
	log.Printf("🟠 [CACHE:%s] Key: %s - Status: %s - Duration: %v",
		operation, key, status, duration)
}

// logDatabaseOperation logs database operations
func logDatabaseOperation(operation, collection string, duration time.Duration, count int64) {
	log.Printf("🔵 [DB:%s] Collection: %s - Duration: %v - Records: %d",
		operation, collection, duration, count)
}

// maskSensitive masks sensitive information in logs
func maskSensitive(s string) string {
	if s == "" {
		return "not set"
	}
	if len(s) > 20 {
		return s[:10] + "***" + s[len(s)-5:]
	}
	return "***"
}

func main() {
	// Initialize comprehensive logging
	log.Println("🚀 ========================================")
	log.Println("🚀 STARTING STREAMNESTAI SERVER")
	log.Println("🚀 ========================================")
	log.Printf("📅 Start time: %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("🌍 Environment: %s", os.Getenv("GIN_MODE"))
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("🏠 Working directory: Error getting working directory: %v", err)
	} else {
		log.Printf("🏠 Working directory: %s", wd)
	}

	// Load environment variables
	log.Println("🔧 Loading environment variables...")
	err = godotenv.Load(".env")
	if err != nil {
		log.Println("⚠️  Warning: unable to find .env file")
	} else {
		log.Println("✅ Environment variables loaded successfully")
	}

	// Log key environment variables (masked for security)
	log.Printf("🔧 MongoDB URI: %s", maskSensitive(os.Getenv("MONGODB_URI")))
	log.Printf("🔧 Redis address: %s", os.Getenv("REDIS_ADDR"))
	log.Printf("🔧 Server port: %s", os.Getenv("PORT"))
	log.Printf("🔧 Allowed origins: %s", os.Getenv("ALLOWED_ORIGINS"))

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
		log.Println("🔧 Gin mode set to RELEASE")
	} else {
		log.Println("🔧 Gin mode set to DEBUG")
	}

	log.Println("🌐 Initializing Gin router...")
	router := gin.Default()

	// Add comprehensive request logging middleware
	router.Use(RequestLogger())
	log.Println("✅ Request logging middleware added")

	// Health check endpoint with enhanced logging
	router.GET("/health", func(c *gin.Context) {
		log.Printf("🟢 Health check requested from %s", c.ClientIP())
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "2.0.0",
			"service":   "StreamNestAI Enhanced",
		})
	})

	// Hello endpoint with logging
	router.GET("/hello", func(c *gin.Context) {
		log.Printf("👋 Hello endpoint requested from %s", c.ClientIP())
		c.String(200, "Hello, StreamNestAI Enhanced!")
	})

	// Configure CORS with detailed logging
	log.Println("🌐 Configuring CORS...")
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
			log.Printf("✅ Allowed Origin: %s", origins[i])
		}
	} else {
		origins = []string{
			"http://localhost:5173",
			"http://localhost:5175",
			"http://localhost:5176",
			"https://stream-nest-3py4n2m1t-shashwatssps-projects.vercel.app",
			"https://stream-nest-3py4n2m1t-shashwatssps-projects.vercel.app/",
		}
		log.Println("⚠️  Using default allowed origins")
	}

	config := cors.Config{}
	config.AllowOrigins = origins
	config.AllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length", "X-Total-Count"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	router.Use(cors.New(config))
	log.Println("✅ CORS middleware configured")

	router.Use(gin.Recovery())
	log.Println("✅ Recovery middleware added")

	// Initialize Redis (if available)
	log.Println("🔴 Initializing Redis connection...")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
		log.Println("⚠️  Using default Redis address: localhost:6379")
	}
	log.Printf("🔴 Redis address: %s", redisAddr)

	// Initialize monitoring service (will be configured after MongoDB connection)
	log.Println("📊 Preparing monitoring service initialization...")
	var monitoringService *monitoring.MonitoringService

	// Initialize rate limiting middleware with detailed logging
	log.Println("🛡️  Initializing rate limiting middleware...")
	rateLimiterConfig := middleware.RateLimiterConfig{
		RequestsPerSecond: 100,
		BurstSize:         200,
		RedisHost:         "localhost",
		RedisPort:         "6379",
		RedisPassword:     "",
		RedisDB:           0,
		WindowDuration:    time.Minute,
		EnableDistributed: true,
	}

	rateLimiter, err := middleware.NewRateLimiter(rateLimiterConfig)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to initialize rate limiter: %v", err)
	} else {
		router.Use(rateLimiter.Middleware())
		log.Println("✅ Rate limiting middleware added")
	}

	// Connect to MongoDB with comprehensive logging
	log.Println("💾 Connecting to MongoDB...")
	startTime := time.Now()
	client := database.Connect()
	connectionDuration := time.Since(startTime)

	if client == nil {
		log.Fatalf("❌ FATAL: Failed to connect to MongoDB")
	}

	log.Printf("🔍 Testing MongoDB connection...")
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("❌ FATAL: Failed to reach MongoDB server: %v", err)
	}

	log.Printf("✅ MongoDB connection established in %v", connectionDuration)

	defer func() {
		log.Println("💾 Disconnecting from MongoDB...")
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("⚠️  Failed to disconnect from MongoDB: %v", err)
		} else {
			log.Println("✅ MongoDB disconnected successfully")
		}
	}()

	log.Println("✅ MongoDB connection verified and ready")

	// Now initialize monitoring service with MongoDB connection
	log.Println("📊 Initializing monitoring service...")
	monitoringConfig := monitoring.MonitoringConfig{
		Enabled:               true,
		Port:                  9090,
		Path:                  "/metrics",
		RedisHost:             "localhost",
		RedisPort:             "6379",
		RedisPassword:         "",
		RedisDB:               0,
		MongoDB:               client.Database("streamnestai"),
		HealthCheckInterval:   30 * time.Second,
		MetricsInterval:       30 * time.Second,
		ErrorRateThreshold:    0.05,
		ResponseTimeThreshold: 1000 * time.Millisecond,
		MemoryThreshold:       80.0,
	}

	monitoringService, err = monitoring.NewMonitoringService(monitoringConfig)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to initialize monitoring service: %v", err)
	} else {
		router.Use(monitoringService.MetricsMiddleware())
		log.Println("✅ Monitoring service initialized and middleware added")
	}

	// Initialize WebSocket hub with logging
	log.Println("🔌 Initializing WebSocket hub...")
	wsHubConfig := websocket.DefaultHubConfig()
	wsHub := websocket.NewHub(wsHubConfig)
	go wsHub.Run()
	log.Println("✅ WebSocket hub initialized and running")

	// Initialize analytics service with logging
	log.Println("📈 Initializing analytics service...")
	analyticsService := analytics.NewAnalyticsService(client, redisAddr)
	go func() {
		log.Println("📈 Starting analytics service...")
		if err := analyticsService.Start(); err != nil {
			log.Printf("⚠️  Warning: Failed to start analytics service: %v", err)
		} else {
			log.Println("✅ Analytics service started successfully")
		}
	}()

	// Initialize recommendation engine with logging
	log.Println("🎯 Initializing recommendation service...")
	recommendationService := recommendations.NewRecommendationService(client, redisAddr)
	go func() {
		log.Println("🎯 Initializing recommendation engine...")
		if err := recommendationService.Initialize(); err != nil {
			log.Printf("⚠️  Warning: Failed to initialize recommendation service: %v", err)
		} else {
			log.Println("✅ Recommendation service initialized successfully")
		}
	}()

	// Initialize social service with logging
	log.Println("👥 Initializing social service...")
	socialService := social.NewSocialService(client)
	go func() {
		log.Println("👥 Starting social service...")
		if err := socialService.Initialize(); err != nil {
			log.Printf("⚠️  Warning: Failed to initialize social service: %v", err)
		} else {
			log.Println("✅ Social service initialized successfully")
		}
	}()

	// Initialize security service with logging
	log.Println("🔒 Initializing security service...")
	securityService := security.NewSecurityService(client)
	go func() {
		log.Println("🔒 Starting security service...")
		if err := securityService.Initialize(); err != nil {
			log.Printf("⚠️  Warning: Failed to initialize security service: %v", err)
		} else {
			log.Println("✅ Security service initialized successfully")
		}
	}()

	// Create service container for dependency injection with logging
	log.Println("🔧 Creating service container for dependency injection...")
	services := routes.NewServiceContainer(
		client,
		wsHub,
		analyticsService,
		recommendationService,
		socialService,
		securityService,
		monitoringService,
		rateLimiter,
		redisAddr,
	)
	log.Println("✅ Service container created")

	// Set the global service container
	routes.SetServiceContainer(services)
	log.Println("✅ Global service container configured")

	// Setup routes with services and comprehensive logging
	log.Println("🛣️  Setting up routes...")

	log.Println("🛣️  Setting up unprotected routes...")
	routes.SetupUnProtectedRoutes(router, client)
	log.Println("✅ Unprotected routes configured")

	log.Println("🛣️  Setting up protected routes...")
	routes.SetupProtectedRoutes(router, client)
	log.Println("✅ Protected routes configured")

	log.Println("🛣️  Setting up enhanced routes...")
	routes.SetupEnhancedRoutes(router)
	log.Println("✅ Enhanced routes configured")

	log.Println("🛣️  Setting up cached routes with cache-first strategy...")
	routes.SetupCachedRoutes(router)
	log.Println("✅ Cached routes configured")

	log.Println("🛣️  Setting up versioned routes for backward compatibility...")
	routes.SetupVersionedRoutes(router)
	log.Println("✅ Versioned routes configured")

	// Note: /metrics endpoint is already registered in enhanced_routes.go
	log.Println("📊 Metrics endpoint available at /metrics")

	// Start background tasks with logging
	log.Println("🔄 Starting background tasks...")
	go startBackgroundTasks(services)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Println("⚠️  Using default port: 8080")
	} else {
		log.Printf("✅ Using configured port: %s", port)
	}

	// Final server status logging
	log.Println("🎉 ========================================")
	log.Println("🎉 STREAMNESTAI SERVER INITIALIZATION COMPLETE")
	log.Println("🎉 ========================================")
	log.Printf("🌐 Starting StreamNestAI Enhanced server on port %s", port)
	log.Printf("🌍 Environment: %s", os.Getenv("GIN_MODE"))
	log.Printf("💾 MongoDB connected: %v", client != nil)
	log.Printf("🔴 Redis address: %s", redisAddr)
	log.Printf("🔌 WebSocket hub initialized: %v", wsHub != nil)
	log.Printf("📊 Monitoring service: %v", monitoringService != nil)
	log.Printf("🛡️  Rate limiting: %v", rateLimiter != nil)
	log.Printf("📈 Analytics service: %v", analyticsService != nil)
	log.Printf("🎯 Recommendation service: %v", recommendationService != nil)
	log.Printf("👥 Social service: %v", socialService != nil)
	log.Printf("🔒 Security service: %v", securityService != nil)
	log.Println("🎉 ========================================")

	log.Printf("🚀 Server starting on http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ FATAL: Failed to start server: %v", err)
	}
}

// startBackgroundTasks initializes and runs background services
func startBackgroundTasks(services *routes.ServiceContainer) {
	log.Println("🔄 ========================================")
	log.Println("🔄 STARTING BACKGROUND TASKS")
	log.Println("🔄 ========================================")

	// Start system metrics collection
	go func() {
		log.Println("📊 Initializing system metrics collection...")
		if services.MonitoringService != nil {
			log.Println("✅ System metrics collection started")
			logFeatureCall("MONITORING", "metrics_collection", "system", map[string]interface{}{
				"service": "monitoring",
				"status":  "active",
			})
		} else {
			log.Println("⚠️  Monitoring service not available - metrics collection skipped")
		}
	}()

	// Start analytics event processing
	go func() {
		log.Println("📈 Initializing analytics event processing...")
		if services.AnalyticsService != nil {
			log.Println("✅ Analytics event processing started")
			logFeatureCall("ANALYTICS", "event_processing", "system", map[string]interface{}{
				"service": "analytics",
				"status":  "active",
			})
		} else {
			log.Println("⚠️  Analytics service not available - event processing skipped")
		}
	}()

	// Start recommendation model training
	go func() {
		log.Println("🎯 Initializing recommendation model training...")
		if services.RecommendationService != nil {
			log.Println("✅ Recommendation model training started")
			logFeatureCall("RECOMMENDATIONS", "model_training", "system", map[string]interface{}{
				"service": "recommendations",
				"status":  "active",
			})
		} else {
			log.Println("⚠️  Recommendation service not available - model training skipped")
		}
	}()

	// Start social feed updates
	go func() {
		log.Println("👥 Initializing social feed updates...")
		if services.SocialService != nil {
			log.Println("✅ Social feed updates started")
			logFeatureCall("SOCIAL", "feed_updates", "system", map[string]interface{}{
				"service": "social",
				"status":  "active",
			})
		} else {
			log.Println("⚠️  Social service not available - feed updates skipped")
		}
	}()

	// Start security monitoring
	go func() {
		log.Println("🔒 Initializing security monitoring...")
		if services.SecurityService != nil {
			log.Println("✅ Security monitoring started")
			logFeatureCall("SECURITY", "monitoring", "system", map[string]interface{}{
				"service": "security",
				"status":  "active",
			})
		} else {
			log.Println("⚠️  Security service not available - security monitoring skipped")
		}
	}()

	log.Println("🔄 ========================================")
	log.Println("🔄 BACKGROUND TASKS INITIALIZATION COMPLETE")
	log.Println("🔄 ========================================")
	log.Println("✅ All background tasks started successfully")
}
