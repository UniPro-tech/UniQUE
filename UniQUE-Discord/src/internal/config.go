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
	BotName           string
	Description       string
	AdminGuildID      string
	AdminRoleID       string
	BotVersion        string
	GitHub            string
	Colors            Colors
	UniqueAPIBaseURL  string
	UniqueFrontendURL string
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
	BotName           = "UniQUE System"
	Description       = "UniQUE SystemのBotです。"
	AdminGuildID      = "1191346186880286770"
	AdminRoleID       = "1390633352360628234"
	GitHubRepo        = "UniPro-tech/UniQUE"
	UniqueAPIBaseURL  = "http://localhost:8001"
	UniqueFrontendURL = "http://localhost:3000"
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
	UniqueAPIBaseURLEnv := os.Getenv("CONFIG_UNIQUE_API_BASE_URL")
	if UniqueAPIBaseURLEnv == "" {
		UniqueAPIBaseURLEnv = UniqueAPIBaseURL
	}
	UniqueFrontendURLEnv := os.Getenv("CONFIG_UNIQUE_FRONTEND_URL")
	if UniqueFrontendURLEnv == "" {
		UniqueFrontendURLEnv = UniqueFrontendURL
	}

	return &Config{
		BotName:           BotNameEnv,
		Description:       DescriptionEnv,
		AdminGuildID:      AdminGuildIDEnv,
		AdminRoleID:       AdminRoleIDEnv,
		BotVersion:        version,
		GitHub:            "https://github.com/" + GitHubRepoEnv,
		UniqueAPIBaseURL:  UniqueAPIBaseURLEnv,
		UniqueFrontendURL: UniqueFrontendURLEnv,
		Colors: Colors{
			Primary: 0x3498DB,
			Success: 0x2ECC71,
			Warning: 0xF1C40F,
			Error:   0xE74C3C,
		},
	}
}
