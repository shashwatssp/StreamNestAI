package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/cache"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/database"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/models"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/gin-gonic/gin"
)

// CachedMovieController handles movie operations with Redis caching
type CachedMovieController struct {
	client       *mongo.Client
	redisService *cache.RedisService
}

// NewCachedMovieController creates a new cached movie controller
func NewCachedMovieController(client *mongo.Client, redisService *cache.RedisService) *CachedMovieController {
	return &CachedMovieController{
		client:       client,
		redisService: redisService,
	}
}

// GetMoviesWithCache handles GET /movies with cache-first strategy
func (cmc *CachedMovieController) GetMoviesWithCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("INFO: GetMoviesWithCache endpoint called")
		startTime := time.Now()

		// Generate cache key with query parameters
		searchQuery := c.Query("search")
		var cacheKey string
		if searchQuery != "" {
			cacheKey = fmt.Sprintf("movies:search:%s", searchQuery)
			log.Printf("DEBUG: Generated cache key: %s", cacheKey)
		} else {
			cacheKey = "movies:all"
			log.Printf("DEBUG: Generated cache key: %s", cacheKey)
		}

		// Try to get from Redis first
		var movies []models.Movie
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := cmc.redisService.Get(ctx, cacheKey, &movies)
		if err == nil {
			// Cache hit
			elapsedTime := time.Since(startTime)
			log.Printf("CACHE HIT: Retrieved %d movies from Redis in %v", len(movies), elapsedTime)
			c.Header("X-Data-Source", "redis")
			c.Header("X-Cache-Status", "hit")
			c.Header("X-Response-Time", elapsedTime.String())
			c.JSON(http.StatusOK, gin.H{
				"data":   movies,
				"source": "redis",
				"count":  len(movies),
				"cached": true,
			})
			return
		}

		log.Printf("CACHE MISS: %s - falling back to MongoDB", err.Error())

		// Cache miss - get from MongoDB
		mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer mongoCancel()

		log.Println("INFO: Opening movies collection from MongoDB")
		var movieCollection *mongo.Collection = database.OpenCollection("movies", cmc.client)

		if movieCollection == nil {
			log.Println("ERROR: Failed to open movies collection")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed To Open Movies Collection"})
			return
		}

		log.Println("INFO: Executing Find query on movies collection")
		cursor, err := movieCollection.Find(mongoCtx, bson.M{})

		if err != nil {
			log.Printf("ERROR: Failed to fetch movies from database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed To Fetch Movies"})
			return
		}

		defer cursor.Close(mongoCtx)

		log.Println("INFO: Decoding cursor results into movies slice")
		if err = cursor.All(mongoCtx, &movies); err != nil {
			log.Printf("ERROR: Failed to decode movies: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed To Decode Movies"})
			return
		}

		elapsedTime := time.Since(startTime)
		log.Printf("MONGODB SUCCESS: Retrieved %d movies from MongoDB in %v", len(movies), elapsedTime)

		// Cache the results for future requests (including empty results)
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cacheCancel()

		// Cache for 30 minutes as requested
		cacheTTL := 30 * time.Minute
		err = cmc.redisService.Set(cacheCtx, cacheKey, movies, cacheTTL)
		if err != nil {
			log.Printf("WARNING: Failed to cache movies in Redis: %v", err)
		} else {
			log.Printf("SUCCESS: Cached %d movies in Redis for %v", len(movies), cacheTTL)
		}

		if len(movies) == 0 {
			log.Println("WARNING: No movies found in database - collection might be empty")
		} else {
			log.Printf("DEBUG: First movie title: %s", movies[0].Title)
		}

		c.Header("X-Data-Source", "mongodb")
		c.Header("X-Cache-Status", "miss")
		c.Header("X-Response-Time", elapsedTime.String())
		c.JSON(http.StatusOK, gin.H{
			"data":   movies,
			"source": "mongodb",
			"count":  len(movies),
			"cached": false,
		})
	}
}

// GetMovieWithCache handles GET /movie/:imdb_id with cache-first strategy
func (cmc *CachedMovieController) GetMovieWithCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		movieID := c.Param("imdb_id")

		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required"})
			return
		}

		log.Printf("INFO: GetMovieWithCache called for movie ID: %s", movieID)

		// Generate cache key
		cacheKey := fmt.Sprintf("movie:%s", movieID)
		log.Printf("DEBUG: Generated cache key: %s", cacheKey)

		// Try to get from Redis first
		var movie models.Movie
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := cmc.redisService.Get(ctx, cacheKey, &movie)
		if err == nil {
			// Cache hit
			elapsedTime := time.Since(startTime)
			log.Printf("CACHE HIT: Retrieved movie %s from Redis in %v", movieID, elapsedTime)
			c.Header("X-Data-Source", "redis")
			c.Header("X-Cache-Status", "hit")
			c.Header("X-Response-Time", elapsedTime.String())
			c.JSON(http.StatusOK, gin.H{
				"data":   movie,
				"source": "redis",
				"cached": true,
			})
			return
		}

		log.Printf("CACHE MISS: %s - falling back to MongoDB", err.Error())

		// Cache miss - get from MongoDB
		mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer mongoCancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", cmc.client)
		err = movieCollection.FindOne(mongoCtx, bson.M{"imdb_id": movieID}).Decode(&movie)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				log.Printf("WARNING: Movie %s not found in MongoDB", movieID)
				c.JSON(http.StatusNotFound, gin.H{"error": "Movie Not Found"})
			} else {
				log.Printf("ERROR: Failed to fetch movie %s from MongoDB: %v", movieID, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed To Fetch Movie"})
			}
			return
		}

		elapsedTime := time.Since(startTime)
		log.Printf("MONGODB SUCCESS: Retrieved movie %s from MongoDB in %v", movieID, elapsedTime)

		// Cache the result for future requests
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cacheCancel()

		// Cache individual movies for 30 minutes
		cacheTTL := 30 * time.Minute
		err = cmc.redisService.Set(cacheCtx, cacheKey, movie, cacheTTL)
		if err != nil {
			log.Printf("WARNING: Failed to cache movie %s in Redis: %v", movieID, err)
		} else {
			log.Printf("SUCCESS: Cached movie %s in Redis for %v", movieID, cacheTTL)
		}

		c.Header("X-Data-Source", "mongodb")
		c.Header("X-Cache-Status", "miss")
		c.Header("X-Response-Time", elapsedTime.String())
		c.JSON(http.StatusOK, gin.H{
			"data":   movie,
			"source": "mongodb",
			"cached": false,
		})
	}
}

// GetRecommendedMoviesWithCache handles GET /recommendedmovies with cache-first strategy
func (cmc *CachedMovieController) GetRecommendedMoviesWithCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		userId, err := utils.GetUserIdFromContext(c)
		if err != nil {
			log.Printf("ERROR: User ID not found in context: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User Id not found in context"})
			return
		}

		log.Printf("INFO: GetRecommendedMoviesWithCache called for user: %s", userId)

		// Generate cache key for user recommendations
		cacheKey := fmt.Sprintf("recommendations:user:%s", userId)
		log.Printf("DEBUG: Generated cache key: %s", cacheKey)

		// Try to get from Redis first
		var recommendedMovies []models.Movie
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = cmc.redisService.Get(ctx, cacheKey, &recommendedMovies)
		if err == nil {
			// Cache hit
			elapsedTime := time.Since(startTime)
			log.Printf("CACHE HIT: Retrieved %d recommended movies for user %s from Redis in %v", len(recommendedMovies), userId, elapsedTime)
			c.Header("X-Data-Source", "redis")
			c.Header("X-Cache-Status", "hit")
			c.Header("X-Response-Time", elapsedTime.String())
			c.JSON(http.StatusOK, gin.H{
				"data":   recommendedMovies,
				"source": "redis",
				"count":  len(recommendedMovies),
				"cached": true,
			})
			return
		}

		log.Printf("CACHE MISS: %s - falling back to MongoDB", err.Error())

		// Cache miss - get from MongoDB using existing logic
		favourite_genres, err := cmc.GetUsersFavouriteGenres(userId)
		if err != nil {
			log.Printf("ERROR: Failed to get user favourite genres: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: .env file not found")
		}

		var recommendedMovieLimitVal int64 = 5
		recommendedMovieLimitStr := os.Getenv("RECOMMENDED_MOVIE_LIMIT")
		if recommendedMovieLimitStr != "" {
			recommendedMovieLimitVal, _ = strconv.ParseInt(recommendedMovieLimitStr, 10, 64)
		}

		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "ranking.ranking_value", Value: 1}})
		findOptions.SetLimit(recommendedMovieLimitVal)

		filter := bson.D{
			{Key: "genre.genre_name", Value: bson.D{
				{Key: "$in", Value: favourite_genres},
			}},
		}

		mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer mongoCancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", cmc.client)
		cursor, err := movieCollection.Find(mongoCtx, filter, findOptions)

		if err != nil {
			log.Printf("ERROR: Failed to fetch recommended movies from database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended movies"})
			return
		}
		defer cursor.Close(mongoCtx)

		if err := cursor.All(mongoCtx, &recommendedMovies); err != nil {
			log.Printf("ERROR: Failed to decode recommended movies: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		elapsedTime := time.Since(startTime)
		log.Printf("MONGODB SUCCESS: Retrieved %d recommended movies for user %s from MongoDB in %v", len(recommendedMovies), userId, elapsedTime)

		// Cache the results for future requests (including empty results)
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cacheCancel()

		// Cache recommendations for 30 minutes as requested
		cacheTTL := 30 * time.Minute
		err = cmc.redisService.Set(cacheCtx, cacheKey, recommendedMovies, cacheTTL)
		if err != nil {
			log.Printf("WARNING: Failed to cache recommended movies for user %s in Redis: %v", userId, err)
		} else {
			log.Printf("SUCCESS: Cached %d recommended movies for user %s in Redis for %v", len(recommendedMovies), userId, cacheTTL)
		}

		c.Header("X-Data-Source", "mongodb")
		c.Header("X-Cache-Status", "miss")
		c.Header("X-Response-Time", elapsedTime.String())
		c.JSON(http.StatusOK, gin.H{
			"data":   recommendedMovies,
			"source": "mongodb",
			"count":  len(recommendedMovies),
			"cached": false,
		})
	}
}

// GetUsersFavouriteGenres retrieves user's favourite genres (helper method)
func (cmc *CachedMovieController) GetUsersFavouriteGenres(userId string) ([]string, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.D{{Key: "user_id", Value: userId}}

	projection := bson.M{
		"favourite_genres.genre_name": 1,
		"_id":                         0,
	}

	opts := options.FindOne().SetProjection(projection)
	var result bson.M

	var userCollection *mongo.Collection = database.OpenCollection("users", cmc.client)
	err := userCollection.FindOne(ctx, filter, opts).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
		return nil, err
	}

	favGenresArray, ok := result["favourite_genres"].(bson.A)

	if !ok {
		return []string{}, errors.New("unable to retrieve favourite genres for user")
	}

	var genreNames []string

	for _, item := range favGenresArray {
		if genreMap, ok := item.(bson.D); ok {
			for _, elem := range genreMap {
				if elem.Key == "genre_name" {
					if name, ok := elem.Value.(string); ok {
						genreNames = append(genreNames, name)
					}
				}
			}
		}
	}

	return genreNames, nil
}

// InvalidateMovieCache clears movie-related cache entries
func (cmc *CachedMovieController) InvalidateMovieCache(movieID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Clear specific movie cache
	movieKey := fmt.Sprintf("movie:%s", movieID)
	err := cmc.redisService.Delete(ctx, movieKey)
	if err != nil {
		log.Printf("WARNING: Failed to delete movie cache for %s: %v", movieID, err)
	}

	// Clear all movies cache
	err = cmc.redisService.Delete(ctx, "movies:all")
	if err != nil {
		log.Printf("WARNING: Failed to delete all movies cache: %v", err)
	}

	// Clear recommendation caches (they might be affected)
	err = cmc.redisService.DeletePattern(ctx, "recommendations:user:*")
	if err != nil {
		log.Printf("WARNING: Failed to delete recommendation caches: %v", err)
	}

	log.Printf("SUCCESS: Invalidated cache entries for movie %s", movieID)
	return nil
}

// GetCacheStats returns cache statistics
func (cmc *CachedMovieController) GetCacheStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stats, err := cmc.redisService.GetStats(ctx)
		if err != nil {
			log.Printf("ERROR: Failed to get Redis stats: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cache stats"})
			return
		}

		log.Println("INFO: Retrieved cache statistics")
		c.JSON(http.StatusOK, gin.H{
			"cache_stats": stats,
			"timestamp":   time.Now(),
		})
	}
}

// ClearCache clears all movie-related cache entries
func (cmc *CachedMovieController) ClearCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Clear all movie-related cache
		patterns := []string{
			"movies:*",
			"movie:*",
			"recommendations:*",
		}

		for _, pattern := range patterns {
			err := cmc.redisService.DeletePattern(ctx, pattern)
			if err != nil {
				log.Printf("WARNING: Failed to clear cache pattern %s: %v", pattern, err)
			} else {
				log.Printf("SUCCESS: Cleared cache pattern: %s", pattern)
			}
		}

		log.Println("SUCCESS: Cleared all movie-related cache entries")
		c.JSON(http.StatusOK, gin.H{
			"message":   "Cache cleared successfully",
			"timestamp": time.Now(),
		})
	}
}
