package routes

import (
	"os"
	"testing"

	appsettings "github.com/UniPro-tech/UniQUE-API/internal/settings"
)

func TestSettingResponseMasksSecrets(t *testing.T) {
	definition, _ := appsettings.Lookup("discord.client_secret")
	response := settingResponse(definition, appsettings.Setting{Value: "database-secret"}, true)
	if response.Value != "" || !response.Configured || !response.Sensitive || response.Source != "database" {
		t.Fatalf("unexpected secret response: %#v", response)
	}
}

func TestSettingResponseReportsEnvironmentOverride(t *testing.T) {
	definition, _ := appsettings.Lookup("discord.guild_id")
	previous, existed := os.LookupEnv(definition.Environment)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(definition.Environment, previous)
		} else {
			_ = os.Unsetenv(definition.Environment)
		}
	})
	_ = os.Setenv(definition.Environment, "environment-guild")
	response := settingResponse(definition, appsettings.Setting{Value: "database-guild"}, true)
	if response.Value != "environment-guild" || response.Source != "environment" || !response.OverriddenByEnvironment {
		t.Fatalf("unexpected environment response: %#v", response)
	}
}
