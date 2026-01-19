package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisService handles all Redis operations with connection pooling
type RedisService struct {
	client     *redis.Client
	keyPrefix  string
	defaultTTL time.Duration
}

// CacheItem represents a cached item with metadata
type CacheItem struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	ExpiresAt   time.Time   `json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
	AccessCount int64       `json:"access_count"`
}

// CacheConfig holds Redis configuration
type CacheConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	KeyPrefix    string
	DefaultTTL   time.Duration
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewRedisService creates a new Redis service with connection pooling
func NewRedisService(config CacheConfig) (*RedisService, error) {
	// Log configuration for debugging (without password)
	log.Printf("🔧 REDIS CONFIG DEBUG:")
	log.Printf("   Host: %s", config.Host)
	log.Printf("   Port: %s", config.Port)
	log.Printf("   DB: %d", config.DB)
	log.Printf("   KeyPrefix: %s", config.KeyPrefix)
	log.Printf("   DefaultTTL: %s", config.DefaultTTL.String())
	log.Printf("   PoolSize: %d", config.PoolSize)
	log.Printf("   MinIdleConns: %d", config.MinIdleConns)
	log.Printf("   DialTimeout: %s", config.DialTimeout.String())
	log.Printf("   ReadTimeout: %s", config.ReadTimeout.String())
	log.Printf("   WriteTimeout: %s", config.WriteTimeout.String())

	// Construct Redis connection string for logging
	redisAddr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	log.Printf("🔗 Connecting to Redis at: %s", redisAddr)

	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection with detailed logging
	log.Printf("🔍 Testing Redis connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	startTime := time.Now()
	pong, err := rdb.Ping(ctx).Result()
	connectionTime := time.Since(startTime)

	if err != nil {
		log.Printf("❌ REDIS CONNECTION FAILED:")
		log.Printf("   Error: %v", err)
		log.Printf("   Connection Time: %s", connectionTime.String())
		log.Printf("   Redis Address: %s", redisAddr)
		log.Printf("   Database: %d", config.DB)
		log.Printf("   Troubleshooting:")
		log.Printf("   1. Check if Redis server is running at %s", redisAddr)
		log.Printf("   2. Verify password is correct")
		log.Printf("   3. Check network connectivity")
		log.Printf("   4. Verify Redis configuration allows remote connections")
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ REDIS CONNECTION SUCCESSFUL:")
	log.Printf("   Response: %s", pong)
	log.Printf("   Connection Time: %s", connectionTime.String())
	log.Printf("   Redis Address: %s", redisAddr)
	log.Printf("   Database: %d", config.DB)
	log.Printf("   Key Prefix: %s", config.KeyPrefix)

	// Test basic operations
	log.Printf("🧪 Testing Redis operations...")
	testKey := fmt.Sprintf("%s:test_connection", config.KeyPrefix)
	testValue := map[string]interface{}{
		"test":      true,
		"timestamp": time.Now(),
		"message":   "Redis connection test",
	}

	err = rdb.Set(ctx, testKey, testValue, 10*time.Second).Err()
	if err != nil {
		log.Printf("⚠️  Redis SET test failed: %v", err)
	} else {
		log.Printf("✅ Redis SET test passed")
	}

	var retrievedValue map[string]interface{}
	err = rdb.Get(ctx, testKey).Scan(&retrievedValue)
	if err != nil {
		log.Printf("⚠️  Redis GET test failed: %v", err)
	} else {
		log.Printf("✅ Redis GET test passed")
	}

	// Clean up test key
	rdb.Del(ctx, testKey)

	log.Printf("🚀 Redis service initialized successfully")

	return &RedisService{
		client:     rdb,
		keyPrefix:  config.KeyPrefix,
		defaultTTL: config.DefaultTTL,
	}, nil
}

// generateKey creates a namespaced key
func (r *RedisService) generateKey(key string) string {
	return fmt.Sprintf("%s:%s", r.keyPrefix, key)
}

// Set stores a value in Redis with TTL
func (r *RedisService) Set(ctx context.Context, key string, value interface{}, ttl ...time.Duration) error {
	duration := r.defaultTTL
	if len(ttl) > 0 {
		duration = ttl[0]
	}

	// Serialize value
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Store in Redis
	fullKey := r.generateKey(key)
	err = r.client.Set(ctx, fullKey, data, duration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	// Update metadata
	metadata := CacheItem{
		Key:         key,
		Value:       value,
		ExpiresAt:   time.Now().Add(duration),
		CreatedAt:   time.Now(),
		AccessCount: 0,
	}

	metadataKey := r.generateKey(fmt.Sprintf("meta:%s", key))
	metadataData, _ := json.Marshal(metadata)
	r.client.Set(ctx, metadataKey, metadataData, duration)

	return nil
}

// Get retrieves a value from Redis
func (r *RedisService) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := r.generateKey(key)

	// Get value
	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key %s not found", key)
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	// Deserialize value
	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	// Update access count
	go r.updateAccessCount(key)

	return nil
}

// GetWithTTL retrieves a value and its remaining TTL
func (r *RedisService) GetWithTTL(ctx context.Context, key string, dest interface{}) (time.Duration, error) {
	fullKey := r.generateKey(key)

	// Get value and TTL
	pipe := r.client.Pipeline()
	getCmd := pipe.Get(ctx, fullKey)
	ttlCmd := pipe.TTL(ctx, fullKey)
	_, err := pipe.Exec(ctx)

	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("key %s not found", key)
		}
		return 0, fmt.Errorf("failed to get key %s: %w", key, err)
	}

	// Deserialize value
	err = json.Unmarshal([]byte(getCmd.Val()), dest)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	// Update access count
	go r.updateAccessCount(key)

	return ttlCmd.Val(), nil
}

// Delete removes a key from Redis
func (r *RedisService) Delete(ctx context.Context, key string) error {
	fullKey := r.generateKey(key)
	metadataKey := r.generateKey(fmt.Sprintf("meta:%s", key))

	pipe := r.client.Pipeline()
	pipe.Del(ctx, fullKey)
	pipe.Del(ctx, metadataKey)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// DeletePattern removes keys matching a pattern
func (r *RedisService) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := r.generateKey(pattern)
	keys, err := r.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	err = r.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete keys for pattern %s: %w", pattern, err)
	}

	return nil
}

// Exists checks if a key exists
func (r *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.generateKey(key)
	count, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return count > 0, nil
}

// Increment atomically increments a numeric value
func (r *RedisService) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := r.generateKey(key)
	result, err := r.client.Incr(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return result, nil
}

// IncrementBy atomically increments a numeric value by a specific amount
func (r *RedisService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	fullKey := r.generateKey(key)
	result, err := r.client.IncrBy(ctx, fullKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, value, err)
	}
	return result, nil
}

// SetWithExpiration sets a key with specific expiration
func (r *RedisService) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration time.Time) error {
	ttl := time.Until(expiration)
	if ttl <= 0 {
		return fmt.Errorf("expiration time must be in the future")
	}
	return r.Set(ctx, key, value, ttl)
}

// GetMultiple retrieves multiple keys at once
func (r *RedisService) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// Generate full keys
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.generateKey(key)
	}

	// Get all values
	values, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple keys: %w", err)
	}

	// Parse results
	result := make(map[string]interface{})
	for i, value := range values {
		if value != nil {
			var dest interface{}
			err := json.Unmarshal([]byte(value.(string)), &dest)
			if err == nil {
				result[keys[i]] = dest
				// Update access count asynchronously
				go r.updateAccessCount(keys[i])
			}
		}
	}

	return result, nil
}

// SetMultiple stores multiple keys at once
func (r *RedisService) SetMultiple(ctx context.Context, items map[string]interface{}, ttl ...time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	duration := r.defaultTTL
	if len(ttl) > 0 {
		duration = ttl[0]
	}

	pipe := r.client.Pipeline()
	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		fullKey := r.generateKey(key)
		pipe.Set(ctx, fullKey, data, duration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to set multiple keys: %w", err)
	}

	return nil
}

// updateAccessCount increments the access count for a key
func (r *RedisService) updateAccessCount(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	metadataKey := r.generateKey(fmt.Sprintf("meta:%s", key))
	r.client.Incr(ctx, metadataKey+":access_count")
}

// GetStats returns cache statistics
func (r *RedisService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Parse basic stats
	stats := map[string]interface{}{
		"info":        info,
		"key_prefix":  r.keyPrefix,
		"default_ttl": r.defaultTTL.String(),
	}

	return stats, nil
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	return r.client.Close()
}

// Ping checks if Redis is available
func (r *RedisService) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// FlushAll clears all keys with the service prefix
func (r *RedisService) FlushAll(ctx context.Context) error {
	pattern := r.generateKey("*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for flush: %w", err)
	}

	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys during flush: %w", err)
		}
	}

	return nil
}

// CacheWarmer interface for preloading cache
type CacheWarmer interface {
	WarmUp(ctx context.Context, cache *RedisService) error
}

// WarmUpCache preloads cache using provided warmers
func (r *RedisService) WarmUpCache(ctx context.Context, warmers []CacheWarmer) error {
	for _, warmer := range warmers {
		err := warmer.WarmUp(ctx, r)
		if err != nil {
			log.Printf("Cache warm-up failed for warmer: %v", err)
			// Continue with other warmers even if one fails
		}
	}
	return nil
}
