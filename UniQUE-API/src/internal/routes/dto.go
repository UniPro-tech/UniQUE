package routes

import (
	"encoding/json"
	"time"
)

// Nullable is a generic type used for PATCH requests to distinguish
// between "field not present" and "field present with null".
type Nullable[T any] struct {
	Set   bool
	Value *T
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	n.Set = true

	if string(data) == "null" {
		n.Value = nil
		return nil
	}

	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	n.Value = &v
	return nil
}

// UserDTO represents user + optional profile for API responses
type UserDTO struct {
	ID                string      `json:"id"`
	CustomID          string      `json:"custom_id"`
	Email             string      `json:"email"`
	ExternalEmail     string      `json:"external_email"`
	PendingEmail      string      `json:"pending_email,omitempty"`
	EmailVerified     bool        `json:"email_verified"`
	AffiliationPeriod string      `json:"affiliation_period,omitempty"`
	Status            string      `json:"status"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	Profile           *ProfileDTO `json:"profile,omitempty"`
	IsTOTPEnabled     bool        `json:"is_totp_enabled"`
}

type UserApproveDTO struct {
	Email               string `json:"email" binding:"required,email"`
	AffiliationPeriod   string `json:"affiliation_period"`
	JoinedAt            string `json:"joined_at"`
	SakuraEmailPassword string `json:"sakura_email_password"`
}

type UserPasswordChangeDTO struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// UserListResponse represents response for listing users
type UserListResponse struct {
	Data []UserDTO `json:"data"`
}

// ProfileDTO is a minimal profile representation embedded in UserDTO
type ProfileDTO struct {
	UserID           string `json:"user_id"`
	DisplayName      string `json:"display_name,omitempty"`
	Bio              string `json:"bio,omitempty"`
	WebsiteURL       string `json:"website_url,omitempty"`
	TwitterHandle    string `json:"twitter_handle,omitempty"`
	Birthdate        string `json:"birthdate,omitempty"`
	BirthdateVisible *bool  `json:"birthdate_visible,omitempty"`
	JoinedAt         string `json:"joined_at,omitempty"`
	IsAdult          *bool  `json:"is_adult,omitempty"`
}

// PatchProfileRequest is used for PATCH /users/:id
type PatchProfileRequest struct {
	DisplayName      Nullable[string] `json:"display_name"`
	Bio              Nullable[string] `json:"bio"`
	WebsiteURL       Nullable[string] `json:"website_url"`
	TwitterHandle    Nullable[string] `json:"twitter_handle"`
	Birthdate        Nullable[string] `json:"birthdate"`
	BirthdateVisible Nullable[bool]   `json:"birthdate_visible"`
	JoinedAt         Nullable[string] `json:"joined_at"`
}

// PatchUserRequest is used for PATCH /users/:id
type PatchUserRequest struct {
	CustomID          Nullable[string]     `json:"custom_id,omitempty"`
	Email             Nullable[string]     `json:"email,omitempty"`
	ExternalEmail     Nullable[string]     `json:"external_email,omitempty"`
	AffiliationPeriod Nullable[string]     `json:"affiliation_period,omitempty"`
	Status            Nullable[string]     `json:"status,omitempty"`
	Profile           *PatchProfileRequest `json:"profile,omitempty"`
}

// CreateUserRequest is used for POST /users
type CreateUserRequest struct {
	CustomID          string      `json:"custom_id" binding:"required"`
	Email             string      `json:"email" binding:"omitempty,email"`
	Password          string      `json:"password,omitempty"`
	ExternalEmail     string      `json:"external_email,omitempty" binding:"required,email"`
	Status            string      `json:"status,omitempty"`
	AffiliationPeriod string      `json:"affiliation_period,omitempty"`
	Profile           *ProfileDTO `json:"profile,omitempty"`
}

// UpdateUserRequest is used for PUT /users/:id
type UpdateUserRequest struct {
	CustomID          *string              `json:"custom_id,omitempty"`
	Email             *string              `json:"email,omitempty"`
	ExternalEmail     *string              `json:"external_email,omitempty"`
	AffiliationPeriod *string              `json:"affiliation_period,omitempty"`
	Status            *string              `json:"status,omitempty"`
	Profile           *PatchProfileRequest `json:"profile,omitempty"`
}

// RoleDTO represents role resource for API responses
type RoleDTO struct {
	ID                string     `json:"id"`
	CustomID          string     `json:"custom_id"`
	Name              string     `json:"name"`
	Description       string     `json:"description,omitempty"`
	PermissionBitmask int64      `json:"permission_bitmask"`
	IsDefault         bool       `json:"is_default"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}

// RoleListResponse wraps a list of roles
type RoleListResponse struct {
	Data []RoleDTO `json:"data"`
}

// CreateRoleRequest is used for POST /roles
type CreateRoleRequest struct {
	CustomID          string `json:"custom_id" binding:"required"`
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description,omitempty"`
	PermissionBitmask int64  `json:"permission_bitmask" binding:"gte=0"`
	IsDefault         *bool  `json:"is_default,omitempty"`
	AssignToExisting  *bool  `json:"assign_to_existing,omitempty"`
}

// UpdateRoleRequest is used for PUT /roles/:id
type UpdateRoleRequest struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	PermissionBitmask *int64  `json:"permission_bitmask,omitempty"`
	IsDefault         *bool   `json:"is_default,omitempty"`
}

// PatchRoleRequest is used for PATCH /roles/:id
type PatchRoleRequest struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	PermissionBitmask *int64  `json:"permission_bitmask,omitempty"`
	IsDefault         *bool   `json:"is_default,omitempty"`
}

// ApplicationDTO represents application resource for API responses
type ApplicationDTO struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description,omitempty"`
	WebsiteURL       string     `json:"website_url,omitempty"`
	PrivacyPolicyURL string     `json:"privacy_policy_url,omitempty"`
	UserID           string     `json:"user_id"`
	PublicClient     bool       `json:"public_client,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

// ApplicationListResponse wraps a list of applications
type ApplicationListResponse struct {
	Data []ApplicationDTO `json:"data"`
}

// ExternalIdentityDTO represents an external identity linked to a user.
// Returns data from ID token claims and provider userinfo only (no raw tokens).
type ExternalIdentityDTO struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	Provider       string `json:"provider"`
	ExternalUserID string `json:"external_user_id"`
	// Common normalised fields from provider userinfo
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Email       string `json:"email,omitempty"`
	// Decoded ID Token claims (JWT payload)
	IDTokenClaims map[string]interface{} `json:"id_token_claims,omitempty"`
	// Raw provider-specific userinfo data
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ExternalIdentityListResponse wraps a list of external identities
type ExternalIdentityListResponse struct {
	Data []ExternalIdentityDTO `json:"data"`
}

// CreateExternalIdentityRequest is used for linking an external account
type CreateExternalIdentityRequest struct {
	Provider       string     `json:"provider" binding:"required"`
	ExternalUserID string     `json:"external_user_id" binding:"required"`
	IDToken        string     `json:"id_token,omitempty"`
	AccessToken    string     `json:"access_token,omitempty"`
	RefreshToken   string     `json:"refresh_token,omitempty"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty"`
}

// EmailVerifyDiscordLinkRequest is used to link Discord during email verification
type EmailVerifyDiscordLinkRequest struct {
	Code           string     `json:"code" binding:"required"`
	ExternalUserID string     `json:"external_user_id" binding:"required"`
	AccessToken    string     `json:"access_token" binding:"required"`
	RefreshToken   string     `json:"refresh_token,omitempty"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty"`
}

// CreateApplicationRequest is used for POST /applications
type CreateApplicationRequest struct {
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description,omitempty"`
	WebsiteURL       string `json:"website_url,omitempty"`
	PrivacyPolicyURL string `json:"privacy_policy_url,omitempty"`
	ClientSecret     string `json:"client_secret,omitempty"`
	PublicClient     bool   `json:"public_client,omitempty"`
	UserID           string `json:"user_id" binding:"required"`
}

// UpdateApplicationRequest is used for PUT /applications/:id
// Allow nullable semantics by aliasing to PatchApplicationRequest so
// PUT can contain explicit nulls for fields.
type UpdateApplicationRequest = PatchApplicationRequest

// PatchApplicationRequest is used for PATCH /applications/:id
type PatchApplicationRequest struct {
	Name             Nullable[string] `json:"name,omitempty"`
	Description      Nullable[string] `json:"description,omitempty"`
	WebsiteURL       Nullable[string] `json:"website_url,omitempty"`
	PrivacyPolicyURL Nullable[string] `json:"privacy_policy_url,omitempty"`
	ClientSecret     Nullable[string] `json:"client_secret,omitempty"`
	PublicClient     Nullable[bool]   `json:"public_client,omitempty"`
}

// CreateUserRoleRequest is used to assign a role to a user
type CreateUserRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

// PermissionsResponse represents user permissions as bitmask and text
type PermissionsResponse struct {
	PermissionBitmask int64    `json:"permission_bitmask"`
	PermissionsText   []string `json:"permissions_text"`
}

// CreateApplicationOwnerRequest is used to assign/replace an application owner
type CreateApplicationOwnerRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// RedirectURIDTO represents a redirect URI for API responses (omits GORM deleted metadata)
type RedirectURIDTO struct {
	URI       string    `json:"uri"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RedirectURIListResponse struct {
	Data []RedirectURIDTO `json:"data"`
}

type CreateRedirectURIRequest struct {
	URI string `json:"uri" binding:"required,url"`
}

// EmailCodeCheckRequest is used to verify email codes
type EmailCodeCheckRequest struct {
	Code string `json:"code" binding:"required"`
}

// EmailCodeCheckResponse is the response for email code verification
type EmailCodeCheckResponse struct {
	Valid bool   `json:"valid"`
	Type  string `json:"type"` // "signup", "change", or "migration"
}

type CreateAnnouncementRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	IsPinned *bool  `json:"is_pinned,omitempty"`
}

type UpdateAnnouncementRequest struct {
	Title    *string `json:"title,omitempty"`
	Content  *string `json:"content,omitempty"`
	IsPinned *bool   `json:"is_pinned,omitempty"`
}

type PatchAnnouncementRequest struct {
	Title    *string `json:"title,omitempty"`
	Content  *string `json:"content,omitempty"`
	IsPinned *bool   `json:"is_pinned,omitempty"`
}

type AnnouncementDTO struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	CreatedBy UserDTO    `json:"created_by"`
	IsPinned  bool       `json:"is_pinned"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type AnnouncementListResponse struct {
	Data []AnnouncementDTO `json:"data"`
}

// formatDate formats a *time.Time as "YYYY-MM-DD" or returns "" if nil.
func formatDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// ptrToString converts *string to string, returning "" if nil
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// stringToPtr converts string to *string, returning nil if empty
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// timeToTimePtr converts time.Time to *time.Time
func timeToTimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
