package settings

import (
	"os"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Definition struct {
	Key         string
	Environment string
	Sensitive   bool
	Description string
}

var definitions = []Definition{
	{Key: "application.name", Environment: "CONFIG_APP_NAME", Description: "Application display name"},
	{Key: "discord.api_version", Environment: "DISCORD_API_VERSION", Description: "Discord API version"},
	{Key: "discord.client_id", Environment: "DISCORD_CLIENT_ID", Description: "Discord OAuth client ID"},
	{Key: "discord.client_secret", Environment: "DISCORD_CLIENT_SECRET", Sensitive: true, Description: "Discord OAuth client secret"},
	{Key: "discord.guild_id", Environment: "DISCORD_GUILD_ID", Description: "Discord guild ID"},
	{Key: "discord.member_role_id", Environment: "DISCORD_MEMBER_ROLE_ID", Description: "Discord member role ID"},
	{Key: "discord.bot_token", Environment: "DISCORD_BOT_TOKEN", Sensitive: true, Description: "Discord bot token used by the API"},
	{Key: "discord.notification_channel_id", Environment: "DISCORD_NOTIFICATION_CHANNEL_ID", Description: "Discord notification channel ID"},
	{Key: "discord.notify_announcements", Environment: "DISCORD_NOTIFY_ANNOUNCEMENTS", Description: "Notify when an announcement is created"},
	{Key: "discord.notify_registration_requests", Environment: "DISCORD_NOTIFY_REGISTRATION_REQUESTS", Description: "Notify when a registration is requested"},
	{Key: "discord.notify_migrations", Environment: "DISCORD_NOTIFY_MIGRATIONS", Description: "Notify when an account migration completes"},
	{Key: "discord.service_token", Environment: "DISCORD_TOKEN", Sensitive: true, Description: "Discord bot service token"},
	{Key: "discord.bot_name", Environment: "CONFIG_BOT_NAME", Description: "Discord bot display name"},
	{Key: "discord.description", Environment: "CONFIG_DESCRIPTION", Description: "Discord bot description"},
	{Key: "discord.admin_guild_id", Environment: "CONFIG_ADMIN_GUILD_ID", Description: "Discord administration guild ID"},
	{Key: "discord.admin_role_id", Environment: "CONFIG_ADMIN_ROLE_ID", Description: "Discord administration role ID"},
	{Key: "discord.github_repository", Environment: "CONFIG_GITHUB_REPO", Description: "GitHub repository displayed by the Discord bot"},
	{Key: "github.client_id", Environment: "GITHUB_CLIENT_ID", Description: "GitHub OAuth client ID"},
	{Key: "github.client_secret", Environment: "GITHUB_CLIENT_SECRET", Sensitive: true, Description: "GitHub OAuth client secret"},
	{Key: "smtp.host", Environment: "SMTP_HOST", Description: "SMTP server host"},
	{Key: "smtp.port", Environment: "SMTP_PORT", Description: "SMTP server port"},
	{Key: "smtp.username", Environment: "SMTP_USERNAME", Sensitive: true, Description: "SMTP username"},
	{Key: "smtp.password", Environment: "SMTP_PASSWORD", Sensitive: true, Description: "SMTP password"},
	{Key: "smtp.from", Environment: "SMTP_FROM", Description: "SMTP envelope sender"},
	{Key: "smtp.secure", Environment: "SMTP_SECURE", Description: "Whether SMTP uses TLS"},
	{Key: "mail.from_name", Environment: "FROM_NAME", Description: "Mail sender display name"},
	{Key: "mail.copyright_name", Environment: "COPYRIGHT_NAME", Description: "Copyright holder shown in email"},
}

type Setting struct {
	Key         string    `gorm:"column:key;primaryKey;size:100"`
	Value       string    `gorm:"column:value;type:text;not null"`
	IsSecret    bool      `gorm:"column:is_secret;not null"`
	Description string    `gorm:"column:description;size:255;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Setting) TableName() string { return "settings" }

func Definitions() []Definition {
	result := append([]Definition(nil), definitions...)
	sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
	return result
}

func Lookup(key string) (Definition, bool) {
	for _, definition := range definitions {
		if definition.Key == key {
			return definition, true
		}
	}
	return Definition{}, false
}

func LoadValues(db *gorm.DB) (map[string]string, error) {
	values := make(map[string]string)
	if db == nil {
		return values, nil
	}
	var rows []Setting
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	return values, nil
}

func LoadRows(db *gorm.DB) (map[string]Setting, error) {
	rowsByKey := make(map[string]Setting)
	if db == nil {
		return rowsByKey, nil
	}
	var rows []Setting
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		rowsByKey[row.Key] = row
	}
	return rowsByKey, nil
}

func Upsert(db *gorm.DB, definition Definition, value string) error {
	row := Setting{
		Key:         definition.Key,
		Value:       value,
		IsSecret:    definition.Sensitive,
		Description: definition.Description,
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"value":       row.Value,
			"is_secret":   row.IsSecret,
			"description": row.Description,
			"updated_at":  time.Now().UTC(),
		}),
	}).Create(&row).Error
}

func Delete(db *gorm.DB, key string) error {
	return db.Where("`key` = ?", key).Delete(&Setting{}).Error
}

// Resolve applies the required precedence: environment, database, fallback.
func Resolve(values map[string]string, key, fallback string) string {
	if definition, ok := Lookup(key); ok {
		if value, exists := os.LookupEnv(definition.Environment); exists {
			return value
		}
	}
	if value, ok := values[key]; ok {
		return value
	}
	return fallback
}
