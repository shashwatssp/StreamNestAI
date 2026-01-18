package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/cache"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/middleware"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/recommendations"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Type alias for Recommendation
type Recommendation = recommendations.Recommendation

// CachedRecommendationController handles recommendation endpoints with Redis caching
type CachedRecommendationController struct {
	client                      *mongo.Client
	redisService                *cache.RedisService
	cachedRecommendationService *recommendations.CachedRecommendationService
}

// NewCachedRecommendationController creates a new cached recommendation controller
func NewCachedRecommendationController(client *mongo.Client, redisService *cache.RedisService) *CachedRecommendationController {
	return &CachedRecommendationController{
		client:                      client,
		redisService:                redisService,
		cachedRecommendationService: recommendations.NewCachedRecommendationService(client, redisService),
	}
}

// GetRecommendedMoviesWithCache returns personalized recommendations with cache-first strategy
func (crc *CachedRecommendationController) GetRecommendedMoviesWithCache(c *gin.Context) {
	log.Printf("INFO: GetRecommendedMoviesWithCache endpoint called")
	startTime := time.Now()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("ERROR: User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("ERROR: Invalid user ID type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("DEBUG: Processing recommendation request for user: %s", userIDStr)

	// Get limit from query params (default 10)
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
		log.Printf("WARNING: Invalid limit parameter, using default: %d", limit)
	}

	if limit > 50 {
		limit = 50
		log.Printf("WARNING: Limit exceeded maximum, setting to: %d", limit)
	}

	log.Printf("DEBUG: Recommendation parameters - user: %s, limit: %d", userIDStr, limit)

	// Get recommendations using cached service
	recommendations, err := crc.cachedRecommendationService.GetRecommendationsWithCache(userIDStr, limit)
	if err != nil {
		log.Printf("ERROR: Failed to get recommendations for user %s: %v", userIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get recommendations",
			"details": err.Error(),
		})
		return
	}

	elapsedTime := time.Since(startTime)
	log.Printf("SUCCESS: Retrieved %d recommendations for user %s in %v", len(recommendations), userIDStr, elapsedTime)

	// Return response with metadata
	c.JSON(http.StatusOK, gin.H{
		"data":          recommendations,
		"count":         len(recommendations),
		"user_id":       userIDStr,
		"generated_at":  time.Now().Format(time.RFC3339),
		"response_time": elapsedTime.String(),
	})
}

// GetTrendingContentWithCache returns trending content with cache-first strategy
func (crc *CachedRecommendationController) GetTrendingContentWithCache(c *gin.Context) {
	log.Printf("INFO: GetTrendingContentWithCache endpoint called")
	startTime := time.Now()

	// Get limit from query params (default 20)
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
		log.Printf("WARNING: Invalid limit parameter, using default: %d", limit)
	}

	if limit > 100 {
		limit = 100
		log.Printf("WARNING: Limit exceeded maximum, setting to: %d", limit)
	}

	// Get time window from query params (default 24h)
	timeWindow := c.DefaultQuery("time_window", "24h")
	if timeWindow != "24h" && timeWindow != "7d" && timeWindow != "30d" {
		timeWindow = "24h"
		log.Printf("WARNING: Invalid time window, using default: %s", timeWindow)
	}

	log.Printf("DEBUG: Trending content parameters - limit: %d, time_window: %s", limit, timeWindow)

	// Get trending content using cached service
	trending, err := crc.cachedRecommendationService.GetTrendingContentWithCache(limit, timeWindow)
	if err != nil {
		log.Printf("ERROR: Failed to get trending content: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get trending content",
			"details": err.Error(),
		})
		return
	}

	elapsedTime := time.Since(startTime)
	log.Printf("SUCCESS: Retrieved %d trending items in %v", len(trending), elapsedTime)

	// Return response with metadata
	c.JSON(http.StatusOK, gin.H{
		"data":          trending,
		"count":         len(trending),
		"time_window":   timeWindow,
		"generated_at":  time.Now().Format(time.RFC3339),
		"response_time": elapsedTime.String(),
	})
}

// GetPersonalizedFeedWithCache returns a personalized feed combining recommendations and trending
func (crc *CachedRecommendationController) GetPersonalizedFeedWithCache(c *gin.Context) {
	log.Printf("INFO: GetPersonalizedFeedWithCache endpoint called")
	startTime := time.Now()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("ERROR: User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("ERROR: Invalid user ID type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("DEBUG: Processing personalized feed request for user: %s", userIDStr)

	// Get recommendation count from query params (default 5)
	recLimitStr := c.DefaultQuery("recommendations", "5")
	recLimit, err := strconv.Atoi(recLimitStr)
	if err != nil || recLimit <= 0 {
		recLimit = 5
	}

	// Get trending count from query params (default 10)
	trendingLimitStr := c.DefaultQuery("trending", "10")
	trendingLimit, err := strconv.Atoi(trendingLimitStr)
	if err != nil || trendingLimit <= 0 {
		trendingLimit = 10
	}

	log.Printf("DEBUG: Feed parameters - user: %s, recommendations: %d, trending: %d", userIDStr, recLimit, trendingLimit)

	// Get personalized recommendations
	recommendations, err := crc.cachedRecommendationService.GetRecommendationsWithCache(userIDStr, recLimit)
	if err != nil {
		log.Printf("WARNING: Failed to get recommendations for feed: %v", err)
		recommendations = []Recommendation{}
	}

	// Get trending content
	trending, err := crc.cachedRecommendationService.GetTrendingContentWithCache(trendingLimit, "24h")
	if err != nil {
		log.Printf("WARNING: Failed to get trending content for feed: %v", err)
		trending = []Recommendation{}
	}

	elapsedTime := time.Since(startTime)
	log.Printf("SUCCESS: Generated personalized feed for user %s in %v (recs: %d, trending: %d)",
		userIDStr, elapsedTime, len(recommendations), len(trending))

	// Return combined feed
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"recommendations": recommendations,
			"trending":        trending,
		},
		"counts": gin.H{
			"recommendations": len(recommendations),
			"trending":        len(trending),
			"total":           len(recommendations) + len(trending),
		},
		"user_id":       userIDStr,
		"generated_at":  time.Now().Format(time.RFC3339),
		"response_time": elapsedTime.String(),
	})
}

// InvalidateUserRecommendationCache clears recommendation cache for a user
func (crc *CachedRecommendationController) InvalidateUserRecommendationCache(c *gin.Context) {
	log.Printf("INFO: InvalidateUserRecommendationCache endpoint called")

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("ERROR: User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("ERROR: Invalid user ID type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("DEBUG: Invalidating recommendation cache for user: %s", userIDStr)

	// Invalidate user's recommendation cache
	err := crc.cachedRecommendationService.InvalidateUserCache(userIDStr)
	if err != nil {
		log.Printf("ERROR: Failed to invalidate cache for user %s: %v", userIDStr, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to invalidate cache",
			"details": err.Error(),
		})
		return
	}

	log.Printf("SUCCESS: Invalidated recommendation cache for user %s", userIDStr)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Cache invalidated successfully",
		"user_id":        userIDStr,
		"invalidated_at": time.Now().Format(time.RFC3339),
	})
}

// GetRecommendationCacheStats returns recommendation cache statistics
func (crc *CachedRecommendationController) GetRecommendationCacheStats(c *gin.Context) {
	log.Printf("INFO: GetRecommendationCacheStats endpoint called")

	// Check if user is admin (optional - for now just require authentication)
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("ERROR: User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	log.Printf("DEBUG: Fetching recommendation cache statistics for user: %s", userID)

	// Get cache statistics
	stats, err := crc.cachedRecommendationService.GetCacheStats(c.Request.Context())
	if err != nil {
		log.Printf("ERROR: Failed to get cache stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get cache statistics",
			"details": err.Error(),
		})
		return
	}

	log.Printf("SUCCESS: Retrieved recommendation cache statistics")

	c.JSON(http.StatusOK, gin.H{
		"data":         stats,
		"service":      "recommendations",
		"generated_at": time.Now().Format(time.RFC3339),
	})
}

// SetupCachedRecommendationRoutes sets up the cached recommendation routes
func SetupCachedRecommendationRoutes(router *gin.Engine, client *mongo.Client, redisService *cache.RedisService) {
	log.Printf("INFO: Setting up cached recommendation routes")

	// Create cached recommendation controller
	cachedRecController := NewCachedRecommendationController(client, redisService)

	// Create auth middleware
	authMiddleware := middleware.AuthMiddleWare()

	// Cached recommendation routes group
	cachedRecRoutes := router.Group("/api/v1/recommendations-cached")
	{
		// Authenticated routes
		cachedRecRoutes.Use(authMiddleware)
		{
			cachedRecRoutes.GET("/recommended", cachedRecController.GetRecommendedMoviesWithCache)
			cachedRecRoutes.GET("/trending", cachedRecController.GetTrendingContentWithCache)
			cachedRecRoutes.GET("/feed", cachedRecController.GetPersonalizedFeedWithCache)
			cachedRecRoutes.DELETE("/cache", cachedRecController.InvalidateUserRecommendationCache)
			cachedRecRoutes.GET("/cache/stats", cachedRecController.GetRecommendationCacheStats)
		}
	}

	log.Printf("SUCCESS: Cached recommendation routes configured at /api/v1/recommendations-cached")
}
