package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/recommendation"
)

type RecommendationController struct {
	engine *recommendation.RecommendationEngine
}

func NewRecommendationController(engine *recommendation.RecommendationEngine) *RecommendationController {
	return &RecommendationController{
		engine: engine,
	}
}

// GetRecommendations handles GET /api/v1/recommendations
func (rc *RecommendationController) GetRecommendations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse query parameters
	recType := c.DefaultQuery("type", "hybrid")
	count, _ := strconv.Atoi(c.DefaultQuery("count", "10"))
	excludeWatched := c.DefaultQuery("exclude_watched", "true") == "true"
	minRating, _ := strconv.ParseFloat(c.DefaultQuery("min_rating", "0"), 64)
	maxDuration, _ := strconv.Atoi(c.DefaultQuery("max_duration", "0"))
	language := c.Query("language")

	// Parse genre filters
	var includeGenres, excludeGenres []string
	if genres := c.Query("include_genres"); genres != "" {
		includeGenres = []string{genres}
	}
	if genres := c.Query("exclude_genres"); genres != "" {
		excludeGenres = []string{genres}
	}

	// Create recommendation request
	req := recommendation.RecommendationRequest{
		UserID:         userID.(string),
		Type:           recommendation.RecommendationType(recType),
		Count:          count,
		ExcludeWatched: excludeWatched,
		IncludeGenres:  includeGenres,
		ExcludeGenres:  excludeGenres,
		MinRating:      minRating,
		MaxDuration:    maxDuration,
		Language:       language,
		Context:        make(map[string]interface{}),
	}

	// Add context information
	if deviceType := c.GetHeader("User-Agent"); deviceType != "" {
		req.Context["device_type"] = deviceType
	}
	if clientIP := c.ClientIP(); clientIP != "" {
		req.Context["client_ip"] = clientIP
	}

	// Get recommendations
	recommendations, err := rc.engine.GetRecommendations(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations: " + err.Error()})
		return
	}

	// Return response
	response := gin.H{
		"recommendations": recommendations,
		"metadata": gin.H{
			"user_id":   userID,
			"type":      recType,
			"count":     len(recommendations),
			"requested": count,
			"timestamp": time.Now(),
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetPersonalizedRecommendations handles GET /api/v1/recommendations/personalized
func (rc *RecommendationController) GetPersonalizedRecommendations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	count, _ := strconv.Atoi(c.DefaultQuery("count", "10"))

	req := recommendation.RecommendationRequest{
		UserID: userID.(string),
		Type:   recommendation.TypePersonalized,
		Count:  count,
		Context: map[string]interface{}{
			"endpoint": "personalized",
		},
	}

	recommendations, err := rc.engine.GetRecommendations(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get personalized recommendations: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"type":            "personalized",
		"count":           len(recommendations),
	})
}

// GetTrendingContent handles GET /api/v1/recommendations/trending
func (rc *RecommendationController) GetTrendingContent(c *gin.Context) {
	count, _ := strconv.Atoi(c.DefaultQuery("count", "20"))
	timeWindow := c.DefaultQuery("time_window", "24h") // 24h, 7d, 30d

	req := recommendation.RecommendationRequest{
		UserID: "anonymous", // Trending doesn't require user
		Type:   recommendation.TypeTrending,
		Count:  count,
		Context: map[string]interface{}{
			"time_window": timeWindow,
			"endpoint":    "trending",
		},
	}

	recommendations, err := rc.engine.GetRecommendations(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending content: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"type":            "trending",
		"time_window":     timeWindow,
		"count":           len(recommendations),
	})
}

// GetSimilarContent handles GET /api/v1/recommendations/similar/:contentId
func (rc *RecommendationController) GetSimilarContent(c *gin.Context) {
	contentID := c.Param("contentId")
	if contentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content ID is required"})
		return
	}

	count, _ := strconv.Atoi(c.DefaultQuery("count", "10"))

	// For similar content, we use content-based recommendations
	req := recommendation.RecommendationRequest{
		UserID: "anonymous", // Similar content doesn't require user
		Type:   recommendation.TypeContentBased,
		Count:  count,
		Context: map[string]interface{}{
			"content_id": contentID,
			"endpoint":   "similar",
		},
	}

	recommendations, err := rc.engine.GetRecommendations(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get similar content: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content_id":      contentID,
		"recommendations": recommendations,
		"type":            "similar",
		"count":           len(recommendations),
	})
}

// GetRecommendationsForGenre handles GET /api/v1/recommendations/genre/:genre
func (rc *RecommendationController) GetRecommendationsForGenre(c *gin.Context) {
	genre := c.Param("genre")
	if genre == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Genre is required"})
		return
	}

	count, _ := strconv.Atoi(c.DefaultQuery("count", "10"))

	req := recommendation.RecommendationRequest{
		UserID:        "anonymous",
		Type:          recommendation.TypeContentBased,
		Count:         count,
		IncludeGenres: []string{genre},
		Context: map[string]interface{}{
			"genre":    genre,
			"endpoint": "genre",
		},
	}

	recommendations, err := rc.engine.GetRecommendations(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get genre recommendations: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"genre":           genre,
		"recommendations": recommendations,
		"type":            "genre",
		"count":           len(recommendations),
	})
}

// UpdateUserPreferences handles POST /api/v1/recommendations/preferences
func (rc *RecommendationController) UpdateUserPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var preferences struct {
		GenrePreferences map[string]float64 `json:"genre_preferences"`
		TagPreferences   map[string]float64 `json:"tag_preferences"`
		LanguagePref     string             `json:"language_pref"`
	}

	if err := c.ShouldBindJSON(&preferences); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// This would update the user profile in the recommendation engine
	// For now, we'll return a success response
	c.JSON(http.StatusOK, gin.H{
		"message": "User preferences updated successfully",
		"user_id": userID,
	})
}

// GetUserRecommendationHistory handles GET /api/v1/recommendations/history
func (rc *RecommendationController) GetUserRecommendationHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// This would fetch the user's recommendation history from the database
	// For now, we'll return a mock response
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"history": []interface{}{}, // Empty for now
		"count":   0,
		"message": "Recommendation history feature coming soon",
	})
}

// GetRecommendationStats handles GET /api/v1/recommendations/stats
func (rc *RecommendationController) GetRecommendationStats(c *gin.Context) {
	// This would return statistics about the recommendation system
	// For now, we'll return mock stats
	c.JSON(http.StatusOK, gin.H{
		"total_recommendations": 0,
		"active_users":          0,
		"content_items":         0,
		"algorithms": []string{
			"collaborative_filtering",
			"content_based",
			"hybrid",
			"trending",
		},
		"last_updated": time.Now(),
	})
}

// RefreshRecommendations handles POST /api/v1/recommendations/refresh
func (rc *RecommendationController) RefreshRecommendations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// This would trigger a refresh of the user's recommendations
	// For now, we'll return a success response
	c.JSON(http.StatusOK, gin.H{
		"message":   "Recommendations refreshed successfully",
		"user_id":   userID,
		"timestamp": time.Now(),
	})
}

// Feedback handles POST /api/v1/recommendations/feedback
func (rc *RecommendationController) Feedback(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var feedback struct {
		ContentID        string  `json:"content_id" binding:"required"`
		RecommendationID string  `json:"recommendation_id"`
		Rating           float64 `json:"rating" binding:"required,min=1,max=5"`
		Feedback         string  `json:"feedback"`
		Useful           bool    `json:"useful"`
	}

	if err := c.ShouldBindJSON(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// This would store the feedback and use it to improve recommendations
	// For now, we'll return a success response
	c.JSON(http.StatusOK, gin.H{
		"message":           "Feedback recorded successfully",
		"user_id":           userID,
		"content_id":        feedback.ContentID,
		"recommendation_id": feedback.RecommendationID,
		"rating":            feedback.Rating,
		"useful":            feedback.Useful,
		"timestamp":         time.Now(),
	})
}
