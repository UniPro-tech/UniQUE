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
	SmtpHost := os.Getenv("SMTP_HOST")
	if SmtpHost == "" {
		panic("SMTP Config not found")
	}
	SmtpHostPort := os.Getenv("SMTP_PORT")
	if SmtpHostPort == "" {
		panic("SMTP Config not found")
	}
	// Convert SmtpHostPort to int
	var SmtpHostPortInt int
	_, err := fmt.Sscanf(SmtpHostPort, "%d", &SmtpHostPortInt)
	if err != nil {
		panic("Invalid SMTP_PORT value")
	}
	SmtpUsername := os.Getenv("SMTP_USERNAME")
	if SmtpUsername == "" {
		panic("SMTP Config not found")
	}
	SmtpPassword := os.Getenv("SMTP_PASSWORD")
	if SmtpPassword == "" {
		panic("SMTP Config not found")
	}
	SmtpFrom := os.Getenv("SMTP_FROM")
	if SmtpFrom == "" {
		panic("SMTP Config not found")
	}
	SmtpSecure := os.Getenv("SMTP_SECURE")
	if SmtpSecure == "" {
		panic("SMTP Config not found")
	}
	FromName := os.Getenv("FROM_NAME")
	if FromName == "" {
		FromName = AppNameEnv
	}
	CopyrightName := os.Getenv("COPYRIGHT_NAME")
	if CopyrightName == "" {
		CopyrightName = AppNameEnv
	}
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
