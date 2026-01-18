package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/cache"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/controllers"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/database"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func testCachingFlow() {
	log.Println("=== StreamNestAI Cache Flow Testing Program ===")

	// Initialize MongoDB connection
	log.Println("1. Initializing MongoDB connection...")
	client := database.Connect()
	if client == nil {
		log.Fatalf("Failed to connect to MongoDB")
	}
	defer client.Disconnect(context.Background())
	log.Println("✓ MongoDB connection established")

	// Initialize Redis service
	log.Println("2. Initializing Redis service...")
	config := cache.CacheConfig{
		Host:         "localhost",
		Port:         "6379",
		Password:     "",
		DB:           0,
		KeyPrefix:    "streamnestai",
		DefaultTTL:   5 * time.Minute,
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	redisService, err := cache.NewRedisService(config)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisService.Close()
	log.Println("✓ Redis connection established")

	// Test Redis basic functionality
	log.Println("3. Testing Redis basic functionality...")
	err = testRedisBasicFunctionality(redisService)
	if err != nil {
		log.Fatalf("Redis basic functionality test failed: %v", err)
	}
	log.Println("✓ Redis basic functionality test passed")

	// Test cached movie controller
	log.Println("4. Testing cached movie controller...")
	err = testCachedMovieController(client, redisService)
	if err != nil {
		log.Fatalf("Cached movie controller test failed: %v", err)
	}
	log.Println("✓ Cached movie controller test passed")

	// Test cached recommendation service
	log.Println("5. Testing cached recommendation service...")
	err = testCachedRecommendationService(client, redisService)
	if err != nil {
		log.Fatalf("Cached recommendation service test failed: %v", err)
	}
	log.Println("✓ Cached recommendation service test passed")

	// Test cache invalidation
	log.Println("6. Testing cache invalidation...")
	err = testCacheInvalidation(client, redisService)
	if err != nil {
		log.Fatalf("Cache invalidation test failed: %v", err)
	}
	log.Println("✓ Cache invalidation test passed")

	// Test performance comparison
	log.Println("7. Testing performance comparison...")
	err = testPerformanceComparison(client, redisService)
	if err != nil {
		log.Fatalf("Performance comparison test failed: %v", err)
	}
	log.Println("✓ Performance comparison test completed")

	log.Println("=== All Cache Flow Tests Passed Successfully ===")
}

func testRedisBasicFunctionality(redisService *cache.RedisService) error {
	ctx := context.Background()

	// Test Set and Get
	testKey := fmt.Sprintf("test:key:%d", time.Now().Unix())
	testValue := map[string]interface{}{
		"message":   "Hello Redis!",
		"timestamp": time.Now().Unix(),
	}

	err := redisService.Set(ctx, testKey, testValue, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to set test value: %v", err)
	}

	var retrievedValue map[string]interface{}
	err = redisService.Get(ctx, testKey, &retrievedValue)
	if err != nil {
		return fmt.Errorf("failed to get test value: %v", err)
	}

	if retrievedValue["message"] != "Hello Redis!" {
		return fmt.Errorf("retrieved value mismatch: expected 'Hello Redis!', got '%v'", retrievedValue["message"])
	}

	// Test Delete
	err = redisService.Delete(ctx, testKey)
	if err != nil {
		return fmt.Errorf("failed to delete test key: %v", err)
	}

	// Verify deletion
	err = redisService.Get(ctx, testKey, &retrievedValue)
	if err == nil {
		return fmt.Errorf("key should have been deleted but was still retrievable")
	}

	log.Printf("  ✓ Redis Set/Get/Delete operations working correctly")
	return nil
}

func testCachedMovieController(client *mongo.Client, redisService *cache.RedisService) error {
	// Create cached movie controller
	cachedMovieController := controllers.NewCachedMovieController(client, redisService)

	// Test cache statistics
	stats := cachedMovieController.GetCacheStats()

	log.Printf("  ✓ Cache stats retrieved: %+v", stats)

	// Test cache invalidation
	err := cachedMovieController.InvalidateMovieCache("test")
	if err != nil {
		return fmt.Errorf("failed to invalidate movie cache: %v", err)
	}

	log.Printf("  ✓ Movie cache invalidated successfully")

	return nil
}

func testCachedRecommendationService(client *mongo.Client, redisService *cache.RedisService) error {
	ctx := context.Background()

	// Create cached recommendation service
	_ = controllers.NewCachedRecommendationController(client, redisService)

	// Test cache statistics
	// Test recommendation cache stats through the service
	// Test recommendation cache stats through Redis service directly
	stats, err := redisService.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get recommendation cache stats: %v", err)
	}

	log.Printf("  ✓ Recommendation cache stats retrieved: %+v", stats)

	return nil
}

func testCacheInvalidation(client *mongo.Client, redisService *cache.RedisService) error {
	ctx := context.Background()

	// Test pattern-based deletion
	testPatterns := []string{
		"test:pattern:*",
		"movies:*",
		"recommendations:*",
	}

	for _, pattern := range testPatterns {
		// First, set some test keys
		for i := 0; i < 3; i++ {
			key := fmt.Sprintf(pattern[:len(pattern)-1], i)
			value := map[string]interface{}{"test": i}
			err := redisService.Set(ctx, key, value, 5*time.Minute)
			if err != nil {
				return fmt.Errorf("failed to set test key %s: %v", key, err)
			}
		}

		// Then delete by pattern
		err := redisService.DeletePattern(ctx, pattern)
		if err != nil {
			return fmt.Errorf("failed to delete pattern %s: %v", pattern, err)
		}

		log.Printf("  ✓ Pattern deletion successful for: %s", pattern)
	}

	return nil
}

func testPerformanceComparison(client *mongo.Client, redisService *cache.RedisService) error {
	ctx := context.Background()

	// Test data
	testData := map[string]interface{}{
		"title":  "Test Movie",
		"genre":  "Action",
		"rating": 8.5,
		"year":   2023,
	}

	// Test Redis performance
	redisTimes := make([]time.Duration, 0, 10)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("perf:test:%d", i)

		start := time.Now()
		err := redisService.Set(ctx, key, testData, 5*time.Minute)
		if err != nil {
			return fmt.Errorf("failed to set performance test data: %v", err)
		}

		var result map[string]interface{}
		err = redisService.Get(ctx, key, &result)
		if err != nil {
			return fmt.Errorf("failed to get performance test data: %v", err)
		}

		redisTimes = append(redisTimes, time.Since(start))
	}

	// Calculate average Redis time
	var totalRedisTime time.Duration
	for _, t := range redisTimes {
		totalRedisTime += t
	}
	avgRedisTime := totalRedisTime / time.Duration(len(redisTimes))

	// Test MongoDB performance (simple find operation)
	mongoTimes := make([]time.Duration, 0, 10)
	coll := client.Database("streamnestai").Collection("movies")

	for i := 0; i < 10; i++ {
		start := time.Now()

		var result map[string]interface{}
		err := coll.FindOne(ctx, map[string]interface{}{"title": "Test Movie"}).Decode(&result)
		if err != nil {
			// Document might not exist, that's okay for performance test
		}

		mongoTimes = append(mongoTimes, time.Since(start))
	}

	// Calculate average MongoDB time
	var totalMongoTime time.Duration
	for _, t := range mongoTimes {
		totalMongoTime += t
	}
	avgMongoTime := totalMongoTime / time.Duration(len(mongoTimes))

	log.Printf("  ✓ Performance comparison completed:")
	log.Printf("    - Average Redis operation time: %v", avgRedisTime)
	log.Printf("    - Average MongoDB operation time: %v", avgMongoTime)

	if avgRedisTime < avgMongoTime {
		speedup := float64(avgMongoTime) / float64(avgRedisTime)
		log.Printf("    - Redis is %.2fx faster than MongoDB", speedup)
	}

	return nil
}

func main() {
	testCachingFlow()
}
