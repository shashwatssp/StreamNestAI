package social

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SocialService handles social features like profiles, follows, and activity feeds
type SocialService struct {
	client *mongo.Client
	db     *mongo.Database
}

// UserProfile represents a user's social profile
type UserProfile struct {
	UserID         string    `bson:"user_id" json:"user_id"`
	Username       string    `bson:"username" json:"username"`
	DisplayName    string    `bson:"display_name" json:"display_name"`
	Bio            string    `bson:"bio" json:"bio"`
	Avatar         string    `bson:"avatar" json:"avatar"`
	FollowersCount int       `bson:"followers_count" json:"followers_count"`
	FollowingCount int       `bson:"following_count" json:"following_count"`
	PostsCount     int       `bson:"posts_count" json:"posts_count"`
	WatchlistCount int       `bson:"watchlist_count" json:"watchlist_count"`
	Watchlists     []string  `bson:"watchlists" json:"watchlists"`
	FavoriteGenres []string  `bson:"favorite_genres" json:"favorite_genres"`
	IsPrivate      bool      `bson:"is_private" json:"is_private"`
	IsVerified     bool      `bson:"is_verified" json:"is_verified"`
	JoinDate       time.Time `bson:"join_date" json:"join_date"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
	Stats          UserStats `bson:"stats" json:"stats"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalWatchTime int     `bson:"total_watch_time" json:"total_watch_time"`
	MoviesWatched  int     `bson:"movies_watched" json:"movies_watched"`
	SeriesWatched  int     `bson:"series_watched" json:"series_watched"`
	AverageRating  float64 `bson:"average_rating" json:"average_rating"`
	StreakDays     int     `bson:"streak_days" json:"streak_days"`
	BadgeCount     int     `bson:"badge_count" json:"badge_count"`
}

// Follow represents a follow relationship
type Follow struct {
	FollowerID  string    `bson:"follower_id" json:"follower_id"`
	FollowingID string    `bson:"following_id" json:"following_id"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
}

// Activity represents an activity in the feed
type Activity struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	Type      string    `bson:"type" json:"type"` // watch, review, follow, list_add
	Content   string    `bson:"content" json:"content"`
	Metadata  bson.M    `bson:"metadata" json:"metadata"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

// Watchlist represents a shared watchlist
type Watchlist struct {
	ID          string    `bson:"_id" json:"id"`
	Name        string    `bson:"name" json:"name"`
	Description string    `bson:"description" json:"description"`
	CreatorID   string    `bson:"creator_id" json:"creator_id"`
	Movies      []string  `bson:"movies" json:"movies"`
	IsPublic    bool      `bson:"is_public" json:"is_public"`
	Followers   []string  `bson:"followers" json:"followers"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

// NewSocialService creates a new social service
func NewSocialService(client *mongo.Client) *SocialService {
	return &SocialService{
		client: client,
		db:     client.Database("streamnestai"),
	}
}

// Initialize sets up the social service
func (ss *SocialService) Initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create indexes for social collections
	collections := map[string][]bson.D{
		"user_profiles": {
			{{"user_id", 1}},
			{{"username", 1}},
		},
		"follows": {
			{{"follower_id", 1}},
			{{"following_id", 1}},
			{{"follower_id", 1}, {"following_id", 1}},
		},
		"activities": {
			{{"user_id", 1}, {"created_at", -1}},
			{{"created_at", -1}},
		},
		"watchlists": {
			{{"creator_id", 1}},
			{{"is_public", 1}},
		},
	}

	for collName, indexes := range collections {
		coll := ss.db.Collection(collName)
		for _, index := range indexes {
			_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: index})
			if err != nil {
				log.Printf("Warning: Failed to create index for %s: %v", collName, err)
			}
		}
	}

	log.Println("Social service initialized successfully")
	return nil
}

// GetProfile retrieves a user's social profile
func (ss *SocialService) GetProfile(userID string) (*UserProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("user_profiles")
	var profile UserProfile

	err := coll.FindOne(ctx, bson.M{"user_id": userID}).Decode(&profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// GetUserProfile retrieves a user's social profile (alias for GetProfile)
func (ss *SocialService) GetUserProfile(userID string) (*UserProfile, error) {
	return ss.GetProfile(userID)
}

// UpdateProfile updates a user's social profile
func (ss *SocialService) UpdateProfile(userID string, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("user_profiles")

	updates["updated_at"] = time.Now()
	_, err := coll.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": updates},
	)

	return err
}

// UpdateUserProfile updates a user's social profile (alias for UpdateProfile)
func (ss *SocialService) UpdateUserProfile(userID string, updates map[string]interface{}) error {
	// Convert map[string]interface{} to bson.M
	bsonUpdates := bson.M{}
	for k, v := range updates {
		bsonUpdates[k] = v
	}
	return ss.UpdateProfile(userID, bsonUpdates)
}

// FollowUser creates a follow relationship
func (ss *SocialService) FollowUser(followerID, followingID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if already following
	coll := ss.db.Collection("follows")
	count, err := coll.CountDocuments(ctx, bson.M{
		"follower_id":  followerID,
		"following_id": followingID,
	})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already following
	}

	// Create follow relationship
	follow := Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
		CreatedAt:   time.Now(),
	}

	_, err = coll.InsertOne(ctx, follow)
	if err != nil {
		return err
	}

	// Update follower counts
	ss.updateFollowCounts(followerID, followingID)

	// Create activity
	ss.createActivity(followerID, "follow", "started following", bson.M{
		"following_id": followingID,
	})

	return nil
}

// UnfollowUser removes a follow relationship
func (ss *SocialService) UnfollowUser(followerID, followingID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("follows")
	_, err := coll.DeleteOne(ctx, bson.M{
		"follower_id":  followerID,
		"following_id": followingID,
	})
	if err != nil {
		return err
	}

	// Update follower counts
	ss.updateFollowCounts(followerID, followingID)

	return nil
}

// GetFollowers returns the list of followers for a user
func (ss *SocialService) GetFollowers(userID string, limit int, offset int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("follows")

	// Apply sorting and limiting using options
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := coll.Find(ctx, bson.M{"following_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var follows []Follow
	if err = cursor.All(ctx, &follows); err != nil {
		return nil, err
	}

	followers := make([]string, 0, len(follows))
	for _, follow := range follows {
		followers = append(followers, follow.FollowerID)
	}

	return followers, nil
}

// GetFollowing returns the list of users a user is following
func (ss *SocialService) GetFollowing(userID string, limit int, offset int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("follows")

	// Apply sorting and limiting using options
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := coll.Find(ctx, bson.M{"follower_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var follows []Follow
	if err = cursor.All(ctx, &follows); err != nil {
		return nil, err
	}

	following := make([]string, 0, len(follows))
	for _, follow := range follows {
		following = append(following, follow.FollowingID)
	}

	return following, nil
}

// GetActivityFeed returns the activity feed for a user
func (ss *SocialService) GetActivityFeed(userID string, limit int, offset int) ([]Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get users that the current user follows
	following, err := ss.GetFollowing(userID, 100, 0)
	if err != nil {
		return nil, err
	}

	// Add current user to the list
	following = append(following, userID)

	coll := ss.db.Collection("activities")

	// Apply sorting and limiting using options
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := coll.Find(ctx, bson.M{
		"user_id": bson.M{"$in": following},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []Activity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// GetUserActivities returns activities for a specific user
func (ss *SocialService) GetUserActivities(userID string, limit int, offset int) ([]Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("activities")

	// Apply sorting and limiting using options
	opts := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := coll.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []Activity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// CreateActivity creates a new activity
func (ss *SocialService) CreateActivity(userID, activityType, contentID, contentType, title, description string, metadata map[string]interface{}) error {
	// Convert map[string]interface{} to bson.M
	bsonMetadata := bson.M{}
	for k, v := range metadata {
		bsonMetadata[k] = v
	}

	// Add additional metadata
	bsonMetadata["content_id"] = contentID
	bsonMetadata["content_type"] = contentType
	bsonMetadata["title"] = title
	bsonMetadata["description"] = description

	return ss.createActivity(userID, activityType, title, bsonMetadata)
}

// CreateWatchlist creates a new watchlist
func (ss *SocialService) CreateWatchlist(userID, name, description string, isPublic bool) (*Watchlist, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("watchlists")

	watchlist := Watchlist{
		ID:          bson.NewObjectID().Hex(),
		Name:        name,
		Description: description,
		CreatorID:   userID,
		Movies:      []string{},
		IsPublic:    isPublic,
		Followers:   []string{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := coll.InsertOne(ctx, watchlist)
	if err != nil {
		return nil, err
	}

	return &watchlist, nil
}

// GetUserWatchlists returns watchlists for a user
func (ss *SocialService) GetUserWatchlists(userID string) ([]Watchlist, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("watchlists")
	cursor, err := coll.Find(ctx, bson.M{"creator_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var watchlists []Watchlist
	if err = cursor.All(ctx, &watchlists); err != nil {
		return nil, err
	}

	return watchlists, nil
}

// AddToWatchlist adds a movie to a watchlist
func (ss *SocialService) AddToWatchlist(watchlistID, contentID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("watchlists")

	_, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": watchlistID},
		bson.M{
			"$addToSet": bson.M{"movies": contentID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)

	return err
}

// RemoveFromWatchlist removes a movie from a watchlist
func (ss *SocialService) RemoveFromWatchlist(watchlistID, contentID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("watchlists")

	_, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": watchlistID},
		bson.M{
			"$pull": bson.M{"movies": contentID},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)

	return err
}

// IsWatchlistOwner checks if a user is the owner of a watchlist
func (ss *SocialService) IsWatchlistOwner(watchlistID, userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("watchlists")
	var watchlist Watchlist

	err := coll.FindOne(ctx, bson.M{"_id": watchlistID}).Decode(&watchlist)
	if err != nil {
		return false, err
	}

	return watchlist.CreatorID == userID, nil
}

// GetWatchlistByID retrieves a watchlist by ID
func (ss *SocialService) GetWatchlistByID(watchlistID string) (*Watchlist, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("watchlists")
	var watchlist Watchlist

	err := coll.FindOne(ctx, bson.M{"_id": watchlistID}).Decode(&watchlist)
	if err != nil {
		return nil, err
	}

	return &watchlist, nil
}

// Helper functions

func (ss *SocialService) updateFollowCounts(followerID, followingID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("user_profiles")

	// Update follower count
	coll.UpdateOne(ctx, bson.M{"user_id": followingID}, bson.M{
		"$inc": bson.M{"followers_count": 1},
	})

	// Update following count
	coll.UpdateOne(ctx, bson.M{"user_id": followerID}, bson.M{
		"$inc": bson.M{"following_count": 1},
	})
}

func (ss *SocialService) createActivity(userID, activityType, content string, metadata bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := ss.db.Collection("activities")

	activity := Activity{
		ID:        bson.NewObjectID().Hex(),
		UserID:    userID,
		Type:      activityType,
		Content:   content,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	_, err := coll.InsertOne(ctx, activity)
	return err
}
