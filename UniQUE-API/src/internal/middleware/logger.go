package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/logging"
	"github.com/UniPro-tech/UniQUE-API/internal/utils"
	"github.com/gin-gonic/gin"
)

// SlogRecovery は panic を slog 形式で記録し、500 を返すミドルウェア。
// Gin 標準の Recovery は標準 log パッケージへ出力するため置き換える。
func SlogRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic_recovered",
			slog.Any("panic", recovered),
			slog.String("stack", string(debug.Stack())),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	})
}

// Gin のアクセスログを slog で JSON 出力するためのミドルウェア
func SlogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestContext := logging.NewRequestContext(c.Request)
		c.Request = c.Request.WithContext(requestContext.WithContext(c.Request.Context()))
		c.Header(logging.HeaderRequestID, requestContext.RequestID)
		c.Header(logging.HeaderTraceID, requestContext.TraceID)

		logger := utils.GetLogger(c)
		c.Set("logger", logger)
		start := time.Now()

		// 先に後続のハンドラー（AuthenticationPostなど）を実行させる
		c.Next()

		// ハンドラー実行中に c.AbortWithError() が呼ばれていたら、ここにエラーが入る
		if len(c.Errors) > 0 {
			err := c.Errors.Last() // 直近のエラーを取得
			status := c.Writer.Status()

			// 1箇所でまとめてエラーログをJSON出力
			logger.Error("http_error",
				slog.Int("status", status),
				slog.Any("error", err),
			)

			// 1箇所でまとめてクライアントにエラーレスポンス（JSON）を返却
			c.JSON(status, gin.H{"error": "Internal Server Error"})
			return
		}

		// エラーがなかった場合は、通常のアクセスログを出す
		logger.Info("http_request",
			slog.Int("status", c.Writer.Status()),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
	}
}
