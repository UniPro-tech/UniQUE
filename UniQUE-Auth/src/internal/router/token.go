package router

// Modified: support public clients without client_secret (PKCE enforced)

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/UniPro-tech/UniQUE-Auth/internal/middleware"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwe"
	"gorm.io/gorm"
)

type TokenGetRequest struct {
	GrantType    string `form:"grant_type" binding:"required,oneof=authorization_code refresh_token client_credentials urn:ietf:params:oauth:grant-type:device_code"`
	Code         string `form:"code"`
	RedirectURI  string `form:"redirect_uri"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	RefreshToken string `form:"refresh_token"`
	CodeVerifier string `form:"code_verifier"`
	DeviceCode   string `form:"device_code"`
}

type TokenGetResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// TokenPost godoc
// @Summary      Token Endpoint
// @Description  OAuth2 Token Endpoint
// @Tags         oauth2
// @Accept       application/x-www-form-urlencoded
// @Produce      json
// @Param        grant_type     formData  string  true   "Grant Type"  Enums(authorization_code, refresh_token, client_credentials, urn:ietf:params:oauth:grant-type:device_code)
// @Param        code           formData  string  false  "Authorization Code"
// @Param        redirect_uri   formData  string  false  "Redirect URI"
// @Param        client_id      formData  string  false  "Client ID"
// @Param        client_secret  formData  string  false  "Client Secret"
// @Param        refresh_token  formData  string  false  "Refresh Token"
// @Param        code_verifier  formData  string  false  "Code Verifier"
// @Param        device_code    formData  string  false  "Device Code"
// @Success      200  {object}  TokenGetResponse
// @Failure      400  {object}  map[string]string
// @Router       /token [post]
func TokenPost(c *gin.Context) {
	req := TokenGetRequest{}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch req.GrantType {
	case "authorization_code":
		clientID := checkClientAuthentication(c, &req, false)
		if clientID == nil {
			return
		}
		handleAuthorizationCodeGrant(c, &req, *clientID)
	case "refresh_token":
		clientID := checkClientAuthentication(c, &req, false)
		if clientID == nil {
			return
		}
		handleRefreshTokenGrant(c, &req, *clientID)
	case "urn:ietf:params:oauth:grant-type:device_code":
		clientID := checkClientAuthentication(c, &req, true)
		if clientID == nil {
			return
		}
		handleDeviceCodeGrant(c, &req, *clientID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
	}
}

// authorization_code グラントの処理
func handleAuthorizationCodeGrant(c *gin.Context, req *TokenGetRequest, clientID string) {
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	var accessToken, idToken, refreshToken string

	err := q.Transaction(func(tx *query.Query) error {
		authReq, err := tx.AuthorizationRequest.Where(
			tx.AuthorizationRequest.Code.Eq(req.Code),
			tx.AuthorizationRequest.ApplicationID.Eq(clientID),
		).First()
		if err != nil {
			return err
		}
		if authReq == nil {
			return gorm.ErrRecordNotFound
		}

		if authReq.RedirectURI == nil || *authReq.RedirectURI != req.RedirectURI {
			return errors.New("redirect_uri_mismatch")
		}

		challenge := derefPtr(authReq.CodeChallenge)
		if challenge != "" {
			verifier := req.CodeVerifier
			if verifier == "" {
				return errors.New("code_verifier_required")
			}
			method := derefPtr(authReq.CodeChallengeMethod)
			if method == "" {
				method = "plain"
			}

			var expected string
			switch method {
			case "S256":
				sum := sha256.Sum256([]byte(verifier))
				expected = base64.RawURLEncoding.EncodeToString(sum[:])
			default:
				expected = verifier
			}

			if subtle.ConstantTimeCompare([]byte(expected), []byte(challenge)) != 1 {
				return errors.New("pkce_verification_failed")
			}
		}

		if authReq.SessionID == nil || *authReq.SessionID == "" {
			return errors.New("invalid_session")
		}

		session, err := tx.Session.Where(tx.Session.ID.Eq(*authReq.SessionID)).First()
		if err != nil {
			return err
		}

		consent, err := tx.Consent.Where(
			tx.Consent.UserID.Eq(session.UserID),
			tx.Consent.ApplicationID.Eq(authReq.ApplicationID),
		).First()
		if err != nil {
			return err
		}

		cfg := *c.MustGet("config").(*config.Config)
		// tx.OauthToken.UnderlyingDB() から、トランザクションコンテキストを持った *gorm.DB を安全に取得
		accessToken, idToken, refreshToken, err = util.GenerateTokens(tx, cfg, consent, authReq.Scope, derefPtr(authReq.Nonce), middleware.GetLogger(c))
		if err != nil {
			return err
		}

		if _, err := tx.AuthorizationRequest.Where(tx.AuthorizationRequest.ID.Eq(authReq.ID)).Delete(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "authorization code not found"})
			return
		}
		switch err.Error() {
		case "redirect_uri_mismatch":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "redirect_uri_mismatch"})
		case "code_verifier_required":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "code_verifier required"})
		case "pkce_verification_failed":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant", "error_description": "pkce_verification_failed"})
		case "invalid_session":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "invalid session"})
		default:
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, TokenGetResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken,
		IDToken:      idToken,
	})
}

// refresh_token グラントの処理
func handleRefreshTokenGrant(c *gin.Context, req *TokenGetRequest, clientID string) {
	cfg := *c.MustGet("config").(*config.Config)
	tokenRaw := req.RefreshToken
	specifiedKid := ""
	if idx := strings.Index(tokenRaw, ":"); idx > 0 {
		maybe := tokenRaw[:idx]
		if len(maybe) == 64 {
			specifiedKid = maybe
			tokenRaw = tokenRaw[idx+1:]
		}
	}

	jweObj, err := jwe.ParseEncrypted(tokenRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	var decryptedObj []byte
	var decErr error
	if specifiedKid != "" {
		found := false
		for _, kp := range cfg.KeyPairs {
			kpKid := util.KidForPublicKey(kp.PublicKey)
			if subtle.ConstantTimeCompare([]byte(kpKid), []byte(specifiedKid)) == 1 {
				decryptedObj, decErr = jweObj.Decrypt(&kp.PrivateKey)
				found = true
				break
			}
		}
		if !found || decErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
			return
		}
	} else {
		for _, kp := range cfg.KeyPairs {
			decryptedObj, decErr = jweObj.Decrypt(&kp.PrivateKey)
			if decErr == nil {
				break
			}
		}
		if decErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
			return
		}
	}

	var claims util.RefreshTokenClaims
	if err := json.Unmarshal(decryptedObj, &claims); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != clientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
		return
	}

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	var accessToken, idToken, refreshToken string

	err = q.Transaction(func(tx *query.Query) error {
		logger := middleware.GetLogger(c)
		tokenset, err := tx.OauthToken.Where(tx.OauthToken.RefreshTokenJti.Eq(claims.ID)).First()
		if err != nil {
			return err
		}
		if tokenset == nil {
			return gorm.ErrRecordNotFound
		}

		consent, err := tx.Consent.Where(tx.Consent.ID.Eq(tokenset.ConsentID)).First()
		if err != nil {
			return err
		}
		if consent == nil {
			return gorm.ErrRecordNotFound
		}

		// tx.OauthToken.UnderlyingDB() から、トランザクションコンテキストを持った *gorm.DB を安全に取得
		accessToken, idToken, refreshToken, err = util.GenerateTokens(tx, cfg, consent, claims.Scope, "", middleware.GetLogger(c))
		if err != nil {
			logger.Error("An error occured in GenTokens")
			return err
		}

		if _, err := tx.OauthToken.Where(tx.OauthToken.RefreshTokenJti.Eq(claims.ID)).Update(tx.OauthToken.DeletedAt, time.Now().UTC()); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant"})
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, TokenGetResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken,
		IDToken:      idToken,
	})
}

// urn:ietf:params:oauth:grant-type:device_code グラントの処理
func handleDeviceCodeGrant(c *gin.Context, req *TokenGetRequest, clientID string) {
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	var accessToken, idToken, refreshToken string

	err := q.Transaction(func(tx *query.Query) error {
		deviceAuthReq, err := tx.AuthorizationRequest.Where(
			tx.AuthorizationRequest.DeviceCode.Eq(req.DeviceCode),
			tx.AuthorizationRequest.ApplicationID.Eq(clientID),
		).First()
		if err != nil {
			return err
		}
		if deviceAuthReq == nil {
			return gorm.ErrRecordNotFound
		}

		if deviceAuthReq.ExpiresAt.Before(time.Now().UTC()) {
			return errors.New("expired_token")
		}

		if !deviceAuthReq.IsConsented {
			return errors.New("authorization_pending")
		}

		if deviceAuthReq.DeviceFlowDenied {
			return errors.New("access_denied")
		}

		// gen tokens
		if deviceAuthReq.SessionID == nil || *deviceAuthReq.SessionID == "" {
			return errors.New("access_denied")
		}

		session, err := tx.Session.Where(tx.Session.ID.Eq(*deviceAuthReq.SessionID)).First()
		if err != nil {
			return err
		}

		consent, err := tx.Consent.Where(
			tx.Consent.UserID.Eq(session.UserID),
			tx.Consent.ApplicationID.Eq(deviceAuthReq.ApplicationID),
		).First()
		if err != nil {
			return err
		}

		cfg := *c.MustGet("config").(*config.Config)
		// tx.OauthToken.UnderlyingDB() から、トランザクションコンテキストを持った *gorm.DB を安全に取得
		accessToken, idToken, refreshToken, err = util.GenerateTokens(tx, cfg, consent, deviceAuthReq.Scope, derefPtr(deviceAuthReq.Nonce), middleware.GetLogger(c))
		if err != nil {
			return err
		}

		if _, err := tx.AuthorizationRequest.Where(tx.AuthorizationRequest.ID.Eq(deviceAuthReq.ID)).Delete(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "device code not found"})
			return
		}
		switch err.Error() {
		case "expired_token":
			c.JSON(http.StatusBadRequest, gin.H{"error": "expired_token"})
		case "authorization_pending":
			c.JSON(http.StatusBadRequest, gin.H{"error": "authorization_pending"})
		case "access_denied":
			c.JSON(http.StatusBadRequest, gin.H{"error": "access_denied"})
		default:
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, TokenGetResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken,
		IDToken:      idToken,
	})
}

// [共通] クライアント認証の検査
func checkClientAuthentication(c *gin.Context, req *TokenGetRequest, isDeviceFlow bool) *string {
	logger := middleware.GetLogger(c)

	if !isDeviceFlow {
		if clientVerifyBasic := c.GetHeader("Authorization"); clientVerifyBasic != "" {
			// Basic認証ヘッダがある場合は、クライアントIDとシークレットを検証する
			clientID, clientSecret, err := parseBasicAuth(clientVerifyBasic)
			if err != nil {
				logger.Warn("not valid authorization header")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
				return nil
			}

			application, err := query.Application.Where(query.Application.ID.Eq(clientID)).First()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
					return nil
				}
				c.AbortWithError(http.StatusInternalServerError, err)
				return nil
			}

			if application.PublicClient {
				return &clientID
			}

			if subtle.ConstantTimeCompare([]byte(application.ClientSecret), []byte(clientSecret)) != 1 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
				return nil
			}
			return &clientID
		} else {
			application, err := query.Application.Where(query.Application.ID.Eq(req.ClientID)).First()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
					return nil
				}
				c.AbortWithError(http.StatusInternalServerError, err)
				return nil
			}

			if application.PublicClient {
				return &req.ClientID
			}

			if subtle.ConstantTimeCompare([]byte(application.ClientSecret), []byte(req.ClientSecret)) != 1 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
				return nil
			}
			return &req.ClientID
		}
	} else {
		// Device Flowの場合は、クライアントIDのみを検証する
		if req.ClientID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
			return nil
		}
		authRequest, err := query.AuthorizationRequest.Where(query.AuthorizationRequest.ApplicationID.Eq(req.ClientID), query.AuthorizationRequest.DeviceCode.Eq(req.DeviceCode)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
				return nil
			}
			c.AbortWithError(http.StatusInternalServerError, err)
			return nil
		}
		if authRequest == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
			return nil
		}
		return &req.ClientID
	}
}

// [共通] Basic認証ヘッダの解析
func parseBasicAuth(authHeader string) (string, string, error) {
	if !strings.HasPrefix(authHeader, "Basic ") {
		return "", "", errors.New("invalid basic auth format")
	}
	base64Decoded, err := base64.StdEncoding.DecodeString(authHeader[len("Basic "):])
	if err != nil {
		return "", "", err
	}

	parts := strings.SplitN(string(base64Decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid basic auth format")
	}
	return parts[0], parts[1], nil
}

// [共通] ポインタの値を取得
func derefPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
