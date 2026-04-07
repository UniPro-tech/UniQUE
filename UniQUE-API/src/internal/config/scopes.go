package config

type ScopesConfig struct {
	AllowedScopes []string
}

var Scopes = ScopesConfig{
	AllowedScopes: []string{"openid", "profile", "email", "announcements.write", "announcements.delete"},
}
