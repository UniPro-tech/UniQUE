package router

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/gin-gonic/gin"
)

type WellKnownOpenIDConfigurationResponse struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	RevocationEndpoint                string   `json:"revocation_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
}

// WellKnownOpenIDConfiguration godoc
// @Summary OpenID Configuration
// @Description .well-known/openid-configuration エンドポイント
// @Produce json
// @Tags metadata
// @Success 200 {object} WellKnownOpenIDConfigurationResponse
// @Router /.well-known/openid-configuration [get]
func WellKnownOpenIDConfiguration(c *gin.Context) {
	environmentConfigs := c.MustGet("config").(*config.Config)

	c.JSON(200, WellKnownOpenIDConfigurationResponse{
		Issuer:                            environmentConfigs.IssuerURL,
		AuthorizationEndpoint:             environmentConfigs.IssuerURL + "/authorization",
		TokenEndpoint:                     environmentConfigs.IssuerURL + "/token",
		UserinfoEndpoint:                  environmentConfigs.IssuerURL + "/userinfo",
		RevocationEndpoint:                environmentConfigs.IssuerURL + "/revocation",
		JwksURI:                           environmentConfigs.IssuerURL + "/.well-known/jwks.json",
		ResponseTypesSupported:            []string{"code" /* TODO: Impliment "code token", etc... */},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		ScopesSupported:                   config.Scopes.AllowedScopes,
		TokenEndpointAuthMethodsSupported: []string{"client_secret_basic", "client_secret_post", "none"},
		ClaimsSupported: []string{
			"sub",
			"iss",
			"aud",
			"exp",
			"iat",
			"name",
			"preferred_username",
			"email",
			"email_verified",
			"roles",
			// TODO: Impliment acr, amr, etc...
		},
		CodeChallengeMethodsSupported: []string{"S256"},
		GrantTypesSupported:           []string{"authorization_code", "refresh_token"},
	})
}

type JWKS struct {
	Keys []JWKSKey `json:"keys"`
}

type JWKSKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// WellKnownJWKS godoc
// @Summary JSON Web Key Set
// @Description .well-known/jwks.json エンドポイント
// @Produce json
// @Tags metadata
// @Success 200 {object} JWKS
// @Router /.well-known/jwks.json [get]
func WellKnownJWKS(c *gin.Context) {
	// Get Keys Path from config
	environmentConfigs := c.MustGet("config").(*config.Config)

	jwks, err := generateJWKSFromConfig(environmentConfigs)
	if err != nil {
		c.JSON(500, gin.H{"error": "internal_server_error"})
		return
	}
	c.JSON(200, jwks)
}

func generateJWKSFromConfig(cfg *config.Config) (*JWKS, error) {
	var keys []JWKSKey

	// iterate over configured key pairs
	for _, kp := range cfg.KeyPairs {
		rsaPub := &kp.PublicKey

		// marshal to PKIX DER to compute kid
		der, err := x509.MarshalPKIXPublicKey(rsaPub)
		if err != nil {
			continue
		}
		thumb := sha256.Sum256(der)
		kid := hex.EncodeToString(thumb[:])

		n := base64.RawURLEncoding.EncodeToString(rsaPub.N.Bytes())

		eInt := rsaPub.E
		eBytes := []byte{}
		for x := eInt; x > 0; x >>= 8 {
			eBytes = append([]byte{byte(x & 0xff)}, eBytes...)
		}
		if len(eBytes) == 0 {
			eBytes = []byte{0}
		}
		e := base64.RawURLEncoding.EncodeToString(eBytes)

		jwk := JWKSKey{
			Kty: "RSA",
			Use: "sig",
			Alg: "RS256",
			Kid: kid,
			N:   n,
			E:   e,
		}
		keys = append(keys, jwk)
	}

	return &JWKS{Keys: keys}, nil
}
