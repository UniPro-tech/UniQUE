package routes

import (
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/constants"
	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	appsettings "github.com/UniPro-tech/UniQUE-API/internal/settings"
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

type UpdateSettingRequest struct {
	Value string `json:"value" binding:"required,max=65535"`
}

type SettingResponse struct {
	Key                     string     `json:"key"`
	Value                   string     `json:"value,omitempty"`
	Sensitive               bool       `json:"sensitive"`
	Configured              bool       `json:"configured"`
	Source                  string     `json:"source"`
	OverriddenByEnvironment bool       `json:"overridden_by_environment"`
	Description             string     `json:"description"`
	UpdatedAt               *time.Time `json:"updated_at,omitempty"`
}

func RegisterSettingRoutes(r *gin.Engine) {
	g := r.Group("/settings", middleware.RequirePermission(constants.CONFIG_UPDATE))
	g.GET("", listSettings)
	g.GET("/:key", getSetting)
	g.PUT("/:key", updateSetting)
	g.DELETE("/:key", deleteSetting)
	g.GET("/discord-notifications", getDiscordNotificationSettings)
	g.PUT("/discord-notifications", updateDiscordNotificationSettings)
}

func settingResponse(definition appsettings.Definition, row appsettings.Setting, stored bool) SettingResponse {
	value, environmentSet := os.LookupEnv(definition.Environment)
	response := SettingResponse{
		Key:                     definition.Key,
		Sensitive:               definition.Sensitive,
		Description:             definition.Description,
		OverriddenByEnvironment: environmentSet,
		Source:                  "unset",
	}
	if environmentSet {
		response.Source = "environment"
		response.Configured = value != ""
		if !definition.Sensitive {
			response.Value = value
		}
	} else if stored {
		response.Source = "database"
		response.Configured = row.Value != ""
		response.UpdatedAt = &row.UpdatedAt
		if !definition.Sensitive {
			response.Value = row.Value
		}
	}
	return response
}

func loadSettingRows(c *gin.Context) (map[string]appsettings.Setting, bool) {
	db := getDB(c)
	if db == nil {
		return nil, false
	}
	rows, err := appsettings.LoadRows(db)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return nil, false
	}
	return rows, true
}

// listSettings returns all supported settings. Secret values are write-only.
func listSettings(c *gin.Context) {
	rows, ok := loadSettingRows(c)
	if !ok {
		return
	}
	responses := make([]SettingResponse, 0, len(appsettings.Definitions()))
	for _, definition := range appsettings.Definitions() {
		row, stored := rows[definition.Key]
		responses = append(responses, settingResponse(definition, row, stored))
	}
	c.JSON(http.StatusOK, gin.H{"data": responses})
}

func getSetting(c *gin.Context) {
	definition, found := appsettings.Lookup(c.Param("key"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown setting"})
		return
	}
	rows, ok := loadSettingRows(c)
	if !ok {
		return
	}
	row, stored := rows[definition.Key]
	c.JSON(http.StatusOK, settingResponse(definition, row, stored))
}

func updateSetting(c *gin.Context) {
	definition, found := appsettings.Lookup(c.Param("key"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown setting"})
		return
	}
	var input UpdateSettingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	if err := appsettings.Upsert(db, definition, input.Value); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	rows, ok := loadSettingRows(c)
	if !ok {
		return
	}
	row := rows[definition.Key]
	c.JSON(http.StatusOK, settingResponse(definition, row, true))
}

func deleteSetting(c *gin.Context) {
	definition, found := appsettings.Lookup(c.Param("key"))
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown setting"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	if err := appsettings.Delete(db, definition.Key); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
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

	_, err := discordutil.SaveNotificationSettings(db, discordutil.NotificationSettings{
		ChannelID:                  input.ChannelID,
		NotifyAnnouncements:        input.NotifyAnnouncements,
		NotifyRegistrationRequests: input.NotifyRegistrationRequests,
		NotifyMigrations:           input.NotifyMigrations,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	cfg := c.MustGet("config").(config.Config)
	effective, err := discordutil.GetNotificationSettings(db, &cfg)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, effective)
}
