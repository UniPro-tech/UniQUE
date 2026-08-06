package router

import (
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"errors"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/UniPro-tech/UniQUE-Auth/internal/constants"
	"github.com/UniPro-tech/UniQUE-Auth/internal/model"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthorizationRequest struct {
	ClientID            string  `form:"client_id" binding:"required"`
	RedirectURI         string  `form:"redirect_uri" binding:"required"`
	ResponseType        string  `form:"response_type" binding:"required,oneof=code token"`
	Scope               string  `form:"scope" binding:"required"`
	State               *string `form:"state"`
	Nonce               *string `form:"nonce"`
	Prompt              *string `form:"prompt"`
	CodeChallenge       *string `form:"code_challenge"`
	CodeChallengeMethod *string `form:"code_challenge_method" binding:"omitempty,oneof=plain S256"`
}

type AuthorizationResponse struct {
	ClientID            string  `json:"client_id" binding:"required"`
	RedirectURI         *string `json:"redirect_uri" binding:"required"`
	ResponseType        *string `json:"response_type" binding:"required,oneof=code token"`
	Scope               string  `json:"scope" binding:"required"`
	State               *string `json:"state"`
	Nonce               *string `json:"nonce"`
	Prompt              *string `json:"prompt"`
	CodeChallenge       *string `json:"code_challenge"`
	CodeChallengeMethod *string `json:"code_challenge_method" binding:"omitempty,oneof=plain S256"`
}

// AuthorizationGet godoc
// @Summary authorization endpoint
// @Schemes
// @Description OAuth 2.0 / OIDC における認可エンドポイントです。クライアントアプリケーションがユーザーの認可を取得するために使用します。
// @Tags authorization
// @Param client_id query string true "Client ID"
// @Param redirect_uri query string true "Redirect URI"
// @Param response_type query string true "Response Type"
// @Param scope query string true "Scope"
// @Param state query string false "State"
// @Param nonce query string false "Nonce"
// @Param code_challenge query string false "Code Challenge"
// @Param code_challenge_method query string false "Code Challenge Method"
// @Success 302 {string} string "Redirect to frontend authorization page"
// @Router /authorization [get]
func AuthorizationGet(c *gin.Context) {
	// Validate query parameters
	var req AuthorizationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	contextConfig := c.MustGet("config").(*config.Config)

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.Redirect(302, config.FrontendURL+"/authorization?error=internal_server_error")
		return
	}
	q := query.Use(db)

	// If client_id, redirect_uri, response_type, scope are valid, redirect to frontend
	client, err := q.Application.Where(q.Application.ID.Eq(req.ClientID)).First()
	if err != nil || client == nil {
		c.Redirect(302, contextConfig.FrontendURL+"/authorization?error=invalid_client")
		return
	}

	redirectURI, err := q.RedirectURI.Where(q.RedirectURI.URI.Eq(req.RedirectURI), q.RedirectURI.ApplicationID.Eq(req.ClientID)).First()
	if err != nil || redirectURI == nil {
		c.Redirect(302, contextConfig.FrontendURL+"/authorization?error=invalid_redirect_uri")
		return
	}

	// Only "code" response type is supported for now
	if req.ResponseType != "code" {
		c.Redirect(302, contextConfig.FrontendURL+"/authorization?error=unsupported_response_type")
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
		c.Redirect(302, contextConfig.FrontendURL+"/authorization?error=invalid_scope")
		return
	}

	// Enforce PKCE for public clients: code_challenge must be present
	if client.PublicClient && req.CodeChallenge == nil {
		c.Redirect(302, contextConfig.FrontendURL+"/authorization?error=invalid_request")
		return
	}

	var authReq *model.AuthorizationRequest
	// save the request parameters in the database transactionally
	err = q.Transaction(func(tx *query.Query) error {
		now := time.Now().UTC()
		authReq = &model.AuthorizationRequest{
			ID:                  ulid.Make().String(),
			ApplicationID:       req.ClientID,
			RedirectURI:         &req.RedirectURI,
			ResponseType:        &req.ResponseType,
			Scope:               req.Scope,
			State:               req.State,
			Nonce:               req.Nonce,
			Prompt:              req.Prompt,
			CodeChallenge:       req.CodeChallenge,
			CodeChallengeMethod: req.CodeChallengeMethod,
			ExpiresAt:           now.Add(20 * time.Minute),
			CreatedAt:           now,
		}
		return tx.AuthorizationRequest.Create(authReq)
	})

	if err != nil || authReq == nil || authReq.ID == "" {
		c.Redirect(302, contextConfig.FrontendURL+"/authorization?error=internal_server_error")
		return
	}

	// Redirect to frontend authorization page with original query parameters
	// preserve original params so frontend can render without extra API calls
	v := url.Values{}
	v.Set("auth_request_id", authReq.ID) // Create によって自動セットされた ID をそのまま利用
	c.Redirect(302, strings.TrimRight(contextConfig.FrontendURL, "/")+"/authorization?"+v.Encode())
}

// AuthPost godoc
// @Summary consent authorization request
// @Schemes
// @Description ユーザーが認可を許可した際に呼び出されるエンドポイントです。Consent レコードを作成し、認可コードを生成してクライアントアプリケーションのリダイレクトURIにリダイレクトします。
// @Tags authorization
// @Accept x-www-form-urlencoded
// @Param auth_request_id formData string false "Authorization Request ID"
// @Param session_jwt formData string false "Session ID"
// @Success 301 {string} string "Redirect to client application with authorization code"
// @Router /authorization [post]
func AuthorizationPost(c *gin.Context) {
	authReqID := c.PostForm("auth_request_id")
	sessionJWT := c.PostForm("session_jwt")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	// locate authorization request
	var authReq *model.AuthorizationRequest
	var err error
	authReq, err = q.AuthorizationRequest.Where(q.AuthorizationRequest.ID.Eq(authReqID)).First()
	if err != nil || authReq == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization request"})
		return
	}

	// sessionJWTからsidを取得し、session, useridを検証
	sessionID, userID, err := util.ValidateSessionJWT(sessionJWT, c)
	cfg := c.MustGet("config").(*config.Config)
	if sessionID == "" || err != nil || userID == "" {
		c.Redirect(302, cfg.FrontendURL+"/signin?error=unauthorized")
		return
	}

	// scope に紐づく権限要件を計算し、ユーザーがそれらの権限を持つか検証
	requiredPerm := constants.Permission(0)
	for _, sc := range splitAndTrim(authReq.Scope) {
		if p, ok := constants.ScopeRequirementsPermissionMap[sc]; ok {
			requiredPerm |= p
		}
	}
	if requiredPerm != 0 {
		perms, gerr := getUserPermissions(userID, db)
		if gerr != nil {
			c.Redirect(302, cfg.FrontendURL+"/authorization?error=internal_server_error")
			return
		}
		if !perms.HasPermission(requiredPerm) {
			c.Redirect(302, cfg.FrontendURL+"/authorization?error=forbidden_scope")
			return
		}
	}

	var newConsentId string
	var authCode string
	now := time.Now().UTC()

	// コンセントの作成・更新と認可コードの生成・保存をすべて単一トランザクション内で行う
	if err := db.Transaction(func(txDB *gorm.DB) error {
		tx := query.Use(txDB)
		var consent model.Consent

		// 行ロックで取得を試みる
		err := txDB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? AND application_id = ?", userID, authReq.ApplicationID).First(&consent).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			// レコードが存在しない -> 作成を試みる
			newConsent := &model.Consent{
				UserID:        userID,
				ApplicationID: authReq.ApplicationID,
				Scope:         authReq.Scope,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := tx.Consent.Create(newConsent); err != nil {
				// 競合で一意制約に引っかかった可能性があるため、再度ロック付きで取得して更新する
				if err2 := txDB.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? AND application_id = ?", userID, authReq.ApplicationID).First(&consent).Error; err2 != nil {
					return err2
				}
				// マージ
				merged := map[string]bool{}
				for _, s := range splitAndTrim(consent.Scope) {
					merged[s] = true
				}
				for _, s := range splitAndTrim(authReq.Scope) {
					merged[s] = true
				}
				var mergedScopes []string
				for s := range merged {
					mergedScopes = append(mergedScopes, s)
				}

				consentUpdate := &model.Consent{
					Scope:     strings.Join(mergedScopes, " "),
					UpdatedAt: now,
				}
				if _, err3 := tx.Consent.Where(tx.Consent.ID.Eq(consent.ID)).Select(tx.Consent.Scope, tx.Consent.UpdatedAt).Updates(consentUpdate); err3 != nil {
					return err3
				}
				newConsentId = consent.ID
			} else {
				newConsentId = newConsent.ID
			}
		} else {
			// 既存レコードが見つかった -> スコープをマージして保存
			merged := map[string]bool{}
			for _, s := range splitAndTrim(consent.Scope) {
				merged[s] = true
			}
			for _, s := range splitAndTrim(authReq.Scope) {
				merged[s] = true
			}
			var mergedScopes []string
			for s := range merged {
				mergedScopes = append(mergedScopes, s)
			}

			consentUpdate := &model.Consent{
				Scope:     strings.Join(mergedScopes, " "),
				UpdatedAt: now,
			}
			if _, err3 := tx.Consent.Where(tx.Consent.ID.Eq(consent.ID)).Select(tx.Consent.Scope, tx.Consent.UpdatedAt).Updates(consentUpdate); err3 != nil {
				return err3
			}
			newConsentId = consent.ID
		}

		// 認可コードの生成
		e := ulid.Monotonic(rand.New(rand.NewSource(now.UnixNano())), 0)
		authCode = ulid.MustNew(ulid.Timestamp(now), e).String()

		// 認可リクエストの更新 (セッションIDおよび認可状態の保存)
		authReqUpdate := &model.AuthorizationRequest{
			Code:        &authCode,
			IsConsented: true,
		}

		// Select() を用いて明示的に対象カラムのみを Update する
		if sessionID != "" {
			authReqUpdate.SessionID = &sessionID
			if _, err := tx.AuthorizationRequest.Where(tx.AuthorizationRequest.ID.Eq(authReq.ID)).Select(
				tx.AuthorizationRequest.Code,
				tx.AuthorizationRequest.IsConsented,
				tx.AuthorizationRequest.SessionID,
			).Updates(authReqUpdate); err != nil {
				return err
			}
		} else {
			if _, err := tx.AuthorizationRequest.Where(tx.AuthorizationRequest.ID.Eq(authReq.ID)).Select(
				tx.AuthorizationRequest.Code,
				tx.AuthorizationRequest.IsConsented,
			).Updates(authReqUpdate); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create or update consent"})
		return
	}

	writeAuditLog(c, "CONSENT", "authorization_requests/"+authReq.ID, &userID, &authReq.ApplicationID, &sessionID, map[string]interface{}{
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"status":     http.StatusMovedPermanently,
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"consent_id": newConsentId,
	})

	// redirect to /consented which will forward to the client's redirect_uri
	config := c.MustGet("config").(*config.Config)
	redirectTo := strings.TrimRight(config.IssuerURL, "/") + "/consented?authorization_id=" + authReq.ID
	c.Redirect(http.StatusMovedPermanently, redirectTo)
}

// splitAndTrim splits a space-separated string into a slice of strings and trims spaces.
func splitAndTrim(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			if start < i {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

// getUserPermissions aggregates a user's permissions from their roles
func getUserPermissions(userID string, db *gorm.DB) (constants.Permission, error) {
	q := query.Use(db)

	userRoles, err := q.UserRole.Where(q.UserRole.UserID.Eq(userID)).Find()
	if err != nil {
		return 0, err
	}
	if len(userRoles) == 0 {
		return 0, nil
	}

	roleIDs := make([]string, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}

	roles, err := q.Role.Where(q.Role.ID.In(roleIDs...)).Find()
	if err != nil {
		return 0, err
	}

	var combined constants.Permission = 0
	for _, r := range roles {
		combined |= constants.Permission(r.PermissionBitmask)
	}
	return combined, nil
}

// InternalAuthorizationGet godoc
// @Summary get authorization request details (internal use only)
// @Schemes
// @Description 内部使用のみのエンドポイントで、(device flow以外の)認可リクエストの詳細情報を取得します。
// @Tags internal
// @Param id path string true "Authorization Request ID"
// @Success 200 {object} AuthorizationResponse
// @Router /internal/auth-requests/:id [get]
func InternalAuthorizationGet(c *gin.Context) {
	authReqID := c.Param("id")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	authReq, err := q.AuthorizationRequest.Where(q.AuthorizationRequest.ID.Eq(authReqID)).First()
	if err != nil || authReq == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization request"})
		return
	}

	if authReq.DeviceCode != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device flow auth req"})
		return
	}

	c.JSON(http.StatusOK, AuthorizationResponse{
		ClientID:            authReq.ApplicationID,
		RedirectURI:         authReq.RedirectURI,
		ResponseType:        authReq.ResponseType,
		Scope:               authReq.Scope,
		State:               authReq.State,
		Nonce:               authReq.Nonce,
		Prompt:              authReq.Prompt,
		CodeChallenge:       authReq.CodeChallenge,
		CodeChallengeMethod: authReq.CodeChallengeMethod,
	})
}

// InternalConsentedPost godoc
// @Summary mark authorization request as consented (internal use only)
// @Schemes
// @Description 内部使用のみのエンドポイントで、認可リクエストをユーザーが同意した状態に更新します。
// @Tags internal
// @Accept json
// @Param id path string true "Authorization Request ID"
// @Success 200 {string} string "Authorization request marked as consented"
// @Router /internal/auth-requests/:id/consented [post]
func InternalConsentedPost(c *gin.Context) {
	authReqID := c.Param("id")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	authReq, err := q.AuthorizationRequest.Where(q.AuthorizationRequest.ID.Eq(authReqID)).First()
	if err != nil || authReq == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization request"})
		return
	}

	// セッション取得
	authorizationHeader := c.GetHeader("Authorization")
	var sessionJWT string
	if after, ok0 := strings.CutPrefix(authorizationHeader, "Bearer "); ok0 {
		sessionJWT = after
	}
	sessionID, userID, err := util.ValidateSessionJWT(sessionJWT, c)
	if err != nil || sessionID == "" || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	// scope に紐づく権限要件を計算し、ユーザーがそれらの権限を持つか検証
	requiredPerm := constants.Permission(0)
	for _, sc := range splitAndTrim(authReq.Scope) {
		if p, ok := constants.ScopeRequirementsPermissionMap[sc]; ok {
			requiredPerm |= p
		}
	}
	if requiredPerm != 0 {
		perms, gerr := getUserPermissions(userID, db)
		if gerr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check permissions"})
			return
		}
		if !perms.HasPermission(requiredPerm) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden_scope"})
			return
		}
	}

	// DBの更新処理をすべて1つのトランザクションにまとめる
	err = q.Transaction(func(tx *query.Query) error {
		authReqUpdate := &model.AuthorizationRequest{
			SessionID:   &sessionID,
			IsConsented: true,
		}

		if authReq.ResponseType != nil && *authReq.ResponseType == "code" {
			now := time.Now().UTC()
			e := ulid.Monotonic(rand.New(rand.NewSource(now.UnixNano())), 0)
			code := ulid.MustNew(ulid.Timestamp(now), e).String()
			authReqUpdate.Code = &code

			if _, err := tx.AuthorizationRequest.Where(tx.AuthorizationRequest.ID.Eq(authReq.ID)).Select(
				tx.AuthorizationRequest.SessionID,
				tx.AuthorizationRequest.IsConsented,
				tx.AuthorizationRequest.Code,
			).Updates(authReqUpdate); err != nil {
				return err
			}
		} else {
			if _, err := tx.AuthorizationRequest.Where(tx.AuthorizationRequest.ID.Eq(authReq.ID)).Select(
				tx.AuthorizationRequest.SessionID,
				tx.AuthorizationRequest.IsConsented,
			).Updates(authReqUpdate); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update auth request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "authorization request marked as consented"})
}

// InternalDeviceAuthorizationGet godoc
// @Summary get authorization request details (internal use only)
// @Schemes
// @Description 内部使用のみのエンドポイントで、device flowの認可リクエストの詳細情報を取得します。
// @Tags internal
// @Param user_code path string true "User Code"
// @Success 200 {object} AuthorizationResponse
// @Router /internal/device-auth-requests/:id [get]
func InternalDeviceAuthorizationGet(c *gin.Context) {
	userCode := c.Param("id")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	authReq, err := q.AuthorizationRequest.Where(q.AuthorizationRequest.UserCode.Eq(userCode)).First()
	if err != nil || authReq == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization request"})
		return
	}

	c.JSON(http.StatusOK, AuthorizationResponse{
		ClientID: authReq.ApplicationID,
		Scope:    authReq.Scope,
	})
}
