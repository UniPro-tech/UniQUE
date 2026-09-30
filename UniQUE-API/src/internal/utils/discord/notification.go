package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	appsettings "github.com/UniPro-tech/UniQUE-API/internal/settings"
	"gorm.io/gorm"
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
	ChannelID                  string `json:"channel_id"`
	NotifyAnnouncements        bool   `json:"notify_announcements"`
	NotifyRegistrationRequests bool   `json:"notify_registration_requests"`
	NotifyMigrations           bool   `json:"notify_migrations"`
}

type legacyNotificationSettings struct {
	ID                         uint8 `gorm:"primaryKey"`
	ChannelID                  string
	NotifyAnnouncements        bool
	NotifyRegistrationRequests bool
	NotifyMigrations           bool
}

type channelMessageRequest struct {
	Content         string          `json:"content"`
	AllowedMentions allowedMentions `json:"allowed_mentions"`
}

type allowedMentions struct {
	Parse []string `json:"parse"`
}

func (legacyNotificationSettings) TableName() string {
	return "discord_notification_settings"
}

func defaultNotificationSettings(cfg *config.Config) NotificationSettings {
	return NotificationSettings{
		ChannelID:                  cfg.DiscordConfig.Guild.NotificationChannelID,
		NotifyAnnouncements:        true,
		NotifyRegistrationRequests: true,
		NotifyMigrations:           true,
	}
}

func resolveNotificationBool(values map[string]string, key string, fallback bool) (bool, error) {
	value := appsettings.Resolve(values, key, strconv.FormatBool(fallback))
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean setting %s: %w", key, err)
	}
	return parsed, nil
}

// GetNotificationSettings applies environment-over-database precedence. The
// Issue #28 table remains a fallback during rolling upgrades.
func GetNotificationSettings(db *gorm.DB, cfg *config.Config) (NotificationSettings, error) {
	defaults := defaultNotificationSettings(cfg)
	values, err := appsettings.LoadValues(db)
	if err != nil {
		legacy := legacyNotificationSettings{}
		legacyErr := db.First(&legacy, discordNotificationSettingsID).Error
		if errors.Is(legacyErr, gorm.ErrRecordNotFound) {
			return defaults, nil
		}
		if legacyErr != nil {
			return NotificationSettings{}, err
		}
		return NotificationSettings{
			ChannelID:                  legacy.ChannelID,
			NotifyAnnouncements:        legacy.NotifyAnnouncements,
			NotifyRegistrationRequests: legacy.NotifyRegistrationRequests,
			NotifyMigrations:           legacy.NotifyMigrations,
		}, nil
	}
	result := defaults
	result.ChannelID = appsettings.Resolve(values, "discord.notification_channel_id", defaults.ChannelID)
	if result.NotifyAnnouncements, err = resolveNotificationBool(values, "discord.notify_announcements", true); err != nil {
		return NotificationSettings{}, err
	}
	if result.NotifyRegistrationRequests, err = resolveNotificationBool(values, "discord.notify_registration_requests", true); err != nil {
		return NotificationSettings{}, err
	}
	if result.NotifyMigrations, err = resolveNotificationBool(values, "discord.notify_migrations", true); err != nil {
		return NotificationSettings{}, err
	}
	return result, nil
}

// SaveNotificationSettings writes the related keys atomically to the general
// settings store.
func SaveNotificationSettings(db *gorm.DB, notificationSettings NotificationSettings) (NotificationSettings, error) {
	values := map[string]string{
		"discord.notification_channel_id":      notificationSettings.ChannelID,
		"discord.notify_announcements":         strconv.FormatBool(notificationSettings.NotifyAnnouncements),
		"discord.notify_registration_requests": strconv.FormatBool(notificationSettings.NotifyRegistrationRequests),
		"discord.notify_migrations":            strconv.FormatBool(notificationSettings.NotifyMigrations),
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		for key, value := range values {
			definition, _ := appsettings.Lookup(key)
			if err := appsettings.Upsert(tx, definition, value); err != nil {
				return err
			}
		}
		legacy := legacyNotificationSettings{
			ID:                         discordNotificationSettingsID,
			ChannelID:                  notificationSettings.ChannelID,
			NotifyAnnouncements:        notificationSettings.NotifyAnnouncements,
			NotifyRegistrationRequests: notificationSettings.NotifyRegistrationRequests,
			NotifyMigrations:           notificationSettings.NotifyMigrations,
		}
		return tx.Where("id = ?", discordNotificationSettingsID).Assign(legacy).FirstOrCreate(&legacy).Error
	})
	return notificationSettings, err
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
