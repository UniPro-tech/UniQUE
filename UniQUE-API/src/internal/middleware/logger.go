package middleware

import (
	"log/slog"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/utils"
	"github.com/gin-gonic/gin"
)

// Gin のアクセスログを slog で JSON 出力するためのミドルウェア
func SlogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c)
		start := time.Now()

		// 先に後続のハンドラー（AuthenticationPostなど）を実行させる
		c.Next()

		// ハンドラー実行中に c.AbortWithError() が呼ばれていたら、ここにエラーが入る
		if len(c.Errors) > 0 {
			err := c.Errors.Last() // 直近のエラーを取得
			status := c.Writer.Status()

			// 1箇所でまとめてエラーログをJSON出力
			logger.Error("An error occurred",
				slog.Int("status", status),
				slog.String("error", err.Error()),
			)

			// 1箇所でまとめてクライアントにエラーレスポンス（JSON）を返却
			c.JSON(status, gin.H{"error": "Internal Server Error"})
			return
		}

		// エラーがなかった場合は、通常のアクセスログを出す
		latency := time.Since(start)
		logger.Info("HTTP Request",
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", latency),
		)
	}
}
