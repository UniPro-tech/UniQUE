package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/UniPro-tech/UniQUE-Auth/internal/model"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwe"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

// ErrNoValidKeyPair indicates that the issuer cannot validate or mint tokens.
var ErrNoValidKeyPair = errors.New("no valid keypair configured")

// kidForKey computes the kid (SHA-256 thumbprint of PKIX DER) matching the JWKS endpoint.
func kidForKey(cfg config.Config) string {
	if len(cfg.KeyPairs) == 0 {
		return ""
	}
	// ensure the keypair appears initialized
	if cfg.KeyPairs[0].PublicKey.N == nil {
		return ""
	}
	rsaPub := &cfg.KeyPairs[0].PublicKey
	der, err := x509.MarshalPKIXPublicKey(rsaPub)
	if err != nil {
		return ""
	}
	thumb := sha256.Sum256(der)
	return hex.EncodeToString(thumb[:])
}

// KidForPublicKey は与えられた RSA 公開鍵の PKIX DER シグネチャの SHA-256 サムを
// hex エンコードした文字列を返す。JWKS の `kid` と互換性を持つ。
func KidForPublicKey(pub rsa.PublicKey) string {
	if pub.N == nil {
		return ""
	}
	der, err := x509.MarshalPKIXPublicKey(&pub)
	if err != nil {
		return ""
	}
	thumb := sha256.Sum256(der)
	return hex.EncodeToString(thumb[:])
}

func hasValidKeyPair(cfg config.Config) bool {
	if len(cfg.KeyPairs) == 0 {
		return false
	}
	kp := cfg.KeyPairs[0]
	if kp.PublicKey.N == nil {
		return false
	}
	if kp.PrivateKey.D == nil {
		return false
	}
	return true
}

func GenerateTokens(q *query.Query, config config.Config, consent *model.Consent, scopes, nonce string, logger *slog.Logger) (accessToken, IDToken, RefreshToken string, err error) {
	scopes = AlphabeticScopeString(scopes)

	t := time.Now()
	entropy := ulid.Monotonic(rand.Reader, 0)
	accessTokenID := ulid.MustNew(ulid.Timestamp(t), entropy).String()

	entropy = ulid.Monotonic(rand.Reader, 0)
	refreshTokenID := ulid.MustNew(ulid.Timestamp(t), entropy).String()

	entropy = ulid.Monotonic(rand.Reader, 0)
	IDTokenIDRaw := ulid.MustNew(ulid.Timestamp(t), entropy).String()
	IDTokenID := &IDTokenIDRaw
	IDTokenString := ""

	entropy = ulid.Monotonic(rand.Reader, 0)
	oauthTokenID := ulid.MustNew(ulid.Timestamp(t), entropy).String()

	if ContainsScope(scopes, "openid") {
		if !hasValidKeyPair(config) {
			return "", "", "", ErrNoValidKeyPair
		}
		IDTokenString, err = GenerateIDToken(q, IDTokenIDRaw, consent.UserID, consent.ApplicationID, nonce, scopes, config)
		if err != nil {
			logger.Error("Un error occured in idtoken gen", slog.String("error", err.Error()))
			return "", "", "", err
		}
	} else {
		IDTokenID = nil
	}

	err = q.OauthToken.Create(&model.OauthToken{
		ID:              oauthTokenID,
		ConsentID:       consent.ID,
		AccessTokenJti:  &accessTokenID,
		RefreshTokenJti: &refreshTokenID,
		IDTokenJti:      IDTokenID,
		ExpiresAt:       time.Now().Add(5 * 24 * time.Hour), // リフレッシュトークン有効期限: 5日
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		return "", "", "", err
	}

	// Generate Access Token (use structured claims per OAuth2/OIDC)
	accessTokenClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   consent.UserID,
			Audience:  jwt.ClaimStrings{consent.ApplicationID},
			Issuer:    config.IssuerURL,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Hour)),
			ID:        accessTokenID,
		},
		Scope: scopes,
	})
	accessTokenClaims.Header["kid"] = kidForKey(config)
	if !hasValidKeyPair(config) {
		return "", "", "", ErrNoValidKeyPair
	}
	accessTokenString, err := accessTokenClaims.SignedString(&config.KeyPairs[0].PrivateKey)
	if err != nil {
		return "", "", "", err
	}

	// Generate Refresh Token
	// For simplicity, using a JWE containing the claims as the refresh token
	refreshClaims := RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   consent.UserID,
			Audience:  jwt.ClaimStrings{consent.ApplicationID},
			Issuer:    config.IssuerURL,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * 24 * time.Hour)),
			ID:        refreshTokenID,
		},
		Scope: scopes,
	}
	// marshal claims to plaintext bytes for JWE
	plaintext, err := json.Marshal(refreshClaims)
	if err != nil {
		return "", "", "", err
	}

	// create JWE with new signature: (alg, key, method, plaintext)
	if !hasValidKeyPair(config) {
		return "", "", "", ErrNoValidKeyPair
	}
	refreshTokenClaim, err := jwe.NewJWE(jwe.KeyAlgorithmRSAOAEP, &config.KeyPairs[0].PublicKey, jwe.EncryptionTypeA256GCM, plaintext)
	if err != nil {
		return "", "", "", err
	}

	// finalize/serialize the JWE to string
	refreshTokenString, err := refreshTokenClaim.CompactSerialize()
	if err != nil {
		return "", "", "", err
	}
	// 後続での安全な復号のため、kid をプレフィックスとして付与する
	kid := kidForKey(config)
	if kid != "" {
		refreshTokenString = kid + ":" + refreshTokenString
	}
	return accessTokenString, IDTokenString, refreshTokenString, nil
}

// ParseRefreshToken verifies that a refresh token was encrypted with one of
// this issuer's keys and returns its claims for rotation or revocation.
func ParseRefreshToken(tokenRaw string, config config.Config) (*RefreshTokenClaims, error) {
	specifiedKid := ""
	if idx := strings.Index(tokenRaw, ":"); idx > 0 {
		maybeKid := tokenRaw[:idx]
		if len(maybeKid) == 64 {
			specifiedKid = maybeKid
			tokenRaw = tokenRaw[idx+1:]
		}
	}

	jweObj, err := jwe.ParseEncrypted(tokenRaw)
	if err != nil {
		return nil, err
	}

	var plaintext []byte
	if specifiedKid != "" {
		for _, keyPair := range config.KeyPairs {
			if subtle.ConstantTimeCompare([]byte(KidForPublicKey(keyPair.PublicKey)), []byte(specifiedKid)) != 1 {
				continue
			}
			plaintext, err = jweObj.Decrypt(&keyPair.PrivateKey)
			break
		}
	} else {
		for _, keyPair := range config.KeyPairs {
			plaintext, err = jweObj.Decrypt(&keyPair.PrivateKey)
			if err == nil {
				break
			}
		}
	}
	if err != nil || plaintext == nil {
		if err == nil {
			err = errors.New("refresh token key not found")
		}
		return nil, err
	}

	claims := &RefreshTokenClaims{}
	if err := json.Unmarshal(plaintext, claims); err != nil {
		return nil, err
	}
	return claims, nil
}

type OIDCTokenClaims struct {
	jwt.RegisteredClaims
	Nonce string `json:"nonce,omitempty"`
	// Standard Profile Claims
	Name              string   `json:"name"`
	Email             string   `json:"email,omitempty"`
	EmailVerified     bool     `json:"email_verified,omitempty"`
	PreferredUsername string   `json:"preferred_username,omitempty"`
	Website           *string  `json:"website,omitempty"`
	Birthdate         *string  `json:"birthdate,omitempty"`
	Roles             []string `json:"roles"`
	UpdatedAt         int64    `json:"updated_at,omitempty"`
}

type AccessTokenClaims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope,omitempty"`
}

type RefreshTokenClaims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope,omitempty"`
}

func GenerateIDToken(q *query.Query, jti, userID, clientID, nonce, scopes string, config config.Config) (string, error) {
	user, err := q.User.Where(q.User.ID.Eq(userID)).First()
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}
	profile, err := q.Profile.Where(q.Profile.UserID.Eq(userID)).First()
	if err != nil {
		return "", err
	}
	if profile == nil {
		return "", errors.New("profile not found")
	}
	var roleCustomID []string

	err = q.Role.
		Join(q.UserRole, q.UserRole.RoleID.EqCol(q.Role.ID)).
		Where(q.UserRole.UserID.Eq(userID)).
		Select(q.Role.CustomID).
		Scan(&roleCustomID)
	if err != nil {
		return "", err
	}

	IDTokenClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, OIDCTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Audience:  jwt.ClaimStrings{clientID},
			Issuer:    config.IssuerURL,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Hour)),
			ID:        jti,
		},
		Nonce: nonce,
		Name:  profile.DisplayName,
		Email: func() string {
			if ContainsScope(scopes, "email") {
				return user.Email
			}
			return ""
		}(),
		EmailVerified: func() bool {
			if ContainsScope(scopes, "email") {
				return user.EmailVerified
			}
			return false
		}(),
		PreferredUsername: user.CustomID,
		Website: func() *string {
			if ContainsScope(scopes, "profile") {
				return profile.WebsiteURL
			}
			return nil
		}(),
		Birthdate: func() *string {
			if ContainsScope(scopes, "profile") && profile.Birthdate != nil {
				dateString := profile.Birthdate.Format("2006-01-02")
				return &dateString
			}
			return nil
		}(),
		Roles:     roleCustomID,
		UpdatedAt: profile.UpdatedAt.Unix(),
	})
	IDTokenClaims.Header["kid"] = kidForKey(config)
	if !hasValidKeyPair(config) {
		return "", ErrNoValidKeyPair
	}
	IDTokenString, err := IDTokenClaims.SignedString(&config.KeyPairs[0].PrivateKey)
	return IDTokenString, err
}

// ParseAccessToken verifies an access token's format, signature, and claims
// without consulting token state in the database.
func ParseAccessToken(tokenString string, config config.Config) (*AccessTokenClaims, error) {
	// 前後の空白を除去
	tokenString = strings.TrimSpace(tokenString)

	// もしBarerトークン形式であれば "Bearer " 部分を取り除く
	if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
		tokenString = tokenString[7:]
	}

	// クライアントが誤ってリフレッシュトークン（kid:... の形式）や
	// その他のプレフィックス付きトークンを送ってきた場合に、
	// ライブラリの生のデコードエラーになるのを避け、わかりやすい
	// エラーメッセージを返す。
	if strings.Contains(tokenString, ":") {
		parts := strings.SplitN(tokenString, ":", 2)
		if len(parts[0]) >= 32 && len(parts[0]) <= 128 {
			// 先頭部分が16進文字列かどうかを簡易確認
			if _, hexErr := hex.DecodeString(parts[0]); hexErr == nil {
				return nil, errors.New("token appears to be a refresh token or contains a kid prefix; expected access token")
			}
		}
	}

	// JWT (JWS) は compact serialization で header.payload.signature の
	// 3 つのパート（ドットが2つ）を持つことを期待する。
	if strings.Count(tokenString, ".") != 2 {
		return nil, errors.New("invalid token format: expected JWS compact serialization")
	}

	// トークンをパースして署名とクレームを検証する
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// パースに失敗すると token が nil の可能性があるためチェックする
		if token == nil {
			return nil, errors.New("invalid token")
		}
		// 署名アルゴリズムが RSA 系であることを期待する
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// 公開鍵が設定されていることを確認して返す
		if !hasValidKeyPair(config) {
			return nil, ErrNoValidKeyPair
		}
		return &config.KeyPairs[0].PublicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// ValidateAccessToken は与えられたアクセストークン文字列を検証し、
// 成功した場合はトークンの JTI、サブジェクト（ユーザID）、スコープを返す。
// - tokenString: 検証する JWT アクセストークン文字列
// - c: Gin コンテキスト（中に設定された `config` と `db` を使用）
// 戻り値は順に (jti, sub, scope, err) で、検証失敗時は err に値が入る。
func ValidateAccessToken(tokenString string, c *gin.Context) (jti, sub, scope string, err error) {
	config := *c.MustGet("config").(*config.Config)
	claims, err := ParseAccessToken(tokenString, config)
	if err != nil {
		return "", "", "", err
	}

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		return "", "", "", errors.New("database not available")
	}
	q := query.Use(db)
	tokenSet, err := q.OauthToken.Where(q.OauthToken.AccessTokenJti.Eq(claims.ID), q.OauthToken.DeletedAt.IsNull()).First()
	if err != nil {
		return "", "", "", err
	}
	if tokenSet == nil {
		return "", "", "", errors.New("invalid token")
	}
	return claims.ID, claims.Subject, claims.Scope, nil
}

type SessionTokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

func GenerateSessionJWT(sessionID, userID string, expiresAt time.Time, config config.Config) (string, error) {
	// Generate Session JWT
	sessionTokenClaims := jwt.NewWithClaims(jwt.SigningMethodRS256, SessionTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "SID_" + sessionID,
			Issuer:    config.IssuerURL,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID: userID,
	})
	sessionTokenClaims.Header["kid"] = kidForKey(config)
	if !hasValidKeyPair(config) {
		return "", ErrNoValidKeyPair
	}
	sessionTokenString, err := sessionTokenClaims.SignedString(&config.KeyPairs[0].PrivateKey)
	if err != nil {
		return "", err
	}
	return sessionTokenString, nil
}

func ValidateSessionJWT(tokenString string, c *gin.Context) (sessionID, userID string, err error) {
	config := *c.MustGet("config").(*config.Config)

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &SessionTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// token may be nil if parsing failed
		if token == nil {
			return nil, errors.New("invalid token")
		}
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// Return the public key for verification
		if !hasValidKeyPair(config) {
			return nil, ErrNoValidKeyPair
		}
		return &config.KeyPairs[0].PublicKey, nil
	})
	if err != nil {
		return "", "", err
	}

	// Validate claims
	if claims, ok := token.Claims.(*SessionTokenClaims); ok && token.Valid {
		if len(claims.Subject) <= 4 || claims.Subject[:4] != "SID_" {
			return "", "", errors.New("invalid session token subject")
		}
		return claims.Subject[4:], claims.UserID, nil
	} else {
		return "", "", errors.New("invalid session token claims")
	}
}
