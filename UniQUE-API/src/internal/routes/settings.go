package routes

import (
	"net/http"
	"regexp"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/constants"
	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	discordutil "github.com/UniPro-tech/UniQUE-API/internal/utils/discord"
	"github.com/gin-gonic/gin"
)

var discordChannelIDPattern = regexp.MustCompile(`^[0-9]{17,20}$`)

type UpdateDiscordNotificationSettingsRequest struct {
	ChannelID                  string `json:"channel_id"`
	NotifyAnnouncements        bool   `json:"notify_announcements"`
	NotifyRegistrationRequests bool   `json:"notify_registration_requests"`
	NotifyMigrations           bool   `json:"notify_migrations"`
}

func RegisterSettingRoutes(r *gin.Engine) {
	g := r.Group("/settings", middleware.RequirePermission(constants.CONFIG_UPDATE))
	g.GET("/discord-notifications", getDiscordNotificationSettings)
	g.PUT("/discord-notifications", updateDiscordNotificationSettings)
}

// getDiscordNotificationSettings godoc
// @Summary Get Discord notification settings
// @Description Get the global Discord channel and enabled notification events
// @Tags settings
// @Produce json
// @Success 200 {object} discord.NotificationSettings
// @Router /settings/discord-notifications [get]
func getDiscordNotificationSettings(c *gin.Context) {
	db := getDB(c)
	if db == nil {
		return
	}
	cfg := c.MustGet("config").(config.Config)
	settings, err := discordutil.GetNotificationSettings(db, &cfg)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

// updateDiscordNotificationSettings godoc
// @Summary Update Discord notification settings
// @Description Update the global Discord channel and enabled notification events. An empty channel disables delivery.
// @Tags settings
// @Accept json
// @Produce json
// @Param settings body routes.UpdateDiscordNotificationSettingsRequest true "Discord notification settings"
// @Success 200 {object} discord.NotificationSettings
// @Router /settings/discord-notifications [put]
func updateDiscordNotificationSettings(c *gin.Context) {
	db := getDB(c)
	if db == nil {
		return
	}
	var input UpdateDiscordNotificationSettingsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.ChannelID != "" && !discordChannelIDPattern.MatchString(input.ChannelID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id must be a 17 to 20 digit Discord snowflake"})
		return
	}

	settings, err := discordutil.SaveNotificationSettings(db, discordutil.NotificationSettings{
		ChannelID:                  input.ChannelID,
		NotifyAnnouncements:        input.NotifyAnnouncements,
		NotifyRegistrationRequests: input.NotifyRegistrationRequests,
		NotifyMigrations:           input.NotifyMigrations,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}
