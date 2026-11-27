package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/logging"
	"github.com/gin-gonic/gin"
)

// SlogRecovery records panic recovery as structured JSON instead of the
// standard library's text logger used by gin.Recovery().
func SlogRecovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		slog.Error("panic_recovered",
			slog.Any("panic", recovered),
			slog.String("stack", string(debug.Stack())),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	})
}

func requestLogger(c *gin.Context) *slog.Logger {
	if ctxLogger, exists := c.Get("logger"); exists {
		if logger, ok := ctxLogger.(*slog.Logger); ok {
			return logger
		}
	}
	attrs := []any{
		slog.String("method", c.Request.Method),
		slog.String("path", strings.ReplaceAll(strings.ReplaceAll(c.Request.URL.Path, "\n", ""), "\r", "")),
		slog.Any("query", logging.RedactedQuery(strings.ReplaceAll(strings.ReplaceAll(c.Request.URL.RawQuery, "\n", ""), "\r", ""))),
		slog.String("ip", strings.ReplaceAll(strings.ReplaceAll(c.ClientIP(), "\n", ""), "\r", "")),
		slog.String("user_agent", strings.ReplaceAll(strings.ReplaceAll(c.Request.UserAgent(), "\n", ""), "\r", "")),
	}
	if requestContext, ok := logging.FromContext(c.Request.Context()); ok {
		attrs = append(attrs, requestContext.LogAttrs()...)
	}
	return slog.Default().With(attrs...)
}

// GetLogger returns the request-scoped logger, adding request attributes when
// the middleware has not populated one yet.
func GetLogger(c *gin.Context) *slog.Logger {
	return requestLogger(c)
}

// SlogMiddleware writes the access log and handler errors as JSON.
func SlogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestContext := logging.NewRequestContext(c.Request)
		c.Request = c.Request.WithContext(requestContext.WithContext(c.Request.Context()))
		c.Header(logging.HeaderRequestID, requestContext.RequestID)
		c.Header(logging.HeaderTraceID, requestContext.TraceID)

		start := time.Now()
		logger := requestLogger(c)
		c.Set("logger", logger)

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logger.Error("http_error",
				slog.Int("status", c.Writer.Status()),
				slog.Any("error", err.Err),
			)
			c.JSON(c.Writer.Status(), gin.H{"error": err.Err.Error()})
			return
		}

		logger.Info("http_request",
			slog.Int("status", c.Writer.Status()),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
	}
}
