package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	// Redis configuration from main.go
	redisAddr := "redis-13554.crce182.ap-south-1-1.ec2.cloud.redislabs.com:13554"
	redisPassword := "HWg60QKmrYFlWmrf3pr0Utv8IC5eU8QG"
	redisDB := 0

	fmt.Println("🔍 Testing Redis Connection...")
	fmt.Printf("📍 Redis Address: %s\n", redisAddr)
	fmt.Printf("🔐 Redis DB: %d\n", redisDB)

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test connection
	fmt.Println("\n🔄 Testing Redis PING...")
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis PING failed: %v", err)
	}
	fmt.Printf("✅ Redis PING successful: %s\n", pong)

	// Test SET operation
	fmt.Println("\n💾 Testing Redis SET...")
	testKey := "streamnest:test:redis"
	testValue := "Redis is working! " + time.Now().Format(time.RFC3339)

	err = rdb.Set(ctx, testKey, testValue, 5*time.Minute).Err()
	if err != nil {
		log.Fatalf("❌ Redis SET failed: %v", err)
	}
	fmt.Printf("✅ Redis SET successful: %s = %s\n", testKey, testValue)

	// Test GET operation
	fmt.Println("\n📖 Testing Redis GET...")
	retrievedValue, err := rdb.Get(ctx, testKey).Result()
	if err != nil {
		log.Fatalf("❌ Redis GET failed: %v", err)
	}
	fmt.Printf("✅ Redis GET successful: %s\n", retrievedValue)

	// Test HSET operation (for user sessions)
	fmt.Println("\n👤 Testing Redis HSET (user sessions)...")
	sessionKey := "session:test_user_123"
	sessionData := map[string]interface{}{
		"user_id":    "test_user_123",
		"username":   "testuser",
		"login_time": time.Now().Unix(),
		"active":     "true",
	}

	for field, value := range sessionData {
		err := rdb.HSet(ctx, sessionKey, field, value).Err()
		if err != nil {
			log.Fatalf("❌ Redis HSET failed for field %s: %v", field, err)
		}
	}
	fmt.Printf("✅ Redis HSET successful for session: %s\n", sessionKey)

	// Test HGETALL operation
	fmt.Println("\n📋 Testing Redis HGETALL...")
	retrievedSession, err := rdb.HGetAll(ctx, sessionKey).Result()
	if err != nil {
		log.Fatalf("❌ Redis HGETALL failed: %v", err)
	}
	fmt.Printf("✅ Redis HGETALL successful: %+v\n", retrievedSession)

	// Test LIST operations (for activity feeds)
	fmt.Println("\n📝 Testing Redis LIST (activity feed)...")
	activityKey := "activity:global"
	activities := []string{
		"user1 watched Movie A",
		"user2 reviewed Movie B",
		"user3 started watching Movie C",
	}

	for _, activity := range activities {
		err := rdb.LPush(ctx, activityKey, activity).Err()
		if err != nil {
			log.Fatalf("❌ Redis LPUSH failed: %v", err)
		}
	}
	fmt.Printf("✅ Redis LPUSH successful for %d activities\n", len(activities))

	// Get list length
	listLen, err := rdb.LLen(ctx, activityKey).Result()
	if err != nil {
		log.Fatalf("❌ Redis LLEN failed: %v", err)
	}
	fmt.Printf("✅ Redis list length: %d\n", listLen)

	// Test ZSET operations (for trending movies)
	fmt.Println("\n🔥 Testing Redis ZSET (trending movies)...")
	trendingKey := "trending:movies"
	trendingMovies := map[string]float64{
		"movie:1": 95.5,
		"movie:2": 87.3,
		"movie:3": 92.1,
	}

	for movie, score := range trendingMovies {
		err := rdb.ZAdd(ctx, trendingKey, &redis.Z{
			Score:  score,
			Member: movie,
		}).Err()
		if err != nil {
			log.Fatalf("❌ Redis ZADD failed for %s: %v", movie, err)
		}
	}
	fmt.Printf("✅ Redis ZADD successful for %d movies\n", len(trendingMovies))

	// Get top trending movies
	topMovies, err := rdb.ZRevRangeWithScores(ctx, trendingKey, 0, 2).Result()
	if err != nil {
		log.Fatalf("❌ Redis ZREVRANGE failed: %v", err)
	}
	fmt.Printf("✅ Top trending movies:\n")
	for _, movie := range topMovies {
		fmt.Printf("   - %s: %.1f\n", movie.Member, movie.Score)
	}

	// Test expiration
	fmt.Println("\n⏰ Testing Redis EXPIRE...")
	expireKey := "streamnest:test:expire"
	rdb.Set(ctx, expireKey, "This will expire in 10 seconds", 0)

	err = rdb.Expire(ctx, expireKey, 10*time.Second).Err()
	if err != nil {
		log.Fatalf("❌ Redis EXPIRE failed: %v", err)
	}
	fmt.Printf("✅ Redis EXPIRE successful for key: %s\n", expireKey)

	// Get TTL
	ttl, err := rdb.TTL(ctx, expireKey).Result()
	if err != nil {
		log.Fatalf("❌ Redis TTL failed: %v", err)
	}
	fmt.Printf("✅ Redis TTL: %v\n", ttl)

	// Clean up test data
	fmt.Println("\n🧹 Cleaning up test data...")
	keysToDelete := []string{testKey, sessionKey, activityKey, trendingKey, expireKey}
	for _, key := range keysToDelete {
		rdb.Del(ctx, key)
	}
	fmt.Printf("✅ Cleaned up %d test keys\n", len(keysToDelete))

	// Test info
	fmt.Println("\n📊 Getting Redis INFO...")
	info, err := rdb.Info(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis INFO failed: %v", err)
	}

	// Extract some key info
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if strings.Contains(line, "redis_version") ||
			strings.Contains(line, "used_memory_human") ||
			strings.Contains(line, "connected_clients") ||
			strings.Contains(line, "total_commands_processed") {
			fmt.Printf("📈 %s\n", line)
		}
	}

	fmt.Println("\n🎉 All Redis tests completed successfully!")
	fmt.Println("✅ Redis is fully functional and ready for production use")
}
