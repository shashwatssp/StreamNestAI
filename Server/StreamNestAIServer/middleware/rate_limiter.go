package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	// Basic rate limiting
	RequestsPerSecond int
	BurstSize         int

	// Redis configuration for distributed rate limiting
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// Key generation
	KeyGenerator func(*gin.Context) string

	// Custom response handler
	OnLimitReached func(*gin.Context, *RateLimitInfo)

	// Time window for sliding window counter
	WindowDuration time.Duration

	// Enable distributed rate limiting
	EnableDistributed bool
}

// RateLimitInfo contains information about the current rate limit status
type RateLimitInfo struct {
	Limit      int       `json:"limit"`
	Remaining  int       `json:"remaining"`
	ResetTime  time.Time `json:"reset_time"`
	RetryAfter int       `json:"retry_after,omitempty"`
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	capacity   int
	tokens     int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity int, refillRate time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed / tb.refillRate)
	if tokensToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// SlidingWindowCounter implements sliding window rate limiting
type SlidingWindowCounter struct {
	windowDuration time.Duration
	maxRequests    int
	requests       []time.Time
	mu             sync.Mutex
}

// NewSlidingWindowCounter creates a new sliding window counter
func NewSlidingWindowCounter(windowDuration time.Duration, maxRequests int) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		windowDuration: windowDuration,
		maxRequests:    maxRequests,
		requests:       make([]time.Time, 0),
	}
}

// Allow checks if a request is allowed
func (swc *SlidingWindowCounter) Allow() bool {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-swc.windowDuration)

	// Remove old requests outside the window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range swc.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	swc.requests = validRequests

	if len(swc.requests) < swc.maxRequests {
		swc.requests = append(swc.requests, now)
		return true
	}

	return false
}

// RedisRateLimiter implements distributed rate limiting using Redis
type RedisRateLimiter struct {
	client         *redis.Client
	windowDuration time.Duration
	maxRequests    int
	keyPrefix      string
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client, windowDuration time.Duration, maxRequests int, keyPrefix string) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:         client,
		windowDuration: windowDuration,
		maxRequests:    maxRequests,
		keyPrefix:      keyPrefix,
	}
}

// Allow checks if a request is allowed using Redis sliding window
func (rrl *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, *RateLimitInfo, error) {
	now := time.Now().Unix()
	windowStart := now - int64(rrl.windowDuration.Seconds())

	redisKey := fmt.Sprintf("%s:%s", rrl.keyPrefix, key)

	// Use Redis pipeline for atomic operations
	pipe := rrl.client.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, redisKey, "0", strconv.FormatInt(windowStart, 10))

	// Count current requests
	countCmd := pipe.ZCard(ctx, redisKey)

	// Add current request
	pipe.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(now),
		Member: now,
	})

	// Set expiration
	pipe.Expire(ctx, redisKey, rrl.windowDuration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("redis pipeline error: %w", err)
	}

	currentCount := countCmd.Val()
	remaining := max(0, rrl.maxRequests-int(currentCount))
	allowed := currentCount < int64(rrl.maxRequests)

	resetTime := time.Now().Add(rrl.windowDuration)

	info := &RateLimitInfo{
		Limit:     rrl.maxRequests,
		Remaining: remaining,
		ResetTime: resetTime,
	}

	return allowed, info, nil
}

// RateLimiter is the main rate limiter middleware
type RateLimiter struct {
	config        RateLimiterConfig
	localBuckets  map[string]*TokenBucket
	localCounters map[string]*SlidingWindowCounter
	redisLimiter  *RedisRateLimiter
	mu            sync.RWMutex
}

// NewRateLimiter creates a new rate limiter middleware
func NewRateLimiter(config RateLimiterConfig) (*RateLimiter, error) {
	rl := &RateLimiter{
		config:        config,
		localBuckets:  make(map[string]*TokenBucket),
		localCounters: make(map[string]*SlidingWindowCounter),
	}

	// Set default key generator if not provided
	if config.KeyGenerator == nil {
		config.KeyGenerator = func(c *gin.Context) string {
			return c.ClientIP()
		}
		rl.config.KeyGenerator = config.KeyGenerator
	}

	// Set default response handler if not provided
	if config.OnLimitReached == nil {
		config.OnLimitReached = func(c *gin.Context, info *RateLimitInfo) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"limit": info.Limit,
				"reset": info.ResetTime,
			})
		}
		rl.config.OnLimitReached = config.OnLimitReached
	}

	// Initialize Redis rate limiter if distributed is enabled
	if config.EnableDistributed {
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		})

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := client.Ping(ctx).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Redis: %w", err)
		}

		rl.redisLimiter = NewRedisRateLimiter(
			client,
			config.WindowDuration,
			config.RequestsPerSecond,
			"rate_limit",
		)
	}

	return rl, nil
}

// Middleware returns the Gin middleware function
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := rl.config.KeyGenerator(c)

		var allowed bool
		var info *RateLimitInfo
		var err error

		if rl.config.EnableDistributed && rl.redisLimiter != nil {
			// Use distributed rate limiting
			allowed, info, err = rl.redisLimiter.Allow(c.Request.Context(), key)
			if err != nil {
				// Fallback to local rate limiting on Redis error
				allowed, info = rl.allowLocal(key)
			}
		} else {
			// Use local rate limiting
			allowed, info = rl.allowLocal(key)
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(info.ResetTime.Unix(), 10))

		if !allowed {
			if info.RetryAfter > 0 {
				c.Header("Retry-After", strconv.Itoa(info.RetryAfter))
			}
			rl.config.OnLimitReached(c, info)
			c.Abort()
			return
		}

		c.Next()
	}
}

// allowLocal performs local rate limiting
func (rl *RateLimiter) allowLocal(key string) (bool, *RateLimitInfo) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Use token bucket for per-second limiting
	bucket, exists := rl.localBuckets[key]
	if !exists {
		bucket = NewTokenBucket(rl.config.BurstSize, time.Second/time.Duration(rl.config.RequestsPerSecond))
		rl.localBuckets[key] = bucket
	}

	allowed := bucket.Allow()

	// Calculate remaining tokens
	remaining := max(0, bucket.tokens)

	info := &RateLimitInfo{
		Limit:     rl.config.RequestsPerSecond,
		Remaining: remaining,
		ResetTime: time.Now().Add(time.Second),
	}

	if !allowed {
		info.RetryAfter = 1
	}

	return allowed, info
}

// Cleanup removes old entries from local rate limiters
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// This is a simple cleanup - in production, you might want more sophisticated cleanup logic
	if len(rl.localBuckets) > 10000 {
		// Clear half of the buckets
		count := 0
		for k := range rl.localBuckets {
			delete(rl.localBuckets, k)
			count++
			if count >= 5000 {
				break
			}
		}
	}
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Predefined key generators

// IPKeyGenerator generates rate limit keys based on client IP
func IPKeyGenerator(c *gin.Context) string {
	return c.ClientIP()
}

// UserKeyGenerator generates rate limit keys based on user ID
func UserKeyGenerator(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return c.ClientIP()
	}
	return fmt.Sprintf("user:%v", userID)
}

// APIKeyGenerator generates rate limit keys based on API key
func APIKeyGenerator(c *gin.Context) string {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return c.ClientIP()
	}
	return fmt.Sprintf("api:%s", apiKey)
}

// CompositeKeyGenerator generates composite keys from multiple sources
func CompositeKeyGenerator(sources ...func(*gin.Context) string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		var parts []string
		for _, source := range sources {
			parts = append(parts, source(c))
		}
		return strings.Join(parts, ":")
	}
}

// Rate limit presets

// APIRateLimit returns a rate limiter for API endpoints
func APIRateLimit(requestsPerMinute int) (*RateLimiter, error) {
	return NewRateLimiter(RateLimiterConfig{
		RequestsPerSecond: requestsPerMinute / 60,
		BurstSize:         requestsPerMinute / 30,
		WindowDuration:    time.Minute,
		KeyGenerator:      APIKeyGenerator,
	})
}

// UserRateLimit returns a rate limiter for user-specific endpoints
func UserRateLimit(requestsPerMinute int) (*RateLimiter, error) {
	return NewRateLimiter(RateLimiterConfig{
		RequestsPerSecond: requestsPerMinute / 60,
		BurstSize:         requestsPerMinute / 30,
		WindowDuration:    time.Minute,
		KeyGenerator:      UserKeyGenerator,
	})
}

// GlobalRateLimit returns a global rate limiter
func GlobalRateLimit(requestsPerSecond int) (*RateLimiter, error) {
	return NewRateLimiter(RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         requestsPerSecond * 2,
		WindowDuration:    time.Second,
		KeyGenerator:      IPKeyGenerator,
	})
}

// DistributedRateLimit returns a distributed rate limiter using Redis
func DistributedRateLimit(redisHost, redisPort, redisPassword string, redisDB int, requestsPerMinute int) (*RateLimiter, error) {
	return NewRateLimiter(RateLimiterConfig{
		RequestsPerSecond: requestsPerMinute / 60,
		BurstSize:         requestsPerMinute / 30,
		WindowDuration:    time.Minute,
		RedisHost:         redisHost,
		RedisPort:         redisPort,
		RedisPassword:     redisPassword,
		RedisDB:           redisDB,
		KeyGenerator:      UserKeyGenerator,
		EnableDistributed: true,
	})
}
