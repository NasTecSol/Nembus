package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware provides structured request logging
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate request duration
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Build log message
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request details
		if errorMessage != "" {
			log.Printf("[%s] %s | %3d | %13v | %15s | %s | %s",
				time.Now().Format("2006/01/02 - 15:04:05"),
				method,
				statusCode,
				latency,
				clientIP,
				path,
				errorMessage,
			)
		} else {
			log.Printf("[%s] %s | %3d | %13v | %15s | %s",
				time.Now().Format("2006/01/02 - 15:04:05"),
				method,
				statusCode,
				latency,
				clientIP,
				path,
			)
		}
	}
}
