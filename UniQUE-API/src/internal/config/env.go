package config

import (
	"os"

	"github.com/UniPro-tech/UniQUE-API/internal/settings"
)

type DiscordGuildConfig struct {
	ID                    string
	MemberRoleID          string
	NotificationChannelID string
}

type DiscordConfig struct {
	ClientID     string
	ClientSecret string
	Guild        DiscordGuildConfig
	BotToken     string
}

type Config struct {
	Env                string
	AppName            string
	Version            string
	FrontendURL        string
	IssuerURL          string
	IssuerInternalURL  string
	EmailSenderURL     string
	DiscordConfig      DiscordConfig
	GitHubClientID     string
	GitHubClientSecret string
	DiscordApiVersion  string
}

// envが設定されていない場合のデフォルト値
var (
	Version   = "latest"
	GitCommit = "unknown"
	GitBranch = "unknown"
)

var (
	Env               = "production"
	AppName           = "UniQUE"
	FrontendURL       = "http://localhost:3000"
	IssuerURL         = "http://localhost:8080"
	IssuerInternalURL = "http://localhost:8080"
	EmailSenderURL    = "http://localhost:8080"
	DiscordApiVersion = "v10"
)

func LoadConfig(databaseValues ...map[string]string) *Config {
	values := map[string]string{}
	if len(databaseValues) > 0 && databaseValues[0] != nil {
		values = databaseValues[0]
	}
	version := Version

	if Version == "latest" {
		version = GitBranch + "@" + GitCommit
	} else {
		version = Version + "+" + GitCommit
	}

	// envから設定を読み込む
	AppNameEnv := settings.Resolve(values, "application.name", AppName)
	FrontendURLEnv := os.Getenv("CONFIG_FRONTEND_URL")
	if FrontendURLEnv == "" {
		FrontendURLEnv = FrontendURL
	}
	IssuerURLEnv := os.Getenv("CONFIG_ISSUER_URL")
	if IssuerURLEnv == "" {
		IssuerURLEnv = IssuerURL
	}
	IssuerInternalURLEnv := os.Getenv("CONFIG_ISSUER_INTERNAL_URL")
	if IssuerInternalURLEnv == "" {
		IssuerInternalURLEnv = IssuerInternalURL
	}
	EmailSenderURLEnv := os.Getenv("CONFIG_EMAIL_SENDER_URL")
	if EmailSenderURLEnv == "" {
		EmailSenderURLEnv = EmailSenderURL
	}
	DiscordConfig := DiscordConfig{
		ClientID:     settings.Resolve(values, "discord.client_id", ""),
		ClientSecret: settings.Resolve(values, "discord.client_secret", ""),
		Guild: DiscordGuildConfig{
			ID:           settings.Resolve(values, "discord.guild_id", ""),
			MemberRoleID: settings.Resolve(values, "discord.member_role_id", ""),
		},
		BotToken: settings.Resolve(values, "discord.bot_token", ""),
	}
	if channelID, exists := os.LookupEnv("DISCORD_NOTIFICATION_CHANNEL_ID"); exists {
		DiscordConfig.Guild.NotificationChannelID = channelID
	} else if channelID, exists := os.LookupEnv("DISCORD_MEMBER_APPLICATION_CHANNEL_ID"); exists {
		// 旧環境変数は移行期間中のフォールバックとして扱う。
		DiscordConfig.Guild.NotificationChannelID = channelID
	} else {
		DiscordConfig.Guild.NotificationChannelID = values["discord.notification_channel_id"]
	}
	if DiscordConfig.ClientID == "" || DiscordConfig.ClientSecret == "" || DiscordConfig.Guild.ID == "" || DiscordConfig.Guild.MemberRoleID == "" || DiscordConfig.BotToken == "" {
		panic("Discord configuration is not fully set in environment variables or database settings")
	}
	EnvEnv := os.Getenv("ENV")
	if EnvEnv == "" {
		EnvEnv = Env
	}
	DiscordApiVersionEnv := settings.Resolve(values, "discord.api_version", DiscordApiVersion)
	return &Config{
		Env:                EnvEnv,
		AppName:            AppNameEnv,
		FrontendURL:        FrontendURLEnv,
		IssuerURL:          IssuerURLEnv,
		IssuerInternalURL:  IssuerInternalURLEnv,
		EmailSenderURL:     EmailSenderURLEnv,
		Version:            version,
		DiscordConfig:      DiscordConfig,
		DiscordApiVersion:  DiscordApiVersionEnv,
		GitHubClientID:     settings.Resolve(values, "github.client_id", ""),
		GitHubClientSecret: settings.Resolve(values, "github.client_secret", ""),
	}
}
