package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/constants"
	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// JWKS キャッシュ
var (
	jwksMu       sync.RWMutex
	jwksKeyCache = map[string]*rsa.PublicKey{}
	jwksFetched  time.Time
	httpClient   = &http.Client{Timeout: 5 * time.Second}
)

const jwksCacheTTL = 24 * time.Hour

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// swagger用のパスはスキップ
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") || c.Request.URL.Path == "/swagger.json" {
			c.Next()
			return
		}

		// health checkはスキップ
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// /internal/users への POST（ユーザー作成）はスキップ
		if c.Request.Method == "POST" && c.Request.URL.Path == "/internal/users" {
			c.Next()
			return
		}
		// /internal/users/email_verify への POST(メール認証)はスキップ
		if c.Request.Method == "POST" && c.Request.URL.Path == "/internal/users/email_verify" {
			c.Next()
			return
		}
		// /internal/users/email_verify/:code への GET(コード参照)はスキップ
		if c.Request.Method == "GET" && strings.HasPrefix(c.Request.URL.Path, "/internal/users/email_verify/") {
			c.Next()
			return
		}
		// /internal/users/email_verify/discord_link への POST(Discord連携)はスキップ
		if c.Request.Method == "POST" && c.Request.URL.Path == "/internal/users/email_verify/discord_link" {
			c.Next()
			return
		}
		// GET /users（ユーザー一覧取得）は認証オプション、認可不要（トークンがあれば検証して権限に応じた情報を返す）
		if c.Request.Method == "GET" && c.Request.URL.Path == "/users" {
			authorization := c.GetHeader("Authorization")
			token := extractToken(authorization)
			if token != "" {
				cfg := c.MustGet("config").(config.Config)
				db, _ := c.MustGet("db").(*gorm.DB)
				claims, user, isOAuth, scope := validateToken(token, cfg, db, c)
				if claims != nil && user != nil {
					c.Set("claims", claims)
					c.Set("user", user)
				}
				if claims != nil {
					c.Set("claims", claims)
				}
				if isOAuth {
					c.Set("isOAuth", true)
					c.Set("scope", scope)
				}
			}
			c.Next()
			return
		}
		// GET /users/:id（単一ユーザー取得）は認証オプション、認可不要（トークンがあれば検証して権限に応じた情報を返す）
		// ただしサブパス（例: /users/:id/roles）には認証が必須
		if c.Request.Method == "GET" && strings.HasPrefix(c.Request.URL.Path, "/users/") {
			// /users/:id のみをマッチ（サブパスは含まない）
			path := strings.TrimPrefix(c.Request.URL.Path, "/users/")
			if !strings.Contains(path, "/") {
				// サブパスがないので認証オプション（トークンがあれば検証）
				authorization := c.GetHeader("Authorization")
				token := extractToken(authorization)
				if token != "" {
					cfg := c.MustGet("config").(config.Config)
					db, _ := c.MustGet("db").(*gorm.DB)
					claims, user, isOAuth, scope := validateToken(token, cfg, db, c)
					if claims != nil && user != nil {
						c.Set("claims", claims)
						c.Set("user", user)
					}
					if claims != nil {
						c.Set("claims", claims)
					}
					if isOAuth {
						c.Set("isOAuth", true)
						c.Set("scope", scope)
					}
				}
				c.Next()
				return
			}
			// サブパスがある場合は認証が必須なので、Authorizationヘッダーをチェック
		}
		// AuthorizationヘッダーからJWTを取得
		authorization := c.GetHeader("Authorization")
		token := extractToken(authorization)
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		// トークン検証
		cfg := c.MustGet("config").(config.Config)
		db, _ := c.MustGet("db").(*gorm.DB)
		claims, user, isOAuth, scope := validateToken(token, cfg, db, c)
		if claims == nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}
		c.Set("claims", claims)
		if user != nil {
			c.Set("user", user)
		}
		if isOAuth {
			c.Set("isOAuth", true)
			c.Set("scope", scope)
		}
		c.Next()
	}
}

func extractToken(authorization string) string {
	// "Bearer <token>"形式からトークン部分を抽出
	const bearerPrefix = "Bearer "
	if len(authorization) > len(bearerPrefix) && authorization[:len(bearerPrefix)] == bearerPrefix {
		return authorization[len(bearerPrefix):]
	}
	return ""
}

// fetchJWKS fetches JWKS JSON and fills jwksKeyCache
func fetchJWKS(jwksURL string) error {
	resp, err := httpClient.Get(jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var body struct {
		Keys []map[string]interface{} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	tmp := map[string]*rsa.PublicKey{}
	for _, k := range body.Keys {
		if kty, _ := k["kty"].(string); kty != "RSA" {
			continue
		}
		kid, _ := k["kid"].(string)
		nStr, _ := k["n"].(string)
		eStr, _ := k["e"].(string)
		if kid == "" || nStr == "" || eStr == "" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
		if err != nil {
			continue
		}
		n := new(big.Int).SetBytes(nBytes)
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}
		tmp[kid] = &rsa.PublicKey{N: n, E: e}
	}
	jwksMu.Lock()
	jwksKeyCache = tmp
	jwksFetched = time.Now()
	jwksMu.Unlock()
	return nil
}

// ensureJWKSFresh makes sure the cache is fresh, fetching JWKS if empty or stale.
func ensureJWKSFresh(jwksURL string) error {
	jwksMu.RLock()
	age := time.Since(jwksFetched)
	cacheLen := len(jwksKeyCache)
	jwksMu.RUnlock()
	if cacheLen == 0 || age >= jwksCacheTTL {
		return fetchJWKS(jwksURL)
	}
	return nil
}

// getPublicKeyByKid returns the RSA public key for a given kid, attempting a refetch if not found.
func getPublicKeyByKid(kid, jwksURL string) (*rsa.PublicKey, error) {
	if kid == "" {
		return nil, fmt.Errorf("empty kid")
	}
	if err := ensureJWKSFresh(jwksURL); err != nil {
		return nil, err
	}
	jwksMu.RLock()
	pub := jwksKeyCache[kid]
	jwksMu.RUnlock()
	if pub != nil {
		return pub, nil
	}
	// try refetch and check again
	if err := fetchJWKS(jwksURL); err != nil {
		return nil, err
	}
	jwksMu.RLock()
	pub = jwksKeyCache[kid]
	jwksMu.RUnlock()
	if pub != nil {
		return pub, nil
	}
	return nil, fmt.Errorf("public key not found for kid %s", kid)
}

// Auth側 jwt.go と同じ定義
type AccessTokenClaims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope,omitempty"`
}

type SessionTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

// パース時に両方のカスタムクレームを受け取るための統合型
type combinedClaims struct {
	jwt.RegisteredClaims
	Scope  string `json:"scope,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

func validateToken(token string, cfg config.Config, db *gorm.DB, c *gin.Context) (*jwt.RegisteredClaims, *model.User, bool, string) {
	issuer := cfg.IssuerURL
	internalIssuer := strings.TrimRight(cfg.IssuerInternalURL, "/")
	jwksURL := strings.TrimRight(internalIssuer, "/") + "/.well-known/jwks.json"

	// keyfunc: JWKSキャッシュを利用して公開鍵を返す
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		// ensure cache is fresh
		if err := ensureJWKSFresh(jwksURL); err != nil {
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}

		// kidがあればそれを使って取得
		if kidRaw, ok := t.Header["kid"]; ok {
			if kid, ok := kidRaw.(string); ok && kid != "" {
				pub, err := getPublicKeyByKid(kid, jwksURL)
				if err != nil {
					return nil, err
				}
				return pub, nil
			}
		}

		// kidなし: 最初のキーを返す
		jwksMu.RLock()
		for _, pub := range jwksKeyCache {
			jwksMu.RUnlock()
			return pub, nil
		}
		jwksMu.RUnlock()
		return nil, fmt.Errorf("no public keys available in JWKS")
	}

	parsed, err := jwt.ParseWithClaims(token, &combinedClaims{}, keyFunc, jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}))
	if err != nil {
		log.Printf("JWT parse error: %v", err)
		return nil, nil, false, ""
	}
	combined, ok := parsed.Claims.(*combinedClaims)
	if !ok {
		log.Printf("JWT claims type assertion failed")
		return nil, nil, false, ""
	}
	claims := &combined.RegisteredClaims

	// claimからの検証
	if claims.Issuer != issuer {
		log.Printf("Issuer mismatch: got %q, expected %q", claims.Issuer, issuer)
		return nil, nil, false, ""
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) <= 0 {
		log.Printf("Token expired: exp=%v", claims.ExpiresAt)
		return nil, nil, false, ""
	}
	if claims.NotBefore != nil && time.Until(claims.NotBefore.Time) < 0 {
		log.Printf("Token not yet valid: nbf=%v", claims.NotBefore)
		return nil, nil, false, ""
	}
	if claims.IssuedAt != nil && time.Until(claims.IssuedAt.Time.Add(time.Hour*24*7)) < 0 {
		log.Printf("Token too old: iat=%v", claims.IssuedAt)
		return nil, nil, false, ""
	}
	var isValidToken bool = true
	var user *model.User
	var userIDFromVerify string
	if strings.HasPrefix(claims.Subject, "SID_") {
		// セッショントークン: jtiなし、subから"SID_"を除いた素のセッションIDで検証
		sessionID := strings.TrimPrefix(claims.Subject, "SID_")
		log.Printf("Session verify: sessionID=%s, path=/internal/session_verify", sessionID)
		isValidToken, userIDFromVerify = verifyJIT(sessionID, cfg, "/internal/session_verify")
		// Auth側から返されたuser_idを優先、なければトークン内のuser_idクレームを使用
		userID := userIDFromVerify
		if userID == "" {
			userID = combined.UserID
		}
		if userID != "" && db != nil {
			q := query.Use(db)
			u, err := q.User.Where(q.User.ID.Eq(userID)).First()
			if err == nil {
				user = u
			}
		}
	} else {
		// アクセストークン: jti(claims.ID)で検証
		isValidToken, _ = verifyJIT(claims.ID, cfg, "/internal/token_verify")
		if !isValidToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return nil, nil, false, ""
		}
		// アクセストークンのsubjectはuser_idの場合がある
		if claims.Subject != "" && !strings.HasPrefix(claims.Subject, "SID_") && db != nil {
			q := query.Use(db)
			u, err := q.User.Where(q.User.ID.Eq(claims.Subject)).First()
			if err == nil {
				user = u
			}
		}
	}
	if !isValidToken {
		log.Printf("Token verification failed for sub=%s", claims.Subject)
		return nil, nil, false, ""
	}

	// determine if this is an OAuth access token by presence of scope
	scope := combined.Scope
	isOAuth := scope != ""

	return claims, user, isOAuth, scope
}

type SessionVerifyResponse struct {
	Valid     bool      `json:"valid"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// verifyJIT calls the issuer's internal endpoint to verify a session/token.
// Auth側は "jti" パラメータ名で受け取る
// 戻り値: (valid, userID)
func verifyJIT(jti string, cfg config.Config, path string) (bool, string) {
	issuer := strings.TrimRight(cfg.IssuerInternalURL, "/")
	url := issuer + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("verifyJIT: failed to create request: %v", err)
		return false, ""
	}
	q := req.URL.Query()
	q.Add("jti", jti)
	req.URL.RawQuery = q.Encode()
	log.Printf("verifyJIT: calling %s?%s", url, req.URL.RawQuery)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("verifyJIT: request failed: %v", err)
		return false, ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("verifyJIT: got status %d from %s", resp.StatusCode, url)
		return false, ""
	}
	var res struct {
		Valid     bool      `json:"valid"`
		UserID    string    `json:"user_id,omitempty"`
		ExpiresAt time.Time `json:"expires_at,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		log.Printf("verifyJIT: failed to decode response: %v", err)
		return false, ""
	}
	log.Printf("verifyJIT: result valid=%v, userID=%s", res.Valid, res.UserID)
	return res.Valid, res.UserID
}

type TokenVerifyResponse struct {
	Valid bool `json:"valid"`
}

// requirePermissionOrScope returns a middleware that allows access when either:
// - the request is from an OAuth token that contains the required scope, or
// - the authenticated user has the required RBAC permission.
func RequirePermissionOrScope(required constants.Permission, requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// OAuth トークンの場合は scope を確認
		if isOAuthI, _ := c.Get("isOAuth"); isOAuthI != nil {
			if isOAuth, ok := isOAuthI.(bool); ok && isOAuth {
				scopeI, _ := c.Get("scope")
				scopeStr, _ := scopeI.(string)
				if strings.Contains(scopeStr, requiredScope) {
					c.Next()
					return
				}
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "OAuth トークンに必要なスコープがありません"})
				return
			}
		}

		// 非 OAuth: RBAC 権限を確認
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
			return
		}
		userModel, ok := user.(*model.User)
		if !ok || userModel == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報が取得できませんでした"})
			return
		}

		dbI, exists := c.Get("db")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "データベース接続エラー"})
			return
		}
		db, ok := dbI.(*gorm.DB)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "データベース接続エラー"})
			return
		}

		perms, err := GetUserPermissions(userModel.ID, db)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "権限の取得に失敗しました"})
			return
		}
		if !perms.HasPermission(required) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "この操作を実行する権限がありません"})
			return
		}
		c.Set("permissions", perms)
		c.Next()
	}
}
