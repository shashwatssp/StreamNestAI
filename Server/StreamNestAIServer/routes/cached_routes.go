package routes

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/cache"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/controllers"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SetupCachedRoutes adds cache-first endpoints to the existing routes
func SetupCachedRoutes(router *gin.Engine) {
	sc := GetServiceContainer()
	if sc == nil {
		return
	}

	// Initialize Redis service for cached controllers using environment variables
	log.Printf("🔧 INITIALIZING CACHED ROUTES - ENVIRONMENT CHECK:")

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisAddr := os.Getenv("REDIS_ADDR")

	log.Printf("   REDIS_HOST: %s", redisHost)
	log.Printf("   REDIS_PORT: %s", redisPort)
	log.Printf("   REDIS_ADDR: %s", redisAddr)
	log.Printf("   REDIS_PASSWORD: %s", func() string {
		if redisPassword == "" {
			return "[EMPTY]"
		}
		return "[SET]"
	}())

	// Validate required environment variables
	if redisHost == "" {
		log.Printf("❌ DEPLOYMENT ERROR: REDIS_HOST environment variable is required")
		log.Printf("   Current environment variables:")
		log.Printf("   - REDIS_HOST: '%s'", redisHost)
		log.Printf("   - REDIS_PORT: '%s'", redisPort)
		log.Printf("   - REDIS_ADDR: '%s'", redisAddr)
		log.Printf("   Fix: Set REDIS_HOST in your deployment environment")
		log.Fatal("REDIS_HOST environment variable is required")
	}
	if redisPort == "" {
		log.Printf("❌ DEPLOYMENT ERROR: REDIS_PORT environment variable is required")
		log.Printf("   Current environment variables:")
		log.Printf("   - REDIS_HOST: '%s'", redisHost)
		log.Printf("   - REDIS_PORT: '%s'", redisPort)
		log.Printf("   - REDIS_ADDR: '%s'", redisAddr)
		log.Printf("   Fix: Set REDIS_PORT in your deployment environment")
		log.Fatal("REDIS_PORT environment variable is required")
	}

	log.Printf("✅ Required Redis environment variables found")

	redisConfig := cache.CacheConfig{
		Host:         redisHost,
		Port:         redisPort,
		Password:     redisPassword, // Can be empty if no password required
		DB:           0,
		KeyPrefix:    "streamnestai",
		DefaultTTL:   30 * time.Minute,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	log.Printf("🚀 Creating Redis service with configuration...")
	redisService, err := cache.NewRedisService(redisConfig)
	if err != nil {
		log.Printf("❌ DEPLOYMENT CRITICAL: Failed to initialize Redis service for cached routes")
		log.Printf("   Error: %v", err)
		log.Printf("   Redis Host: %s", redisHost)
		log.Printf("   Redis Port: %s", redisPort)
		log.Printf("   Redis Address: %s", redisAddr)
		log.Printf("   Troubleshooting steps:")
		log.Printf("   1. Verify Redis server is running at %s:%s", redisHost, redisPort)
		log.Printf("   2. Check if Redis password is correct")
		log.Printf("   3. Ensure network connectivity to Redis server")
		log.Printf("   4. Verify Redis allows remote connections")
		log.Printf("   5. Check firewall settings")
		log.Printf("   Cached routes will not be available. Check Redis configuration.")
		// Log error but don't fail - cached routes won't be available
		return
	}

	log.Printf("✅ Redis service initialized successfully for cached routes")

	// Initialize cached controllers
	cachedMovieController := controllers.NewCachedMovieController(sc.MongoClient, redisService)
	cachedRecommendationController := controllers.NewCachedRecommendationController(sc.MongoClient, redisService)

	// Cached public routes (no authentication required)
	setupCachedPublicRoutes(router, cachedMovieController)

	// Cached protected routes (authentication required)
	setupCachedProtectedRoutes(router, cachedMovieController, cachedRecommendationController)
}

func setupCachedPublicRoutes(router *gin.Engine, cmc *controllers.CachedMovieController) {
	// Cached movies endpoints
	cached := router.Group("/cached")
	{
		// Get all movies with cache-first strategy
		cached.GET("/movies", cmc.GetMoviesWithCache())

		// Get movie by ID with cache-first strategy
		cached.GET("/movies/:imdb_id", cmc.GetMovieWithCache())

		// Cache management endpoints
		cached.DELETE("/cache/movies", cmc.ClearCache())
		cached.GET("/cache/stats", cmc.GetCacheStats())
	}
}

func setupCachedProtectedRoutes(router *gin.Engine, cmc *controllers.CachedMovieController, crc *controllers.CachedRecommendationController) {
	// Apply authentication middleware for protected cached routes
	protected := router.Group("/cached")
	protected.Use(middleware.AuthMiddleWare())
	{
		// Cached recommendation endpoints
		protected.GET("/recommendedmovies", crc.GetRecommendedMoviesWithCache)

		// Get trending content with cache
		protected.GET("/recommendations/trending", crc.GetTrendingContentWithCache)

		// Get personalized feed with cache
		protected.GET("/recommendations/feed", crc.GetPersonalizedFeedWithCache)

		// Cache management for recommendations
		protected.DELETE("/cache/recommendations", crc.InvalidateUserRecommendationCache)
		protected.GET("/cache/recommendations/stats", crc.GetRecommendationCacheStats)

		// Get recommended movies with cache (using movie controller)
		protected.GET("/user/recommended", cmc.GetRecommendedMoviesWithCache())
	}
}

// Additional helper functions for cached routes

// GetCacheKey generates a consistent cache key for different data types
func GetCacheKey(dataType string, params ...string) string {
	key := "streamnestai:" + dataType
	for _, param := range params {
		key += ":" + param
	}
	return key
}

// SetCacheHeaders sets appropriate cache-related headers
func SetCacheHeaders(c *gin.Context, dataSource string, cacheStatus string, responseTime time.Duration) {
	c.Header("X-Data-Source", dataSource)
	c.Header("X-Cache-Status", cacheStatus)
	c.Header("X-Response-Time", responseTime.String())
	c.Header("X-Cache-Control", "max-age=300") // 5 minutes
}

// HandleCacheError handles cache-related errors gracefully
func HandleCacheError(c *gin.Context, err error, operation string) {
	// Log the error but don't expose internal details to client
	// In production, you might want to use a proper logging framework
	c.Header("X-Cache-Error", operation+" failed")

	// Continue with the request even if caching fails
	// The controllers will fall back to MongoDB
}

// ValidateCacheRequest validates common cache request parameters
func ValidateCacheRequest(c *gin.Context) (limit int, offset int, err error) {
	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "20")
	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Prevent excessive requests
	}

	// Parse offset parameter
	offsetStr := c.DefaultQuery("offset", "0")
	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset, nil
}

// GetCacheTTL returns appropriate TTL based on data type
func GetCacheTTL(dataType string) time.Duration {
	switch dataType {
	case "movies":
		return 30 * time.Minute // Movies change less frequently - updated to 30 mins
	case "recommendations":
		return 30 * time.Minute // Recommendations - updated to 30 mins
	case "user_data":
		return 30 * time.Minute // User data - updated to 30 mins
	case "popular":
		return 30 * time.Minute // Popular movies list is stable
	case "trending":
		return 30 * time.Minute // Trending - updated to 30 mins
	default:
		return 30 * time.Minute // Default TTL - updated to 30 mins
	}
}

// InvalidateRelatedCache invalidates cache entries related to a specific movie
func InvalidateRelatedCache(redisService *cache.RedisService, movieID string) {
	if redisService == nil {
		return
	}

	// Keys to invalidate when a movie is updated
	keysToInvalidate := []string{
		GetCacheKey("movies"),
		GetCacheKey("movie", movieID),
		GetCacheKey("popular"),
		GetCacheKey("latest"),
		GetCacheKey("recommendations"),
	}

	// Delete each key from cache
	for _, key := range keysToInvalidate {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		redisService.Delete(ctx, key)
	}
}

// WarmupCacheData pre-populates cache with frequently accessed data
func WarmupCacheData(redisService *cache.RedisService, mongoClient *mongo.Client) {
	if redisService == nil || mongoClient == nil {
		return
	}

	// This would typically be called by a background job or admin endpoint
	// For now, it's a placeholder for the warmup functionality

	// TODO: Implement actual warmup logic
	// 1. Get popular movies from MongoDB
	// 2. Cache them with appropriate TTL
	// 3. Get trending recommendations
	// 4. Cache them with appropriate TTL
}
