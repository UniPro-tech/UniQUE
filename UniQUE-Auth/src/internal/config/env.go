package config

import (
	"crypto/rsa"
	"os"
)

type Config struct {
	AppName       string
	Version       string
	FrontendURL   string
	IssuerURL     string
	MailServerURL string
	KeyPaths      KeyPathsConfig
	KeyPairs      []KeyPairConfig
}

type KeyPathsConfig struct {
	KeyType         string
	PublicKeysPath  string
	PrivateKeysPath string
}

type KeyPairConfig struct {
	PublicKey  rsa.PublicKey
	PrivateKey rsa.PrivateKey
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
	KeyPaths    = KeyPathsConfig{
		KeyType:         "RSA",
		PublicKeysPath:  "../keys/public",
		PrivateKeysPath: "../keys/private",
	}
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
	MailServerURLEnv := os.Getenv("CONFIG_MAIL_SERVER_URL")
	if MailServerURLEnv == "" {
		MailServerURLEnv = "http://mailserver:8080"
	}
	KeyPathEnv := KeyPathsConfig{
		KeyType:         os.Getenv("CONFIG_KEY_TYPE"),
		PublicKeysPath:  os.Getenv("CONFIG_KEY_PUBLIC_PATH"),
		PrivateKeysPath: os.Getenv("CONFIG_KEY_PRIVATE_PATH"),
	}
	if KeyPathEnv.PublicKeysPath == "" {
		KeyPathEnv.PublicKeysPath = KeyPaths.PublicKeysPath
	}
	if KeyPathEnv.PrivateKeysPath == "" {
		KeyPathEnv.PrivateKeysPath = KeyPaths.PrivateKeysPath
	}
	if KeyPathEnv.KeyType == "" {
		KeyPathEnv.KeyType = KeyPaths.KeyType
	}
	keyPairs, err := ParseKeys(KeyPathEnv)
	if err != nil {
		panic(err)
	}
	return &Config{
		AppName:       AppNameEnv,
		FrontendURL:   FrontendURLEnv,
		IssuerURL:     IssuerURLEnv,
		MailServerURL: MailServerURLEnv,
		Version:       version,
		KeyPaths:      KeyPathEnv,
		KeyPairs:      keyPairs,
	}
}
