package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/analytics"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/middleware"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/websocket"
)

// SetupEnhancedRoutes adds new enhanced endpoints to the existing routes
func SetupEnhancedRoutes(router *gin.Engine) {
	sc := GetServiceContainer()
	if sc == nil {
		return
	}

	// Enhanced public routes
	setupEnhancedPublicRoutes(router, sc)

	// Enhanced protected routes
	setupEnhancedProtectedRoutes(router, sc)
}

func setupEnhancedPublicRoutes(router *gin.Engine, sc *ServiceContainer) {
	// Health and monitoring endpoints
	router.GET("/metrics", func(c *gin.Context) {
		if sc.MonitoringService != nil {
			// Return basic metrics since GetAllMetrics doesn't exist
			c.JSON(http.StatusOK, gin.H{
				"message": "Monitoring service is active",
				"status":  "running",
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Monitoring service unavailable"})
		}
	})

	// Basic analytics endpoints
	router.GET("/analytics/overview", func(c *gin.Context) {
		if sc.AnalyticsService != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "Analytics service is running",
				"status":  "active",
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analytics service unavailable"})
		}
	})

	// Basic recommendation endpoints
	router.GET("/recommendations/trending", func(c *gin.Context) {
		if sc.RecommendationService != nil {
			// Return basic trending info since GetPersonalizedRecommendations doesn't exist
			c.JSON(http.StatusOK, gin.H{
				"message":  "Recommendation service is active",
				"trending": []string{}, // Empty array for now
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Recommendation service unavailable"})
		}
	})

	// WebSocket endpoint - protected with authentication middleware
	if sc.WebSocketHub != nil {
		protectedWS := router.Group("/ws")
		protectedWS.Use(middleware.AuthMiddleWare())
		{
			protectedWS.GET("", func(c *gin.Context) {
				// Extract authenticated user information from context (set by auth middleware)
				userID := c.GetString("user_id")
				if userID == "" {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
					return
				}

				// Use the existing WebSocket hub to handle the connection
				websocket.HandleWebSocketConnection(sc.WebSocketHub, c)
			})
		}
	} else {
		router.GET("/ws", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "WebSocket service unavailable"})
		})
	}
}

func setupEnhancedProtectedRoutes(router *gin.Engine, sc *ServiceContainer) {
	// Apply authentication middleware
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleWare())
	{
		// User profile endpoints
		protected.GET("/profile/:userId", func(c *gin.Context) {
			userID := c.Param("userId")
			if sc.SocialService != nil {
				profile, err := sc.SocialService.GetUserProfile(userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, profile)
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		protected.PUT("/profile", func(c *gin.Context) {
			userID := c.GetString("user_id") // Assuming user_id is set by auth middleware
			var updates map[string]interface{}
			if err := c.ShouldBindJSON(&updates); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if sc.SocialService != nil {
				err := sc.SocialService.UpdateUserProfile(userID, updates)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Follow/Unfollow endpoints
		protected.POST("/follow/:userId", func(c *gin.Context) {
			followerID := c.GetString("user_id")
			followingID := c.Param("userId")

			if sc.SocialService != nil {
				err := sc.SocialService.FollowUser(followerID, followingID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "User followed successfully"})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		protected.DELETE("/follow/:userId", func(c *gin.Context) {
			followerID := c.GetString("user_id")
			followingID := c.Param("userId")

			if sc.SocialService != nil {
				err := sc.SocialService.UnfollowUser(followerID, followingID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "User unfollowed successfully"})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Activity feed
		protected.GET("/activity/feed", func(c *gin.Context) {
			userID := c.GetString("user_id")
			limitStr := c.DefaultQuery("limit", "20")
			offsetStr := c.DefaultQuery("offset", "0")

			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				limit = 20
			}
			offset, err := strconv.Atoi(offsetStr)
			if err != nil {
				offset = 0
			}

			if sc.SocialService != nil {
				activities, err := sc.SocialService.GetActivityFeed(userID, limit, offset)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"activities": activities})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Enhanced recommendations
		protected.GET("/recommendations/personalized", func(c *gin.Context) {
			userID := c.GetString("user_id")

			if sc.RecommendationService != nil {
				// Use a basic response since GetPersonalizedRecommendations doesn't exist
				c.JSON(http.StatusOK, gin.H{
					"user_id":         userID,
					"message":         "Personalized recommendations would be generated here",
					"recommendations": []string{}, // Empty array for now
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Recommendation service unavailable"})
			}
		})

		// Analytics tracking
		protected.POST("/analytics/track", func(c *gin.Context) {
			var event analytics.AnalyticsEvent
			if err := c.ShouldBindJSON(&event); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Set user ID from context
			event.UserID = c.GetString("user_id")

			if sc.AnalyticsService != nil {
				err := sc.AnalyticsService.TrackEvent(event)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Event tracked successfully"})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analytics service unavailable"})
			}
		})

		// Watchlist management
		protected.POST("/watchlist/add/:movieId", func(c *gin.Context) {
			userID := c.GetString("user_id")
			movieID := c.Param("movieId")
			watchlistID := userID + "_watchlist" // Simple watchlist ID generation

			if sc.SocialService != nil {
				err := sc.SocialService.AddToWatchlist(watchlistID, movieID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Added to watchlist"})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		protected.DELETE("/watchlist/remove/:movieId", func(c *gin.Context) {
			userID := c.GetString("user_id")
			movieID := c.Param("movieId")
			watchlistID := userID + "_watchlist"

			if sc.SocialService != nil {
				err := sc.SocialService.RemoveFromWatchlist(watchlistID, movieID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Removed from watchlist"})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Security endpoints
		protected.GET("/security/audit-log", func(c *gin.Context) {
			userID := c.GetString("user_id")

			if sc.SecurityService != nil {
				// Return basic audit log info since GetUserAuditLog doesn't exist
				c.JSON(http.StatusOK, gin.H{
					"user_id":       userID,
					"message":       "Audit log would be retrieved here",
					"audit_entries": []string{}, // Empty array for now
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Security service unavailable"})
			}
		})
	}

	// Admin endpoints (simplified)
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleWare())
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message":  "Admin dashboard",
				"services": sc.GetServiceStatus(),
			})
		})

		admin.GET("/users", func(c *gin.Context) {
			if sc.SocialService != nil {
				// Return basic user info since GetAllUsers doesn't exist
				c.JSON(http.StatusOK, gin.H{
					"message": "User list would be retrieved here",
					"users":   []string{}, // Empty array for now
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})
	}
}
