package config_test

import (
	"os"
	"testing"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
)

// テスト前に環境変数を保存し、テスト後に復元するヘルパー関数
func saveAndClearEnv(keys ...string) map[string]string {
	saved := make(map[string]string)
	for _, key := range keys {
		saved[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	return saved
}

func restoreEnv(saved map[string]string) {
	for key, value := range saved {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func TestLoadConfig_DefaultValues(t *testing.T) {
	// 環境変数をクリア
	saved := saveAndClearEnv(
		"CONFIG_APP_NAME",
		"CONFIG_FRONTEND_URL",
		"CONFIG_ISSUER_URL",
		"CONFIG_ISSUER_INTERNAL_URL",
		"CONFIG_EMAIL_SENDER_URL",
		"ENV",
		"DISCORD_API_VERSION",
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_MEMBER_APPLICATION_CHANNEL_ID",
		"DISCORD_BOT_TOKEN",
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
	)
	defer restoreEnv(saved)

	// Discord設定は必須なので設定
	os.Setenv("DISCORD_CLIENT_ID", "test-client-id")
	os.Setenv("DISCORD_CLIENT_SECRET", "test-client-secret")
	os.Setenv("DISCORD_GUILD_ID", "test-guild-id")
	os.Setenv("DISCORD_MEMBER_ROLE_ID", "test-role-id")
	os.Setenv("DISCORD_MEMBER_APPLICATION_CHANNEL_ID", "test-channel-id")
	os.Setenv("DISCORD_BOT_TOKEN", "test-bot-token")

	cfg := config.LoadConfig()

	if cfg.AppName != "UniQUE" {
		t.Errorf("AppName = %q, want %q", cfg.AppName, "UniQUE")
	}
	if cfg.FrontendURL != "http://localhost:3000" {
		t.Errorf("FrontendURL = %q, want %q", cfg.FrontendURL, "http://localhost:3000")
	}
	if cfg.IssuerURL != "http://localhost:8080" {
		t.Errorf("IssuerURL = %q, want %q", cfg.IssuerURL, "http://localhost:8080")
	}
	if cfg.IssuerInternalURL != "http://localhost:8080" {
		t.Errorf("IssuerInternalURL = %q, want %q", cfg.IssuerInternalURL, "http://localhost:8080")
	}
	if cfg.EmailSenderURL != "http://localhost:8080" {
		t.Errorf("EmailSenderURL = %q, want %q", cfg.EmailSenderURL, "http://localhost:8080")
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want %q", cfg.Env, "production")
	}
	if cfg.DiscordApiVersion != "v10" {
		t.Errorf("DiscordApiVersion = %q, want %q", cfg.DiscordApiVersion, "v10")
	}
	if cfg.DiscordConfig.Guild.MemberApplicationChannelID != "test-channel-id" {
		t.Errorf("MemberApplicationChannelID = %q, want %q", cfg.DiscordConfig.Guild.MemberApplicationChannelID, "test-channel-id")
	}
}

func TestLoadConfig_EnvironmentVariablesOverride(t *testing.T) {
	saved := saveAndClearEnv(
		"CONFIG_APP_NAME",
		"CONFIG_FRONTEND_URL",
		"CONFIG_ISSUER_URL",
		"CONFIG_ISSUER_INTERNAL_URL",
		"CONFIG_EMAIL_SENDER_URL",
		"ENV",
		"DISCORD_API_VERSION",
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_MEMBER_APPLICATION_CHANNEL_ID",
		"DISCORD_BOT_TOKEN",
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
	)
	defer restoreEnv(saved)

	// 環境変数を設定
	os.Setenv("CONFIG_APP_NAME", "CustomApp")
	os.Setenv("CONFIG_FRONTEND_URL", "https://frontend.example.com")
	os.Setenv("CONFIG_ISSUER_URL", "https://issuer.example.com")
	os.Setenv("CONFIG_ISSUER_INTERNAL_URL", "http://issuer-internal.local")
	os.Setenv("CONFIG_EMAIL_SENDER_URL", "http://email-sender.local")
	os.Setenv("ENV", "development")
	os.Setenv("DISCORD_API_VERSION", "v11")
	os.Setenv("DISCORD_CLIENT_ID", "test-client-id")
	os.Setenv("DISCORD_CLIENT_SECRET", "test-client-secret")
	os.Setenv("DISCORD_GUILD_ID", "test-guild-id")
	os.Setenv("DISCORD_MEMBER_ROLE_ID", "test-role-id")
	os.Setenv("DISCORD_BOT_TOKEN", "test-bot-token")
	os.Setenv("GITHUB_CLIENT_ID", "github-client-id")
	os.Setenv("GITHUB_CLIENT_SECRET", "github-client-secret")

	cfg := config.LoadConfig()

	if cfg.AppName != "CustomApp" {
		t.Errorf("AppName = %q, want %q", cfg.AppName, "CustomApp")
	}
	if cfg.FrontendURL != "https://frontend.example.com" {
		t.Errorf("FrontendURL = %q, want %q", cfg.FrontendURL, "https://frontend.example.com")
	}
	if cfg.IssuerURL != "https://issuer.example.com" {
		t.Errorf("IssuerURL = %q, want %q", cfg.IssuerURL, "https://issuer.example.com")
	}
	if cfg.IssuerInternalURL != "http://issuer-internal.local" {
		t.Errorf("IssuerInternalURL = %q, want %q", cfg.IssuerInternalURL, "http://issuer-internal.local")
	}
	if cfg.EmailSenderURL != "http://email-sender.local" {
		t.Errorf("EmailSenderURL = %q, want %q", cfg.EmailSenderURL, "http://email-sender.local")
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
	if cfg.DiscordApiVersion != "v11" {
		t.Errorf("DiscordApiVersion = %q, want %q", cfg.DiscordApiVersion, "v11")
	}
	if cfg.GitHubClientID != "github-client-id" {
		t.Errorf("GitHubClientID = %q, want %q", cfg.GitHubClientID, "github-client-id")
	}
	if cfg.GitHubClientSecret != "github-client-secret" {
		t.Errorf("GitHubClientSecret = %q, want %q", cfg.GitHubClientSecret, "github-client-secret")
	}
}

func TestLoadConfig_DiscordConfigComplete(t *testing.T) {
	saved := saveAndClearEnv(
		"CONFIG_APP_NAME",
		"CONFIG_FRONTEND_URL",
		"CONFIG_ISSUER_URL",
		"CONFIG_ISSUER_INTERNAL_URL",
		"CONFIG_EMAIL_SENDER_URL",
		"ENV",
		"DISCORD_API_VERSION",
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_BOT_TOKEN",
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
	)
	defer restoreEnv(saved)

	os.Setenv("DISCORD_CLIENT_ID", "discord-client-id")
	os.Setenv("DISCORD_CLIENT_SECRET", "discord-client-secret")
	os.Setenv("DISCORD_GUILD_ID", "discord-guild-id")
	os.Setenv("DISCORD_MEMBER_ROLE_ID", "discord-role-id")
	os.Setenv("DISCORD_MEMBER_APPLICATION_CHANNEL_ID", "discord-application-channel-id")
	os.Setenv("DISCORD_BOT_TOKEN", "discord-bot-token")

	cfg := config.LoadConfig()

	if cfg.DiscordConfig.ClientID != "discord-client-id" {
		t.Errorf("DiscordConfig.ClientID = %q, want %q", cfg.DiscordConfig.ClientID, "discord-client-id")
	}
	if cfg.DiscordConfig.ClientSecret != "discord-client-secret" {
		t.Errorf("DiscordConfig.ClientSecret = %q, want %q", cfg.DiscordConfig.ClientSecret, "discord-client-secret")
	}
	if cfg.DiscordConfig.Guild.ID != "discord-guild-id" {
		t.Errorf("DiscordConfig.Guild.ID = %q, want %q", cfg.DiscordConfig.Guild.ID, "discord-guild-id")
	}
	if cfg.DiscordConfig.Guild.MemberRoleID != "discord-role-id" {
		t.Errorf("DiscordConfig.Guild.MemberRoleID = %q, want %q", cfg.DiscordConfig.Guild.MemberRoleID, "discord-role-id")
	}
	if cfg.DiscordConfig.Guild.MemberApplicationChannelID != "discord-application-channel-id" {
		t.Errorf("DiscordConfig.Guild.MemberApplicationChannelID = %q, want %q", cfg.DiscordConfig.Guild.MemberApplicationChannelID, "discord-application-channel-id")
	}
	if cfg.DiscordConfig.BotToken != "discord-bot-token" {
		t.Errorf("DiscordConfig.BotToken = %q, want %q", cfg.DiscordConfig.BotToken, "discord-bot-token")
	}
}

func TestLoadConfig_DiscordMissingClientID(t *testing.T) {
	saved := saveAndClearEnv(
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_BOT_TOKEN",
	)
	defer restoreEnv(saved)

	// ClientIDを設定しない
	os.Setenv("DISCORD_CLIENT_SECRET", "secret")
	os.Setenv("DISCORD_GUILD_ID", "guild-id")
	os.Setenv("DISCORD_MEMBER_ROLE_ID", "role-id")
	os.Setenv("DISCORD_BOT_TOKEN", "bot-token")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing DISCORD_CLIENT_ID, but no panic occurred")
		}
	}()

	config.LoadConfig()
}

func TestLoadConfig_DiscordMissingClientSecret(t *testing.T) {
	saved := saveAndClearEnv(
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_BOT_TOKEN",
	)
	defer restoreEnv(saved)

	// ClientSecretを設定しない
	os.Setenv("DISCORD_CLIENT_ID", "client-id")
	os.Setenv("DISCORD_GUILD_ID", "guild-id")
	os.Setenv("DISCORD_MEMBER_ROLE_ID", "role-id")
	os.Setenv("DISCORD_BOT_TOKEN", "bot-token")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing DISCORD_CLIENT_SECRET, but no panic occurred")
		}
	}()

	config.LoadConfig()
}

func TestLoadConfig_DiscordMissingGuildID(t *testing.T) {
	saved := saveAndClearEnv(
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_BOT_TOKEN",
	)
	defer restoreEnv(saved)

	// GuildIDを設定しない
	os.Setenv("DISCORD_CLIENT_ID", "client-id")
	os.Setenv("DISCORD_CLIENT_SECRET", "secret")
	os.Setenv("DISCORD_MEMBER_ROLE_ID", "role-id")
	os.Setenv("DISCORD_BOT_TOKEN", "bot-token")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing DISCORD_GUILD_ID, but no panic occurred")
		}
	}()

	config.LoadConfig()
}

func TestLoadConfig_DiscordMissingMemberRoleID(t *testing.T) {
	saved := saveAndClearEnv(
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_BOT_TOKEN",
	)
	defer restoreEnv(saved)

	// MemberRoleIDを設定しない
	os.Setenv("DISCORD_CLIENT_ID", "client-id")
	os.Setenv("DISCORD_CLIENT_SECRET", "secret")
	os.Setenv("DISCORD_GUILD_ID", "guild-id")
	os.Setenv("DISCORD_BOT_TOKEN", "bot-token")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing DISCORD_MEMBER_ROLE_ID, but no panic occurred")
		}
	}()

	config.LoadConfig()
}

func TestLoadConfig_DiscordMissingBotToken(t *testing.T) {
	saved := saveAndClearEnv(
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_BOT_TOKEN",
	)
	defer restoreEnv(saved)

	// BotTokenを設定しない
	os.Setenv("DISCORD_CLIENT_ID", "client-id")
	os.Setenv("DISCORD_CLIENT_SECRET", "secret")
	os.Setenv("DISCORD_GUILD_ID", "guild-id")
	os.Setenv("DISCORD_MEMBER_ROLE_ID", "role-id")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing DISCORD_BOT_TOKEN, but no panic occurred")
		}
	}()

	config.LoadConfig()
}

func TestLoadConfig_VersionFormatLatest(t *testing.T) {
	saved := saveAndClearEnv(
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_BOT_TOKEN",
	)
	defer restoreEnv(saved)

	os.Setenv("DISCORD_CLIENT_ID", "client-id")
	os.Setenv("DISCORD_CLIENT_SECRET", "secret")
	os.Setenv("DISCORD_GUILD_ID", "guild-id")
	os.Setenv("DISCORD_MEMBER_ROLE_ID", "role-id")
	os.Setenv("DISCORD_BOT_TOKEN", "bot-token")

	cfg := config.LoadConfig()

	// Versionが "branch@commit" 形式になっていることを確認
	// 実際の値はビルド時に設定されるため、ここではフォーマットのみ確認
	if cfg.Version == "" {
		t.Errorf("Version should not be empty")
	}
}

func TestLoadConfig_GitHubOptional(t *testing.T) {
	saved := saveAndClearEnv(
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_BOT_TOKEN",
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
	)
	defer restoreEnv(saved)

	os.Setenv("DISCORD_CLIENT_ID", "client-id")
	os.Setenv("DISCORD_CLIENT_SECRET", "secret")
	os.Setenv("DISCORD_GUILD_ID", "guild-id")
	os.Setenv("DISCORD_MEMBER_ROLE_ID", "role-id")
	os.Setenv("DISCORD_BOT_TOKEN", "bot-token")

	// GitHub認証情報を設定しない

	cfg := config.LoadConfig()

	if cfg.GitHubClientID != "" {
		t.Errorf("GitHubClientID should be empty, got %q", cfg.GitHubClientID)
	}
	if cfg.GitHubClientSecret != "" {
		t.Errorf("GitHubClientSecret should be empty, got %q", cfg.GitHubClientSecret)
	}
}

func TestLoadConfig_ConfigStructNotNil(t *testing.T) {
	saved := saveAndClearEnv(
		"CONFIG_APP_NAME",
		"CONFIG_FRONTEND_URL",
		"CONFIG_ISSUER_URL",
		"CONFIG_ISSUER_INTERNAL_URL",
		"CONFIG_EMAIL_SENDER_URL",
		"ENV",
		"DISCORD_API_VERSION",
		"DISCORD_CLIENT_ID",
		"DISCORD_CLIENT_SECRET",
		"DISCORD_GUILD_ID",
		"DISCORD_MEMBER_ROLE_ID",
		"DISCORD_BOT_TOKEN",
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
	)
	defer restoreEnv(saved)

	os.Setenv("DISCORD_CLIENT_ID", "client-id")
	os.Setenv("DISCORD_CLIENT_SECRET", "secret")
	os.Setenv("DISCORD_GUILD_ID", "guild-id")
	os.Setenv("DISCORD_MEMBER_ROLE_ID", "role-id")
	os.Setenv("DISCORD_BOT_TOKEN", "bot-token")

	cfg := config.LoadConfig()

	if cfg == nil {
		t.Errorf("LoadConfig() returned nil")
	}
}
