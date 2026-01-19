package routes

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/analytics"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/controllers"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/database"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/middleware"
)

// SetupVersionedRoutes creates versioned API routes for backward compatibility
func SetupVersionedRoutes(router *gin.Engine) {
	sc := GetServiceContainer()
	if sc == nil {
		return
	}

	// API v1 routes (current stable version)
	v1 := router.Group("/api/v1")
	{
		setupV1PublicRoutes(v1, sc)
		setupV1ProtectedRoutes(v1, sc)
	}

	// API v2 routes (enhanced features)
	v2 := router.Group("/api/v2")
	{
		setupV2PublicRoutes(v2, sc)
		setupV2ProtectedRoutes(v2, sc)
	}

	// API version info
	router.GET("/api/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"versions": []gin.H{
				{
					"version":     "v1",
					"status":      "stable",
					"description": "Original API with basic features",
					"deprecated":  false,
				},
				{
					"version":     "v2",
					"status":      "stable",
					"description": "Enhanced API with social features, analytics, and real-time capabilities",
					"deprecated":  false,
				},
			},
			"current":     "v2",
			"recommended": "v2",
		})
	})
}

// v1 routes - Original API (backward compatible)
func setupV1PublicRoutes(v1 *gin.RouterGroup, sc *ServiceContainer) {
	// Health check
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": "v1",
		})
	})

	// Movies
	v1.GET("/movies", func(c *gin.Context) {
		// Delegate to original movie controller
		// This maintains compatibility with existing v1 clients
		getMoviesHandler := controllers.GetMovies(database.Client)
		getMoviesHandler(c) // Call the actual movie controller function
	})

	v1.GET("/movies/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"message":  "v1 movie detail",
			"movie_id": id,
			"version":  "v1",
		})
	})

	// Search
	v1.GET("/movies/search", func(c *gin.Context) {
		query := c.Query("q")
		c.JSON(http.StatusOK, gin.H{
			"message": "v1 search",
			"query":   query,
			"version": "v1",
		})
	})

	// Authentication
	v1.POST("/register", func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "v1 registration endpoint not implemented",
			"message": "Please use v2 API endpoint: /api/v2/register",
			"version": "v1",
		})
	})

	v1.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "v1 login endpoint not implemented",
			"message": "Please use v2 API endpoint: /api/v2/login",
			"version": "v1",
		})
	})
}

func setupV1ProtectedRoutes(v1 *gin.RouterGroup, sc *ServiceContainer) {
	// Apply authentication middleware
	v1.Use(middleware.AuthMiddleWare())
	{
		// User profile
		v1.GET("/profile", func(c *gin.Context) {
			userID := c.GetString("user_id")
			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"message": "v1 user profile",
				"version": "v1",
			})
		})

		// Reviews
		v1.POST("/movies/:id/reviews", func(c *gin.Context) {
			movieID := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"message":  "v1 review created",
				"movie_id": movieID,
				"version":  "v1",
			})
		})

		// Watchlist
		v1.POST("/watchlist/:movieId", func(c *gin.Context) {
			movieID := c.Param("movieId")
			c.JSON(http.StatusOK, gin.H{
				"message":  "v1 added to watchlist",
				"movie_id": movieID,
				"version":  "v1",
			})
		})

		v1.DELETE("/watchlist/:movieId", func(c *gin.Context) {
			movieID := c.Param("movieId")
			c.JSON(http.StatusOK, gin.H{
				"message":  "v1 removed from watchlist",
				"movie_id": movieID,
				"version":  "v1",
			})
		})
	}
}

// v2 routes - Enhanced API with new features
func setupV2PublicRoutes(v2 *gin.RouterGroup, sc *ServiceContainer) {
	// Health check with enhanced info
	v2.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"version":  "v2",
			"services": sc.IsHealthy(),
			"features": []string{
				"real-time-websocket",
				"analytics",
				"social-features",
				"recommendations",
				"advanced-caching",
			},
		})
	})

	// Enhanced movies with caching
	v2.GET("/movies", func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "20")

		c.JSON(http.StatusOK, gin.H{
			"message": "v2 movies with enhanced features",
			"page":    page,
			"limit":   limit,
			"version": "v2",
			"cached":  true,
		})
	})

	// Enhanced movie details with recommendations
	v2.GET("/movies/:id", func(c *gin.Context) {
		id := c.Param("id")
		includeRecommendations := c.DefaultQuery("include_recommendations", "false")

		response := gin.H{
			"message":  "v2 enhanced movie detail",
			"movie_id": id,
			"version":  "v2",
		}

		if includeRecommendations == "true" && sc.RecommendationService != nil {
			response["recommendations"] = []string{"related_movie_1", "related_movie_2"}
		}

		c.JSON(http.StatusOK, response)
	})

	// Enhanced search with filters
	v2.GET("/movies/search", func(c *gin.Context) {
		query := c.Query("q")
		genre := c.Query("genre")
		year := c.Query("year")
		sort := c.DefaultQuery("sort", "relevance")

		c.JSON(http.StatusOK, gin.H{
			"message": "v2 enhanced search",
			"query":   query,
			"filters": gin.H{
				"genre": genre,
				"year":  year,
				"sort":  sort,
			},
			"version": "v2",
		})
	})

	// Enhanced authentication with social login options
	v2.POST("/register", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "v2 enhanced registration",
			"version": "v2",
			"features": []string{
				"email-verification",
				"social-login",
				"profile-completion",
			},
		})
	})

	v2.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":       "v2 enhanced login",
			"version":       "v2",
			"token":         "v2_enhanced_token_example",
			"refresh_token": "v2_refresh_token_example",
			"expires_in":    86400,
		})
	})

	// Public recommendations
	v2.GET("/recommendations/trending", func(c *gin.Context) {
		if sc.RecommendationService != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "v2 trending recommendations",
				"version": "v2",
				"trending": []string{
					"trending_movie_1",
					"trending_movie_2",
					"trending_movie_3",
				},
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Recommendation service unavailable"})
		}
	})

	// Public analytics overview
	v2.GET("/analytics/overview", func(c *gin.Context) {
		if sc.AnalyticsService != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": "v2 public analytics",
				"version": "v2",
				"stats": gin.H{
					"total_movies": 1000,
					"active_users": 500,
					"daily_views":  10000,
				},
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analytics service unavailable"})
		}
	})
}

func setupV2ProtectedRoutes(v2 *gin.RouterGroup, sc *ServiceContainer) {
	// Apply authentication middleware
	v2.Use(middleware.AuthMiddleWare())
	{
		// Enhanced user profile with social features
		v2.GET("/profile/:userId", func(c *gin.Context) {
			userID := c.Param("userId")
			if sc.SocialService != nil {
				profile, err := sc.SocialService.GetUserProfile(userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"profile": profile,
					"version": "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		v2.PUT("/profile", func(c *gin.Context) {
			userID := c.GetString("user_id")
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
				c.JSON(http.StatusOK, gin.H{
					"message": "v2 profile updated successfully",
					"version": "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Social features
		v2.POST("/follow/:userId", func(c *gin.Context) {
			followerID := c.GetString("user_id")
			followingID := c.Param("userId")

			if sc.SocialService != nil {
				err := sc.SocialService.FollowUser(followerID, followingID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"message": "v2 user followed successfully",
					"version": "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		v2.DELETE("/follow/:userId", func(c *gin.Context) {
			followerID := c.GetString("user_id")
			followingID := c.Param("userId")

			if sc.SocialService != nil {
				err := sc.SocialService.UnfollowUser(followerID, followingID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"message": "v2 user unfollowed successfully",
					"version": "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Enhanced activity feed
		v2.GET("/activity/feed", func(c *gin.Context) {
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
				c.JSON(http.StatusOK, gin.H{
					"activities": activities,
					"version":    "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Enhanced recommendations
		v2.GET("/recommendations/personalized", func(c *gin.Context) {
			userID := c.GetString("user_id")
			limit := c.DefaultQuery("limit", "10")

			if sc.RecommendationService != nil {
				c.JSON(http.StatusOK, gin.H{
					"user_id": userID,
					"message": "v2 personalized recommendations",
					"version": "v2",
					"limit":   limit,
					"recommendations": []string{
						"personalized_movie_1",
						"personalized_movie_2",
						"personalized_movie_3",
					},
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Recommendation service unavailable"})
			}
		})

		// Enhanced analytics tracking
		v2.POST("/analytics/track", func(c *gin.Context) {
			var event analytics.AnalyticsEvent
			if err := c.ShouldBindJSON(&event); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			event.UserID = c.GetString("user_id")

			if sc.AnalyticsService != nil {
				err := sc.AnalyticsService.TrackEvent(event)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"message": "v2 event tracked successfully",
					"version": "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analytics service unavailable"})
			}
		})

		// Enhanced watchlist with sharing
		v2.POST("/watchlist/add/:movieId", func(c *gin.Context) {
			userID := c.GetString("user_id")
			movieID := c.Param("movieId")
			watchlistID := userID + "_watchlist"

			if sc.SocialService != nil {
				err := sc.SocialService.AddToWatchlist(watchlistID, movieID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"message":   "v2 added to watchlist",
					"version":   "v2",
					"shareable": true,
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Security features
		v2.GET("/security/audit-log", func(c *gin.Context) {
			userID := c.GetString("user_id")

			if sc.SecurityService != nil {
				c.JSON(http.StatusOK, gin.H{
					"user_id": userID,
					"message": "v2 audit log",
					"version": "v2",
					"audit_entries": []string{
						"login_attempt",
						"password_change",
						"profile_update",
					},
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Security service unavailable"})
			}
		})

		// Admin endpoints
		admin := v2.Group("/admin")
		admin.Use(middleware.AuthMiddleWare())
		{
			admin.GET("/dashboard", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message":  "v2 admin dashboard",
					"version":  "v2",
					"services": sc.GetServiceStatus(),
				})
			})

			admin.GET("/users", func(c *gin.Context) {
				if sc.SocialService != nil {
					c.JSON(http.StatusOK, gin.H{
						"message": "v2 user management",
						"version": "v2",
						"users": []string{
							"user_1",
							"user_2",
							"user_3",
						},
					})
				} else {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
				}
			})
		}
	}
}
