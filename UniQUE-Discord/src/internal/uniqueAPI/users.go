package uniqueapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE/Discord/internal"
	"github.com/disgoorg/snowflake/v2"
)

type ExternalUserDTO struct {
	AvatarURL      string         `json:"avatar_url"`
	CreatedAt      string         `json:"created_at"`
	DisplayName    string         `json:"display_name"`
	Email          string         `json:"email"`
	ExternalUserID string         `json:"external_user_id"`
	ID             string         `json:"id"`
	IDTokenClaims  map[string]any `json:"id_token_claims"`
	Provider       string         `json:"provider"`
	ProviderData   map[string]any `json:"provider_data"`
	UpdatedAt      string         `json:"updated_at"`
	UserID         string         `json:"user_id"`
	Username       string         `json:"username"`
}

type UserProfileDTO struct {
	Bio              string `json:"bio"`
	Birthdate        string `json:"birthdate"`
	BirthdateVisible bool   `json:"birthdate_visible"`
	DisplayName      string `json:"display_name"`
	IsAdult          bool   `json:"is_adult"`
	JoinedAt         string `json:"joined_at"`
	TwitterHandle    string `json:"twitter_handle"`
	UserID           string `json:"user_id"`
	WebsiteURL       string `json:"website_url"`
}

type UserDTO struct {
	AffiliationPeriod string         `json:"affiliation_period"`
	CreatedAt         string         `json:"created_at"`
	CustomID          string         `json:"custom_id"`
	Email             string         `json:"email"`
	EmailVerified     bool           `json:"email_verified"`
	ExternalEmail     string         `json:"external_email"`
	ID                string         `json:"id"`
	IsTotpEnabled     bool           `json:"is_totp_enabled"`
	PendingEmail      string         `json:"pending_email"`
	Profile           UserProfileDTO `json:"profile"`
	Status            string         `json:"status"`
	UpdatedAt         string         `json:"updated_at"`
}

func GetUserInfo(ctx *internal.BotContext, userId snowflake.ID) (*UserDTO, error) {
	var uniqueAPIClient = &http.Client{Timeout: 10 * time.Second}
	requestID, traceID := newID(16), newID(16)
	req, err := http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("%s/internal/users/external_identities/search?provider=discord&external_user_id=%s", ctx.Config.UniqueAPIBaseURL, userId.String()), nil)
	if err != nil {
		return nil, err
	}
	setCorrelation(req, requestID, traceID)
	resp, err := uniqueAPIClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	slog.Debug("unique_api_response", slog.String("operation", "search_external_identity"), slog.Int("status", resp.StatusCode), slog.String("request_id", requestID), slog.String("trace_id", traceID))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s, status: %d", string(body), resp.StatusCode)
	}

	var result struct {
		Data []ExternalUserDTO `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	targetUserId := result.Data[0].UserID
	req, err = http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("%s/internal/users/%s", ctx.Config.UniqueAPIBaseURL, targetUserId), nil)
	if err != nil {
		return nil, err
	}
	setCorrelation(req, requestID, traceID)
	resp, err = uniqueAPIClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	slog.Debug("unique_api_response", slog.String("operation", "get_user"), slog.Int("status", resp.StatusCode), slog.String("request_id", requestID), slog.String("trace_id", traceID))

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s, status: %d", string(body), resp.StatusCode)
	}

	var userResult *UserDTO

	if err := json.Unmarshal(body, &userResult); err != nil {
		return nil, err
	}

	return userResult, nil
}

func newID(size int) string {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return strings.Repeat("0", size*2)
	}
	return hex.EncodeToString(buffer)
}

func setCorrelation(req *http.Request, requestID, traceID string) {
	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("X-Trace-ID", traceID)
	req.Header.Set("traceparent", "00-"+traceID+"-"+newID(8)+"-01")
}
