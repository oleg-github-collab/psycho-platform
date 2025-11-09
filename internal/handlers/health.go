package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

func NewHealthHandler(db *sql.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Check(c *gin.Context) {
	health := map[string]interface{}{
		"status":   "healthy",
		"database": "unknown",
		"redis":    "unknown",
	}

	// Check database
	if err := h.db.Ping(); err != nil {
		health["database"] = "unhealthy"
		health["status"] = "degraded"
	} else {
		health["database"] = "healthy"
	}

	// Check Redis
	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		health["redis"] = "unhealthy"
		health["status"] = "degraded"
	} else {
		health["redis"] = "healthy"
	}

	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

func (h *HealthHandler) Ready(c *gin.Context) {
	// Check if database is ready
	if err := h.db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready": false,
			"error": "Database not ready",
		})
		return
	}

	// Check if Redis is ready
	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready": false,
			"error": "Redis not ready",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ready": true})
}
