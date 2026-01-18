package routes

import (
	"context"
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

	// Initialize Redis service for cached controllers using the same Redis Cloud config
	redisConfig := cache.CacheConfig{
		Host:         "redis-13554.crce182.ap-south-1-1.ec2.cloud.redislabs.com",
		Port:         "13554",
		Password:     "", // Set if Redis Cloud requires password
		DB:           0,
		KeyPrefix:    "streamnestai",
		DefaultTTL:   5 * time.Minute,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	redisService, err := cache.NewRedisService(redisConfig)
	if err != nil {
		// Log error but don't fail - cached routes won't be available
		return
	}

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
		return 15 * time.Minute // Movies change less frequently
	case "recommendations":
		return 10 * time.Minute // Recommendations are more dynamic
	case "user_data":
		return 5 * time.Minute // User data changes frequently
	case "popular":
		return 30 * time.Minute // Popular movies list is stable
	case "trending":
		return 5 * time.Minute // Trending changes frequently
	default:
		return 10 * time.Minute // Default TTL
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
