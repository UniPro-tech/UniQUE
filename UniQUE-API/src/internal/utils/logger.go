package utils

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// Ginのコンテキストからロガーを安全に取り出すためのヘルパー関数
func GetLogger(c *gin.Context) *slog.Logger {
	if ctxLogger, exists := c.Get("logger"); exists {
		if logger, ok := ctxLogger.(*slog.Logger); ok {
			return logger
		}
	}
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
	return logger
}
