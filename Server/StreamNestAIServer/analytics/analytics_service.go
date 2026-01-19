package analytics

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AnalyticsService handles analytics and event tracking
type AnalyticsService struct {
	client         *mongo.Client
	db             *mongo.Database
	redisAddr      string
	eventChannel   chan AnalyticsEvent
	metricsChannel chan MetricEvent
}

// AnalyticsEvent represents a user activity event
type AnalyticsEvent struct {
	ID        string                 `bson:"_id,omitempty" json:"id"`
	UserID    string                 `bson:"user_id" json:"user_id"`
	SessionID string                 `bson:"session_id" json:"session_id"`
	Event     string                 `bson:"event" json:"event"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata" json:"metadata"`
}

// MetricEvent represents a system metric
type MetricEvent struct {
	Name      string                 `bson:"name" json:"name"`
	Value     float64                `bson:"value" json:"value"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Tags      map[string]string      `bson:"tags" json:"tags"`
	Metadata  map[string]interface{} `bson:"metadata" json:"metadata"`
}

// UserBehavior represents user behavior analytics
type UserBehavior struct {
	UserID           string    `bson:"user_id" json:"user_id"`
	TotalWatchTime   int64     `bson:"total_watch_time" json:"total_watch_time"`
	MoviesWatched    int       `bson:"movies_watched" json:"movies_watched"`
	GenresWatched    []string  `bson:"genres_watched" json:"genres_watched"`
	AverageRating    float64   `bson:"average_rating" json:"average_rating"`
	LastActive       time.Time `bson:"last_active" json:"last_active"`
	PreferredQuality string    `bson:"preferred_quality" json:"preferred_quality"`
}

// ContentAnalytics represents content performance analytics
type ContentAnalytics struct {
	MovieID       string    `bson:"movie_id" json:"movie_id"`
	Title         string    `bson:"title" json:"title"`
	TotalViews    int       `bson:"total_views" json:"total_views"`
	UniqueViewers int       `bson:"unique_viewers" json:"unique_viewers"`
	AverageRating float64   `bson:"average_rating" json:"average_rating"`
	WatchTime     int64     `bson:"watch_time" json:"watch_time"`
	DropOffRate   float64   `bson:"drop_off_rate" json:"drop_off_rate"`
	LastUpdated   time.Time `bson:"last_updated" json:"last_updated"`
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(client *mongo.Client, redisAddr string) *AnalyticsService {
	return &AnalyticsService{
		client:         client,
		db:             client.Database("streamnestai"),
		redisAddr:      redisAddr,
		eventChannel:   make(chan AnalyticsEvent, 1000),
		metricsChannel: make(chan MetricEvent, 1000),
	}
}

// Initialize sets up the analytics service
func (as *AnalyticsService) Initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create indexes for analytics collections
	collections := []string{
		"analytics_events",
		"user_behavior",
		"content_analytics",
		"system_metrics",
	}

	for _, collName := range collections {
		coll := as.db.Collection(collName)

		// Create timestamp index
		_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{"timestamp", -1}},
		})
		if err != nil {
			log.Printf("Warning: Failed to create timestamp index for %s: %v", collName, err)
		}

		// Create user_id index for relevant collections
		if collName == "analytics_events" || collName == "user_behavior" {
			_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: bson.D{{"user_id", 1}},
			})
			if err != nil {
				log.Printf("Warning: Failed to create user_id index for %s: %v", collName, err)
			}
		}
	}

	log.Println("Analytics service initialized successfully")
	return nil
}

// TrackEvent tracks a user activity event
func (as *AnalyticsService) TrackEvent(event AnalyticsEvent) error {
	select {
	case as.eventChannel <- event:
		return nil
	default:
		return fmt.Errorf("event channel is full")
	}
}

// TrackMetric tracks a system metric
func (as *AnalyticsService) TrackMetric(metric MetricEvent) error {
	select {
	case as.metricsChannel <- metric:
		return nil
	default:
		return fmt.Errorf("metrics channel is full")
	}
}

// ProcessEvents processes analytics events in the background
func (as *AnalyticsService) ProcessEvents() {
	log.Println("Starting analytics event processing...")

	for {
		select {
		case event := <-as.eventChannel:
			if err := as.saveEvent(event); err != nil {
				log.Printf("Failed to save analytics event: %v", err)
			}
		case metric := <-as.metricsChannel:
			if err := as.saveMetric(metric); err != nil {
				log.Printf("Failed to save metric: %v", err)
			}
		}
	}
}

// saveEvent saves an analytics event to the database
func (as *AnalyticsService) saveEvent(event AnalyticsEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := as.db.Collection("analytics_events")
	_, err := coll.InsertOne(ctx, event)
	return err
}

// saveMetric saves a metric to the database
func (as *AnalyticsService) saveMetric(metric MetricEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := as.db.Collection("system_metrics")
	_, err := coll.InsertOne(ctx, metric)
	return err
}

// GetUserBehavior returns analytics for a specific user
func (as *AnalyticsService) GetUserBehavior(userID string) (*UserBehavior, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := as.db.Collection("user_behavior")
	var behavior UserBehavior

	err := coll.FindOne(ctx, bson.M{"user_id": userID}).Decode(&behavior)
	if err != nil {
		return nil, err
	}

	return &behavior, nil
}

// UpdateUserBehavior updates user behavior analytics
func (as *AnalyticsService) UpdateUserBehavior(userID string, event AnalyticsEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := as.db.Collection("user_behavior")

	update := bson.M{
		"$set": bson.M{
			"last_active": time.Now(),
		},
		"$inc": bson.M{
			"total_watch_time": event.Metadata["watch_time"].(int64),
		},
		"$setOnInsert": bson.M{
			"user_id":        userID,
			"movies_watched": 0,
			"average_rating": 0.0,
		},
	}

	// Handle specific event types
	switch event.Event {
	case "movie_watched":
		update["$inc"].(bson.M)["movies_watched"] = 1

		// Add genre to watched genres
		if genre, ok := event.Metadata["genre"].(string); ok {
			update["$addToSet"] = bson.M{"genres_watched": genre}
		}

		// Update average rating
		if rating, ok := event.Metadata["rating"].(float64); ok {
			update["$avg"] = bson.M{"average_rating": rating}
		}
	}

	_, err := coll.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		update,
		nil, // options
	)

	return err
}

// GetContentAnalytics returns analytics for a specific movie
func (as *AnalyticsService) GetContentAnalytics(movieID string) (*ContentAnalytics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := as.db.Collection("content_analytics")
	var analytics ContentAnalytics

	err := coll.FindOne(ctx, bson.M{"movie_id": movieID}).Decode(&analytics)
	if err != nil {
		return nil, err
	}

	return &analytics, nil
}

// UpdateContentAnalytics updates content performance analytics
func (as *AnalyticsService) UpdateContentAnalytics(movieID string, title string, event AnalyticsEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := as.db.Collection("content_analytics")

	update := bson.M{
		"$set": bson.M{
			"last_updated": time.Now(),
		},
		"$inc": bson.M{
			"total_views": 1,
		},
		"$setOnInsert": bson.M{
			"movie_id":       movieID,
			"title":          title,
			"unique_viewers": 0,
			"average_rating": 0.0,
			"watch_time":     0,
			"drop_off_rate":  0.0,
		},
	}

	// Handle specific event data
	if watchTime, ok := event.Metadata["watch_time"].(int64); ok {
		update["$inc"].(bson.M)["watch_time"] = watchTime
	}

	if rating, ok := event.Metadata["rating"].(float64); ok {
		update["$avg"] = bson.M{"average_rating": rating}
	}

	_, err := coll.UpdateOne(
		ctx,
		bson.M{"movie_id": movieID},
		update,
		nil, // options
	)

	return err
}

// GetAnalyticsDashboard returns dashboard analytics data
func (as *AnalyticsService) GetAnalyticsDashboard() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dashboard := make(map[string]interface{})

	// Get total users
	usersColl := as.db.Collection("users")
	totalUsers, err := usersColl.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	dashboard["total_users"] = totalUsers

	// Get active users in last 24 hours
	eventsColl := as.db.Collection("analytics_events")
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	// Get active users in last 24 hours using aggregation
	activeUsersPipeline := []bson.M{
		{
			"$match": bson.M{
				"timestamp": bson.M{"$gte": twentyFourHoursAgo},
			},
		},
		{
			"$group": bson.M{
				"_id": "$user_id",
			},
		},
		{
			"$count": "active_users",
		},
	}

	cursor, err := eventsColl.Aggregate(ctx, activeUsersPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var countResult []bson.M
	if err = cursor.All(ctx, &countResult); err != nil {
		return nil, err
	}

	activeUsersCount := 0
	if len(countResult) > 0 {
		activeUsersCount = int(countResult[0]["active_users"].(int64))
	}

	dashboard["active_users_24h"] = activeUsersCount

	// Get total movies watched
	totalWatched, err := eventsColl.CountDocuments(ctx, bson.M{
		"event": "movie_watched",
	})
	if err != nil {
		return nil, err
	}
	dashboard["total_movies_watched"] = totalWatched

	// Get average session duration
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"event":     "session_end",
				"timestamp": bson.M{"$gte": twentyFourHoursAgo},
			},
		},
		{
			"$group": bson.M{
				"_id":          nil,
				"avg_duration": bson.M{"$avg": "$metadata.duration"},
			},
		},
	}

	sessionCursor, err := eventsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer sessionCursor.Close(ctx)

	var result []bson.M
	if err = sessionCursor.All(ctx, &result); err != nil {
		return nil, err
	}

	if len(result) > 0 {
		dashboard["avg_session_duration"] = result[0]["avg_duration"]
	}

	return dashboard, nil
}

// Start starts the analytics service
func (as *AnalyticsService) Start() error {
	log.Println("Analytics service started")
	return nil
}

// Stop stops the analytics service
func (as *AnalyticsService) Stop() error {
	close(as.eventChannel)
	close(as.metricsChannel)
	log.Println("Analytics service stopped")
	return nil
}
