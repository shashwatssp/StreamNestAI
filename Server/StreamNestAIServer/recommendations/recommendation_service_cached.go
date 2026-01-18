package recommendations

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/cache"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// CachedRecommendationService handles ML-powered recommendations with Redis caching
type CachedRecommendationService struct {
	client       *mongo.Client
	db           *mongo.Database
	redisService *cache.RedisService
}

// NewCachedRecommendationService creates a new cached recommendation service
func NewCachedRecommendationService(client *mongo.Client, redisService *cache.RedisService) *CachedRecommendationService {
	return &CachedRecommendationService{
		client:       client,
		db:           client.Database("streamnestai"),
		redisService: redisService,
	}
}

// GetRecommendationsWithCache returns personalized recommendations with cache-first strategy
func (crs *CachedRecommendationService) GetRecommendationsWithCache(userID string, limit int) ([]Recommendation, error) {
	log.Printf("INFO: GetRecommendationsWithCache called for user: %s, limit: %d", userID, limit)
	startTime := time.Now()

	// Generate cache key
	cacheKey := fmt.Sprintf("recommendations:user:%s:limit:%d", userID, limit)
	log.Printf("DEBUG: Generated cache key: %s", cacheKey)

	// Try to get from Redis first
	var recommendations []Recommendation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := crs.redisService.Get(ctx, cacheKey, &recommendations)
	if err == nil {
		// Cache hit
		elapsedTime := time.Since(startTime)
		log.Printf("CACHE HIT: Retrieved %d recommendations for user %s from Redis in %v", len(recommendations), userID, elapsedTime)
		return recommendations, nil
	}

	log.Printf("CACHE MISS: %s - falling back to MongoDB generation", err.Error())

	// Cache miss - generate new recommendations
	recommendations, err = crs.generateRecommendationsWithCache(userID, limit)
	if err != nil {
		log.Printf("ERROR: Failed to generate recommendations: %v", err)
		return nil, err
	}

	elapsedTime := time.Since(startTime)
	log.Printf("MONGODB SUCCESS: Generated %d recommendations for user %s in %v", len(recommendations), userID, elapsedTime)

	// Cache the recommendations for future requests
	if len(recommendations) > 0 {
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cacheCancel()

		// Cache recommendations for 10 minutes (shorter TTL as they change more frequently)
		cacheTTL := 10 * time.Minute
		err = crs.redisService.Set(cacheCtx, cacheKey, recommendations, cacheTTL)
		if err != nil {
			log.Printf("WARNING: Failed to cache recommendations for user %s in Redis: %v", userID, err)
		} else {
			log.Printf("SUCCESS: Cached %d recommendations for user %s in Redis for %v", len(recommendations), userID, cacheTTL)
		}
	}

	return recommendations, nil
}

// GetTrendingContentWithCache returns trending content with cache-first strategy
func (crs *CachedRecommendationService) GetTrendingContentWithCache(limit int, timeWindow string) ([]Recommendation, error) {
	log.Printf("INFO: GetTrendingContentWithCache called with limit: %d, timeWindow: %s", limit, timeWindow)
	startTime := time.Now()

	// Generate cache key
	cacheKey := fmt.Sprintf("trending:limit:%d:window:%s", limit, timeWindow)
	log.Printf("DEBUG: Generated cache key: %s", cacheKey)

	// Try to get from Redis first
	var recommendations []Recommendation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := crs.redisService.Get(ctx, cacheKey, &recommendations)
	if err == nil {
		// Cache hit
		elapsedTime := time.Since(startTime)
		log.Printf("CACHE HIT: Retrieved %d trending recommendations from Redis in %v", len(recommendations), elapsedTime)
		return recommendations, nil
	}

	log.Printf("CACHE MISS: %s - falling back to MongoDB generation", err.Error())

	// Cache miss - generate trending content
	recommendations, err = crs.generateTrendingContent(limit, timeWindow)
	if err != nil {
		log.Printf("ERROR: Failed to generate trending content: %v", err)
		return nil, err
	}

	elapsedTime := time.Since(startTime)
	log.Printf("MONGODB SUCCESS: Generated %d trending recommendations in %v", len(recommendations), elapsedTime)

	// Cache the trending content for future requests
	if len(recommendations) > 0 {
		cacheCtx, cacheCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cacheCancel()

		// Cache trending content for 5 minutes (very dynamic)
		cacheTTL := 5 * time.Minute
		err = crs.redisService.Set(cacheCtx, cacheKey, recommendations, cacheTTL)
		if err != nil {
			log.Printf("WARNING: Failed to cache trending content in Redis: %v", err)
		} else {
			log.Printf("SUCCESS: Cached %d trending recommendations in Redis for %v", len(recommendations), cacheTTL)
		}
	}

	return recommendations, nil
}

// generateRecommendationsWithCache generates recommendations using hybrid approach with logging
func (crs *CachedRecommendationService) generateRecommendationsWithCache(userID string, limit int) ([]Recommendation, error) {
	log.Printf("INFO: Generating recommendations for user %s with limit %d", userID, limit)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user preferences
	log.Printf("DEBUG: Fetching user preferences for user %s", userID)
	prefs, err := crs.getUserPreferencesWithCache(ctx, userID)
	if err != nil {
		log.Printf("WARNING: Could not get user preferences: %v", err)
		prefs = &UserPreferences{UserID: userID}
	} else {
		log.Printf("DEBUG: Found %d preferred genres for user %s", len(prefs.PreferredGenres), userID)
	}

	// Get collaborative filtering recommendations
	log.Printf("DEBUG: Generating collaborative filtering recommendations")
	collabRecs, err := crs.getCollaborativeRecommendationsWithCache(ctx, userID, limit/2)
	if err != nil {
		log.Printf("WARNING: Collaborative filtering failed: %v", err)
		collabRecs = []Recommendation{}
	} else {
		log.Printf("DEBUG: Generated %d collaborative recommendations", len(collabRecs))
	}

	// Get content-based recommendations
	log.Printf("DEBUG: Generating content-based recommendations")
	contentRecs, err := crs.getContentBasedRecommendationsWithCache(ctx, prefs, limit/2)
	if err != nil {
		log.Printf("WARNING: Content-based filtering failed: %v", err)
		contentRecs = []Recommendation{}
	} else {
		log.Printf("DEBUG: Generated %d content-based recommendations", len(contentRecs))
	}

	// Combine and rank recommendations
	log.Printf("DEBUG: Combining recommendations with 60% collaborative, 40% content weight")
	combined := crs.combineRecommendations(collabRecs, contentRecs, 0.6)

	// Sort by score and limit results
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Score > combined[j].Score
	})

	if len(combined) > limit {
		combined = combined[:limit]
	}

	log.Printf("SUCCESS: Final recommendation count for user %s: %d", userID, len(combined))
	return combined, nil
}

// getUserPreferencesWithCache gets user preferences with caching
func (crs *CachedRecommendationService) getUserPreferencesWithCache(ctx context.Context, userID string) (*UserPreferences, error) {
	cacheKey := fmt.Sprintf("user_preferences:%s", userID)

	var prefs UserPreferences
	err := crs.redisService.Get(ctx, cacheKey, &prefs)
	if err == nil {
		log.Printf("CACHE HIT: Retrieved user preferences for %s from Redis", userID)
		return &prefs, nil
	}

	log.Printf("CACHE MISS: User preferences for %s - fetching from MongoDB", userID)

	coll := crs.db.Collection("user_preferences")
	err = coll.FindOne(ctx, bson.M{"user_id": userID}).Decode(&prefs)
	if err != nil {
		return nil, err
	}

	// Cache user preferences for 30 minutes
	cacheTTL := 30 * time.Minute
	err = crs.redisService.Set(ctx, cacheKey, prefs, cacheTTL)
	if err != nil {
		log.Printf("WARNING: Failed to cache user preferences for %s: %v", userID, err)
	} else {
		log.Printf("SUCCESS: Cached user preferences for %s for %v", userID, cacheTTL)
	}

	return &prefs, nil
}

// getCollaborativeRecommendationsWithCache generates collaborative recommendations with caching
func (crs *CachedRecommendationService) getCollaborativeRecommendationsWithCache(ctx context.Context, userID string, limit int) ([]Recommendation, error) {
	cacheKey := fmt.Sprintf("collab_recs:%s:limit:%d", userID, limit)

	var recommendations []Recommendation
	err := crs.redisService.Get(ctx, cacheKey, &recommendations)
	if err == nil {
		log.Printf("CACHE HIT: Retrieved collaborative recommendations for %s from Redis", userID)
		return recommendations, nil
	}

	log.Printf("CACHE MISS: Collaborative recommendations for %s - generating from MongoDB", userID)

	// Find similar users based on viewing history
	similarUsers, err := crs.findSimilarUsers(ctx, userID, 50)
	if err != nil {
		return nil, err
	}
	log.Printf("DEBUG: Found %d similar users for %s", len(similarUsers), userID)

	// Get movies liked by similar users that the current user hasn't watched
	recommendations = make([]Recommendation, 0)
	userWatched, err := crs.getUserWatchedMovies(ctx, userID)
	if err != nil {
		return nil, err
	}
	log.Printf("DEBUG: User %s has watched %d movies", userID, len(userWatched))

	for _, similarUser := range similarUsers {
		userMovies, err := crs.getUserWatchedMovies(ctx, similarUser)
		if err != nil {
			continue
		}

		for _, movieID := range userMovies {
			if !contains(userWatched, movieID) {
				movie, err := crs.getMovieDetails(ctx, movieID)
				if err != nil {
					continue
				}

				score := crs.calculateCollaborativeScore(userID, similarUser, movieID)
				title, _ := movie["title"].(string)
				genre, _ := movie["genre"].(string)
				year, _ := movie["year"].(int32)
				rating, _ := movie["rating"].(float64)
				poster, _ := movie["poster"].(string)

				recommendations = append(recommendations, Recommendation{
					MovieID: movieID,
					Title:   title,
					Score:   score,
					Reason:  "Users with similar taste also liked this",
					Genre:   genre,
					Year:    int(year),
					Rating:  rating,
					Poster:  poster,
				})
			}
		}
	}

	// Sort and limit
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	// Cache collaborative recommendations for 15 minutes
	cacheTTL := 15 * time.Minute
	err = crs.redisService.Set(ctx, cacheKey, recommendations, cacheTTL)
	if err != nil {
		log.Printf("WARNING: Failed to cache collaborative recommendations for %s: %v", userID, err)
	} else {
		log.Printf("SUCCESS: Cached %d collaborative recommendations for %s for %v", len(recommendations), userID, cacheTTL)
	}

	return recommendations, nil
}

// getContentBasedRecommendationsWithCache generates content-based recommendations with caching
func (crs *CachedRecommendationService) getContentBasedRecommendationsWithCache(ctx context.Context, prefs *UserPreferences, limit int) ([]Recommendation, error) {
	cacheKey := fmt.Sprintf("content_recs:%s:limit:%d", prefs.UserID, limit)

	var recommendations []Recommendation
	err := crs.redisService.Get(ctx, cacheKey, &recommendations)
	if err == nil {
		log.Printf("CACHE HIT: Retrieved content-based recommendations for %s from Redis", prefs.UserID)
		return recommendations, nil
	}

	log.Printf("CACHE MISS: Content-based recommendations for %s - generating from MongoDB", prefs.UserID)

	recommendations = make([]Recommendation, 0)

	// Get movies based on preferred genres
	// Guard against division by zero when PreferredGenres is empty
	if len(prefs.PreferredGenres) == 0 {
		log.Printf("DEBUG: User %s has no preferred genres, returning empty recommendations", prefs.UserID)
		return recommendations, nil
	}

	perGenreLimit := limit / len(prefs.PreferredGenres)
	if perGenreLimit < 1 {
		perGenreLimit = 1 // Ensure at least 1 movie per genre
	}

	for _, genre := range prefs.PreferredGenres {
		movies, err := crs.getMoviesByGenre(ctx, genre, perGenreLimit)
		if err != nil {
			continue
		}

		for _, movie := range movies {
			score := crs.calculateContentScore(prefs, movie)
			movieID, _ := movie["_id"].(string)
			title, _ := movie["title"].(string)
			movieGenre, _ := movie["genre"].(string)
			year, _ := movie["year"].(int32)
			rating, _ := movie["rating"].(float64)
			poster, _ := movie["poster"].(string)

			recommendations = append(recommendations, Recommendation{
				MovieID: movieID,
				Title:   title,
				Score:   score,
				Reason:  fmt.Sprintf("Based on your interest in %s", genre),
				Genre:   movieGenre,
				Year:    int(year),
				Rating:  rating,
				Poster:  poster,
			})
		}
	}

	// Sort and limit
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	// Cache content-based recommendations for 20 minutes
	cacheTTL := 20 * time.Minute
	err = crs.redisService.Set(ctx, cacheKey, recommendations, cacheTTL)
	if err != nil {
		log.Printf("WARNING: Failed to cache content-based recommendations for %s: %v", prefs.UserID, err)
	} else {
		log.Printf("SUCCESS: Cached %d content-based recommendations for %s for %v", len(recommendations), prefs.UserID, cacheTTL)
	}

	return recommendations, nil
}

// generateTrendingContent generates trending content based on recent activity
func (crs *CachedRecommendationService) generateTrendingContent(limit int, timeWindow string) ([]Recommendation, error) {
	log.Printf("INFO: Generating trending content with limit %d, time window %s", limit, timeWindow)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get recent movie watching events
	analyticsColl := crs.db.Collection("analytics_events")

	// Calculate time threshold based on timeWindow
	var timeThreshold time.Time
	switch timeWindow {
	case "24h":
		timeThreshold = time.Now().Add(-24 * time.Hour)
	case "7d":
		timeThreshold = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		timeThreshold = time.Now().Add(-30 * 24 * time.Hour)
	default:
		timeThreshold = time.Now().Add(-24 * time.Hour)
	}

	// Aggregate movie views in the time window
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"event":     "movie_watched",
				"timestamp": bson.M{"$gte": timeThreshold},
			},
		},
		{
			"$group": bson.M{
				"_id":         "$metadata.movie_id",
				"watch_count": bson.M{"$sum": 1},
				"avg_rating":  bson.M{"$avg": "$metadata.rating"},
			},
		},
		{
			"$sort": bson.M{"watch_count": -1, "avg_rating": -1},
		},
		{
			"$limit": int64(limit),
		},
	}

	cursor, err := analyticsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trendingMovies []bson.M
	if err = cursor.All(ctx, &trendingMovies); err != nil {
		return nil, err
	}

	log.Printf("DEBUG: Found %d trending movies in time window %s", len(trendingMovies), timeWindow)

	// Convert to recommendations
	recommendations := make([]Recommendation, 0)
	for _, trending := range trendingMovies {
		movieID := trending["_id"].(string)
		watchCount := trending["watch_count"].(int32)
		avgRating := trending["avg_rating"].(float64)

		movie, err := crs.getMovieDetails(ctx, movieID)
		if err != nil {
			continue
		}

		title, _ := movie["title"].(string)
		genre, _ := movie["genre"].(string)
		year, _ := movie["year"].(int32)
		rating, _ := movie["rating"].(float64)
		poster, _ := movie["poster"].(string)

		// Calculate trending score based on watch count and rating
		score := float64(watchCount)*0.7 + avgRating*0.3

		recommendations = append(recommendations, Recommendation{
			MovieID: movieID,
			Title:   title,
			Score:   math.Min(score, 1.0),
			Reason:  fmt.Sprintf("Trending with %d views in the last %s", watchCount, timeWindow),
			Genre:   genre,
			Year:    int(year),
			Rating:  rating,
			Poster:  poster,
		})
	}

	log.Printf("SUCCESS: Generated %d trending recommendations", len(recommendations))
	return recommendations, nil
}

// Helper methods (reuse from original service)
func (crs *CachedRecommendationService) combineRecommendations(collab, content []Recommendation, collabWeight float64) []Recommendation {
	combined := make([]Recommendation, 0)
	seen := make(map[string]bool)

	// Add collaborative recommendations with weight
	for _, rec := range collab {
		if !seen[rec.MovieID] {
			rec.Score = rec.Score * collabWeight
			combined = append(combined, rec)
			seen[rec.MovieID] = true
		}
	}

	// Add content-based recommendations with weight
	contentWeight := 1.0 - collabWeight
	for _, rec := range content {
		if !seen[rec.MovieID] {
			rec.Score = rec.Score * contentWeight
			combined = append(combined, rec)
			seen[rec.MovieID] = true
		}
	}

	return combined
}

func (crs *CachedRecommendationService) getUserWatchedMovies(ctx context.Context, userID string) ([]string, error) {
	coll := crs.db.Collection("analytics_events")

	cursor, err := coll.Find(ctx, bson.M{
		"user_id": userID,
		"event":   "movie_watched",
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []bson.M
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	movies := make([]string, 0, len(events))
	for _, event := range events {
		if movieID, ok := event["metadata"].(bson.M)["movie_id"].(string); ok {
			movies = append(movies, movieID)
		}
	}

	return movies, nil
}

func (crs *CachedRecommendationService) findSimilarUsers(ctx context.Context, userID string, limit int) ([]string, error) {
	similarityColl := crs.db.Collection("similarity_matrix")

	cursor, err := similarityColl.Find(ctx, bson.M{
		"$or": []bson.M{
			{"user_id1": userID},
			{"user_id2": userID},
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var similarities []SimilarityMatrix
	if err = cursor.All(ctx, &similarities); err != nil {
		return nil, err
	}

	// Sort by similarity and return top users
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})

	users := make([]string, 0, limit)
	for _, sim := range similarities {
		var similarUser string
		if sim.UserID1 == userID {
			similarUser = sim.UserID2
		} else {
			similarUser = sim.UserID1
		}

		users = append(users, similarUser)
		if len(users) >= limit {
			break
		}
	}

	return users, nil
}

func (crs *CachedRecommendationService) getMovieDetails(ctx context.Context, movieID string) (bson.M, error) {
	coll := crs.db.Collection("movies")
	var movie bson.M

	err := coll.FindOne(ctx, bson.M{"_id": movieID}).Decode(&movie)
	return movie, err
}

func (crs *CachedRecommendationService) getMoviesByGenre(ctx context.Context, genre string, limit int) ([]bson.M, error) {
	coll := crs.db.Collection("movies")

	cursor, err := coll.Find(ctx, bson.M{
		"genre": genre,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var movies []bson.M
	if err = cursor.All(ctx, &movies); err != nil {
		return nil, err
	}

	// Apply limit
	if len(movies) > limit {
		movies = movies[:limit]
	}

	return movies, nil
}

func (crs *CachedRecommendationService) calculateCollaborativeScore(userID, similarUser, movieID string) float64 {
	// Simplified scoring based on user similarity
	return 0.8 // Placeholder
}

func (crs *CachedRecommendationService) calculateContentScore(prefs *UserPreferences, movie bson.M) float64 {
	score := 0.0

	// Genre matching
	if genre, ok := movie["genre"].(string); ok {
		for _, prefGenre := range prefs.PreferredGenres {
			if genre == prefGenre {
				score += 0.3
			}
		}
	}

	// Rating consideration
	if rating, ok := movie["rating"].(float64); ok {
		score += rating * 0.1
	}

	return math.Min(score, 1.0)
}

// InvalidateUserCache clears recommendation cache for a specific user
func (crs *CachedRecommendationService) InvalidateUserCache(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	patterns := []string{
		fmt.Sprintf("recommendations:user:%s:*", userID),
		fmt.Sprintf("collab_recs:%s:*", userID),
		fmt.Sprintf("content_recs:%s:*", userID),
		fmt.Sprintf("user_preferences:%s", userID),
	}

	for _, pattern := range patterns {
		err := crs.redisService.DeletePattern(ctx, pattern)
		if err != nil {
			log.Printf("WARNING: Failed to delete cache pattern %s: %v", pattern, err)
		} else {
			log.Printf("SUCCESS: Cleared cache pattern: %s", pattern)
		}
	}

	log.Printf("SUCCESS: Invalidated recommendation cache for user %s", userID)
	return nil
}

// GetCacheStats returns recommendation cache statistics
func (crs *CachedRecommendationService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	return crs.redisService.GetStats(ctx)
}

