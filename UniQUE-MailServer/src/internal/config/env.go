package config

import (
	"fmt"
	"os"

	"github.com/UniPro-tech/UniQUE-MailServer/internal/settings"
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
	AppNameEnv := settings.Resolve(values, "CONFIG_APP_NAME", "application.name", AppName)
	FrontendURLEnv := os.Getenv("CONFIG_FRONTEND_URL")
	if FrontendURLEnv == "" {
		FrontendURLEnv = FrontendURL
	}
	IssuerURLEnv := os.Getenv("CONFIG_ISSUER_URL")
	if IssuerURLEnv == "" {
		IssuerURLEnv = IssuerURL
	}
	SmtpHost := settings.Resolve(values, "SMTP_HOST", "smtp.host", "")
	if SmtpHost == "" {
		panic("SMTP Config not found")
	}
	SmtpHostPort := settings.Resolve(values, "SMTP_PORT", "smtp.port", "")
	if SmtpHostPort == "" {
		panic("SMTP Config not found")
	}
	// Convert SmtpHostPort to int
	var SmtpHostPortInt int
	_, err := fmt.Sscanf(SmtpHostPort, "%d", &SmtpHostPortInt)
	if err != nil {
		panic("Invalid SMTP_PORT value")
	}
	SmtpUsername := settings.Resolve(values, "SMTP_USERNAME", "smtp.username", "")
	if SmtpUsername == "" {
		panic("SMTP Config not found")
	}
	SmtpPassword := settings.Resolve(values, "SMTP_PASSWORD", "smtp.password", "")
	if SmtpPassword == "" {
		panic("SMTP Config not found")
	}
	SmtpFrom := settings.Resolve(values, "SMTP_FROM", "smtp.from", "")
	if SmtpFrom == "" {
		panic("SMTP Config not found")
	}
	SmtpSecure := settings.Resolve(values, "SMTP_SECURE", "smtp.secure", "")
	if SmtpSecure == "" {
		panic("SMTP Config not found")
	}
	FromName := settings.Resolve(values, "FROM_NAME", "mail.from_name", AppNameEnv)
	CopyrightName := settings.Resolve(values, "COPYRIGHT_NAME", "mail.copyright_name", AppNameEnv)
	return &Config{
		AppName:       AppNameEnv,
		FrontendURL:   FrontendURLEnv,
		CopyrightName: CopyrightName,
		IssuerURL:     IssuerURLEnv,
		Version:       version,
		SmtpConfig: SmtpConfig{
			Host:     SmtpHost,
			Port:     SmtpHostPortInt,
			Username: SmtpUsername,
			Password: SmtpPassword,
			From:     SmtpFrom,
			FromName: FromName,
			Secure:   SmtpSecure == "true",
		},
	}
}
