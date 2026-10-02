package settings_test

import (
	"os"
	"testing"

	"github.com/UniPro-tech/UniQUE-Auth/internal/settings"
)

func TestResolvePrefersEnvironment(t *testing.T) {
	const key = "CONFIG_APP_NAME"
	previous, existed := os.LookupEnv(key)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, previous)
		} else {
			_ = os.Unsetenv(key)
		}
	})
	_ = os.Setenv(key, "environment")
	if got := settings.Resolve(map[string]string{"application.name": "database"}, key, "application.name", "fallback"); got != "environment" {
		t.Fatalf("Resolve() = %q", got)
	}
}
