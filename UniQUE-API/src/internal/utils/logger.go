package utils

import (
	"log/slog"
	"strings"

	"github.com/UniPro-tech/UniQUE-API/internal/logging"
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
	attrs := []any{
		slog.String("method", c.Request.Method),
		slog.String("path", strings.ReplaceAll(strings.ReplaceAll(path, "\n", ""), "\r", "")),
		slog.Any("query", RedactedURL(strings.ReplaceAll(strings.ReplaceAll(query, "\n", ""), "\r", ""))),
		slog.String("ip", strings.ReplaceAll(strings.ReplaceAll(c.ClientIP(), "\n", ""), "\r", "")),
		slog.String("user_agent", strings.ReplaceAll(strings.ReplaceAll(c.Request.UserAgent(), "\n", ""), "\r", "")),
	}
	if requestContext, ok := logging.FromContext(c.Request.Context()); ok {
		attrs = append(attrs, requestContext.LogAttrs()...)
	}
	return slog.Default().With(attrs...)
}
