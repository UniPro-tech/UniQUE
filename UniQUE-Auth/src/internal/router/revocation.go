package router

import (
	"errors"
	"net/http"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RevocationRequest struct {
	Token         string  `form:"token" binding:"required"`
	TokenTypeHint *string `form:"token_type_hint" binding:"omitempty"`
	ClientID      string  `form:"client_id"`
	ClientSecret  string  `form:"client_secret"`
}

// Revocation godoc
// @Summary Revoke a token RFC7009
// @Description RFC7009に基づいてアクセストークンを失効させる。
// @Tags oauth2
// @Param token formData string true "Token to revoke"
// @Success 200 {object} map[string]string "OK"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /revocation [post]
func Revocation(c *gin.Context) {
	var req RevocationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	clientID := checkClientAuthentication(c, &TokenGetRequest{
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
	}, false)
	if clientID == nil {
		return
	}

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	config := *c.MustGet("config").(*config.Config)
	tokenJTI := ""
	if req.TokenTypeHint == nil || *req.TokenTypeHint == "access_token" || *req.TokenTypeHint == "refresh_token" {
		tokenJTI, _, _, _ = util.ValidateAccessToken(req.Token, c)
	}
	if tokenJTI == "" && (req.TokenTypeHint == nil || *req.TokenTypeHint == "refresh_token" || *req.TokenTypeHint == "access_token") {
		if claims, err := util.ParseRefreshToken(req.Token, config); err == nil {
			tokenJTI = claims.ID
		}
	}
	if tokenJTI == "" {
		// RFC 7009 requires a successful response for unknown or invalid tokens.
		c.Status(http.StatusOK)
		return
	}

	// Token revocation is transactional and cannot affect another client.
	err := q.Transaction(func(tx *query.Query) error {
		tokenset, err := tx.OauthToken.Where(tx.OauthToken.AccessTokenJti.Eq(tokenJTI)).Or(tx.OauthToken.RefreshTokenJti.Eq(tokenJTI)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if tokenset == nil {
			return nil
		}
		consent, err := tx.Consent.Where(tx.Consent.ID.Eq(tokenset.ConsentID)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if consent == nil || consent.ApplicationID != *clientID {
			return nil
		}
		if _, err := tx.OauthToken.Where(tx.OauthToken.ID.Eq(tokenset.ID)).Delete(); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// RFC7009: トークンが元々存在しない（またはすでに失効している）場合でも、安全のため 200 OK を返す
	c.Status(http.StatusOK)
}
