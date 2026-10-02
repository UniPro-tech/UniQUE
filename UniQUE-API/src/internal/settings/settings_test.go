package settings_test

import (
	"os"
	"testing"

	"github.com/UniPro-tech/UniQUE-API/internal/settings"
)

func TestResolvePrecedence(t *testing.T) {
	const environment = "DISCORD_CLIENT_ID"
	previous, existed := os.LookupEnv(environment)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(environment, previous)
		} else {
			_ = os.Unsetenv(environment)
		}
	})

	_ = os.Unsetenv(environment)
	values := map[string]string{"discord.client_id": "database"}
	if got := settings.Resolve(values, "discord.client_id", "fallback"); got != "database" {
		t.Fatalf("database value = %q, want database", got)
	}

	_ = os.Setenv(environment, "environment")
	if got := settings.Resolve(values, "discord.client_id", "fallback"); got != "environment" {
		t.Fatalf("environment value = %q, want environment", got)
	}

	_ = os.Unsetenv(environment)
	if got := settings.Resolve(nil, "discord.client_id", "fallback"); got != "fallback" {
		t.Fatalf("fallback value = %q, want fallback", got)
	}
}

func TestSensitiveDefinitions(t *testing.T) {
	for _, key := range []string{"discord.client_secret", "discord.bot_token", "github.client_secret", "smtp.password"} {
		definition, ok := settings.Lookup(key)
		if !ok || !definition.Sensitive {
			t.Fatalf("%s must be a sensitive setting", key)
		}
	}
}
