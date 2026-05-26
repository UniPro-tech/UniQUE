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
	GrantType    string `form:"grant_type" binding:"required,oneof=authorization_code refresh_token client_credentials"`
	Code         string `form:"code"`
	RedirectURI  string `form:"redirect_uri"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	RefreshToken string `form:"refresh_token"`
	CodeVerifier string `form:"code_verifier"`
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
// @Param        grant_type     formData  string  true   "Grant Type"  Enums(authorization_code, refresh_token, client_credentials)
// @Param        code           formData  string  false  "Authorization Code"
// @Param        redirect_uri   formData  string  false  "Redirect URI"
// @Param        client_id      formData  string  false  "Client ID"
// @Param        client_secret  formData  string  false  "Client Secret"
// @Param        refresh_token  formData  string  false  "Refresh Token"
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
		clientID := checkClientAuthentication(c, &req)
		if clientID == nil {
			return
		}
		handleAuthorizationCodeGrant(c, &req, *clientID)
	case "refresh_token":
		clientID := checkClientAuthentication(c, &req)
		if clientID == nil {
			return
		}
		handleRefreshTokenGrant(c, &req, *clientID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported grant_type"})
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

		if authReq.RedirectURI != req.RedirectURI {
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
		accessToken, idToken, refreshToken, err = util.GenerateTokens(tx.OauthToken.UnderlyingDB(), cfg, consent, authReq.Scope, derefPtr(authReq.Nonce))
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization code"})
			return
		}
		switch err.Error() {
		case "redirect_uri_mismatch":
			c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri does not match"})
		case "code_verifier_required":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "code_verifier required"})
		case "pkce_verification_failed":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_grant", "error_description": "pkce_verification_failed"})
		case "invalid_session":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refresh token"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refresh token"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refresh token"})
			return
		}
	}

	var claims util.RefreshTokenClaims
	if err := json.Unmarshal(decryptedObj, &claims); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refresh token"})
		return
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != clientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refresh token"})
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
		accessToken, idToken, refreshToken, err = util.GenerateTokens(tx.OauthToken.UnderlyingDB(), cfg, consent, claims.Scope, "")
		if err != nil {
			return err
		}

		if _, err := tx.OauthToken.Where(tx.OauthToken.RefreshTokenJti.Eq(claims.ID)).Update(tx.OauthToken.DeletedAt, time.Now().UTC()); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refresh token"})
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

// クライアント認証の検査
func checkClientAuthentication(c *gin.Context, req *TokenGetRequest) *string {
	logger := middleware.GetLogger(c)
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return nil
	}
	q := query.Use(db)

	if clientVerifyBasic := c.GetHeader("Authorization"); clientVerifyBasic != "" {
		clientID, clientSecret, err := parseBasicAuth(clientVerifyBasic)
		if err != nil {
			logger.Warn("not valid authorization header")
			c.JSON(http.StatusBadRequest, gin.H{"error": "not valid authorization header"})
			return nil
		}

		application, err := q.Application.Where(q.Application.ID.Eq(clientID)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client credentials"})
				return nil
			}
			c.AbortWithError(http.StatusInternalServerError, err)
			return nil
		}

		if application.PublicClient {
			return &clientID
		}

		if subtle.ConstantTimeCompare([]byte(application.ClientSecret), []byte(clientSecret)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client credentials"})
			return nil
		}
		return &clientID
	} else {
		application, err := q.Application.Where(q.Application.ID.Eq(req.ClientID)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client credentials"})
				return nil
			}
			c.AbortWithError(http.StatusInternalServerError, err)
			return nil
		}

		if application.PublicClient {
			return &req.ClientID
		}

		if subtle.ConstantTimeCompare([]byte(application.ClientSecret), []byte(req.ClientSecret)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid client credentials"})
			return nil
		}
		return &req.ClientID
	}
}

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

func derefPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
