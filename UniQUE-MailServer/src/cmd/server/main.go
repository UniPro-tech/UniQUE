package main

import (
	"log/slog"
	"os"

	"github.com/UniPro-tech/UniQUE-MailServer/docs"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/config"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/logging"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/middleware"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/routes"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/settings"
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
	logging.Init("mail")
	gin.DefaultWriter = logging.Writer()
	gin.DefaultErrorWriter = logging.Writer()

	databaseSettings, err := settings.LoadValues()
	if err != nil {
		slog.Warn("settings_load_failed", slog.Any("error", err))
		databaseSettings = map[string]string{}
	}
	environmentConfigs := config.LoadConfig(databaseSettings)

	// SMTPメーラーの初期化
	utils.InitMailer(&environmentConfigs.SmtpConfig)

	// Swagger Info
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Title = environmentConfigs.AppName + " Mail Server API"
	docs.SwaggerInfo.Version = environmentConfigs.Version

	r := gin.New()
	r.Use(middleware.SlogRecovery(), middleware.SlogMiddleware())

	r.Use(func(c *gin.Context) {
		c.Set("config", environmentConfigs)
		c.Next()
	})

	r.GET("/health", healthCheck)

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Mail sending endpoint
	r.POST("/email-change", routes.EmailChangeEmail)
	r.POST("/register", routes.RegistrationEmail)
	r.POST("/password-reset", routes.PasswordResetEmail)

	slog.Info("service_starting", slog.String("addr", ":8080"))
	if err := r.Run(); err != nil {
		slog.Error("service_start_failed", slog.Any("error", err))
		os.Exit(1)
	}
}
