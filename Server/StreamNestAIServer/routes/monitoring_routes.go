package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/controllers"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/middleware"
)

// SetupMonitoringRoutes configures monitoring-related routes
func SetupMonitoringRoutes(router *gin.Engine, monitoringController *controllers.MonitoringController) {
	// Monitoring group
	monitoringGroup := router.Group("/api/v1/monitoring")
	{
		// Health check endpoints
		monitoringGroup.GET("/health", monitoringController.GetHealth)
		monitoringGroup.POST("/health/check", monitoringController.TriggerHealthCheck)
		monitoringGroup.GET("/health/:service", monitoringController.GetServiceStatus)

		// Metrics endpoints
		monitoringGroup.GET("/metrics", monitoringController.GetMetrics)
		monitoringGroup.GET("/metrics/prometheus", monitoringController.GetMetricsHandler)
		monitoringGroup.GET("/metrics/system", monitoringController.GetSystemInfo)
		monitoringGroup.GET("/metrics/business", monitoringController.GetBusinessMetrics)
		monitoringGroup.GET("/metrics/performance", monitoringController.GetPerformanceMetrics)
		monitoringGroup.GET("/metrics/websocket", monitoringController.GetWebSocketMetrics)
		monitoringGroup.GET("/metrics/errors", monitoringController.GetErrorMetrics)

		// Advanced monitoring endpoints
		monitoringGroup.GET("/metrics/history", monitoringController.GetMetricsHistory)
		monitoringGroup.GET("/alerts", monitoringController.GetAlerts)
		monitoringGroup.GET("/dashboard", monitoringController.GetDashboardData)
	}

	// Admin monitoring group (with authentication)
	adminMonitoringGroup := router.Group("/api/v1/admin/monitoring")
	adminMonitoringGroup.Use(middleware.AuthMiddleWare())
	{
		// Protected health and metrics endpoints
		adminMonitoringGroup.GET("/health", monitoringController.GetHealth)
		adminMonitoringGroup.POST("/health/check", monitoringController.TriggerHealthCheck)
		adminMonitoringGroup.GET("/health/:service", monitoringController.GetServiceStatus)

		// Protected metrics endpoints
		adminMonitoringGroup.GET("/metrics", monitoringController.GetMetrics)
		adminMonitoringGroup.GET("/metrics/prometheus", monitoringController.GetMetricsHandler)
		adminMonitoringGroup.GET("/metrics/system", monitoringController.GetSystemInfo)
		adminMonitoringGroup.GET("/metrics/business", monitoringController.GetBusinessMetrics)
		adminMonitoringGroup.GET("/metrics/performance", monitoringController.GetPerformanceMetrics)
		adminMonitoringGroup.GET("/metrics/websocket", monitoringController.GetWebSocketMetrics)
		adminMonitoringGroup.GET("/metrics/errors", monitoringController.GetErrorMetrics)

		// Protected advanced monitoring endpoints
		adminMonitoringGroup.GET("/metrics/history", monitoringController.GetMetricsHistory)
		adminMonitoringGroup.GET("/alerts", monitoringController.GetAlerts)
		adminMonitoringGroup.GET("/dashboard", monitoringController.GetDashboardData)
	}

	// Legacy health check endpoint for compatibility
	router.GET("/health", monitoringController.GetHealth)
	router.GET("/healthz", monitoringController.GetHealth)

	// Legacy metrics endpoint for compatibility
	router.GET("/metrics", monitoringController.GetMetricsHandler)
}

// SetupMonitoringMiddleware adds monitoring middleware to the router
func SetupMonitoringMiddleware(router *gin.Engine, monitoringController *controllers.MonitoringController) {
	if monitoringController != nil {
		// Add metrics middleware to all routes
		// Note: Metrics middleware should be added at the application level
		// This is handled in the main.go file
	}
}
