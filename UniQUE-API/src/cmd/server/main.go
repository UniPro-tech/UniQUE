package main

import (
	"log"
	"net/http"

	"github.com/UniPro-tech/UniQUE-API/docs"
	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/db"
	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/UniPro-tech/UniQUE-API/internal/routes"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

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
	environmentConfigs := config.LoadConfig()

	// Initialize database
	dbConnection, err := db.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	// gormgen のグローバル変数を初期化 (query.User.ID 等の参照に必要)
	query.SetDefault(dbConnection)

	// loggerとrecoveryミドルウェア付きGinルーター作成
	r := gin.Default()

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
