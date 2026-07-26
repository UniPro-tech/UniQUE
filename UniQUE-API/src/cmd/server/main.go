package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/UniPro-tech/UniQUE-API/docs"
	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/db"
	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/UniPro-tech/UniQUE-API/internal/routes"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/cors"
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
	// Initialize AWS SDK configuration
	ctx := context.Background()

	endpoint := os.Getenv("CONFIG_S3_ENDPOINT")
	accessKey := os.Getenv("RUSTFS_ACCESS_KEY")
	secretKey := os.Getenv("RUSTFS_SECRET_KEY")

	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			),
		),
	)
	if err != nil {
		slog.Error("Failed to initialize AWS SDK", slog.String("error", err.Error()))
		os.Exit(1)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
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

	query.SetDefault(dbConnection)

	// loggerとrecoveryミドルウェア付きGinルーター作成
	r := gin.Default()

	// Swagger Info
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Title = environmentConfigs.AppName + " API"
	docs.SwaggerInfo.Version = environmentConfigs.Version

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add contexts (AuthMiddlewareより先にセットする必要がある)
	r.Use(func(c *gin.Context) {
		c.Set("config", *environmentConfigs)
		c.Set("db", dbConnection)
		c.Set("s3_client", client)
		c.Next()
	})

	r.Use(middleware.SlogMiddleware())
	r.Use(middleware.AuthMiddleware())
	r.Use(middleware.AuditLogMiddleware())

	// Routes
	r.GET("/health", healthCheck)

	// Register resource routes
	routes.RegisterUserRoutes(r)
	routes.RegisterRoleRoutes(r)
	routes.RegisterApplicationRoutes(r)
	routes.RegisterAnnouncementRoutes(r)

	// Start server
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.Run()
}
