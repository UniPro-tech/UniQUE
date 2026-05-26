package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Ginのコンテキストからロガーを安全に取り出すためのヘルパー関数
func GetLogger(c *gin.Context) *slog.Logger {
	if ctxLogger, exists := c.Get("logger"); exists {
		if logger, ok := ctxLogger.(*slog.Logger); ok {
			return logger
		}
	}
	return slog.Default() // fallback
}

func SlogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// リクエスト固有の情報をあらかじめ埋め込んだロガーを作る
		logger := slog.Default().With(
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		)

		// Ginのコンテキストにロガーをセット
		c.Set("logger", logger)

		c.Next()

		// ハンドラーで c.AbortWithError() が呼ばれていた場合の共通エラー処理
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			status := c.Writer.Status()

			// 埋め込み済みのロガーを使うので、errorを足すだけでOK！
			logger.Error("An error occurred",
				slog.Int("status", status),
				slog.String("error", err.Error()),
			)

			c.JSON(status, gin.H{"error": err.Err.Error()})
			return
		}

		// 通常のアクセスログ（成功時）
		latency := time.Since(start)
		logger.Info("HTTP Request",
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", latency),
		)
	}
}
