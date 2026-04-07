package config

import (
	"os"
)

type DiscordGuildConfig struct {
	ID           string
	MemberRoleID string
}

type DiscordConfig struct {
	ClientID     string
	ClientSecret string
	Guild        DiscordGuildConfig
	BotToken     string
}

type Config struct {
	AppName            string
	Version            string
	FrontendURL        string
	IssuerURL          string
	IssuerInternalURL  string
	EmailSenderURL     string
	DiscordConfig      DiscordConfig
	GitHubClientID     string
	GitHubClientSecret string
}

// envが設定されていない場合のデフォルト値
var (
	Version   = "latest"
	GitCommit = "unknown"
	GitBranch = "unknown"
)

var (
	AppName           = "UniQUE"
	FrontendURL       = "http://localhost:3000"
	IssuerURL         = "http://localhost:8080"
	IssuerInternalURL = "http://localhost:8080"
	EmailSenderURL    = "http://localhost:8080"
)

func LoadConfig() *Config {
	version := Version

	if Version == "latest" {
		version = GitBranch + "@" + GitCommit
	} else {
		version = Version + "+" + GitCommit
	}

	// envから設定を読み込む
	AppNameEnv := os.Getenv("CONFIG_APP_NAME")
	if AppNameEnv == "" {
		AppNameEnv = AppName
	}
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
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		Guild: DiscordGuildConfig{
			ID:           os.Getenv("DISCORD_GUILD_ID"),
			MemberRoleID: os.Getenv("DISCORD_MEMBER_ROLE_ID"),
		},
		BotToken: os.Getenv("DISCORD_BOT_TOKEN"),
	}
	if DiscordConfig.ClientID == "" || DiscordConfig.ClientSecret == "" || DiscordConfig.Guild.ID == "" || DiscordConfig.Guild.MemberRoleID == "" || DiscordConfig.BotToken == "" {
		panic("Discord configuration is not fully set in environment variables")
	}
	return &Config{
		AppName:            AppNameEnv,
		FrontendURL:        FrontendURLEnv,
		IssuerURL:          IssuerURLEnv,
		IssuerInternalURL:  IssuerInternalURLEnv,
		EmailSenderURL:     EmailSenderURLEnv,
		Version:            version,
		DiscordConfig:      DiscordConfig,
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	}
}
