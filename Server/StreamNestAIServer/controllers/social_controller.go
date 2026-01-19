package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/social"
)

type SocialController struct {
	socialService *social.SocialService
}

func NewSocialController(socialService *social.SocialService) *SocialController {
	return &SocialController{
		socialService: socialService,
	}
}

// GetUserProfile handles GET /api/v1/social/profile
func (sc *SocialController) GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	profile, err := sc.socialService.GetUserProfile(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": profile,
	})
}

// UpdateUserProfile handles PUT /api/v1/social/profile
func (sc *SocialController) UpdateUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Remove sensitive fields that shouldn't be updated directly
	delete(updates, "user_id")
	delete(updates, "join_date")
	delete(updates, "followers_count")
	delete(updates, "following_count")
	delete(updates, "posts_count")
	delete(updates, "watchlist_count")

	err := sc.socialService.UpdateUserProfile(userID.(string), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

// GetUserProfileByUserID handles GET /api/v1/social/profile/:userId
func (sc *SocialController) GetUserProfileByUserID(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	profile, err := sc.socialService.GetUserProfile(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile: " + err.Error()})
		return
	}

	// Remove private information if profile is private and user is not the owner
	currentUserID, exists := c.Get("user_id")
	if !exists || currentUserID.(string) != targetUserID {
		if profile.IsPrivate {
			// Return only public information
			publicProfile := map[string]interface{}{
				"user_id":         profile.UserID,
				"username":        profile.Username,
				"display_name":    profile.DisplayName,
				"avatar":          profile.Avatar,
				"bio":             profile.Bio,
				"followers_count": profile.FollowersCount,
				"following_count": profile.FollowingCount,
				"posts_count":     profile.PostsCount,
				"is_verified":     profile.IsVerified,
				"join_date":       profile.JoinDate,
			}
			c.JSON(http.StatusOK, gin.H{"profile": publicProfile})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": profile,
	})
}

// FollowUser handles POST /api/v1/social/follow/:userId
func (sc *SocialController) FollowUser(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	err := sc.socialService.FollowUser(currentUserID.(string), targetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "User followed successfully",
		"follower_id":  currentUserID,
		"following_id": targetUserID,
	})
}

// UnfollowUser handles DELETE /api/v1/social/follow/:userId
func (sc *SocialController) UnfollowUser(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	err := sc.socialService.UnfollowUser(currentUserID.(string), targetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "User unfollowed successfully",
		"follower_id":  currentUserID,
		"following_id": targetUserID,
	})
}

// GetFollowers handles GET /api/v1/social/followers/:userId
func (sc *SocialController) GetFollowers(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	followers, err := sc.socialService.GetFollowers(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get followers: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":   userID,
		"followers": followers,
		"count":     len(followers),
		"limit":     limit,
		"offset":    offset,
	})
}

// GetFollowing handles GET /api/v1/social/following/:userId
func (sc *SocialController) GetFollowing(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	following, err := sc.socialService.GetFollowing(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get following: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":   userID,
		"following": following,
		"count":     len(following),
		"limit":     limit,
		"offset":    offset,
	})
}

// GetActivityFeed handles GET /api/v1/social/feed
func (sc *SocialController) GetActivityFeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	activities, err := sc.socialService.GetActivityFeed(userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get activity feed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"count":      len(activities),
		"limit":      limit,
		"offset":     offset,
	})
}

// GetUserActivities handles GET /api/v1/social/activities/:userId
func (sc *SocialController) GetUserActivities(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	activities, err := sc.socialService.GetUserActivities(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user activities: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    userID,
		"activities": activities,
		"count":      len(activities),
		"limit":      limit,
		"offset":     offset,
	})
}

// CreateActivity handles POST /api/v1/social/activity
func (sc *SocialController) CreateActivity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var activity struct {
		Type        string                 `json:"type" binding:"required"`
		ContentID   string                 `json:"content_id" binding:"required"`
		ContentType string                 `json:"content_type" binding:"required"`
		Title       string                 `json:"title" binding:"required"`
		Description string                 `json:"description"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	err := sc.socialService.CreateActivity(
		userID.(string),
		activity.Type,
		activity.ContentID,
		activity.ContentType,
		activity.Title,
		activity.Description,
		activity.Metadata,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create activity: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Activity created successfully",
	})
}

// CreateWatchlist handles POST /api/v1/social/watchlist
func (sc *SocialController) CreateWatchlist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var watchlist struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&watchlist); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	newWatchlist, err := sc.socialService.CreateWatchlist(
		userID.(string),
		watchlist.Name,
		watchlist.Description,
		watchlist.IsPublic,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create watchlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Watchlist created successfully",
		"watchlist": newWatchlist,
	})
}

// GetUserWatchlists handles GET /api/v1/social/watchlists
func (sc *SocialController) GetUserWatchlists(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	watchlists, err := sc.socialService.GetUserWatchlists(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get watchlists: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"watchlists": watchlists,
		"count":      len(watchlists),
	})
}

// AddToWatchlist handles POST /api/v1/social/watchlist/:watchlistId/add
func (sc *SocialController) AddToWatchlist(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	watchlistID := c.Param("watchlistId")
	if watchlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Watchlist ID is required"})
		return
	}

	var request struct {
		ContentID string `json:"content_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Check if user owns the watchlist
	isOwner, err := sc.socialService.IsWatchlistOwner(watchlistID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify watchlist ownership: " + err.Error()})
		return
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this watchlist"})
		return
	}

	err = sc.socialService.AddToWatchlist(watchlistID, request.ContentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to watchlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Content added to watchlist successfully",
		"watchlist_id": watchlistID,
		"content_id":   request.ContentID,
	})
}

// RemoveFromWatchlist handles DELETE /api/v1/social/watchlist/:watchlistId/remove/:contentId
func (sc *SocialController) RemoveFromWatchlist(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	watchlistID := c.Param("watchlistId")
	contentID := c.Param("contentId")

	if watchlistID == "" || contentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Watchlist ID and Content ID are required"})
		return
	}

	// Check if user owns the watchlist
	isOwner, err := sc.socialService.IsWatchlistOwner(watchlistID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify watchlist ownership: " + err.Error()})
		return
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this watchlist"})
		return
	}

	err = sc.socialService.RemoveFromWatchlist(watchlistID, contentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from watchlist: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Content removed from watchlist successfully",
		"watchlist_id": watchlistID,
		"content_id":   contentID,
	})
}

// GetSocialStats handles GET /api/v1/social/stats
func (sc *SocialController) GetSocialStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	profile, err := sc.socialService.GetUserProfile(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile: " + err.Error()})
		return
	}

	stats := map[string]interface{}{
		"followers_count":  profile.FollowersCount,
		"following_count":  profile.FollowingCount,
		"posts_count":      profile.PostsCount,
		"watchlist_count":  profile.WatchlistCount,
		"total_watch_time": profile.Stats.TotalWatchTime,
		"movies_watched":   profile.Stats.MoviesWatched,
		"series_watched":   profile.Stats.SeriesWatched,
		"average_rating":   profile.Stats.AverageRating,
		"streak_days":      profile.Stats.StreakDays,
		"badge_count":      profile.Stats.BadgeCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"stats":   stats,
	})
}

// SearchUsers handles GET /api/v1/social/search/users
func (sc *SocialController) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 50 {
		limit = 50
	}

	// This would implement user search functionality
	// For now, return empty results
	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"users":   []interface{}{},
		"count":   0,
		"limit":   limit,
		"offset":  offset,
		"message": "User search feature coming soon",
	})
}
