package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/UniPro-tech/UniQUE-API/docs"
	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/db"
	"github.com/UniPro-tech/UniQUE-API/internal/logging"
	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/UniPro-tech/UniQUE-API/internal/routes"
	"github.com/UniPro-tech/UniQUE-API/internal/settings"
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
	logging.Init("api")
	gin.DefaultWriter = logging.Writer()
	gin.DefaultErrorWriter = logging.Writer()

	// Initialize database
	dbConnection, err := db.NewDB()
	if err != nil {
		slog.Error("database_init_failed", slog.Any("error", err))
		os.Exit(1)
	}
	databaseSettings, err := settings.LoadValues(dbConnection)
	if err != nil {
		slog.Warn("settings_load_failed", slog.Any("error", err))
		databaseSettings = map[string]string{}
	}
	environmentConfigs := config.LoadConfig(databaseSettings)

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

	// アクセスログは SlogMiddleware が JSON で出力するため、Gin 標準の
	// テキストログは有効化しない。panic は SlogRecovery で slog に記録する。
	r := gin.New()
	r.Use(middleware.SlogRecovery())

	// Swagger Info
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Title = environmentConfigs.AppName + " API"
	docs.SwaggerInfo.Version = environmentConfigs.Version

	// Add contexts (AuthMiddlewareより先にセットする必要がある)
	r.Use(func(c *gin.Context) {
		c.Set("config", *environmentConfigs)
		c.Set("db", dbConnection)
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
	routes.RegisterSettingRoutes(r)

	// Start server
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	slog.Info("service_starting", slog.String("addr", ":8080"))
	if err := r.Run(); err != nil {
		slog.Error("service_start_failed", slog.Any("error", err))
		os.Exit(1)
	}
}
