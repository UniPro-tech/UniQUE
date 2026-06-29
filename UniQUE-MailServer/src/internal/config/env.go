package config

import (
	"fmt"
	"os"
)

type SmtpConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Secure   bool
	FromName string
}

type Config struct {
	AppName       string
	CopyrightName string
	Version       string
	FrontendURL   string
	IssuerURL     string
	SmtpConfig    SmtpConfig
	LogLevel      string
	Debug         bool
}

// envが設定されていない場合のデフォルト値
var (
	Version   = "latest"
	GitCommit = "unknown"
	GitBranch = "unknown"
)

var (
	AppName     = "UniQUE"
	FrontendURL = "http://localhost:3000"
	IssuerURL   = "http://localhost:8080"
	LogLevel    = "info"
	Debug       = false
)

func LoadConfig() *Config {
	// Constract
	version := Version

	if Version == "latest" {
		version = GitBranch + "@" + GitCommit
	} else {
		version = Version + "+" + GitCommit
	}

	// envからロード
	AppNameEnv := os.Getenv("CONFIG_APP_NAME")
	if AppNameEnv == "" {
		AppNameEnv = AppName
	}

	CopyrightName := os.Getenv("COPYRIGHT_NAME")
	if CopyrightName == "" {
		CopyrightName = AppNameEnv
	}

	DebugEnv := os.Getenv("DEBUG") == "true"
	if DebugEnv {
		LogLevel = "debug"
	}

	FrontendURLEnv := os.Getenv("CONFIG_FRONTEND_URL")
	if FrontendURLEnv == "" {
		FrontendURLEnv = FrontendURL
	}

	IssuerURLEnv := os.Getenv("CONFIG_ISSUER_URL")
	if IssuerURLEnv == "" {
		IssuerURLEnv = IssuerURL
	}

	LogLevelEnv := os.Getenv("LOG_LEVEL")
	if LogLevelEnv == "" {
		LogLevelEnv = LogLevel
	}

	SmtpHost := os.Getenv("SMTP_HOST")
	if SmtpHost == "" {
		panic("SMTP Config not found")
	}

	SmtpHostPort := os.Getenv("SMTP_PORT")
	if SmtpHostPort == "" {
		panic("SMTP Config not found")
	}

	// SMTP Settings
	SmtpFrom := os.Getenv("SMTP_FROM")
	if SmtpFrom == "" {
		panic("SMTP Config not found")
	}

	FromName := os.Getenv("FROM_NAME")
	if FromName == "" {
		FromName = AppNameEnv
	}

	// ポート番号をintに変換
	var SmtpHostPortInt int
	_, err := fmt.Sscanf(SmtpHostPort, "%d", &SmtpHostPortInt)
	if err != nil {
		panic("Invalid SMTP_PORT value")
	}

	SmtpPassword := os.Getenv("SMTP_PASSWORD")
	if SmtpPassword == "" {
		panic("SMTP Config not found")
	}

	SmtpSecure := os.Getenv("SMTP_SECURE")
	if SmtpSecure == "" {
		panic("SMTP Config not found")
	}

	SmtpUsername := os.Getenv("SMTP_USERNAME")
	if SmtpUsername == "" {
		panic("SMTP Config not found")
	}

	return &Config{
		AppName:       AppNameEnv,
		CopyrightName: CopyrightName,
		Debug:         DebugEnv,
		FrontendURL:   FrontendURLEnv,
		IssuerURL:     IssuerURLEnv,
		LogLevel:      LogLevelEnv,
		SmtpConfig: SmtpConfig{
			From:     SmtpFrom,
			FromName: FromName,
			Host:     SmtpHost,
			Port:     SmtpHostPortInt,
			Password: SmtpPassword,
			Secure:   SmtpSecure == "true",
			Username: SmtpUsername,
		},
		Version: version,
	}
}
