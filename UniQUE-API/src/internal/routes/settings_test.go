package routes

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	appsettings "github.com/UniPro-tech/UniQUE-API/internal/settings"
	"github.com/gin-gonic/gin"
)

func TestSettingResponseMasksSecrets(t *testing.T) {
	definition, _ := appsettings.Lookup("discord.client_secret")
	response := settingResponse(definition, appsettings.Setting{Value: "database-secret"}, true)
	if response.Value != "" || !response.Configured || !response.Sensitive || response.Source != "database" {
		t.Fatalf("unexpected secret response: %#v", response)
	}
}

func TestUpdateSettingRequestAcceptsEmptyValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest("PUT", "/settings/application.name", strings.NewReader(`{"value":""}`))
	request.Header.Set("Content-Type", "application/json")
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = request

	var input UpdateSettingRequest
	if err := context.ShouldBindJSON(&input); err != nil {
		t.Fatalf("empty value must be accepted: %v", err)
	}
	if input.Value == nil || *input.Value != "" {
		t.Fatalf("unexpected value: %#v", input.Value)
	}
}

func TestUpdateSettingRequestRequiresValueField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest("PUT", "/settings/application.name", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = request

	var input UpdateSettingRequest
	if err := context.ShouldBindJSON(&input); err == nil {
		t.Fatal("missing value field must be rejected")
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
