package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		traceID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Set("trace_id", traceID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		requestID, _ := c.Get("request_id")
		traceID, _ := c.Get("trace_id")
		logger.Info("request",
			slog.String("request_id", requestID.(string)),
			slog.String("trace_id", traceID.(string)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
		)
	}
}
