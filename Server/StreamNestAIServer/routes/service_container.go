package routes

import (
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/analytics"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/middleware"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/monitoring"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/recommendations"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/security"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/social"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/websocket"
)

// ServiceContainer holds all the services for dependency injection
type ServiceContainer struct {
	MongoClient           *mongo.Client
	WebSocketHub          *websocket.Hub
	AnalyticsService      *analytics.AnalyticsService
	RecommendationService *recommendations.RecommendationService
	SocialService         *social.SocialService
	SecurityService       *security.SecurityService
	MonitoringService     *monitoring.MonitoringService
	RateLimiter           *middleware.RateLimiter
	RedisAddr             string
}

// Global service container instance
var globalServiceContainer *ServiceContainer

// NewServiceContainer creates a new service container
func NewServiceContainer(
	client *mongo.Client,
	wsHub *websocket.Hub,
	analyticsService *analytics.AnalyticsService,
	recommendationService *recommendations.RecommendationService,
	socialService *social.SocialService,
	securityService *security.SecurityService,
	monitoringService *monitoring.MonitoringService,
	rateLimiter *middleware.RateLimiter,
	redisAddr string,
) *ServiceContainer {
	return &ServiceContainer{
		MongoClient:           client,
		WebSocketHub:          wsHub,
		AnalyticsService:      analyticsService,
		RecommendationService: recommendationService,
		SocialService:         socialService,
		SecurityService:       securityService,
		MonitoringService:     monitoringService,
		RateLimiter:           rateLimiter,
		RedisAddr:             redisAddr,
	}
}

// SetServiceContainer sets the global service container instance
func SetServiceContainer(sc *ServiceContainer) {
	globalServiceContainer = sc
}

// GetServiceContainer returns the global service container instance
func GetServiceContainer() *ServiceContainer {
	return globalServiceContainer
}

// GetMongoClient returns the MongoDB client
func (sc *ServiceContainer) GetMongoClient() *mongo.Client {
	return sc.MongoClient
}

// GetWebSocketHub returns the WebSocket hub
func (sc *ServiceContainer) GetWebSocketHub() *websocket.Hub {
	return sc.WebSocketHub
}

// GetAnalyticsService returns the analytics service
func (sc *ServiceContainer) GetAnalyticsService() *analytics.AnalyticsService {
	return sc.AnalyticsService
}

// GetRecommendationService returns the recommendation service
func (sc *ServiceContainer) GetRecommendationService() *recommendations.RecommendationService {
	return sc.RecommendationService
}

// GetSocialService returns the social service
func (sc *ServiceContainer) GetSocialService() *social.SocialService {
	return sc.SocialService
}

// GetSecurityService returns the security service
func (sc *ServiceContainer) GetSecurityService() *security.SecurityService {
	return sc.SecurityService
}

// GetMonitoringService returns the monitoring service
func (sc *ServiceContainer) GetMonitoringService() *monitoring.MonitoringService {
	return sc.MonitoringService
}

// GetRedisAddr returns the Redis address
func (sc *ServiceContainer) GetRedisAddr() string {
	return sc.RedisAddr
}

// IsHealthy checks if all services are healthy
func (sc *ServiceContainer) IsHealthy() map[string]bool {
	health := make(map[string]bool)

	// Check MongoDB
	if sc.MongoClient != nil {
		health["mongodb"] = true
	} else {
		health["mongodb"] = false
	}

	// Check WebSocket Hub
	if sc.WebSocketHub != nil {
		health["websocket"] = true
	} else {
		health["websocket"] = false
	}

	// Check Analytics Service
	if sc.AnalyticsService != nil {
		health["analytics"] = true
	} else {
		health["analytics"] = false
	}

	// Check Recommendation Service
	if sc.RecommendationService != nil {
		health["recommendations"] = true
	} else {
		health["recommendations"] = false
	}

	// Check Social Service
	if sc.SocialService != nil {
		health["social"] = true
	} else {
		health["social"] = false
	}

	// Check Security Service
	if sc.SecurityService != nil {
		health["security"] = true
	} else {
		health["security"] = false
	}

	// Check Monitoring Service
	if sc.MonitoringService != nil {
		health["monitoring"] = true
	} else {
		health["monitoring"] = false
	}

	return health
}

// GetServiceStatus returns detailed status of all services
func (sc *ServiceContainer) GetServiceStatus() map[string]interface{} {
	status := make(map[string]interface{})

	status["mongodb"] = map[string]interface{}{
		"connected": sc.MongoClient != nil,
		"type":      "database",
	}

	status["websocket"] = map[string]interface{}{
		"initialized": sc.WebSocketHub != nil,
		"type":        "realtime",
	}

	status["analytics"] = map[string]interface{}{
		"initialized": sc.AnalyticsService != nil,
		"type":        "processing",
	}

	status["recommendations"] = map[string]interface{}{
		"initialized": sc.RecommendationService != nil,
		"type":        "ml",
	}

	status["social"] = map[string]interface{}{
		"initialized": sc.SocialService != nil,
		"type":        "social",
	}

	status["security"] = map[string]interface{}{
		"initialized": sc.SecurityService != nil,
		"type":        "security",
	}

	status["monitoring"] = map[string]interface{}{
		"initialized": sc.MonitoringService != nil,
		"type":        "observability",
	}

	status["redis"] = map[string]interface{}{
		"address": sc.RedisAddr,
		"type":    "cache",
	}

	return status
}
