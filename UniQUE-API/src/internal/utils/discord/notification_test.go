package discord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
