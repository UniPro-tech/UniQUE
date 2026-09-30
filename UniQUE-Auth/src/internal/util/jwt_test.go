package util

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

func TestParseAccessTokenDoesNotRequireDatabaseState(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	cfg := config.Config{KeyPairs: []config.KeyPairConfig{{
		PublicKey:  privateKey.PublicKey,
		PrivateKey: *privateKey,
	}}}
	want := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "access-token-jti",
			Subject:   "user-id",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Scope: "openid profile",
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, want).SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}

	got, err := ParseAccessToken(token, cfg)
	if err != nil {
		t.Fatalf("ParseAccessToken returned an error: %v", err)
	}
	if got.ID != want.ID || got.Subject != want.Subject || got.Scope != want.Scope {
		t.Fatalf("unexpected claims: got %#v", got)
	}
}

func TestParseAccessTokenRejectsRefreshTokenFormat(t *testing.T) {
	_, err := ParseAccessToken("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef:not-a-jwt", config.Config{})
	if err == nil {
		t.Fatal("ParseAccessToken accepted a refresh-token-shaped value")
	}
}

func TestParseAccessTokenReportsMissingIssuerKeys(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, AccessTokenClaims{})
	unsigned, err := token.SignedString(mustTestRSAKey(t))
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}

	_, err = ParseAccessToken(unsigned, config.Config{})
	if !errors.Is(err, ErrNoValidKeyPair) {
		t.Fatalf("expected ErrNoValidKeyPair, got %v", err)
	}
}

func TestParseAccessTokenForRevocationAllowsExpiredToken(t *testing.T) {
	privateKey := mustTestRSAKey(t)
	now := time.Now()
	cfg := config.Config{
		IssuerURL: "https://issuer.example.com",
		KeyPairs: []config.KeyPairConfig{{
			PublicKey:  privateKey.PublicKey,
			PrivateKey: *privateKey,
		}},
	}
	claims := AccessTokenClaims{RegisteredClaims: jwt.RegisteredClaims{
		ID:        "expired-access-token-jti",
		Subject:   "user-id",
		Audience:  jwt.ClaimStrings{"client-id"},
		Issuer:    cfg.IssuerURL,
		IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
	}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}

	got, err := ParseAccessTokenForRevocation(token, cfg)
	if err != nil {
		t.Fatalf("ParseAccessTokenForRevocation returned an error: %v", err)
	}
	if got.ID != claims.ID {
		t.Fatalf("unexpected JTI: got %q, want %q", got.ID, claims.ID)
	}
	if _, err := ParseAccessToken(token, cfg); err == nil {
		t.Fatal("ParseAccessToken accepted an expired token")
	}
}

func TestParseAccessTokenForRevocationRejectsInvalidSignature(t *testing.T) {
	trustedKey := mustTestRSAKey(t)
	untrustedKey := mustTestRSAKey(t)
	now := time.Now()
	cfg := config.Config{
		IssuerURL: "https://issuer.example.com",
		KeyPairs: []config.KeyPairConfig{{
			PublicKey:  trustedKey.PublicKey,
			PrivateKey: *trustedKey,
		}},
	}
	claims := AccessTokenClaims{RegisteredClaims: jwt.RegisteredClaims{
		ID:        "expired-access-token-jti",
		Subject:   "user-id",
		Audience:  jwt.ClaimStrings{"client-id"},
		Issuer:    cfg.IssuerURL,
		IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
	}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(untrustedKey)
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}

	if _, err := ParseAccessTokenForRevocation(token, cfg); err == nil {
		t.Fatal("ParseAccessTokenForRevocation accepted an invalid signature")
	}
}

func mustTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return key
}
