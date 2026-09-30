package settings_test

import (
	"os"
	"testing"

	"github.com/UniPro-tech/UniQUE/Discord/internal/settings"
)

func TestResolvePrecedence(t *testing.T) {
	const environment = "DISCORD_TOKEN"
	previous, existed := os.LookupEnv(environment)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(environment, previous)
		} else {
			_ = os.Unsetenv(environment)
		}
	})
	_ = os.Unsetenv(environment)
	values := map[string]string{"discord.service_token": "database"}
	if got := settings.Resolve(values, environment, "discord.service_token", "fallback"); got != "database" {
		t.Fatalf("database value = %q", got)
	}
	_ = os.Setenv(environment, "environment")
	if got := settings.Resolve(values, environment, "discord.service_token", "fallback"); got != "environment" {
		t.Fatalf("environment value = %q", got)
	}
}
