package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shashwatssp/StreamNestAI/Server/StreamNestAIServer/monitoring"
)

// MonitoringController handles monitoring endpoints
type MonitoringController struct {
	monitoringService *monitoring.MonitoringService
}

// NewMonitoringController creates a new monitoring controller
func NewMonitoringController(monitoringService *monitoring.MonitoringService) *MonitoringController {
	return &MonitoringController{
		monitoringService: monitoringService,
	}
}

// GetHealth returns the health status of all services
func (mc *MonitoringController) GetHealth(c *gin.Context) {
	if mc.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"message": "Monitoring service not available",
		})
		return
	}

	healthStatus := mc.monitoringService.GetHealthStatus()

	// Determine overall health
	overallStatus := "healthy"
	for _, status := range healthStatus {
		if status.Status == "unhealthy" {
			overallStatus = "unhealthy"
			break
		} else if status.Status == "degraded" {
			overallStatus = "degraded"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    overallStatus,
		"timestamp": time.Now(),
		"services":  healthStatus,
	})
}

// GetMetrics returns application metrics
func (mc *MonitoringController) GetMetrics(c *gin.Context) {
	if mc.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	// Get metrics summary
	metrics := mc.monitoringService.GetMetricsSummary()

	c.JSON(http.StatusOK, gin.H{
		"timestamp": time.Now(),
		"metrics":   metrics,
	})
}

// GetMetricsHandler returns the Prometheus metrics handler
func (mc *MonitoringController) GetMetricsHandler(c *gin.Context) {
	if mc.monitoringService == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	handler := mc.monitoringService.GetMetricsHandler()
	handler.ServeHTTP(c.Writer, c.Request)
}

// GetSystemInfo returns system information
func (mc *MonitoringController) GetSystemInfo(c *gin.Context) {
	if mc.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	// Get system metrics
	metrics := mc.monitoringService.GetMetricsSummary()

	systemInfo := gin.H{
		"timestamp": time.Now(),
		"system": gin.H{
			"active_connections": metrics["active_connections"],
			"memory_usage":       metrics["memory_usage"],
			"goroutines":         metrics["goroutines"],
		},
		"websocket": gin.H{
			"connections": metrics["websocket_connections"],
		},
		"cache": gin.H{
			"hits":   metrics["cache_hits"],
			"misses": metrics["cache_misses"],
		},
		"errors": gin.H{
			"total": metrics["errors_total"],
		},
	}

	c.JSON(http.StatusOK, systemInfo)
}

// GetBusinessMetrics returns business-related metrics
func (mc *MonitoringController) GetBusinessMetrics(c *gin.Context) {
	if mc.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	// Get metrics summary
	metrics := mc.monitoringService.GetMetricsSummary()

	businessMetrics := gin.H{
		"timestamp": time.Now(),
		"users": gin.H{
			"registrations": metrics["user_registrations"],
			"logins":        metrics["user_logins"],
		},
		"content": gin.H{
			"movie_views":    metrics["movie_views"],
			"movie_ratings":  metrics["movie_ratings"],
			"watchlist_adds": metrics["watchlist_adds"],
		},
		"social": gin.H{
			"actions":           metrics["social_actions"],
			"follow_operations": metrics["follow_operations"],
		},
		"recommendations": gin.H{
			"requests": metrics["recommendation_requests"],
		},
	}

	c.JSON(http.StatusOK, businessMetrics)
}

// GetPerformanceMetrics returns performance-related metrics
func (mc *MonitoringController) GetPerformanceMetrics(c *gin.Context) {
	if mc.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	// Get metrics summary
	metrics := mc.monitoringService.GetMetricsSummary()

	performanceMetrics := gin.H{
		"timestamp": time.Now(),
		"http": gin.H{
			"requests_total":   metrics["http_requests_total"],
			"request_duration": metrics["http_request_duration"],
			"request_size":     metrics["http_request_size"],
			"response_size":    metrics["http_response_size"],
		},
		"cache": gin.H{
			"hits":       metrics["cache_hits"],
			"misses":     metrics["cache_misses"],
			"operations": metrics["cache_operations"],
		},
		"recommendations": gin.H{
			"latency": metrics["recommendation_latency"],
		},
	}

	c.JSON(http.StatusOK, performanceMetrics)
}

// GetWebSocketMetrics returns WebSocket-related metrics
func (mc *MonitoringController) GetWebSocketMetrics(c *gin.Context) {
	if mc.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	// Get metrics summary
	metrics := mc.monitoringService.GetMetricsSummary()

	wsMetrics := gin.H{
		"timestamp":   time.Now(),
		"connections": metrics["websocket_connections"],
		"messages":    metrics["websocket_messages"],
	}

	c.JSON(http.StatusOK, wsMetrics)
}

// GetErrorMetrics returns error-related metrics
func (mc *MonitoringController) GetErrorMetrics(c *gin.Context) {
	if mc.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	// Get metrics summary
	metrics := mc.monitoringService.GetMetricsSummary()

	errorMetrics := gin.H{
		"timestamp":    time.Now(),
		"errors_total": metrics["errors_total"],
		"error_rate":   metrics["error_rate"],
	}

	c.JSON(http.StatusOK, errorMetrics)
}

// TriggerHealthCheck triggers an immediate health check
func (mc *MonitoringController) TriggerHealthCheck(c *gin.Context) {
	if mc.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	// Get current health status (health checks run automatically in background)
	healthStatus := mc.monitoringService.GetHealthStatus()

	c.JSON(http.StatusOK, gin.H{
		"message":   "Health check completed",
		"timestamp": time.Now(),
		"services":  healthStatus,
	})
}

// GetServiceStatus returns the status of a specific service
func (mc *MonitoringController) GetServiceStatus(c *gin.Context) {
	if mc.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	serviceName := c.Param("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service name is required",
		})
		return
	}

	healthStatus := mc.monitoringService.GetHealthStatus()
	serviceStatus, exists := healthStatus[serviceName]

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Service not found",
			"service": serviceName,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service":   serviceName,
		"status":    serviceStatus,
		"timestamp": time.Now(),
	})
}

// GetMetricsHistory returns metrics history (placeholder for future implementation)
func (mc *MonitoringController) GetMetricsHistory(c *gin.Context) {
	// Parse query parameters
	durationStr := c.DefaultQuery("duration", "1h")
	intervalStr := c.DefaultQuery("interval", "1m")

	// Parse duration
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid duration format",
			"example": "1h, 30m, 1d",
		})
		return
	}

	// Parse interval
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid interval format",
			"example": "1m, 5m, 1h",
		})
		return
	}

	// This is a placeholder implementation
	// In a real implementation, you would query a time-series database
	// like Prometheus, InfluxDB, or your own metrics storage

	c.JSON(http.StatusOK, gin.H{
		"message":   "Metrics history endpoint",
		"duration":  duration.String(),
		"interval":  interval.String(),
		"note":      "This endpoint requires a time-series database for full implementation",
		"timestamp": time.Now(),
	})
}

// GetAlerts returns active alerts (placeholder for future implementation)
func (mc *MonitoringController) GetAlerts(c *gin.Context) {
	// Parse query parameters
	status := c.DefaultQuery("status", "active")
	severity := c.Query("severity")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	// This is a placeholder implementation
	// In a real implementation, you would query your alerting system
	// like Prometheus AlertManager, PagerDuty, or your own alert system

	c.JSON(http.StatusOK, gin.H{
		"message":   "Alerts endpoint",
		"status":    status,
		"severity":  severity,
		"limit":     limit,
		"alerts":    []gin.H{}, // Empty array for now
		"note":      "This endpoint requires an alerting system for full implementation",
		"timestamp": time.Now(),
	})
}

// GetDashboardData returns consolidated dashboard data
func (mc *MonitoringController) GetDashboardData(c *gin.Context) {
	if mc.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	// Get all metrics
	metrics := mc.monitoringService.GetMetricsSummary()
	healthStatus := mc.monitoringService.GetHealthStatus()

	// Determine overall health
	overallStatus := "healthy"
	for _, status := range healthStatus {
		if status.Status == "unhealthy" {
			overallStatus = "unhealthy"
			break
		} else if status.Status == "degraded" {
			overallStatus = "degraded"
		}
	}

	dashboardData := gin.H{
		"timestamp": time.Now(),
		"overview": gin.H{
			"status":                overallStatus,
			"active_connections":    metrics["active_connections"],
			"websocket_connections": metrics["websocket_connections"],
			"total_errors":          metrics["errors_total"],
		},
		"services": healthStatus,
		"metrics": gin.H{
			"http_requests": metrics["http_requests_total"],
			"user_activity": gin.H{
				"registrations": metrics["user_registrations"],
				"logins":        metrics["user_logins"],
			},
			"content_activity": gin.H{
				"movie_views":   metrics["movie_views"],
				"movie_ratings": metrics["movie_ratings"],
			},
			"cache_performance": gin.H{
				"hits":   metrics["cache_hits"],
				"misses": metrics["cache_misses"],
			},
		},
		"system": gin.H{
			"memory_usage": metrics["memory_usage"],
			"goroutines":   metrics["goroutines"],
		},
	}

	c.JSON(http.StatusOK, dashboardData)
}

// GetMonitoringMiddleware returns the metrics middleware
func (mc *MonitoringController) GetMonitoringMiddleware() gin.HandlerFunc {
	if mc.monitoringService == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return mc.monitoringService.MetricsMiddleware()
}
