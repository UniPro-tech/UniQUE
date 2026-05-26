package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/docs"
	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/UniPro-tech/UniQUE-Auth/internal/db"
	"github.com/UniPro-tech/UniQUE-Auth/internal/router"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm/logger"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status string `json:"status"`
}

// @BasePath /

// HealthCheck godoc
// @Summary health check endpoint
// @Schemes
// @Description システムの稼働状況を確認するためのエンドポイントです。
// @Tags system info
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "ok",
	})
}

// Gin のアクセスログを slog で JSON 出力するためのミドルウェア
func slogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 先に後続のハンドラー（AuthenticationPostなど）を実行させる
		c.Next()

		// ハンドラー実行中に c.AbortWithError() が呼ばれていたら、ここにエラーが入る
		if len(c.Errors) > 0 {
			err := c.Errors.Last() // 直近のエラーを取得
			status := c.Writer.Status()

			// 1箇所でまとめてエラーログをJSON出力
			slog.Error("An error occurred",
				slog.Int("status", status),
				slog.String("method", c.Request.Method),
				slog.String("path", path),
				slog.String("query", query),
				slog.String("ip", c.ClientIP()),
				slog.String("user_agent", c.Request.UserAgent()),
				slog.String("error", err.Error()),
			)

			// 1箇所でまとめてクライアントにエラーレスポンス（JSON）を返却
			c.JSON(status, gin.H{"error": "Internal Server Error"})
			return
		}

		// エラーがなかった場合は、通常のアクセスログを出す
		latency := time.Since(start)
		slog.Info("HTTP Request",
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.Duration("latency", latency),
		)
	}
}

func main() {
	// --- slog の初期化 ---
	// JSONフォーマットで標準出力へログを書き出すように設定
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // 出力レベルを変更したい場合はここで調整
	})
	slog.SetDefault(slog.New(handler))

	environmentConfigs := config.LoadConfig()

	// Initialize database
	dbConnection, err := db.NewDB()
	if err != nil {
		// 標準の log.Fatal ではなく slog.Error を使用して JSON 形式を維持
		slog.Error("Failed to initialize database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	dbConnection.Logger = dbConnection.Logger.LogMode(logger.Info)

	// ログレベルの決定（環境変数などで切り替えるイメージ）
	var gormLogLevel logger.LogLevel

	if environmentConfigs.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
		gormLogLevel = logger.Error
		dbConnection.Logger = dbConnection.Logger.LogMode(gormLogLevel)
	} else {
		gormLogLevel = logger.Info
		dbConnection.Logger = dbConnection.Logger.LogMode(gormLogLevel)
	}

	r := gin.New()
	// カスタム slog ミドルウェアと、パニックリカバリーを登録
	r.Use(slogMiddleware(), gin.Recovery())

	// Swagger Info
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Title = environmentConfigs.AppName + " Auth API"
	docs.SwaggerInfo.Version = environmentConfigs.Version

	// Add contexts
	r.Use(func(c *gin.Context) {
		c.Set("config", environmentConfigs)
		c.Set("db", dbConnection)
		c.Next()
	})

	// Routes
	r.GET("/health", healthCheck)
	r.GET("/authorization", router.AuthorizationGet)
	r.POST("/authorization", router.AuthorizationPost)
	r.GET("/.well-known/openid-configuration", router.WellKnownOpenIDConfiguration)
	r.GET("/.well-known/jwks.json", router.WellKnownJWKS)
	r.POST("/token", router.TokenPost)
	r.GET("/userinfo", router.UserInfoGet)
	r.GET("/consented", router.ConsentedGet)

	// Internal routes
	ig := r.Group("/internal")
	{
		ig.POST("/authentication", router.AuthenticationPost)
		ig.GET("/consents", router.ConsentList)
		ig.POST("/consents", router.ConsentCreate)
		ig.DELETE("/consents/:id", router.ConsentDeleteByID)
		ig.POST("/password_hash", router.PasswordHashPost)
		ig.GET("/sessions", router.SessionsGet)
		ig.GET("/sessions/:sid", router.GetSessionById)
		ig.DELETE("/sessions/:sid", router.SessionsDelete)
		ig.GET("/session_verify", router.SessionVerifyGet)
		ig.GET("/token_verify", router.TokenVerifyGet)
		ig.GET("/auth-requests/:id", router.InternalAuthorizationGet)
		ig.POST("/auth-requests/:id/consented", router.InternalConsentedPost)
		ig.POST("/totp/:uid", router.GenerateTOTP)
		ig.POST("/totp/:uid/verify", router.VerifyTOTP)
		ig.POST("/totp/:uid/disable", router.DisableTOTP)
		ig.POST("/password_reset/request", router.PasswordResetRequestPost)
		ig.POST("/password_reset/confirm", router.PasswordResetConfirmPost)
	}

	// Start server
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	slog.Info("Starting server on :8080")
	r.Run()
}
