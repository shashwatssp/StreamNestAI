package recommendations

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RecommendationService handles ML-powered recommendations
type RecommendationService struct {
	client    *mongo.Client
	db        *mongo.Database
	redisAddr string
}

// Recommendation represents a movie recommendation
type Recommendation struct {
	MovieID string  `bson:"movie_id" json:"movie_id"`
	Title   string  `bson:"title" json:"title"`
	Score   float64 `bson:"score" json:"score"`
	Reason  string  `bson:"reason" json:"reason"`
	Genre   string  `bson:"genre" json:"genre"`
	Year    int     `bson:"year" json:"year"`
	Rating  float64 `bson:"rating" json:"rating"`
	Poster  string  `bson:"poster" json:"poster"`
}

// UserPreferences represents user preference data
type UserPreferences struct {
	UserID             string    `bson:"user_id" json:"user_id"`
	PreferredGenres    []string  `bson:"preferred_genres" json:"preferred_genres"`
	PreferredActors    []string  `bson:"preferred_actors" json:"preferred_actors"`
	PreferredDirectors []string  `bson:"preferred_directors" json:"preferred_directors"`
	AverageRating      float64   `bson:"average_rating" json:"average_rating"`
	WatchTimeWeight    float64   `bson:"watch_time_weight" json:"watch_time_weight"`
	LastUpdated        time.Time `bson:"last_updated" json:"last_updated"`
}

// SimilarityMatrix stores similarity scores between users
type SimilarityMatrix struct {
	UserID1    string  `bson:"user_id1" json:"user_id1"`
	UserID2    string  `bson:"user_id2" json:"user_id2"`
	Similarity float64 `bson:"similarity" json:"similarity"`
}

// RecommendationModel represents the ML model configuration
type RecommendationModel struct {
	Name        string                 `bson:"name" json:"name"`
	Version     string                 `bson:"version" json:"version"`
	Type        string                 `bson:"type" json:"type"` // collaborative, content, hybrid
	Parameters  map[string]interface{} `bson:"parameters" json:"parameters"`
	Accuracy    float64                `bson:"accuracy" json:"accuracy"`
	LastTrained time.Time              `bson:"last_trained" json:"last_trained"`
	IsActive    bool                   `bson:"is_active" json:"is_active"`
}

// NewRecommendationService creates a new recommendation service
func NewRecommendationService(client *mongo.Client, redisAddr string) *RecommendationService {
	return &RecommendationService{
		client:    client,
		db:        client.Database("streamnestai"),
		redisAddr: redisAddr,
	}
}

// Initialize sets up the recommendation service
func (rs *RecommendationService) Initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create indexes for recommendation collections
	collections := []string{
		"user_preferences",
		"recommendations",
		"similarity_matrix",
		"recommendation_models",
	}

	for _, collName := range collections {
		coll := rs.db.Collection(collName)

		// Create user_id index
		_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{"user_id", 1}},
		})
		if err != nil {
			log.Printf("Warning: Failed to create user_id index for %s: %v", collName, err)
		}

		// Create score index for recommendations
		if collName == "recommendations" {
			_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: bson.D{{"score", -1}},
			})
			if err != nil {
				log.Printf("Warning: Failed to create score index for %s: %v", collName, err)
			}
		}
	}

	log.Println("Recommendation service initialized successfully")
	return nil
}

// GetRecommendations returns personalized recommendations for a user
func (rs *RecommendationService) GetRecommendations(userID string, limit int) ([]Recommendation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to get cached recommendations first
	cached := rs.getCachedRecommendations(userID, limit)
	if len(cached) > 0 {
		return cached, nil
	}

	// Generate new recommendations
	recommendations, err := rs.generateRecommendations(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	// Cache the recommendations
	rs.cacheRecommendations(userID, recommendations)

	return recommendations, nil
}

// generateRecommendations generates recommendations using hybrid approach
func (rs *RecommendationService) generateRecommendations(ctx context.Context, userID string, limit int) ([]Recommendation, error) {
	// Get user preferences
	prefs, err := rs.getUserPreferences(ctx, userID)
	if err != nil {
		log.Printf("Warning: Could not get user preferences: %v", err)
		prefs = &UserPreferences{UserID: userID}
	}

	// Get collaborative filtering recommendations
	collabRecs, err := rs.getCollaborativeRecommendations(ctx, userID, limit/2)
	if err != nil {
		log.Printf("Warning: Collaborative filtering failed: %v", err)
		collabRecs = []Recommendation{}
	}

	// Get content-based recommendations
	contentRecs, err := rs.getContentBasedRecommendations(ctx, prefs, limit/2)
	if err != nil {
		log.Printf("Warning: Content-based filtering failed: %v", err)
		contentRecs = []Recommendation{}
	}

	// Combine and rank recommendations
	combined := rs.combineRecommendations(collabRecs, contentRecs, 0.6) // 60% collaborative, 40% content

	// Sort by score and limit results
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Score > combined[j].Score
	})

	if len(combined) > limit {
		combined = combined[:limit]
	}

	return combined, nil
}

// getCollaborativeRecommendations generates recommendations using collaborative filtering
func (rs *RecommendationService) getCollaborativeRecommendations(ctx context.Context, userID string, limit int) ([]Recommendation, error) {
	// Find similar users based on viewing history
	similarUsers, err := rs.findSimilarUsers(ctx, userID, 50)
	if err != nil {
		return nil, err
	}

	// Get movies liked by similar users that the current user hasn't watched
	recommendations := make([]Recommendation, 0)
	userWatched, err := rs.getUserWatchedMovies(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, similarUser := range similarUsers {
		userMovies, err := rs.getUserWatchedMovies(ctx, similarUser)
		if err != nil {
			continue
		}

		for _, movieID := range userMovies {
			if !contains(userWatched, movieID) {
				movie, err := rs.getMovieDetails(ctx, movieID)
				if err != nil {
					continue
				}

				score := rs.calculateCollaborativeScore(userID, similarUser, movieID)
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

	return recommendations, nil
}

// getContentBasedRecommendations generates recommendations based on content
func (rs *RecommendationService) getContentBasedRecommendations(ctx context.Context, prefs *UserPreferences, limit int) ([]Recommendation, error) {
	recommendations := make([]Recommendation, 0)

	// Get movies based on preferred genres
	for _, genre := range prefs.PreferredGenres {
		movies, err := rs.getMoviesByGenre(ctx, genre, limit/len(prefs.PreferredGenres))
		if err != nil {
			continue
		}

		for _, movie := range movies {
			score := rs.calculateContentScore(prefs, movie)
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

	return recommendations, nil
}

// combineRecommendations combines recommendations from different sources
func (rs *RecommendationService) combineRecommendations(collab, content []Recommendation, collabWeight float64) []Recommendation {
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

// updateUserPreferences updates user preference data
func (rs *RecommendationService) UpdateUserPreferences(userID string, event map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := rs.db.Collection("user_preferences")

	update := bson.M{
		"$set": bson.M{
			"last_updated": time.Now(),
		},
		"$setOnInsert": bson.M{
			"user_id":             userID,
			"preferred_genres":    []string{},
			"preferred_actors":    []string{},
			"preferred_directors": []string{},
			"average_rating":      0.0,
			"watch_time_weight":   1.0,
		},
	}

	// Update based on event data
	if genre, ok := event["genre"].(string); ok {
		update["$addToSet"] = bson.M{"preferred_genres": genre}
	}

	if rating, ok := event["rating"].(float64); ok {
		update["$avg"] = bson.M{"average_rating": rating}
	}

	_, err := coll.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		update,
	)

	return err
}

// TrainModels trains the recommendation models
func (rs *RecommendationService) TrainModels() {
	log.Println("Starting recommendation model training...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Train collaborative filtering model
	if err := rs.trainCollaborativeModel(ctx); err != nil {
		log.Printf("Error training collaborative model: %v", err)
	}

	// Train content-based model
	if err := rs.trainContentModel(ctx); err != nil {
		log.Printf("Error training content model: %v", err)
	}

	log.Println("Recommendation model training completed")
}

// trainCollaborativeModel trains the collaborative filtering model
func (rs *RecommendationService) trainCollaborativeModel(ctx context.Context) error {
	// Get all user-movie interactions
	interactionsColl := rs.db.Collection("analytics_events")

	// Build user-item matrix
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"event": "movie_watched",
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"user_id":  "$user_id",
					"movie_id": "$metadata.movie_id",
				},
				"rating":     bson.M{"$avg": "$metadata.rating"},
				"watch_time": bson.M{"$sum": "$metadata.watch_time"},
			},
		},
	}

	cursor, err := interactionsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var interactions []bson.M
	if err = cursor.All(ctx, &interactions); err != nil {
		return err
	}

	// Calculate similarity matrix
	similarityColl := rs.db.Collection("similarity_matrix")

	// Clear existing matrix
	_, err = similarityColl.DeleteMany(ctx, bson.M{})
	if err != nil {
		return err
	}

	// Build user-movie rating matrix
	userMovieRatings := make(map[string]map[string]float64)
	for _, interaction := range interactions {
		userID := interaction["_id"].(bson.M)["user_id"].(string)
		movieID := interaction["_id"].(bson.M)["movie_id"].(string)
		rating := interaction["rating"].(float64)

		if userMovieRatings[userID] == nil {
			userMovieRatings[userID] = make(map[string]float64)
		}
		userMovieRatings[userID][movieID] = rating
	}

	// Calculate cosine similarity between users
	users := make([]string, 0, len(userMovieRatings))
	for userID := range userMovieRatings {
		users = append(users, userID)
	}

	for i, user1 := range users {
		for j, user2 := range users {
			if i >= j {
				continue
			}

			similarity := rs.calculateCosineSimilarity(userMovieRatings[user1], userMovieRatings[user2])
			if similarity > 0.1 { // Only store significant similarities
				_, err := similarityColl.InsertOne(ctx, SimilarityMatrix{
					UserID1:    user1,
					UserID2:    user2,
					Similarity: similarity,
				})
				if err != nil {
					log.Printf("Error inserting similarity: %v", err)
				}
			}
		}
	}

	log.Printf("Trained collaborative model with %d users and %d interactions", len(users), len(interactions))
	return nil
}

// trainContentModel trains the content-based model
func (rs *RecommendationService) trainContentModel(ctx context.Context) error {
	// Get all movies and extract features
	moviesColl := rs.db.Collection("movies")

	cursor, err := moviesColl.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var movies []bson.M
	if err = cursor.All(ctx, &movies); err != nil {
		return err
	}

	// Extract content features and calculate similarities
	log.Printf("Training content model with %d movies", len(movies))

	// This would typically involve TF-IDF, word embeddings, etc.
	// For simplicity, we'll use genre-based similarity

	return nil
}

// Helper functions

func (rs *RecommendationService) getUserPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	coll := rs.db.Collection("user_preferences")
	var prefs UserPreferences

	err := coll.FindOne(ctx, bson.M{"user_id": userID}).Decode(&prefs)
	if err != nil {
		return nil, err
	}

	return &prefs, nil
}

func (rs *RecommendationService) getUserWatchedMovies(ctx context.Context, userID string) ([]string, error) {
	coll := rs.db.Collection("analytics_events")

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

func (rs *RecommendationService) findSimilarUsers(ctx context.Context, userID string, limit int) ([]string, error) {
	similarityColl := rs.db.Collection("similarity_matrix")

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

func (rs *RecommendationService) getMovieDetails(ctx context.Context, movieID string) (bson.M, error) {
	coll := rs.db.Collection("movies")
	var movie bson.M

	err := coll.FindOne(ctx, bson.M{"_id": movieID}).Decode(&movie)
	return movie, err
}

func (rs *RecommendationService) getMoviesByGenre(ctx context.Context, genre string, limit int) ([]bson.M, error) {
	coll := rs.db.Collection("movies")

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

func (rs *RecommendationService) calculateCollaborativeScore(userID, similarUser, movieID string) float64 {
	// Simplified scoring based on user similarity
	return 0.8 // Placeholder
}

func (rs *RecommendationService) calculateContentScore(prefs *UserPreferences, movie bson.M) float64 {
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

func (rs *RecommendationService) calculateCosineSimilarity(user1, user2 map[string]float64) float64 {
	// Calculate cosine similarity between two users
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for movie, rating1 := range user1 {
		if rating2, exists := user2[movie]; exists {
			dotProduct += rating1 * rating2
		}
		norm1 += rating1 * rating1
	}

	for _, rating2 := range user2 {
		norm2 += rating2 * rating2
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

func (rs *RecommendationService) getCachedRecommendations(userID string, limit int) []Recommendation {
	// Redis caching implementation would go here
	return []Recommendation{}
}

func (rs *RecommendationService) cacheRecommendations(userID string, recommendations []Recommendation) {
	// Redis caching implementation would go here
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
