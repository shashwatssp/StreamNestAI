package recommendation

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RecommendationType represents different types of recommendations
type RecommendationType string

const (
	TypeCollaborative RecommendationType = "collaborative"
	TypeContentBased  RecommendationType = "content_based"
	TypeHybrid        RecommendationType = "hybrid"
	TypeTrending      RecommendationType = "trending"
	TypePersonalized  RecommendationType = "personalized"
)

// Content represents a content item (movie, series, etc.)
type Content struct {
	ID          string                 `json:"id" bson:"_id"`
	Title       string                 `json:"title" bson:"title"`
	Genre       []string               `json:"genre" bson:"genre"`
	Tags        []string               `json:"tags" bson:"tags"`
	Description string                 `json:"description" bson:"description"`
	Duration    int                    `json:"duration" bson:"duration"` // in minutes
	ReleaseYear int                    `json:"release_year" bson:"release_year"`
	Rating      float64                `json:"rating" bson:"rating"`
	Language    string                 `json:"language" bson:"language"`
	Cast        []string               `json:"cast" bson:"cast"`
	Director    []string               `json:"director" bson:"director"`
	Features    map[string]float64     `json:"features" bson:"features"`
	Metadata    map[string]interface{} `json:"metadata" bson:"metadata"`
}

// UserInteraction represents a user's interaction with content
type UserInteraction struct {
	UserID      string    `json:"user_id" bson:"user_id"`
	ContentID   string    `json:"content_id" bson:"content_id"`
	Interaction string    `json:"interaction" bson:"interaction"` // view, like, share, rating, etc.
	Rating      float64   `json:"rating" bson:"rating,omitempty"`
	Duration    int       `json:"duration" bson:"duration,omitempty"` // watch time in seconds
	Timestamp   time.Time `json:"timestamp" bson:"timestamp"`
	Weight      float64   `json:"weight" bson:"weight"`
}

// UserProfile represents a user's preference profile
type UserProfile struct {
	UserID           string                 `json:"user_id" bson:"user_id"`
	GenrePreferences map[string]float64     `json:"genre_preferences" bson:"genre_preferences"`
	TagPreferences   map[string]float64     `json:"tag_preferences" bson:"tag_preferences"`
	LanguagePref     string                 `json:"language_pref" bson:"language_pref"`
	AvgRating        float64                `json:"avg_rating" bson:"avg_rating"`
	ActivityLevel    float64                `json:"activity_level" bson:"activity_level"`
	Features         map[string]float64     `json:"features" bson:"features"`
	LastUpdated      time.Time              `json:"last_updated" bson:"last_updated"`
	Metadata         map[string]interface{} `json:"metadata" bson:"metadata"`
}

// Recommendation represents a recommended content item
type Recommendation struct {
	ContentID   string                 `json:"content_id"`
	Content     *Content               `json:"content,omitempty"`
	Score       float64                `json:"score"`
	Reason      string                 `json:"reason"`
	Type        RecommendationType     `json:"type"`
	Confidence  float64                `json:"confidence"`
	Explanation map[string]interface{} `json:"explanation,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// RecommendationRequest represents a request for recommendations
type RecommendationRequest struct {
	UserID         string                 `json:"user_id"`
	Type           RecommendationType     `json:"type"`
	Count          int                    `json:"count"`
	ExcludeWatched bool                   `json:"exclude_watched"`
	IncludeGenres  []string               `json:"include_genres,omitempty"`
	ExcludeGenres  []string               `json:"exclude_genres,omitempty"`
	MinRating      float64                `json:"min_rating,omitempty"`
	MaxDuration    int                    `json:"max_duration,omitempty"`
	Language       string                 `json:"language,omitempty"`
	Context        map[string]interface{} `json:"context,omitempty"`
}

// RecommendationConfig holds configuration for the recommendation engine
type RecommendationConfig struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	MongoDB       *mongo.Database

	// Algorithm weights
	CollaborativeWeight float64
	ContentWeight       float64
	TrendingWeight      float64

	// Model parameters
	MinInteractions     int
	SimilarityThreshold float64
	UpdateInterval      time.Duration
	CacheTTL            time.Duration

	// Feature weights
	GenreWeight      float64
	TagWeight        float64
	RatingWeight     float64
	RecencyWeight    float64
	PopularityWeight float64
}

// RecommendationEngine handles content recommendations
type RecommendationEngine struct {
	config      RecommendationConfig
	redisClient *redis.Client
	mongoDB     *mongo.Database
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine(config RecommendationConfig) (*RecommendationEngine, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	engine := &RecommendationEngine{
		config:      config,
		redisClient: redisClient,
		mongoDB:     config.MongoDB,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start background processes
	go engine.updateModels()
	go engine.precomputeRecommendations()

	return engine, nil
}

// GetRecommendations generates recommendations for a user
func (re *RecommendationEngine) GetRecommendations(req RecommendationRequest) ([]Recommendation, error) {
	// Validate request
	if req.Count <= 0 {
		req.Count = 10
	}
	if req.Count > 100 {
		req.Count = 100
	}

	var recommendations []Recommendation

	switch req.Type {
	case TypeCollaborative:
		recs, err := re.getCollaborativeRecommendations(req)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, recs...)

	case TypeContentBased:
		recs, err := re.getContentBasedRecommendations(req)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, recs...)

	case TypeTrending:
		recs, err := re.getTrendingRecommendations(req)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, recs...)

	case TypeHybrid:
		recs, err := re.getHybridRecommendations(req)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, recs...)

	case TypePersonalized:
		recs, err := re.getPersonalizedRecommendations(req)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, recs...)

	default:
		// Default to hybrid
		req.Type = TypeHybrid
		recs, err := re.getHybridRecommendations(req)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, recs...)
	}

	// Apply filters
	recommendations = re.applyFilters(recommendations, req)

	// Sort by score and limit
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > req.Count {
		recommendations = recommendations[:req.Count]
	}

	// Enrich with content details
	recommendations = re.enrichWithContent(recommendations)

	return recommendations, nil
}

// getCollaborativeRecommendations generates collaborative filtering recommendations
func (re *RecommendationEngine) getCollaborativeRecommendations(req RecommendationRequest) ([]Recommendation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user's interaction history
	interactions, err := re.getUserInteractions(ctx, req.UserID, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to get user interactions: %w", err)
	}

	if len(interactions) < re.config.MinInteractions {
		// Fall back to content-based for new users
		return re.getContentBasedRecommendations(req)
	}

	// Find similar users
	similarUsers, err := re.findSimilarUsers(ctx, req.UserID, interactions, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar users: %w", err)
	}

	// Get content liked by similar users
	contentScores := make(map[string]float64)
	userContent := make(map[string]bool)

	for _, interaction := range interactions {
		userContent[interaction.ContentID] = true
	}

	for _, similarUser := range similarUsers {
		userInteractions, err := re.getUserInteractions(ctx, similarUser.UserID, 100)
		if err != nil {
			continue
		}

		for _, interaction := range userInteractions {
			if !userContent[interaction.ContentID] && interaction.Rating >= 4.0 {
				contentScores[interaction.ContentID] += similarUser.Similarity * interaction.Weight
			}
		}
	}

	// Convert to recommendations
	var recommendations []Recommendation
	for contentID, score := range contentScores {
		if score > re.config.SimilarityThreshold {
			rec := Recommendation{
				ContentID:  contentID,
				Score:      score,
				Reason:     "Users with similar taste also liked this",
				Type:       TypeCollaborative,
				Confidence: math.Min(score, 1.0),
				Timestamp:  time.Now(),
				Explanation: map[string]interface{}{
					"similar_users": len(similarUsers),
					"method":        "user_based_collaborative_filtering",
				},
			}
			recommendations = append(recommendations, rec)
		}
	}

	return recommendations, nil
}

// getContentBasedRecommendations generates content-based recommendations
func (re *RecommendationEngine) getContentBasedRecommendations(req RecommendationRequest) ([]Recommendation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user's profile
	profile, err := re.getUserProfile(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Get user's interaction history
	interactions, err := re.getUserInteractions(ctx, req.UserID, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to get user interactions: %w", err)
	}

	// Get content the user liked
	likedContent := make(map[string]bool)
	for _, interaction := range interactions {
		if interaction.Rating >= 4.0 {
			likedContent[interaction.ContentID] = true
		}
	}

	// Get all content
	allContent, err := re.getAllContent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all content: %w", err)
	}

	// Calculate similarity scores
	var recommendations []Recommendation
	for _, content := range allContent {
		if likedContent[content.ID] {
			continue
		}

		score := re.calculateContentSimilarity(profile, content, interactions)
		if score > 0.1 {
			rec := Recommendation{
				ContentID:  content.ID,
				Score:      score,
				Reason:     "Similar to content you've enjoyed",
				Type:       TypeContentBased,
				Confidence: score,
				Timestamp:  time.Now(),
				Explanation: map[string]interface{}{
					"genre_match":   re.calculateGenreMatch(profile, content),
					"tag_match":     re.calculateTagMatch(profile, content),
					"feature_match": re.calculateFeatureMatch(profile, content),
				},
			}
			recommendations = append(recommendations, rec)
		}
	}

	return recommendations, nil
}

// getTrendingRecommendations generates trending content recommendations
func (re *RecommendationEngine) getTrendingRecommendations(req RecommendationRequest) ([]Recommendation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get trending content from Redis
	trendingKeys, err := re.redisClient.ZRevRangeWithScores(ctx, "trending:content", 0, int64(req.Count*2)-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get trending content: %w", err)
	}

	var recommendations []Recommendation
	for _, item := range trendingKeys {
		contentID := item.Member.(string)
		score := item.Score

		rec := Recommendation{
			ContentID:  contentID,
			Score:      score,
			Reason:     "Trending now",
			Type:       TypeTrending,
			Confidence: math.Min(score/100, 1.0),
			Timestamp:  time.Now(),
			Explanation: map[string]interface{}{
				"trending_score": score,
				"method":         "popularity_based",
			},
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations, nil
}

// getHybridRecommendations generates hybrid recommendations combining multiple approaches
func (re *RecommendationEngine) getHybridRecommendations(req RecommendationRequest) ([]Recommendation, error) {
	// Get recommendations from different approaches
	collaborativeRecs, _ := re.getCollaborativeRecommendations(req)
	contentRecs, _ := re.getContentBasedRecommendations(req)
	trendingRecs, _ := re.getTrendingRecommendations(req)

	// Combine and weight recommendations
	contentScores := make(map[string]float64)
	contentTypes := make(map[string]RecommendationType)
	contentReasons := make(map[string]string)

	// Add collaborative recommendations
	for _, rec := range collaborativeRecs {
		contentScores[rec.ContentID] += rec.Score * re.config.CollaborativeWeight
		if contentTypes[rec.ContentID] == "" {
			contentTypes[rec.ContentID] = TypeCollaborative
			contentReasons[rec.ContentID] = rec.Reason
		}
	}

	// Add content-based recommendations
	for _, rec := range contentRecs {
		contentScores[rec.ContentID] += rec.Score * re.config.ContentWeight
		if contentTypes[rec.ContentID] == "" {
			contentTypes[rec.ContentID] = TypeContentBased
			contentReasons[rec.ContentID] = rec.Reason
		}
	}

	// Add trending recommendations
	for _, rec := range trendingRecs {
		contentScores[rec.ContentID] += rec.Score * re.config.TrendingWeight
		if contentTypes[rec.ContentID] == "" {
			contentTypes[rec.ContentID] = TypeTrending
			contentReasons[rec.ContentID] = rec.Reason
		}
	}

	// Convert to recommendations
	var recommendations []Recommendation
	for contentID, score := range contentScores {
		rec := Recommendation{
			ContentID:  contentID,
			Score:      score,
			Reason:     contentReasons[contentID],
			Type:       TypeHybrid,
			Confidence: math.Min(score, 1.0),
			Timestamp:  time.Now(),
			Explanation: map[string]interface{}{
				"components": []string{string(contentTypes[contentID])},
				"method":     "hybrid_approach",
			},
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations, nil
}

// getPersonalizedRecommendations generates highly personalized recommendations
func (re *RecommendationEngine) getPersonalizedRecommendations(req RecommendationRequest) ([]Recommendation, error) {
	// Get user profile and recent activity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	profile, err := re.getUserProfile(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Adjust weights based on user activity level
	weights := re.calculatePersonalizedWeights(profile)

	// Create modified request with adjusted weights
	modifiedReq := req
	modifiedReq.Type = TypeHybrid

	// Get hybrid recommendations
	recommendations, err := re.getHybridRecommendations(modifiedReq)
	if err != nil {
		return nil, err
	}

	// Apply personalized scoring
	for i := range recommendations {
		recommendations[i].Score = re.applyPersonalizedScoring(recommendations[i], profile, weights)
		recommendations[i].Type = TypePersonalized
		recommendations[i].Reason = "Personalized for you"
		recommendations[i].Explanation["personalization_factors"] = weights
	}

	return recommendations, nil
}

// updateUserProfile updates user's preference profile based on interactions
func (re *RecommendationEngine) updateUserProfile(userID string, interactions []UserInteraction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	profile := &UserProfile{
		UserID:      userID,
		LastUpdated: time.Now(),
	}

	// Calculate genre preferences
	genreCounts := make(map[string]int)
	genreRatings := make(map[string]float64)

	// Calculate tag preferences
	tagCounts := make(map[string]int)
	tagRatings := make(map[string]float64)

	var totalRating float64
	var ratingCount int

	for _, interaction := range interactions {
		if interaction.Rating > 0 {
			totalRating += interaction.Rating
			ratingCount++
		}

		// Get content details to extract genres and tags
		content, err := re.getContentByID(ctx, interaction.ContentID)
		if err != nil {
			continue
		}

		// Update genre preferences
		for _, genre := range content.Genre {
			genreCounts[genre]++
			if interaction.Rating > 0 {
				genreRatings[genre] += interaction.Rating
			}
		}

		// Update tag preferences
		for _, tag := range content.Tags {
			tagCounts[tag]++
			if interaction.Rating > 0 {
				tagRatings[tag] += interaction.Rating
			}
		}
	}

	// Normalize preferences
	profile.GenrePreferences = make(map[string]float64)
	for genre, count := range genreCounts {
		if rating, exists := genreRatings[genre]; exists {
			profile.GenrePreferences[genre] = (rating / float64(count)) / 5.0
		}
	}

	profile.TagPreferences = make(map[string]float64)
	for tag, count := range tagCounts {
		if rating, exists := tagRatings[tag]; exists {
			profile.TagPreferences[tag] = (rating / float64(count)) / 5.0
		}
	}

	// Set other profile attributes
	if ratingCount > 0 {
		profile.AvgRating = totalRating / float64(ratingCount)
	}
	profile.ActivityLevel = float64(len(interactions)) / 30.0 // interactions per day

	// Save profile to MongoDB
	collection := re.mongoDB.Collection("user_profiles")
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": profile}

	_, err := collection.UpdateOne(ctx, filter, update)

	return err
}

// Helper methods

func (re *RecommendationEngine) getUserInteractions(ctx context.Context, userID string, limit int) ([]UserInteraction, error) {
	collection := re.mongoDB.Collection("user_interactions")
	filter := bson.M{"user_id": userID}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var interactions []UserInteraction
	if err := cursor.All(ctx, &interactions); err != nil {
		return nil, err
	}

	return interactions, nil
}

func (re *RecommendationEngine) getUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	collection := re.mongoDB.Collection("user_profiles")
	filter := bson.M{"user_id": userID}

	var profile UserProfile
	err := collection.FindOne(ctx, filter).Decode(&profile)
	if err != nil {
		// Return default profile if not found
		return &UserProfile{
			UserID:           userID,
			GenrePreferences: make(map[string]float64),
			TagPreferences:   make(map[string]float64),
			Features:         make(map[string]float64),
			LastUpdated:      time.Now(),
		}, nil
	}

	return &profile, nil
}

func (re *RecommendationEngine) getAllContent(ctx context.Context) ([]Content, error) {
	collection := re.mongoDB.Collection("content")
	filter := bson.M{}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var content []Content
	if err := cursor.All(ctx, &content); err != nil {
		return nil, err
	}

	return content, nil
}

func (re *RecommendationEngine) getContentByID(ctx context.Context, contentID string) (*Content, error) {
	collection := re.mongoDB.Collection("content")
	filter := bson.M{"_id": contentID}

	var content Content
	err := collection.FindOne(ctx, filter).Decode(&content)
	if err != nil {
		return nil, err
	}

	return &content, nil
}

func (re *RecommendationEngine) findSimilarUsers(ctx context.Context, userID string, interactions []UserInteraction, limit int) ([]SimilarUser, error) {
	// This is a simplified implementation
	// In production, you would use more sophisticated similarity algorithms

	return []SimilarUser{}, nil
}

func (re *RecommendationEngine) calculateContentSimilarity(profile *UserProfile, content Content, interactions []UserInteraction) float64 {
	score := 0.0

	// Genre similarity
	genreScore := re.calculateGenreMatch(profile, content)
	score += genreScore * re.config.GenreWeight

	// Tag similarity
	tagScore := re.calculateTagMatch(profile, content)
	score += tagScore * re.config.TagWeight

	// Rating similarity
	if content.Rating > 0 {
		ratingScore := content.Rating / 5.0
		score += ratingScore * re.config.RatingWeight
	}

	// Recency score
	recencyScore := re.calculateRecencyScore(content.ReleaseYear)
	score += recencyScore * re.config.RecencyWeight

	return score
}

func (re *RecommendationEngine) calculateGenreMatch(profile *UserProfile, content Content) float64 {
	if len(profile.GenrePreferences) == 0 || len(content.Genre) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, genre := range content.Genre {
		if score, exists := profile.GenrePreferences[genre]; exists {
			totalScore += score
		}
	}

	return totalScore / float64(len(content.Genre))
}

func (re *RecommendationEngine) calculateTagMatch(profile *UserProfile, content Content) float64 {
	if len(profile.TagPreferences) == 0 || len(content.Tags) == 0 {
		return 0.0
	}

	totalScore := 0.0
	matchCount := 0

	for _, tag := range content.Tags {
		if score, exists := profile.TagPreferences[tag]; exists {
			totalScore += score
			matchCount++
		}
	}

	if matchCount == 0 {
		return 0.0
	}

	return totalScore / float64(matchCount)
}

func (re *RecommendationEngine) calculateFeatureMatch(profile *UserProfile, content Content) float64 {
	// Simplified feature matching
	return 0.5
}

func (re *RecommendationEngine) calculateRecencyScore(releaseYear int) float64 {
	currentYear := time.Now().Year()
	age := currentYear - releaseYear

	// Newer content gets higher score
	if age <= 1 {
		return 1.0
	} else if age <= 3 {
		return 0.8
	} else if age <= 5 {
		return 0.6
	} else if age <= 10 {
		return 0.4
	} else {
		return 0.2
	}
}

func (re *RecommendationEngine) calculatePersonalizedWeights(profile *UserProfile) map[string]float64 {
	weights := make(map[string]float64)

	// Adjust weights based on user activity level
	if profile.ActivityLevel > 1.0 {
		// Active users get more collaborative recommendations
		weights["collaborative"] = 0.5
		weights["content"] = 0.3
		weights["trending"] = 0.2
	} else {
		// Less active users get more content-based recommendations
		weights["collaborative"] = 0.2
		weights["content"] = 0.6
		weights["trending"] = 0.2
	}

	return weights
}

func (re *RecommendationEngine) applyPersonalizedScoring(rec Recommendation, profile *UserProfile, weights map[string]float64) float64 {
	// Apply personalized scoring based on user profile
	score := rec.Score

	// Boost score based on genre preferences
	if content, err := re.getContentByID(context.Background(), rec.ContentID); err == nil {
		genreBonus := re.calculateGenreMatch(profile, *content) * 0.2
		score += genreBonus
	}

	return score
}

func (re *RecommendationEngine) applyFilters(recommendations []Recommendation, req RecommendationRequest) []Recommendation {
	var filtered []Recommendation

	for _, rec := range recommendations {
		// Apply genre filters
		if len(req.IncludeGenres) > 0 {
			// Check if content matches included genres
			// This is simplified - in production, you'd fetch content details
		}

		if len(req.ExcludeGenres) > 0 {
			// Check if content doesn't match excluded genres
		}

		filtered = append(filtered, rec)
	}

	return filtered
}

func (re *RecommendationEngine) enrichWithContent(recommendations []Recommendation) []Recommendation {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := range recommendations {
		content, err := re.getContentByID(ctx, recommendations[i].ContentID)
		if err == nil {
			recommendations[i].Content = content
		}
	}

	return recommendations
}

// Background processes

func (re *RecommendationEngine) updateModels() {
	ticker := time.NewTicker(re.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			re.updateUserProfiles()
		case <-re.ctx.Done():
			return
		}
	}
}

func (re *RecommendationEngine) updateUserProfiles() {
	// Update user profiles based on recent interactions
	log.Println("Updating user profiles...")
}

func (re *RecommendationEngine) precomputeRecommendations() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			re.precomputeForActiveUsers()
		case <-re.ctx.Done():
			return
		}
	}
}

func (re *RecommendationEngine) precomputeForActiveUsers() {
	// Precompute recommendations for active users
	log.Println("Precomputing recommendations for active users...")
}

// Close stops the recommendation engine
func (re *RecommendationEngine) Close() error {
	re.cancel()
	re.wg.Wait()
	return re.redisClient.Close()
}

// SimilarUser represents a user similar to another user
type SimilarUser struct {
	UserID     string
	Similarity float64
}
