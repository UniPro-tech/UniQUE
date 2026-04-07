package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/model"
)

// ProviderUserInfo holds normalised common fields + raw provider-specific data.
type ProviderUserInfo struct {
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Email       string `json:"email,omitempty"`
	// Raw response from the providers userinfo endpoint.
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

// FetchProviderUserInfo calls the providers userinfo API using the stored
// access token and returns normalised common fields plus the raw JSON.
func FetchProviderUserInfo(ei *model.ExternalIdentity) (*ProviderUserInfo, error) {
	switch ei.Provider {
	case "discord":
		return fetchDiscordUserInfo(ei)
	case "github":
		return fetchGitHubUserInfo(ei)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", ei.Provider)
	}
}

// ----- Discord -----

func fetchDiscordUserInfo(ei *model.ExternalIdentity) (*ProviderUserInfo, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("discord userinfo request creation failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+ei.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discord userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord userinfo returned status %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("discord userinfo decode failed: %w", err)
	}

	info := &ProviderUserInfo{ProviderData: raw}

	// Normalise common fields
	if v, ok := raw["username"].(string); ok {
		info.Username = v
	}
	// global_name is the user's display name on Discord (nullable).
	if v, ok := raw["global_name"].(string); ok && v != "" {
		info.DisplayName = v
	}
	if v, ok := raw["email"].(string); ok {
		info.Email = v
	}

	// Build avatar URL per Discord docs:
	// https://discord.com/developers/docs/reference#image-formatting
	uid, _ := raw["id"].(string)
	if avatarHash, ok := raw["avatar"].(string); ok && avatarHash != "" && uid != "" {
		ext := "png"
		if strings.HasPrefix(avatarHash, "a_") {
			ext = "gif"
		}
		info.AvatarURL = fmt.Sprintf(
			"https://cdn.discordapp.com/avatars/%s/%s.%s?size=256",
			uid, avatarHash, ext,
		)
	} else if uid != "" {
		// avatar is null → use Discord default avatar.
		// Index = (user_id >> 22) % 6  for the new username system.
		id := new(big.Int)
		if _, ok := id.SetString(uid, 10); ok {
			index := new(big.Int).Rsh(id, 22)
			index.Mod(index, big.NewInt(6))
			info.AvatarURL = fmt.Sprintf(
				"https://cdn.discordapp.com/embed/avatars/%d.png",
				index.Int64(),
			)
		}
	}

	return info, nil
}

// ----- GitHub -----

func fetchGitHubUserInfo(ei *model.ExternalIdentity) (*ProviderUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("github userinfo request creation failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+ei.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github userinfo returned status %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("github userinfo decode failed: %w", err)
	}

	info := &ProviderUserInfo{ProviderData: raw}

	// Normalise common fields
	if v, ok := raw["login"].(string); ok {
		info.Username = v
	}
	if v, ok := raw["name"].(string); ok {
		info.DisplayName = v
	}
	if v, ok := raw["email"].(string); ok {
		info.Email = v
	}
	if v, ok := raw["avatar_url"].(string); ok {
		info.AvatarURL = v
	}

	return info, nil
}

// ----- ID Token decoding -----

// DecodeIDTokenClaims decodes the payload of a JWT ID token without
// signature verification (the token is already stored in our DB,
// so we trust it). Returns the claims as a generic map.
func DecodeIDTokenClaims(idToken string) (map[string]interface{}, error) {
	if idToken == "" {
		return nil, nil
	}

	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid id_token format: expected 3 parts, got %d", len(parts))
	}

	// base64url decode the payload (2nd part)
	payload := parts[1]
	// Add padding if necessary
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode id_token payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal id_token claims: %w", err)
	}

	return claims, nil
}
