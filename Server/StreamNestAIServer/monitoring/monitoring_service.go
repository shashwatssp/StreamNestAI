package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	httpRequestsTotal   map[string]int64
	httpRequestDuration map[string]float64
	httpRequestSize     map[string]int64
	httpResponseSize    map[string]int64

	// Business metrics
	userRegistrations map[string]int64
	userLogins        map[string]int64
	movieViews        map[string]int64
	movieRatings      map[string]int64
	watchlistAdds     map[string]int64

	// System metrics
	activeConnections   int64
	memoryUsage         int64
	goroutines          int64
	databaseConnections int64
	redisConnections    int64

	// Cache metrics
	cacheHits       map[string]int64
	cacheMisses     map[string]int64
	cacheOperations map[string]int64

	// Recommendation metrics
	recommendationRequests map[string]int64
	recommendationLatency  map[string]float64

	// Social metrics
	socialActions    map[string]int64
	followOperations map[string]int64

	// WebSocket metrics
	websocketConnections int64
	websocketMessages    map[string]int64

	// Error metrics
	errorsTotal map[string]int64
	errorRate   float64

	mu sync.RWMutex
}

// MonitoringService handles monitoring and metrics collection
type MonitoringService struct {
	metrics     *Metrics
	redisClient *redis.Client
	mongoDB     *mongo.Database
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex

	// Health check status
	healthStatus map[string]HealthStatus
}

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status    string                 `json:"status"` // healthy, unhealthy, degraded
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// MonitoringConfig holds configuration for the monitoring service
type MonitoringConfig struct {
	Enabled       bool
	Port          int
	Path          string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	MongoDB       *mongo.Database

	// Health check intervals
	HealthCheckInterval time.Duration

	// Metrics collection intervals
	MetricsInterval time.Duration

	// Alert thresholds
	ErrorRateThreshold    float64
	ResponseTimeThreshold time.Duration
	MemoryThreshold       float64
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(config MonitoringConfig) (*MonitoringService, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize Redis client for health checks
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Create metrics
	metrics := &Metrics{
		httpRequestsTotal:      make(map[string]int64),
		httpRequestDuration:    make(map[string]float64),
		httpRequestSize:        make(map[string]int64),
		httpResponseSize:       make(map[string]int64),
		userRegistrations:      make(map[string]int64),
		userLogins:             make(map[string]int64),
		movieViews:             make(map[string]int64),
		movieRatings:           make(map[string]int64),
		watchlistAdds:          make(map[string]int64),
		cacheHits:              make(map[string]int64),
		cacheMisses:            make(map[string]int64),
		cacheOperations:        make(map[string]int64),
		recommendationRequests: make(map[string]int64),
		recommendationLatency:  make(map[string]float64),
		socialActions:          make(map[string]int64),
		followOperations:       make(map[string]int64),
		websocketMessages:      make(map[string]int64),
		errorsTotal:            make(map[string]int64),
	}

	service := &MonitoringService{
		metrics:      metrics,
		redisClient:  redisClient,
		mongoDB:      config.MongoDB,
		ctx:          ctx,
		cancel:       cancel,
		healthStatus: make(map[string]HealthStatus),
	}

	// Start background processes
	go service.collectSystemMetrics()
	go service.performHealthChecks(config.HealthCheckInterval)

	return service, nil
}

// MetricsMiddleware returns Gin middleware for metrics collection
func (ms *MonitoringService) MetricsMiddleware() gin.HandlerFunc {
	if ms == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())
		key := fmt.Sprintf("%s_%s_%s", c.Request.Method, c.FullPath(), status)

		ms.metrics.mu.Lock()
		ms.metrics.httpRequestsTotal[key]++
		ms.metrics.httpRequestDuration[key] = duration
		if c.Request.ContentLength > 0 {
			ms.metrics.httpRequestSize[key] = int64(c.Request.ContentLength)
		}
		ms.metrics.httpResponseSize[key] = int64(c.Writer.Size())

		// Record errors
		if c.Writer.Status() >= 400 {
			errorKey := fmt.Sprintf("http_%s", c.FullPath())
			ms.metrics.errorsTotal[errorKey]++
		}
		ms.metrics.mu.Unlock()
	}
}

// RecordUserRegistration records a user registration event
func (ms *MonitoringService) RecordUserRegistration(source string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	ms.metrics.mu.Lock()
	ms.metrics.userRegistrations[source]++
	ms.metrics.mu.Unlock()
}

// RecordUserLogin records a user login event
func (ms *MonitoringService) RecordUserLogin(method, status string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	key := fmt.Sprintf("%s_%s", method, status)
	ms.metrics.mu.Lock()
	ms.metrics.userLogins[key]++
	ms.metrics.mu.Unlock()
}

// RecordMovieView records a movie view event
func (ms *MonitoringService) RecordMovieView(genre, source string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	key := fmt.Sprintf("%s_%s", genre, source)
	ms.metrics.mu.Lock()
	ms.metrics.movieViews[key]++
	ms.metrics.mu.Unlock()
}

// RecordMovieRating records a movie rating event
func (ms *MonitoringService) RecordMovieRating(ratingRange string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	ms.metrics.mu.Lock()
	ms.metrics.movieRatings[ratingRange]++
	ms.metrics.mu.Unlock()
}

// RecordWatchlistAdd records a watchlist add event
func (ms *MonitoringService) RecordWatchlistAdd(contentType string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	ms.metrics.mu.Lock()
	ms.metrics.watchlistAdds[contentType]++
	ms.metrics.mu.Unlock()
}

// RecordCacheHit records a cache hit
func (ms *MonitoringService) RecordCacheHit(cacheType string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	ms.metrics.mu.Lock()
	ms.metrics.cacheHits[cacheType]++
	ms.metrics.mu.Unlock()
}

// RecordCacheMiss records a cache miss
func (ms *MonitoringService) RecordCacheMiss(cacheType string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	ms.metrics.mu.Lock()
	ms.metrics.cacheMisses[cacheType]++
	ms.metrics.mu.Unlock()
}

// RecordCacheOperation records a cache operation
func (ms *MonitoringService) RecordCacheOperation(operation, cacheType string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	key := fmt.Sprintf("%s_%s", operation, cacheType)
	ms.metrics.mu.Lock()
	ms.metrics.cacheOperations[key]++
	ms.metrics.mu.Unlock()
}

// RecordRecommendationRequest records a recommendation request
func (ms *MonitoringService) RecordRecommendationRequest(algorithm, status string, duration time.Duration) {
	if ms == nil || ms.metrics == nil {
		return
	}
	requestKey := fmt.Sprintf("%s_%s", algorithm, status)
	latencyKey := algorithm

	ms.metrics.mu.Lock()
	ms.metrics.recommendationRequests[requestKey]++
	ms.metrics.recommendationLatency[latencyKey] = duration.Seconds()
	ms.metrics.mu.Unlock()
}

// RecordSocialAction records a social action
func (ms *MonitoringService) RecordSocialAction(action, status string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	key := fmt.Sprintf("%s_%s", action, status)
	ms.metrics.mu.Lock()
	ms.metrics.socialActions[key]++
	ms.metrics.mu.Unlock()
}

// RecordFollowOperation records a follow operation
func (ms *MonitoringService) RecordFollowOperation(operation, status string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	key := fmt.Sprintf("%s_%s", operation, status)
	ms.metrics.mu.Lock()
	ms.metrics.followOperations[key]++
	ms.metrics.mu.Unlock()
}

// IncrementWebSocketConnections increments WebSocket connection count
func (ms *MonitoringService) IncrementWebSocketConnections() {
	if ms == nil || ms.metrics == nil {
		return
	}
	ms.metrics.mu.Lock()
	ms.metrics.websocketConnections++
	ms.metrics.mu.Unlock()
}

// DecrementWebSocketConnections decrements WebSocket connection count
func (ms *MonitoringService) DecrementWebSocketConnections() {
	if ms == nil || ms.metrics == nil {
		return
	}
	ms.metrics.mu.Lock()
	ms.metrics.websocketConnections--
	ms.metrics.mu.Unlock()
}

// RecordWebSocketMessage records a WebSocket message
func (ms *MonitoringService) RecordWebSocketMessage(direction, messageType string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	key := fmt.Sprintf("%s_%s", direction, messageType)
	ms.metrics.mu.Lock()
	ms.metrics.websocketMessages[key]++
	ms.metrics.mu.Unlock()
}

// RecordError records an error
func (ms *MonitoringService) RecordError(errorType, component string) {
	if ms == nil || ms.metrics == nil {
		return
	}
	key := fmt.Sprintf("%s_%s", errorType, component)
	ms.metrics.mu.Lock()
	ms.metrics.errorsTotal[key]++
	ms.metrics.mu.Unlock()
}

// GetMetricsHandler returns the metrics handler
func (ms *MonitoringService) GetMetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if ms == nil || ms.metrics == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		summary := ms.GetMetricsSummary()
		json.NewEncoder(w).Encode(summary)
	})
}

// GetHealthStatus returns the current health status
func (ms *MonitoringService) GetHealthStatus() map[string]HealthStatus {
	if ms == nil {
		return map[string]HealthStatus{}
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Return a copy to avoid concurrent access issues
	status := make(map[string]HealthStatus)
	for k, v := range ms.healthStatus {
		status[k] = v
	}

	return status
}

// Background processes

func (ms *MonitoringService) collectSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ms.updateSystemMetrics()
		case <-ms.ctx.Done():
			return
		}
	}
}

func (ms *MonitoringService) updateSystemMetrics() {
	if ms == nil || ms.metrics == nil {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ms.metrics.mu.Lock()
	ms.metrics.memoryUsage = int64(m.Alloc)
	ms.metrics.goroutines = int64(runtime.NumGoroutine())

	// Update database connections (placeholder - would need actual connection pool monitoring)
	ms.metrics.databaseConnections = 10 // Example value

	// Update Redis connections
	if ms.redisClient != nil {
		// This would require Redis connection pool monitoring
		ms.metrics.redisConnections = 5 // Example value
	}
	ms.metrics.mu.Unlock()
}

func (ms *MonitoringService) performHealthChecks(interval time.Duration) {
	if interval == 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ms.runHealthChecks()
		case <-ms.ctx.Done():
			return
		}
	}
}

func (ms *MonitoringService) runHealthChecks() {
	if ms == nil {
		return
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check Redis health
	if ms.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := ms.redisClient.Ping(ctx).Err()
		cancel()

		if err != nil {
			ms.healthStatus["redis"] = HealthStatus{
				Status:    "unhealthy",
				Message:   fmt.Sprintf("Redis connection failed: %v", err),
				Timestamp: time.Now(),
			}
		} else {
			ms.healthStatus["redis"] = HealthStatus{
				Status:    "healthy",
				Message:   "Redis connection successful",
				Timestamp: time.Now(),
			}
		}
	}

	// Check MongoDB health
	if ms.mongoDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// Use a simple ping command instead of Ping() method
		err := ms.mongoDB.RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
		cancel()

		if err != nil {
			ms.healthStatus["mongodb"] = HealthStatus{
				Status:    "unhealthy",
				Message:   fmt.Sprintf("MongoDB connection failed: %v", err),
				Timestamp: time.Now(),
			}
		} else {
			ms.healthStatus["mongodb"] = HealthStatus{
				Status:    "healthy",
				Message:   "MongoDB connection successful",
				Timestamp: time.Now(),
			}
		}
	}

	// Check memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsageMB := float64(m.Alloc) / 1024 / 1024

	if memoryUsageMB > 500 { // 500MB threshold
		ms.healthStatus["memory"] = HealthStatus{
			Status:    "degraded",
			Message:   fmt.Sprintf("High memory usage: %.2f MB", memoryUsageMB),
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"memory_mb": memoryUsageMB,
			},
		}
	} else {
		ms.healthStatus["memory"] = HealthStatus{
			Status:    "healthy",
			Message:   fmt.Sprintf("Memory usage normal: %.2f MB", memoryUsageMB),
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"memory_mb": memoryUsageMB,
			},
		}
	}

	// Overall system health
	overallStatus := "healthy"
	unhealthyCount := 0
	degradedCount := 0

	for _, status := range ms.healthStatus {
		switch status.Status {
		case "unhealthy":
			unhealthyCount++
		case "degraded":
			degradedCount++
		}
	}

	if unhealthyCount > 0 {
		overallStatus = "unhealthy"
	} else if degradedCount > 0 {
		overallStatus = "degraded"
	}

	ms.healthStatus["system"] = HealthStatus{
		Status:    overallStatus,
		Message:   fmt.Sprintf("System status: %d unhealthy, %d degraded components", unhealthyCount, degradedCount),
		Timestamp: time.Now(),
	}
}

// Close stops the monitoring service
func (ms *MonitoringService) Close() error {
	if ms == nil {
		return nil
	}

	ms.cancel()
	ms.wg.Wait()

	if ms.redisClient != nil {
		return ms.redisClient.Close()
	}

	return nil
}

// GetMetricsSummary returns a summary of current metrics
func (ms *MonitoringService) GetMetricsSummary() map[string]interface{} {
	if ms == nil {
		return map[string]interface{}{}
	}

	ms.metrics.mu.RLock()
	defer ms.metrics.mu.RUnlock()

	return map[string]interface{}{
		"http_requests_total":   ms.metrics.httpRequestsTotal,
		"user_registrations":    ms.metrics.userRegistrations,
		"movie_views":           ms.metrics.movieViews,
		"cache_hits":            ms.metrics.cacheHits,
		"cache_misses":          ms.metrics.cacheMisses,
		"websocket_connections": ms.metrics.websocketConnections,
		"errors_total":          ms.metrics.errorsTotal,
		"active_connections":    ms.metrics.activeConnections,
		"memory_usage":          ms.metrics.memoryUsage,
		"goroutines":            ms.metrics.goroutines,
	}
}
