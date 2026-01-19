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
	startTime := time.Now()
	duration := r.defaultTTL
	if len(ttl) > 0 {
		duration = ttl[0]
	}

	log.Printf("🗄️  CACHE SET START: key=%s, ttl=%s", key, duration.String())

	// Serialize value
	marshalStart := time.Now()
	data, err := json.Marshal(value)
	marshalTime := time.Since(marshalStart)

	if err != nil {
		log.Printf("❌ CACHE SET MARSHAL ERROR: key=%s, error=%v, marshal_time=%s", key, err, marshalTime.String())
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	log.Printf("✅ CACHE SET MARSHAL SUCCESS: key=%s, data_size=%d bytes, marshal_time=%s", key, len(data), marshalTime.String())

	// Store in Redis
	fullKey := r.generateKey(key)
	setStart := time.Now()
	err = r.client.Set(ctx, fullKey, data, duration).Err()
	setTime := time.Since(setStart)
	totalTime := time.Since(startTime)

	if err != nil {
		log.Printf("❌ CACHE SET ERROR: key=%s, full_key=%s, error=%v, set_time=%s, total_time=%s",
			key, fullKey, err, setTime.String(), totalTime.String())
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	log.Printf("✅ CACHE SET SUCCESS: key=%s, full_key=%s, data_size=%d bytes, ttl=%s, set_time=%s, total_time=%s",
		key, fullKey, len(data), duration.String(), setTime.String(), totalTime.String())

	// Update metadata
	metadataStart := time.Now()
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
	metadataTime := time.Since(metadataStart)

	log.Printf("📊 CACHE METADATA SET: key=%s, metadata_key=%s, metadata_time=%s",
		key, metadataKey, metadataTime.String())

	return nil
}

// Get retrieves a value from Redis
func (r *RedisService) Get(ctx context.Context, key string, dest interface{}) error {
	startTime := time.Now()
	fullKey := r.generateKey(key)

	log.Printf("🔍 CACHE GET START: key=%s, full_key=%s", key, fullKey)

	// Get value
	getStart := time.Now()
	data, err := r.client.Get(ctx, fullKey).Result()
	getTime := time.Since(getStart)

	if err != nil {
		totalTime := time.Since(startTime)
		if err == redis.Nil {
			log.Printf("⚠️  CACHE GET MISS: key=%s, full_key=%s, get_time=%s, total_time=%s",
				key, fullKey, getTime.String(), totalTime.String())
			return fmt.Errorf("key %s not found", key)
		}
		log.Printf("❌ CACHE GET ERROR: key=%s, full_key=%s, error=%v, get_time=%s, total_time=%s",
			key, fullKey, err, getTime.String(), totalTime.String())
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	log.Printf("✅ CACHE GET HIT: key=%s, full_key=%s, data_size=%d bytes, get_time=%s",
		key, fullKey, len(data), getTime.String())

	// Deserialize value
	unmarshalStart := time.Now()
	err = json.Unmarshal([]byte(data), dest)
	unmarshalTime := time.Since(unmarshalStart)
	totalTime := time.Since(startTime)

	if err != nil {
		log.Printf("❌ CACHE GET UNMARSHAL ERROR: key=%s, error=%v, unmarshal_time=%s, total_time=%s",
			key, err, unmarshalTime.String(), totalTime.String())
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	log.Printf("✅ CACHE GET SUCCESS: key=%s, data_size=%d bytes, get_time=%s, unmarshal_time=%s, total_time=%s",
		key, len(data), getTime.String(), unmarshalTime.String(), totalTime.String())

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
	startTime := time.Now()
	fullKey := r.generateKey(key)
	metadataKey := r.generateKey(fmt.Sprintf("meta:%s", key))

	log.Printf("🗑️  CACHE DELETE START: key=%s, full_key=%s, metadata_key=%s", key, fullKey, metadataKey)

	pipe := r.client.Pipeline()
	pipe.Del(ctx, fullKey)
	pipe.Del(ctx, metadataKey)

	deleteStart := time.Now()
	result, err := pipe.Exec(ctx)
	deleteTime := time.Since(deleteStart)
	totalTime := time.Since(startTime)

	if err != nil {
		log.Printf("❌ CACHE DELETE ERROR: key=%s, error=%v, delete_time=%s, total_time=%s",
			key, err, deleteTime.String(), totalTime.String())
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	deletedCount := len(result)
	log.Printf("✅ CACHE DELETE SUCCESS: key=%s, deleted_count=%d, delete_time=%s, total_time=%s",
		key, deletedCount, deleteTime.String(), totalTime.String())

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
	startTime := time.Now()

	log.Printf("🔍 CACHE GET MULTIPLE START: key_count=%d, keys=%v", len(keys), keys)

	if len(keys) == 0 {
		log.Printf("⚠️  CACHE GET MULTIPLE EMPTY: no keys requested")
		return make(map[string]interface{}), nil
	}

	// Generate full keys
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.generateKey(key)
	}

	// Get all values
	mgetStart := time.Now()
	values, err := r.client.MGet(ctx, fullKeys...).Result()
	mgetTime := time.Since(mgetStart)

	if err != nil {
		totalTime := time.Since(startTime)
		log.Printf("❌ CACHE GET MULTIPLE ERROR: key_count=%d, error=%v, mget_time=%s, total_time=%s",
			len(keys), err, mgetTime.String(), totalTime.String())
		return nil, fmt.Errorf("failed to get multiple keys: %w", err)
	}

	// Parse results
	parseStart := time.Now()
	result := make(map[string]interface{})
	hitCount := 0
	missCount := 0

	for i, value := range values {
		if value != nil {
			var dest interface{}
			err := json.Unmarshal([]byte(value.(string)), &dest)
			if err == nil {
				result[keys[i]] = dest
				hitCount++
				// Update access count asynchronously
				go r.updateAccessCount(keys[i])
			} else {
				log.Printf("⚠️  CACHE GET MULTIPLE UNMARSHAL ERROR: key=%s, error=%v", keys[i], err)
				missCount++
			}
		} else {
			missCount++
		}
	}

	parseTime := time.Since(parseStart)
	totalTime := time.Since(startTime)

	log.Printf("✅ CACHE GET MULTIPLE SUCCESS: key_count=%d, hit_count=%d, miss_count=%d, hit_rate=%.2f%%, mget_time=%s, parse_time=%s, total_time=%s",
		len(keys), hitCount, missCount, float64(hitCount)/float64(len(keys))*100,
		mgetTime.String(), parseTime.String(), totalTime.String())

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
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	metadataKey := r.generateKey(fmt.Sprintf("meta:%s", key))
	accessCountKey := metadataKey + ":access_count"

	incrStart := time.Now()
	newCount, err := r.client.Incr(ctx, accessCountKey).Result()
	incrTime := time.Since(incrStart)
	totalTime := time.Since(startTime)

	if err != nil {
		log.Printf("⚠️  CACHE ACCESS COUNT ERROR: key=%s, access_key=%s, error=%v, time=%s",
			key, accessCountKey, err, totalTime.String())
	} else {
		log.Printf("📊 CACHE ACCESS COUNT UPDATED: key=%s, access_key=%s, new_count=%d, incr_time=%s, total_time=%s",
			key, accessCountKey, newCount, incrTime.String(), totalTime.String())
	}
}

// GetStats returns cache statistics
func (r *RedisService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	startTime := time.Now()

	log.Printf("📊 CACHE STATS START: getting Redis statistics")

	infoStart := time.Now()
	info, err := r.client.Info(ctx).Result()
	infoTime := time.Since(infoStart)

	if err != nil {
		totalTime := time.Since(startTime)
		log.Printf("❌ CACHE STATS ERROR: failed to get Redis info, error=%v, info_time=%s, total_time=%s",
			err, infoTime.String(), totalTime.String())
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Parse basic stats
	stats := map[string]interface{}{
		"info":        info,
		"key_prefix":  r.keyPrefix,
		"default_ttl": r.defaultTTL.String(),
		"info_size":   len(info),
	}

	totalTime := time.Since(startTime)
	log.Printf("✅ CACHE STATS SUCCESS: info_size=%d bytes, info_time=%s, total_time=%s",
		len(info), infoTime.String(), totalTime.String())

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
