package internal

import (
	"os"
)

type Colors struct {
	Primary int
	Success int
	Warning int
	Error   int
}

type Config struct {
	BotName      string
	Description  string
	AdminGuildID string
	AdminRoleID  string
	BotVersion   string
	URL          string
	GitHub       string
	Colors       Colors
}

type BotContext struct {
	Config *Config
}

var (
	Version   = "latest"
	GitCommit = "unknown"
	GitBranch = "unknown"
)

// envが設定されていない場合のデフォルト値
var (
	BotName        = "UniBot"
	Description    = "UniBotはデジタル創作サークルUniProjectの内製Discord Botです。"
	AdminGuildID   = "1191346186880286770"
	AdminRoleID    = "1390633352360628234"
	GitHubRepo     = "UniPro-tech/UniBot"
	HomePage       = "https://unibot.uniproject.jp"
	SupportServer  = "https://discord.gg/HYWB2aztr8"
	VoiceVoxURI    = "http://localhost:50021"
	VoiceVoxAPIKey = ""
)

func LoadConfig() *Config {
	version := Version

	if Version == "latest" {
		version = GitBranch + "@" + GitCommit
	} else {
		version = Version + "+" + GitCommit
	}

	// envから設定を読み込む
	BotNameEnv := os.Getenv("CONFIG_BOT_NAME")
	if BotNameEnv == "" {
		BotNameEnv = BotName
	}
	DescriptionEnv := os.Getenv("CONFIG_DESCRIPTION")
	if DescriptionEnv == "" {
		DescriptionEnv = Description
	}
	AdminGuildIDEnv := os.Getenv("CONFIG_ADMIN_GUILD_ID")
	if AdminGuildIDEnv == "" {
		AdminGuildIDEnv = AdminGuildID
	}
	AdminRoleIDEnv := os.Getenv("CONFIG_ADMIN_ROLE_ID")
	if AdminRoleIDEnv == "" {
		AdminRoleIDEnv = AdminRoleID
	}
	GitHubRepoEnv := os.Getenv("CONFIG_GITHUB_REPO")
	if GitHubRepoEnv == "" {
		GitHubRepoEnv = GitHubRepo
	}
	HomePageEnv := os.Getenv("CONFIG_HOME_PAGE")
	if HomePageEnv == "" {
		HomePageEnv = HomePage
	}

	return &Config{
		BotName:      BotNameEnv,
		Description:  DescriptionEnv,
		AdminGuildID: AdminGuildIDEnv,
		AdminRoleID:  AdminRoleIDEnv,
		BotVersion:   version,
		URL:          HomePageEnv,
		GitHub:       "https://github.com/" + GitHubRepoEnv,
		Colors: Colors{
			Primary: 0x3498DB,
			Success: 0x2ECC71,
			Warning: 0xF1C40F,
			Error:   0xE74C3C,
		},
	}
}
