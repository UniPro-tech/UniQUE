package routes

import (
	"testing"

	discordutil "github.com/UniPro-tech/UniQUE-API/internal/utils/discord"
)

func TestNotificationEventForCreationSource(t *testing.T) {
	tests := []struct {
		source string
		want   discordutil.NotificationEvent
	}{
		{"registration", discordutil.NotificationRegistrationRequested},
		{"migration", discordutil.NotificationMigrationCompleted},
		{"admin", ""},
		{"unknown", ""},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			if got := notificationEventForCreationSource(test.source); got != test.want {
				t.Errorf("notificationEventForCreationSource(%q) = %q, want %q", test.source, got, test.want)
			}
		})
	}
}

func TestDiscordChannelIDPattern(t *testing.T) {
	valid := []string{"12345678901234567", "12345678901234567890"}
	invalid := []string{"1234567890123456", "123456789012345678901", "channel-id"}

	for _, value := range valid {
		if !discordChannelIDPattern.MatchString(value) {
			t.Errorf("expected %q to be a valid Discord channel ID", value)
		}
	}
	for _, value := range invalid {
		if discordChannelIDPattern.MatchString(value) {
			t.Errorf("expected %q to be rejected as a Discord channel ID", value)
		}
	}
}
