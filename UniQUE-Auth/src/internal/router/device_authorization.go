package router

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/UniPro-tech/UniQUE-Auth/internal/model"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

// These implement the OAuth 2.0 Device Authorization Grant (RFC 8628) flow.
// See also https://datatracker.ietf.org/doc/html/rfc8628

type DeviceAuthorizationRequest struct {
	ClientID string `form:"client_id" binding:"required"`
	Scope    string `form:"scope" binding:"required"`
}

type DeviceAuthorizationResponse struct {
	DeviceCode    string `json:"device_code" binding:"required"`
	UserCode      string `json:"user_code" binding:"required"`
	ValidationURI string `json:"verification_uri" binding:"required"`
	// ExpiresIn is the number of seconds the device_code and user_code are valid for.
	ExpiresIn int `json:"expires_in" binding:"required"`
	// Interval is the minimum number of seconds that the client SHOULD wait between polling requests to the token endpoint.
	Interval int `json:"interval" binding:"required"`
}

// DeviceAuthorizationGet godoc
// @Summary device authorization request
// @Schemes
// @Description device authorization request(RFC 8628)を行うためのエンドポイントです。
// @Tags authorization
// @Accept x-www-form-urlencoded
// @Param client_id formData string true "Client ID"
// @Param scope formData string false "Scope"
// @Success 200 {object} DeviceAuthorizationResponse "Device authorization response"
// @Router /device_authorization [POST]
func DeviceAuthorizationGet(c *gin.Context) {
	var req DeviceAuthorizationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// validate client_id
	application, err := query.Application.Where(query.Application.ID.Eq(req.ClientID)).First()
	if err != nil || application == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client"})
		return
	}

	// Validate scopes
	requestedScopes := map[string]bool{}
	for _, scope := range splitAndTrim(req.Scope) {
		requestedScopes[scope] = true
	}
	for _, allowedScope := range config.Scopes.AllowedScopes {
		delete(requestedScopes, allowedScope)
	}
	if len(requestedScopes) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	userCode, err := util.MakeRandomStr(8)
	now := time.Now().UTC()
	e := ulid.Monotonic(rand.New(rand.NewSource(now.UnixNano())), 0)
	deviceCode := ulid.MustNew(ulid.Timestamp(now), e).String()

	var authReq *model.AuthorizationRequest
	// save the request parameters in the database transactionally
	err = query.Q.Transaction(func(tx *query.Query) error {
		now := time.Now().UTC()
		authReq = &model.AuthorizationRequest{
			ID:               ulid.Make().String(),
			ApplicationID:    req.ClientID,
			Scope:            req.Scope,
			UserCode:         &userCode,
			DeviceCode:       &deviceCode,
			DeviceFlowDenied: false,
			ExpiresAt:        now.Add(10 * time.Minute),
			CreatedAt:        now,
		}
		return tx.AuthorizationRequest.Create(authReq)
	})

	if err != nil || authReq == nil || authReq.ID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, DeviceAuthorizationResponse{
		DeviceCode:    deviceCode,
		UserCode:      userCode,
		ValidationURI: "https://unique.uniproject.jp/device",
		Interval:      7,
		ExpiresIn:     10 * 60, // 10 minutes
	})
}
