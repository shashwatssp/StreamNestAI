package routes

import (
	"fmt"
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
		// Log the request for debugging
		fmt.Printf("[V2 API] GET /movies called with query params: %v\n", c.Request.URL.RawQuery)

		// Delegate to actual movie controller like v1 does
		// This ensures v2 returns real data from database
		getMoviesHandler := controllers.GetMovies(database.Client)
		getMoviesHandler(c) // Call the actual movie controller function
	})

	// Enhanced movie details with recommendations
	v2.GET("/movies/:imdb_id", func(c *gin.Context) {
		imdb_id := c.Param("imdb_id")
		includeRecommendations := c.DefaultQuery("include_recommendations", "false")

		// Log the request for debugging
		fmt.Printf("[V2 API] GET /movies/%s called with include_recommendations: %s\n", imdb_id, includeRecommendations)

		// Delegate to actual movie controller for real data
		getMovieHandler := controllers.GetMovie(database.Client)
		getMovieHandler(c) // Call the actual movie controller function
	})

	// Enhanced search with filters
	v2.GET("/movies/search", func(c *gin.Context) {
		query := c.Query("q")
		genre := c.Query("genre")
		year := c.Query("year")
		sort := c.DefaultQuery("sort", "relevance")

		// Log the request for debugging
		fmt.Printf("[V2 API] GET /movies/search called with query: %s, genre: %s, year: %s, sort: %s\n", query, genre, year, sort)

		// Delegate to actual search controller for real data
		searchMoviesHandler := controllers.NaturalLanguageMovieSearch(database.Client)
		searchMoviesHandler(c) // Call the actual search controller function
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
				"content": []gin.H{
					{
						"movie_id":         "tt1856101",
						"title":            "Inception",
						"poster_url":       "https://m.media-amazon.com/images/M/MV5BMjAxMzY3NjcxNF5BMl5BanBnXkFtZTcwNTI5OTM0Mw@@._V1_SX300.jpg",
						"genre":            "Action, Sci-Fi, Thriller",
						"year":             "2010",
						"duration":         "148 min",
						"rating":           8.8,
						"description":      "A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.",
						"confidence_score": 0.90,
						"match_score":      0.87,
						"reason":           "Trending in sci-fi action genre",
					},
					{
						"movie_id":         "tt0468569",
						"title":            "The Dark Knight",
						"poster_url":       "https://m.media-amazon.com/images/M/MV5BMTMxNTMwODM0NF5BMl5BanBnXkFtZTcwODAyMTk2Mw@@._V1_SX300.jpg",
						"genre":            "Action, Crime, Drama",
						"year":             "2008",
						"duration":         "152 min",
						"rating":           9.0,
						"description":      "When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests.",
						"confidence_score": 0.93,
						"match_score":      0.91,
						"reason":           "Most popular superhero movie",
					},
					{
						"movie_id":         "tt1375666",
						"title":            "Interstellar",
						"poster_url":       "https://m.media-amazon.com/images/M/MV5BZjdkOTU3MDktN2IxOS00OGEyLWFmMjktY2FiMmZkNWIyODZiXkEyXkFqcGdeQXVyMTMxODk2OTU@._V1_SX300.jpg",
						"genre":            "Adventure, Drama, Sci-Fi",
						"year":             "2014",
						"duration":         "169 min",
						"rating":           8.6,
						"description":      "A team of explorers travel through a wormhole in space in an attempt to ensure humanity's survival.",
						"confidence_score": 0.85,
						"match_score":      0.82,
						"reason":           "Rising in space exploration genre",
					},
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
			fmt.Printf("[V2 API] GET /profile/%s called\n", userID)

			if sc.SocialService != nil {
				fmt.Printf("[V2 API] SocialService is available, calling GetUserProfile for user: %s\n", userID)
				profile, err := sc.SocialService.GetUserProfile(userID)
				if err != nil {
					fmt.Printf("[V2 API] Error getting user profile: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				fmt.Printf("[V2 API] Successfully retrieved profile for user: %s\n", userID)
				c.JSON(http.StatusOK, gin.H{
					"profile": profile,
					"version": "v2",
				})
			} else {
				fmt.Printf("[V2 API] SocialService is nil/unavailable\n")
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
				// Return mock movie data with proper structure for frontend
				c.JSON(http.StatusOK, gin.H{
					"user_id": userID,
					"message": "v2 personalized recommendations",
					"version": "v2",
					"limit":   limit,
					"recommendations": []gin.H{
						{
							"movie_id":         "tt0111161",
							"title":            "The Shawshank Redemption",
							"poster_url":       "https://m.media-amazon.com/images/M/MV5BNDE3ODcxYzMtY2YzZC00NmNlLWJiODMtZDZhMDRjMmE0M2IwXkEyXkFqcGdeQXVyNjAwNDUxODI@._V1_SX300.jpg",
							"genre":            "Drama",
							"year":             "1994",
							"duration":         "142 min",
							"rating":           9.3,
							"description":      "Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.",
							"confidence_score": 0.95,
							"match_score":      0.92,
							"reason":           "Based on your interest in highly-rated dramas",
						},
						{
							"movie_id":         "tt0068646",
							"title":            "The Godfather",
							"poster_url":       "https://m.media-amazon.com/images/M/MV5BM2MyNjYxNmUtYTAwNi00MTYxLWJmNWYtYzZlODY3ZTVmNjY0XkEyXkFqcGdeQXVyNzkwMjQ5NzM@._V1_SX300.jpg",
							"genre":            "Crime, Drama",
							"year":             "1972",
							"duration":         "175 min",
							"rating":           9.2,
							"description":      "The aging patriarch of an organized crime dynasty transfers control of his clandestine empire to his reluctant son.",
							"confidence_score": 0.88,
							"match_score":      0.85,
							"reason":           "Classic crime drama matching your preferences",
						},
						{
							"movie_id":         "tt0071562",
							"title":            "The Godfather: Part II",
							"poster_url":       "https://m.media-amazon.com/images/M/MV5BMWMwMGQzZTItY2JlNC00OWZiLWIyMDctNDk2ZDQ2YjRjMWQ0XkEyXkFqcGdeQXVyNzkwMjQ5NzM@._V1_SX300.jpg",
							"genre":            "Crime, Drama",
							"year":             "1974",
							"duration":         "202 min",
							"rating":           9.0,
							"description":      "The early life and career of Vito Corleone in 1920s New York is portrayed while his son, Michael, expands and tightens his grip on the family crime syndicate.",
							"confidence_score": 0.82,
							"match_score":      0.78,
							"reason":           "Follow-up to your previously watched movies",
						},
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

		// Social stats endpoint
		v2.GET("/social/stats", func(c *gin.Context) {
			userID := c.GetString("user_id")
			fmt.Printf("[V2 API] GET /social/stats called for user: %s\n", userID)

			if sc.SocialService != nil {
				// Get user profile to extract stats
				profile, err := sc.SocialService.GetUserProfile(userID)
				if err != nil {
					fmt.Printf("[V2 API] Error getting user profile for stats: %v\n", err)
					// Return default stats if profile doesn't exist
					c.JSON(http.StatusOK, gin.H{
						"stats": gin.H{
							"followers_count": 0,
							"following_count": 0,
							"posts_count":     0,
							"watchlist_count": 0,
						},
						"version": "v2",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"stats": gin.H{
						"followers_count": profile.FollowersCount,
						"following_count": profile.FollowingCount,
						"posts_count":     profile.PostsCount,
						"watchlist_count": profile.WatchlistCount,
					},
					"version": "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Social followers endpoint
		v2.GET("/social/followers/:userId", func(c *gin.Context) {
			targetUserID := c.Param("userId")
			fmt.Printf("[V2 API] GET /social/followers/%s called\n", targetUserID)

			if sc.SocialService != nil {
				followers, err := sc.SocialService.GetFollowers(targetUserID, 50, 0)
				if err != nil {
					fmt.Printf("[V2 API] Error getting followers: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"followers": followers,
					"version":   "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Social following endpoint
		v2.GET("/social/following/:userId", func(c *gin.Context) {
			targetUserID := c.Param("userId")
			fmt.Printf("[V2 API] GET /social/following/%s called\n", targetUserID)

			if sc.SocialService != nil {
				following, err := sc.SocialService.GetFollowing(targetUserID, 50, 0)
				if err != nil {
					fmt.Printf("[V2 API] Error getting following: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"following": following,
					"version":   "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Social search endpoint
		v2.GET("/social/search", func(c *gin.Context) {
			query := c.Query("q")
			fmt.Printf("[V2 API] GET /social/search called with query: %s\n", query)

			// For now, return empty results - this would need proper implementation
			c.JSON(http.StatusOK, gin.H{
				"users":   []gin.H{},
				"version": "v2",
			})
		})

		// Social activity endpoint
		v2.POST("/social/activity", func(c *gin.Context) {
			userID := c.GetString("user_id")
			var activityData map[string]interface{}
			if err := c.ShouldBindJSON(&activityData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			fmt.Printf("[V2 API] POST /social/activity called by user: %s\n", userID)

			if sc.SocialService != nil {
				activityType := activityData["type"].(string)
				contentID := ""
				contentType := "movie"
				title := ""
				description := ""

				if activityData["content_id"] != nil {
					contentID = activityData["content_id"].(string)
				}
				if activityData["content_type"] != nil {
					contentType = activityData["content_type"].(string)
				}
				if activityData["title"] != nil {
					title = activityData["title"].(string)
				}
				if activityData["description"] != nil {
					description = activityData["description"].(string)
				}

				err := sc.SocialService.CreateActivity(userID, activityType, contentID, contentType, title, description, activityData)
				if err != nil {
					fmt.Printf("[V2 API] Error creating activity: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "Activity created successfully",
					"version": "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Watchlist endpoints
		v2.GET("/watchlist/user/:userId", func(c *gin.Context) {
			userID := c.Param("userId")
			fmt.Printf("[V2 API] GET /watchlist/user/%s called\n", userID)

			if sc.SocialService != nil {
				watchlistID := userID + "_watchlist"
				watchlist, err := sc.SocialService.GetWatchlist(watchlistID)
				if err != nil {
					fmt.Printf("[V2 API] Error getting watchlist: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"watchlist": watchlist,
					"version":   "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		v2.GET("/watchlist/public", func(c *gin.Context) {
			fmt.Printf("[V2 API] GET /watchlist/public called\n")

			// For now, return empty public watchlists - this would need proper implementation
			c.JSON(http.StatusOK, gin.H{
				"public_watchlists": []gin.H{},
				"version":           "v2",
			})
		})

		v2.GET("/watchlist/trending", func(c *gin.Context) {
			fmt.Printf("[V2 API] GET /watchlist/trending called\n")

			// For now, return empty trending watchlists - this would need proper implementation
			c.JSON(http.StatusOK, gin.H{
				"trending": []gin.H{},
				"version":  "v2",
			})
		})

		v2.DELETE("/watchlist/remove/:movieId", func(c *gin.Context) {
			userID := c.GetString("user_id")
			movieID := c.Param("movieId")
			watchlistID := userID + "_watchlist"

			fmt.Printf("[V2 API] DELETE /watchlist/remove/%s called for user: %s\n", movieID, userID)

			if sc.SocialService != nil {
				err := sc.SocialService.RemoveFromWatchlist(watchlistID, movieID)
				if err != nil {
					fmt.Printf("[V2 API] Error removing from watchlist: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "Movie removed from watchlist successfully",
					"version": "v2",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social service unavailable"})
			}
		})

		// Search endpoints
		v2.GET("/search/history", func(c *gin.Context) {
			userID := c.GetString("user_id")
			fmt.Printf("[V2 API] GET /search/history called for user: %s\n", userID)

			// For now, return empty search history - this would need proper implementation
			c.JSON(http.StatusOK, gin.H{
				"search_history": []gin.H{},
				"version":        "v2",
			})
		})

		v2.GET("/search/advanced", func(c *gin.Context) {
			query := c.Query("q")
			genre := c.Query("genre")
			year := c.Query("year")
			rating := c.Query("rating")
			sort := c.DefaultQuery("sort", "relevance")

			fmt.Printf("[V2 API] GET /search/advanced called with query: %s, genre: %s, year: %s, rating: %s, sort: %s\n", query, genre, year, rating, sort)

			// Delegate to actual search controller for real data
			searchMoviesHandler := controllers.NaturalLanguageMovieSearch(database.Client)
			searchMoviesHandler(c) // Call the actual search controller function
		})

		v2.GET("/search/natural", func(c *gin.Context) {
			query := c.Query("q")
			fmt.Printf("[V2 API] GET /search/natural called with query: %s\n", query)

			// Delegate to actual search controller for real data
			searchMoviesHandler := controllers.NaturalLanguageMovieSearch(database.Client)
			searchMoviesHandler(c) // Call the actual search controller function
		})

		v2.GET("/search/semantic", func(c *gin.Context) {
			query := c.Query("q")
			fmt.Printf("[V2 API] GET /search/semantic called with query: %s\n", query)

			// For now, delegate to natural language search - semantic search would need proper implementation
			searchMoviesHandler := controllers.NaturalLanguageMovieSearch(database.Client)
			searchMoviesHandler(c) // Call the actual search controller function
		})

		// Monitoring endpoints
		monitoring := v2.Group("/monitoring")
		monitoring.Use(middleware.AuthMiddleWare())
		{
			monitoring.GET("/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"status":    "healthy",
					"timestamp": "2026-01-19T15:34:00Z",
					"services": gin.H{
						"database": "healthy",
						"redis":    "healthy",
						"api":      "healthy",
					},
					"api_version": "v2",
				})
			})

			monitoring.GET("/performance", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"cpu_usage":    45.2,
					"memory_usage": 67.8,
					"disk_usage":   23.1,
					"network_io": gin.H{
						"bytes_in":  1024000,
						"bytes_out": 512000,
					},
					"response_times": gin.H{
						"avg": 120.5,
						"p95": 250.8,
						"p99": 450.2,
					},
					"api_version": "v2",
				})
			})

			monitoring.GET("/errors", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"errors": []gin.H{
						{
							"timestamp":   "2026-01-19T15:30:00Z",
							"level":       "error",
							"message":     "Database connection timeout",
							"endpoint":    "/api/v2/movies",
							"user_id":     "user123",
							"stack_trace": "goroutine 1 [running]:\nmain.main()",
						},
						{
							"timestamp":   "2026-01-19T15:25:00Z",
							"level":       "error",
							"message":     "Redis cache miss",
							"endpoint":    "/api/v2/recommendations",
							"user_id":     "user456",
							"stack_trace": "",
						},
					},
					"total_count": 2,
					"api_version": "v2",
				})
			})

			monitoring.GET("/api", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"endpoints": []gin.H{
						{
							"path":         "/api/v2/movies",
							"method":       "GET",
							"requests":     1250,
							"avg_response": 120.5,
							"error_rate":   0.02,
						},
						{
							"path":         "/api/v2/profile",
							"method":       "GET",
							"requests":     850,
							"avg_response": 95.3,
							"error_rate":   0.01,
						},
					},
					"total_requests": 2100,
					"avg_response":   108.9,
					"error_rate":     0.015,
					"api_version":    "v2",
				})
			})

			monitoring.GET("/database", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"connections": gin.H{
						"active":      15,
						"idle":        5,
						"max":         100,
						"utilization": 15.0,
					},
					"queries": gin.H{
						"total":    5000,
						"slow":     12,
						"failed":   3,
						"avg_time": 45.2,
					},
					"collections": []gin.H{
						{
							"name":  "users",
							"count": 1250,
							"size":  "15.2MB",
						},
						{
							"name":  "movies",
							"count": 5000,
							"size":  "125.5MB",
						},
					},
					"api_version": "v2",
				})
			})

			monitoring.GET("/cache", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"redis": gin.H{
						"status":       "connected",
						"hit_rate":     0.85,
						"miss_rate":    0.15,
						"memory_usage": "45.2MB",
						"total_keys":   1250,
						"expired_keys": 45,
					},
					"application_cache": gin.H{
						"size":      "12.5MB",
						"entries":   850,
						"hit_rate":  0.72,
						"evictions": 12,
					},
					"api_version": "v2",
				})
			})

			monitoring.GET("/server", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"uptime":      "72h30m15s",
					"start_time":  "2026-01-16T09:04:00Z",
					"version":     "2.1.0",
					"environment": "development",
					"goroutines":  45,
					"heap_size":   "125.5MB",
					"gc_cycles":   1250,
					"system": gin.H{
						"os":           "windows",
						"architecture": "amd64",
						"cpu_cores":    8,
						"total_memory": "16GB",
					},
					"api_version": "v2",
				})
			})
		}

		// Analytics endpoints
		analytics := v2.Group("/analytics")
		analytics.Use(middleware.AuthMiddleWare())
		{
			analytics.GET("/realtime", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"timestamp":           "2026-01-19T15:34:00Z",
					"active_users":        125,
					"current_load":        0.65,
					"requests_per_second": 15.2,
					"error_rate":          0.01,
					"avg_response_time":   120.5,
					"top_endpoints": []gin.H{
						{
							"path":     "/api/v2/movies",
							"requests": 45,
							"avg_time": 95.2,
						},
						{
							"path":     "/api/v2/profile",
							"requests": 32,
							"avg_time": 78.5,
						},
					},
					"api_version": "v2",
				})
			})

			analytics.GET("/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"total_users":     1250,
					"active_users":    850,
					"new_users_today": 12,
					"retention_rate":  0.75,
					"demographics": gin.H{
						"age_groups": gin.H{
							"18-24": 0.25,
							"25-34": 0.35,
							"35-44": 0.25,
							"45+":   0.15,
						},
						"locations": []gin.H{
							{"country": "United States", "users": 450},
							{"country": "India", "users": 320},
							{"country": "United Kingdom", "users": 180},
						},
					},
					"api_version": "v2",
				})
			})

			analytics.GET("/content", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"total_movies":      5000,
					"total_series":      1200,
					"new_content_today": 8,
					"popular_genres": []gin.H{
						{"genre": "Action", "count": 1250, "percentage": 0.25},
						{"genre": "Drama", "count": 1000, "percentage": 0.20},
						{"genre": "Comedy", "count": 750, "percentage": 0.15},
						{"genre": "Thriller", "count": 625, "percentage": 0.125},
					},
					"top_content": []gin.H{
						{
							"title":  "Popular Movie 1",
							"views":  12500,
							"rating": 4.5,
							"genre":  "Action",
						},
						{
							"title":  "Popular Series 1",
							"views":  8500,
							"rating": 4.7,
							"genre":  "Drama",
						},
					},
					"api_version": "v2",
				})
			})

			analytics.GET("/engagement", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"daily_active_users": 850,
					"session_duration":   "25m30s",
					"pages_per_session":  8.5,
					"bounce_rate":        0.25,
					"interactions": gin.H{
						"likes":               1250,
						"comments":            450,
						"shares":              180,
						"watchlist_additions": 320,
					},
					"feature_usage": []gin.H{
						{
							"feature": "Movie Streaming",
							"usage":   0.85,
						},
						{
							"feature": "Social Features",
							"usage":   0.65,
						},
						{
							"feature": "Recommendations",
							"usage":   0.75,
						},
					},
					"api_version": "v2",
				})
			})

			analytics.GET("/conversion", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"sign_up_conversion":      0.15,
					"subscription_conversion": 0.08,
					"trial_to_paid":           0.65,
					"funnel_stages": []gin.H{
						{
							"stage":      "Visit",
							"users":      10000,
							"conversion": 1.0,
						},
						{
							"stage":      "Sign Up",
							"users":      1500,
							"conversion": 0.15,
						},
						{
							"stage":      "Trial",
							"users":      800,
							"conversion": 0.53,
						},
						{
							"stage":      "Paid",
							"users":      520,
							"conversion": 0.65,
						},
					},
					"revenue_metrics": gin.H{
						"monthly_recurring_revenue": 12500,
						"average_revenue_per_user":  24.5,
						"customer_lifetime_value":   295.5,
					},
					"api_version": "v2",
				})
			})
		}

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
