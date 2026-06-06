package uniqueapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func GetUserInfo(ctx *internal.BotContext, userId snowflake.ID) (UserDTO, error) {
	var uniqueAPIClient = &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/users/external_identities/search?provider=discord&external_user_id=%s", ctx.Config.UniqueAPIBaseURL, userId.String()), nil)
	if err != nil {
		return UserDTO{}, err
	}
	resp, err := uniqueAPIClient.Do(req)
	if err != nil {
		return UserDTO{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UserDTO{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return UserDTO{}, fmt.Errorf("failed to get user info: %s, status: %d", string(body), resp.StatusCode)
	}

	var result struct {
		Data []ExternalUserDTO `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return UserDTO{}, err
	}

	if len(result.Data) == 0 {
		return UserDTO{}, nil
	}

	targetUserId := result.Data[0].UserID
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/internal/users/%s", ctx.Config.UniqueAPIBaseURL, targetUserId), nil)
	if err != nil {
		return UserDTO{}, err
	}
	resp, err = uniqueAPIClient.Do(req)
	if err != nil {
		return UserDTO{}, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return UserDTO{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return UserDTO{}, fmt.Errorf("failed to get user info: %s, status: %d", string(body), resp.StatusCode)
	}

	var userResult UserDTO

	if err := json.Unmarshal(body, &userResult); err != nil {
		return UserDTO{}, err
	}

	return userResult, nil
}
