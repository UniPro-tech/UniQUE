package discord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	appsettings "github.com/UniPro-tech/UniQUE-API/internal/settings"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestNotificationContent(t *testing.T) {
	tests := []struct {
		event NotificationEvent
		path  string
	}{
		{NotificationAnnouncementCreated, "/dashboard/announcements"},
		{NotificationRegistrationRequested, "/dashboard/requests"},
		{NotificationMigrationCompleted, "/dashboard/members"},
	}

	for _, test := range tests {
		t.Run(string(test.event), func(t *testing.T) {
			content, err := notificationContent(test.event, "https://unique.example/")
			if err != nil {
				t.Fatalf("notificationContent returned an error: %v", err)
			}
			if !strings.Contains(content, "https://unique.example"+test.path) {
				t.Fatalf("content %q does not contain expected path %q", content, test.path)
			}
		})
	}
}

func TestNotificationEnabled(t *testing.T) {
	settings := NotificationSettings{
		NotifyAnnouncements:        true,
		NotifyRegistrationRequests: false,
		NotifyMigrations:           true,
	}
	if !notificationEnabled(settings, NotificationAnnouncementCreated) {
		t.Error("announcement notification should be enabled")
	}
	if notificationEnabled(settings, NotificationRegistrationRequested) {
		t.Error("registration notification should be disabled")
	}
	if !notificationEnabled(settings, NotificationMigrationCompleted) {
		t.Error("migration notification should be enabled")
	}
}

func TestSendChannelMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/channels/12345678901234567/messages" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bot test-token" {
			t.Errorf("Authorization = %q", got)
		}
		var body channelMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body.Content != "test message" {
			t.Errorf("content = %q", body.Content)
		}
		if body.AllowedMentions.Parse == nil || len(body.AllowedMentions.Parse) != 0 {
			t.Errorf("allowed mentions must explicitly disable parsing")
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	err := sendChannelMessage(
		server.Client(),
		server.URL,
		"test-token",
		"12345678901234567",
		"test message",
	)
	if err != nil {
		t.Fatalf("sendChannelMessage returned an error: %v", err)
	}
}

func TestSendChannelMessageRejectsDiscordError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	err := sendChannelMessage(
		server.Client(),
		server.URL,
		"test-token",
		"12345678901234567",
		"test message",
	)
	if err == nil {
		t.Fatal("sendChannelMessage should return an error for a Discord error response")
	}
}

func TestGetNotificationSettingsChannelPrecedence(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "test:test@tcp(localhost:3306)/test",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
		rows := tx.Statement.Dest.(*[]appsettings.Setting)
		*rows = []appsettings.Setting{{Key: "discord.notification_channel_id", Value: "database"}}
	}); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name            string
		current, legacy *string
		want            string
	}{
		{name: "database", want: "database"},
		{name: "legacy overrides database", legacy: stringPointer("legacy"), want: "legacy"},
		{name: "current overrides legacy", current: stringPointer("current"), legacy: stringPointer("legacy"), want: "current"},
		{name: "empty current is explicit", current: stringPointer(""), legacy: stringPointer("legacy"), want: ""},
		{name: "empty legacy is explicit", legacy: stringPointer(""), want: ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			for key, value := range map[string]*string{
				"DISCORD_NOTIFICATION_CHANNEL_ID":       test.current,
				"DISCORD_MEMBER_APPLICATION_CHANNEL_ID": test.legacy,
			} {
				t.Setenv(key, "")
				if value == nil {
					if err := os.Unsetenv(key); err != nil {
						t.Fatal(err)
					}
				} else {
					t.Setenv(key, *value)
				}
			}
			got, err := GetNotificationSettings(db, &config.Config{})
			if err != nil {
				t.Fatal(err)
			}
			if got.ChannelID != test.want {
				t.Fatalf("channel = %q, want %q", got.ChannelID, test.want)
			}
		})
	}
}

func stringPointer(value string) *string { return &value }
