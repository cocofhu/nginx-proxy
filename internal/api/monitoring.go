package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	// 检查数据库连接
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "database ping failed",
		})
		return
	}

	// 检查Nginx状态
	if err := h.nginxManager.TestConfig(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "nginx config test failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
	})
}
