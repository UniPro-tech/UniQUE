package main

import (
	"log/slog"
	"os"

	"github.com/UniPro-tech/UniQUE-MailServer/docs"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/config"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/middleware"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/routes"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/utils"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @BasePath /

// HealthCheck godoc
// @Summary Health Check
// @Description get the health status
// @Tags health
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]string
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func main() {
	// envのロード
	environmentConfigs := config.LoadConfig()

	// slogの設定
	var programLevel = new(slog.LevelVar)
	slogHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(slogHandler))
	// ginのモードも同時に設定
	if environmentConfigs.Debug {
		gin.SetMode(gin.DebugMode)
		programLevel.Set(slog.LevelDebug)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	// ログレベルの設定をenvから参照する
	switch environmentConfigs.LogLevel {
	case "debug":
		programLevel.Set(slog.LevelDebug)
	default:
	case "info":
		programLevel.Set(slog.LevelInfo)
	case "warn":
		programLevel.Set(slog.LevelWarn)
	case "error":
		programLevel.Set(slog.LevelError)
	}

	logger := slog.Default()
	logger.With(
		slog.String("package", "main"),
		slog.String("file", "cmd/server/main.go"),
		slog.String("function", "main"),
		slog.String("layer", "command"),
	)
	logger.Info("Config loaded")

	// SMTPメーラーの初期化
	logger.Info("Loading SMTP config")
	utils.InitMailer(&environmentConfigs.SmtpConfig)
	logger.Info("SMTP config Loaded")

	// Swagger Info
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Title = environmentConfigs.AppName + " Mail Server API"
	docs.SwaggerInfo.Version = environmentConfigs.Version

	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("config", environmentConfigs)
		c.Next()
	})

	ginLogger := slog.Default()
	ginLogger.With(slog.String("layer", "gin"))
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.SlogMiddleware(ginLogger))

	r.GET("/health", healthCheck)

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Mail sending endpoint
	r.POST("/email-change", routes.EmailChangeEmail)
	r.POST("/register", routes.RegistrationEmail)
	r.POST("/password-reset", routes.PasswordResetEmail)

	r.Run()
	logger.Info("UniQUE Mail Server started now 🚀")
}
