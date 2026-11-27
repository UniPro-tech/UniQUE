package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/logging"
	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
)

// stringToPtr converts string to *string, returning nil if empty
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// RefreshExternalToken checks if the token is expired and refreshes it using the refresh token.
// Returns the (possibly updated) ExternalIdentity and any error.
func RefreshExternalToken(ctx context.Context, ei *model.ExternalIdentity, q *query.Query, cfg *config.Config) (*model.ExternalIdentity, error) {
	if ei.RefreshToken == "" {
		return ei, nil // no refresh token available
	}
	if ei.TokenExpiresAt != nil && !ei.TokenExpiresAt.IsZero() && ei.TokenExpiresAt.After(time.Now()) {
		return ei, nil // not expired yet
	}

	switch ei.Provider {
	case "discord":
		return refreshDiscordToken(ctx, ei, q, cfg)
	case "github":
		return refreshGitHubToken(ctx, ei, q, cfg)
	default:
		return ei, fmt.Errorf("unsupported provider: %s", ei.Provider)
	}
}

// refreshDiscordToken refreshes tokens via Discord's OAuth2 endpoint.
func refreshDiscordToken(ctx context.Context, ei *model.ExternalIdentity, q *query.Query, cfg *config.Config) (*model.ExternalIdentity, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {ei.RefreshToken},
		"client_id":     {cfg.DiscordConfig.ClientID},
		"client_secret": {cfg.DiscordConfig.ClientSecret},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://discord.com/api/%s/oauth2/token", cfg.DiscordApiVersion), strings.NewReader(data.Encode()))
	if err != nil {
		return ei, fmt.Errorf("discord token refresh request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	logging.Inject(ctx, req.Header)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ei, fmt.Errorf("discord token refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxLoggedBodyBytes+1))
		slog.Warn("discord_token_refresh_failed",
			slog.Int("status", resp.StatusCode),
			slog.Any("response_body", MaskedBody(body)),
		)
		return ei, fmt.Errorf("discord token refresh returned status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		IDToken      string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ei, fmt.Errorf("discord token refresh decode failed: %w", err)
	}

	updates := map[string]interface{}{
		"access_token":     result.AccessToken,
		"token_expires_at": time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}
	if result.RefreshToken != "" {
		updates["refresh_token"] = result.RefreshToken
	}
	if result.IDToken != "" {
		updates["id_token"] = result.IDToken
	}

	if _, err := q.ExternalIdentity.Where(
		query.ExternalIdentity.ID.Eq(ei.ID),
	).Updates(updates); err != nil {
		return ei, fmt.Errorf("failed to update discord tokens in db: %w", err)
	}

	// reflect updates in returned struct
	ei.AccessToken = result.AccessToken
	newExpiry := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	ei.TokenExpiresAt = &newExpiry
	if result.RefreshToken != "" {
		ei.RefreshToken = result.RefreshToken
	}
	if result.IDToken != "" {
		ei.IDToken = stringToPtr(result.IDToken)
	}
	return ei, nil
}

// refreshGitHubToken refreshes tokens via GitHub's OAuth2 endpoint.
func refreshGitHubToken(ctx context.Context, ei *model.ExternalIdentity, q *query.Query, cfg *config.Config) (*model.ExternalIdentity, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {ei.RefreshToken},
		"client_id":     {cfg.GitHubClientID},
		"client_secret": {cfg.GitHubClientSecret},
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return ei, fmt.Errorf("github token refresh request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	logging.Inject(ctx, req.Header)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ei, fmt.Errorf("github token refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxLoggedBodyBytes+1))
		slog.Warn("github_token_refresh_failed",
			slog.Int("status", resp.StatusCode),
			slog.Any("response_body", MaskedBody(body)),
		)
		return ei, fmt.Errorf("github token refresh returned status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		IDToken      string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ei, fmt.Errorf("github token refresh decode failed: %w", err)
	}

	if result.AccessToken == "" {
		return ei, fmt.Errorf("github token refresh returned empty access_token")
	}

	updates := map[string]interface{}{
		"access_token": result.AccessToken,
	}
	if result.ExpiresIn > 0 {
		updates["token_expires_at"] = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	}
	if result.RefreshToken != "" {
		updates["refresh_token"] = result.RefreshToken
	}
	if result.IDToken != "" {
		updates["id_token"] = result.IDToken
	}

	if _, err := q.ExternalIdentity.Where(
		query.ExternalIdentity.ID.Eq(ei.ID),
	).Updates(updates); err != nil {
		return ei, fmt.Errorf("failed to update github tokens in db: %w", err)
	}

	// reflect updates in returned struct
	ei.AccessToken = result.AccessToken
	if result.ExpiresIn > 0 {
		newExpiry := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
		ei.TokenExpiresAt = &newExpiry
	}
	if result.RefreshToken != "" {
		ei.RefreshToken = result.RefreshToken
	}
	if result.IDToken != "" {
		ei.IDToken = stringToPtr(result.IDToken)
	}
	return ei, nil
}
