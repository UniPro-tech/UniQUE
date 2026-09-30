package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const discordNotificationSettingsID uint8 = 1

// NotificationEvent identifies an event that can be forwarded to Discord.
type NotificationEvent string

const (
	NotificationAnnouncementCreated   NotificationEvent = "announcement_created"
	NotificationRegistrationRequested NotificationEvent = "registration_requested"
	NotificationMigrationCompleted    NotificationEvent = "migration_completed"
)

// NotificationSettings is the singleton, administrator-managed Discord
// notification configuration.
type NotificationSettings struct {
	ID                         uint8     `json:"-" gorm:"primaryKey"`
	ChannelID                  string    `json:"channel_id"`
	NotifyAnnouncements        bool      `json:"notify_announcements"`
	NotifyRegistrationRequests bool      `json:"notify_registration_requests"`
	NotifyMigrations           bool      `json:"notify_migrations"`
	CreatedAt                  time.Time `json:"-"`
	UpdatedAt                  time.Time `json:"-"`
}

type channelMessageRequest struct {
	Content         string          `json:"content"`
	AllowedMentions allowedMentions `json:"allowed_mentions"`
}

type allowedMentions struct {
	Parse []string `json:"parse"`
}

func (NotificationSettings) TableName() string {
	return "discord_notification_settings"
}

func defaultNotificationSettings(cfg *config.Config) NotificationSettings {
	return NotificationSettings{
		ID:                         discordNotificationSettingsID,
		ChannelID:                  cfg.DiscordConfig.Guild.NotificationChannelID,
		NotifyAnnouncements:        true,
		NotifyRegistrationRequests: true,
		NotifyMigrations:           true,
	}
}

// GetNotificationSettings returns the saved setting. Until an administrator
// saves it for the first time, the environment variable is used as a fallback.
func GetNotificationSettings(db *gorm.DB, cfg *config.Config) (NotificationSettings, error) {
	settings := NotificationSettings{}
	err := db.First(&settings, discordNotificationSettingsID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return defaultNotificationSettings(cfg), nil
	}
	return settings, err
}

// SaveNotificationSettings creates or replaces the singleton setting.
func SaveNotificationSettings(db *gorm.DB, settings NotificationSettings) (NotificationSettings, error) {
	now := time.Now().UTC()
	settings.ID = discordNotificationSettingsID
	settings.UpdatedAt = now
	if settings.CreatedAt.IsZero() {
		settings.CreatedAt = now
	}

	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"channel_id":                   settings.ChannelID,
			"notify_announcements":         settings.NotifyAnnouncements,
			"notify_registration_requests": settings.NotifyRegistrationRequests,
			"notify_migrations":            settings.NotifyMigrations,
			"updated_at":                   settings.UpdatedAt,
		}),
	}).Create(&settings).Error
	return settings, err
}

func notificationEnabled(settings NotificationSettings, event NotificationEvent) bool {
	switch event {
	case NotificationAnnouncementCreated:
		return settings.NotifyAnnouncements
	case NotificationRegistrationRequested:
		return settings.NotifyRegistrationRequests
	case NotificationMigrationCompleted:
		return settings.NotifyMigrations
	default:
		return false
	}
}

func notificationContent(event NotificationEvent, frontendURL string) (string, error) {
	baseURL := strings.TrimRight(frontendURL, "/")

	switch event {
	case NotificationAnnouncementCreated:
		return fmt.Sprintf("# 📢 新しいアナウンスが追加されました\n%s/dashboard/announcements", baseURL), nil
	case NotificationRegistrationRequested:
		return fmt.Sprintf("# 📝 新しいメンバー登録申請が届きました\n%s/dashboard/requests", baseURL), nil
	case NotificationMigrationCompleted:
		return fmt.Sprintf("# ✅ アカウント移行によりメンバーが追加されました\n%s/dashboard/members", baseURL), nil
	default:
		return "", fmt.Errorf("unsupported discord notification event: %s", event)
	}
}

// SendNotification sends a configured event to Discord. Callers should log an
// error but must not roll back the domain operation that already succeeded.
func SendNotification(event NotificationEvent, db *gorm.DB, cfg *config.Config) error {
	settings, err := GetNotificationSettings(db, cfg)
	if err != nil {
		return fmt.Errorf("load discord notification settings: %w", err)
	}
	if settings.ChannelID == "" || !notificationEnabled(settings, event) {
		return nil
	}
	content, err := notificationContent(event, cfg.FrontendURL)
	if err != nil {
		return err
	}
	return sendChannelMessage(
		&http.Client{Timeout: 5 * time.Second},
		"https://discord.com/api/"+cfg.DiscordApiVersion,
		cfg.DiscordConfig.BotToken,
		settings.ChannelID,
		content,
	)
}

func sendChannelMessage(client *http.Client, apiBaseURL, botToken, channelID, content string) error {
	if botToken == "" {
		return errors.New("discord bot token not configured")
	}
	body, err := json.Marshal(channelMessageRequest{
		Content: content,
		AllowedMentions: allowedMentions{
			Parse: []string{},
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/channels/%s/messages", strings.TrimRight(apiBaseURL, "/"), channelID),
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send discord channel message, status code: %d", resp.StatusCode)
	}
	return nil
}
