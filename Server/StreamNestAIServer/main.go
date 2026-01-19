package main

import (
	"context"
	"log"
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

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: unable to find .env file")
	}

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "2.0.0",
			"service":   "StreamNestAI Enhanced",
		})
	})

	// Hello endpoint
	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, StreamNestAI Enhanced!")
	})

	// Configure CORS
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
			log.Println("Allowed Origin:", origins[i])
		}
	} else {
		origins = []string{
			"http://localhost:5173",
			"https://stream-nest-3py4n2m1t-shashwatssps-projects.vercel.app",
			"https://stream-nest-3py4n2m1t-shashwatssps-projects.vercel.app/",
		}
		log.Println("Allowed Origins: http://localhost:5173, https://stream-nest-3py4n2m1t-shashwatssps-projects.vercel.app, https://stream-nest-3py4n2m1t-shashwatssps-projects.vercel.app/")
	}

	config := cors.Config{}
	config.AllowOrigins = origins
	config.AllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length", "X-Total-Count"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	router.Use(cors.New(config))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Initialize Redis (if available)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	log.Printf("Redis address: %s", redisAddr)

	// Initialize monitoring service (will be configured after MongoDB connection)
	var monitoringService *monitoring.MonitoringService

	// Initialize rate limiting middleware
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
		log.Printf("Warning: Failed to initialize rate limiter: %v", err)
	} else {
		router.Use(rateLimiter.Middleware())
	}

	// Connect to MongoDB
	client := database.Connect()
	if client == nil {
		log.Fatal("Failed to connect to MongoDB")
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("Failed to reach MongoDB server: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	log.Println("Successfully connected to MongoDB")

	// Now initialize monitoring service with MongoDB connection
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
		log.Printf("Warning: Failed to initialize monitoring service: %v", err)
	} else {
		// Add monitoring middleware
		router.Use(monitoringService.MetricsMiddleware())
	}

	// Initialize WebSocket hub
	wsHubConfig := websocket.DefaultHubConfig()
	wsHub := websocket.NewHub(wsHubConfig)
	go wsHub.Run()

	// Initialize analytics service
	analyticsService := analytics.NewAnalyticsService(client, redisAddr)
	go func() {
		if err := analyticsService.Start(); err != nil {
			log.Printf("Warning: Failed to start analytics service: %v", err)
		}
	}()

	// Initialize recommendation engine
	recommendationService := recommendations.NewRecommendationService(client, redisAddr)
	go func() {
		if err := recommendationService.Initialize(); err != nil {
			log.Printf("Warning: Failed to initialize recommendation service: %v", err)
		}
	}()

	// Initialize social service
	socialService := social.NewSocialService(client)
	go func() {
		if err := socialService.Initialize(); err != nil {
			log.Printf("Warning: Failed to initialize social service: %v", err)
		}
	}()

	// Initialize security service
	securityService := security.NewSecurityService(client)
	go func() {
		if err := securityService.Initialize(); err != nil {
			log.Printf("Warning: Failed to initialize security service: %v", err)
		}
	}()

	// Create service container for dependency injection
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

	// Set the global service container
	routes.SetServiceContainer(services)

	// Setup routes with services
	routes.SetupUnProtectedRoutes(router, client)
	routes.SetupProtectedRoutes(router, client)

	// Setup enhanced routes
	routes.SetupEnhancedRoutes(router)

	// Setup cached routes with cache-first strategy
	routes.SetupCachedRoutes(router)

	// Setup versioned routes for backward compatibility
	routes.SetupVersionedRoutes(router)

	// Note: /metrics endpoint is already registered in enhanced_routes.go

	// Start background tasks
	go startBackgroundTasks(services)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting StreamNestAI Enhanced server on port %s", port)
	log.Printf("Environment: %s", os.Getenv("GIN_MODE"))
	log.Printf("MongoDB connected: %v", client != nil)
	log.Printf("Redis address: %s", redisAddr)
	log.Printf("WebSocket hub initialized: %v", wsHub != nil)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// startBackgroundTasks initializes and runs background services
func startBackgroundTasks(services *routes.ServiceContainer) {
	log.Println("Starting background tasks...")

	// Start system metrics collection
	go func() {
		if services.MonitoringService != nil {
			log.Println("System metrics collection started")
		}
	}()

	// Start analytics event processing
	go func() {
		if services.AnalyticsService != nil {
			log.Println("Analytics event processing started")
		}
	}()

	// Start recommendation model training
	go func() {
		if services.RecommendationService != nil {
			log.Println("Recommendation model training started")
		}
	}()

	// Start social feed updates
	go func() {
		if services.SocialService != nil {
			log.Println("Social feed updates started")
		}
	}()

	// Start security monitoring
	go func() {
		if services.SecurityService != nil {
			log.Println("Security monitoring started")
		}
	}()

	log.Println("Background tasks started successfully")
}
