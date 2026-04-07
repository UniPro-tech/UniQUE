package router

// Modified: support public clients without client_secret (PKCE enforced)

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
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
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	switch req.GrantType {
	case "authorization_code":
		clientID := checkClientAuthentication(c, &req)
		if clientID == nil {
			c.JSON(401, gin.H{"error": "invalid client credentials"})
			return
		}
		handleAuthorizationCodeGrant(c, &req, *clientID)
	case "refresh_token":
		clientID := checkClientAuthentication(c, &req)
		if clientID == nil {
			c.JSON(401, gin.H{"error": "invalid client credentials"})
			return
		}
		handleRefreshTokenGrant(c, &req, *clientID)
	default:
		c.JSON(400, gin.H{"error": "unsupported grant_type"})
	}
}

// authorization_code グラントの処理
// PKCE検証もここで行う
// 成功時はアクセストークン、IDトークン、リフレッシュトークンを発行して返す
func handleAuthorizationCodeGrant(c *gin.Context, req *TokenGetRequest, clientID string) {
	// get DB + query instance
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(500, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	// Authorization RequestをCodeから取得する
	// ここでClientIDもWhereの中に入れて検証してしまう
	authReq, err := q.AuthorizationRequest.Where(q.AuthorizationRequest.Code.Eq(req.Code), q.AuthorizationRequest.ApplicationID.Eq(clientID)).First()
	if err != nil || authReq == nil {
		c.JSON(400, gin.H{"error": "invalid authorization code"})
		return
	}

	if authReq.RedirectURI != req.RedirectURI {
		c.JSON(400, gin.H{"error": "redirect_uri does not match"})
		return
	}

	// PKCE検証: AuthorizationRequest に code_challenge がある場合は code_verifier を検証する
	challenge := derefPtr(authReq.CodeChallenge)
	if challenge != "" {
		verifier := req.CodeVerifier
		if verifier == "" {
			c.JSON(400, gin.H{"error": "invalid_request", "error_description": "code_verifier required"})
			return
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
			c.JSON(400, gin.H{"error": "invalid_grant", "error_description": "pkce_verification_failed"})
			return
		}
	}

	// generate tokens and respond
	if authReq.SessionID == nil || *authReq.SessionID == "" {
		c.JSON(400, gin.H{"error": "invalid session"})
		return
	}

	session, err := q.Session.Where(q.Session.ID.Eq(*authReq.SessionID)).First()
	if err != nil || session == nil {
		c.JSON(400, gin.H{"error": "invalid session"})
		return
	}
	consent, err := q.Consent.Where(q.Consent.UserID.Eq(session.UserID), q.Consent.ApplicationID.Eq(authReq.ApplicationID)).First()
	if err != nil || consent == nil {
		c.JSON(400, gin.H{"error": "invalid consent"})
		return
	}
	accessToken, idToken, refreshToken, err := util.GenerateTokens(db, *c.MustGet("config").(*config.Config), consent, authReq.Scope, derefPtr(authReq.Nonce))
	if err != nil {
		log.Printf("GenerateTokens error: %v", err)
		c.JSON(400, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Authorization Codeは使い捨てなので削除する
	q.AuthorizationRequest.Where(q.AuthorizationRequest.ID.Eq(authReq.ID)).Delete()

	c.JSON(200, TokenGetResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken,
		IDToken:      idToken,
	})
}

// refresh_token グラントの処理
// リフレッシュトークンを復号して JTI を取得し、その JTI に紐づくトークンセットと同じユーザー・アプリケーションの新しいトークンを発行して返す
// リフレッシュトークンの復号に失敗した場合はエラーを返す（クライアントがトークンを改ざんしている可能性があるため、他の処理は行わない）
func handleRefreshTokenGrant(c *gin.Context, req *TokenGetRequest, clientID string) {
	config := *c.MustGet("config").(*config.Config)
	// parse refresh token; 新しい形式では先頭に "<kid>:<compact-jwe>" を付与している
	tokenRaw := req.RefreshToken
	specifiedKid := ""
	if idx := strings.Index(tokenRaw, ":"); idx > 0 {
		maybe := tokenRaw[:idx]
		// 簡易チェック: hex 長が期待値(64)なら kid とみなす
		if len(maybe) == 64 {
			specifiedKid = maybe
			tokenRaw = tokenRaw[idx+1:]
		}
	}

	jweObj, err := jwe.ParseEncrypted(tokenRaw)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid refresh token"})
		return
	}

	var decryptedObj []byte
	var decErr error
	if specifiedKid != "" {
		// 指定kidがある場合はその鍵のみで復号を試みる（タイミング攻撃緩和）
		found := false
		for _, kp := range config.KeyPairs {
			kpKid := util.KidForPublicKey(kp.PublicKey)
			if subtle.ConstantTimeCompare([]byte(kpKid), []byte(specifiedKid)) == 1 {
				decryptedObj, decErr = jweObj.Decrypt(&kp.PrivateKey)
				found = true
				break
			}
		}
		if !found || decErr != nil {
			c.JSON(400, gin.H{"error": "invalid refresh token"})
			return
		}
	} else {
		// 従来トークン互換性のために全鍵で試行
		for _, kp := range config.KeyPairs {
			decryptedObj, decErr = jweObj.Decrypt(&kp.PrivateKey)
			if decErr == nil {
				break
			}
		}
		if decErr != nil {
			c.JSON(400, gin.H{"error": "invalid refresh token"})
			return
		}
	}

	// parse decrypted payload into refresh token claims to extract JTI
	var claims util.RefreshTokenClaims
	if err := json.Unmarshal(decryptedObj, &claims); err != nil {
		c.JSON(400, gin.H{"error": "invalid refresh token"})
		return
	}

	// get tokenset by JTI from decrypted claims
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(500, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	// Clientチェック: クレームの Audience に ClientID が含まれていることを確認する
	if len(claims.Audience) == 0 || claims.Audience[0] != clientID {
		c.JSON(400, gin.H{"error": "invalid refresh token"})
		return
	}

	tokenset, err := q.OauthToken.Where(q.OauthToken.RefreshTokenJti.Eq(claims.ID)).First()

	if err != nil || tokenset == nil {
		c.JSON(400, gin.H{"error": "invalid refresh token"})
		return
	}

	// get consent
	consent, err := q.Consent.Where(q.Consent.ID.Eq(tokenset.ConsentID)).First()
	if err != nil || consent == nil {
		c.JSON(400, gin.H{"error": "invalid consent"})
		return
	}

	accessToken, idToken, refreshToken, err := util.GenerateTokens(db, config, consent, claims.Scope, "")
	if err != nil {
		log.Printf("GenerateTokens error (refresh): %v", err)
		c.JSON(400, gin.H{"error": "failed to generate tokens"})
		return
	}

	// 旧トークンを無効化する
	q.OauthToken.Where(q.OauthToken.RefreshTokenJti.Eq(claims.ID)).Update(q.OauthToken.DeletedAt, time.Now())

	c.JSON(200, TokenGetResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken,
		IDToken:      idToken,
	})
}

// クライアント認証の検査
// Authorization ヘッダーの Basic 認証を優先して検査し、なければフォームの client_id/client_secret を検査する
// 成功すれば client_id を返し、失敗すればエラー応答を返して false を返す
func checkClientAuthentication(c *gin.Context, req *TokenGetRequest) *string {
	// check client_secret
	if clientVerifyBasic := c.GetHeader("Authorization"); clientVerifyBasic != "" {
		// if authorization header
		clientID, clientSecret, err := parseBasicAuth(clientVerifyBasic)
		if err != nil {
			c.JSON(400, gin.H{"error": "not valid authorization header"})
			return nil
		}
		dbAny := c.MustGet("db")
		db, ok := dbAny.(*gorm.DB)
		if !ok || db == nil {
			c.JSON(500, gin.H{"error": "database not available"})
			return nil
		}
		q := query.Use(db)
		// DB側で平文比較するのではなく、IDで取得してから定数時間比較を行う
		application, err := q.Application.Where(q.Application.ID.Eq(clientID)).First()
		if err != nil || application == nil {
			c.JSON(400, gin.H{"error": "invalid client credentials"})
			return nil
		}
		// If application is public client, skip secret validation (PKCE enforced elsewhere)
		if application.PublicClient {
			return &clientID
		}
		// 定数時間比較
		if subtle.ConstantTimeCompare([]byte(application.ClientSecret), []byte(clientSecret)) != 1 {
			c.JSON(400, gin.H{"error": "invalid client credentials"})
			return nil
		}
		return &clientID
	} else {
		// if form body
		dbAny := c.MustGet("db")
		db, ok := dbAny.(*gorm.DB)
		if !ok || db == nil {
			c.JSON(500, gin.H{"error": "database not available"})
			return nil
		}
		q := query.Use(db)
		application, err := q.Application.Where(q.Application.ID.Eq(req.ClientID)).First()
		if err != nil || application == nil {
			c.JSON(400, gin.H{"error": "invalid client credentials"})
			return nil
		}
		// If application is public client, allow token endpoint without client_secret (PKCE must be used)
		if application.PublicClient {
			return &req.ClientID
		}
		if subtle.ConstantTimeCompare([]byte(application.ClientSecret), []byte(req.ClientSecret)) != 1 {
			c.JSON(400, gin.H{"error": "invalid client credentials"})
			return nil
		}
	}
	return &req.ClientID
}

// Basic 認証ヘッダーをパースして client_id と client_secret を取得する
func parseBasicAuth(authHeader string) (string, string, error) {
	base64Decoded, err := base64.StdEncoding.DecodeString(authHeader[len("Basic "):])
	if err != nil {
		return "", "", err
	}
	parts := string(base64Decoded)
	for i := 0; i < len(parts); i++ {
		if parts[i] == ':' {
			return parts[:i], parts[i+1:], nil
		}
	}
	return "", "", errors.New("invalid basic auth format")
}
