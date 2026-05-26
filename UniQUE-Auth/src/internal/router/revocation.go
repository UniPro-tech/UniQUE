package router

import (
	"errors"
	"net/http"

	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RevocationRequest struct {
	Token         string  `form:"token" binding:"required"`
	TokenTypeHint *string `form:"token_type_hint" binding:"omitempty"`
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

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	// トークンの失効処理をトランザクション化
	err := q.Transaction(func(tx *query.Query) error {
		// token_type_hint が指定されている場合はその種別を優先して探索
		if req.TokenTypeHint != nil {
			switch *req.TokenTypeHint {
			case "access_token":
				res, err := tx.OauthToken.Where(tx.OauthToken.AccessTokenJti.Eq(req.Token)).Delete()
				if err != nil {
					return err
				}
				// access_token が見つからなければ refresh_token 側を試す (RFC7009 フォールバック仕様)
				if res.RowsAffected == 0 {
					if _, err := tx.OauthToken.Where(tx.OauthToken.RefreshTokenJti.Eq(req.Token)).Delete(); err != nil {
						return err
					}
				}
			case "refresh_token":
				res, err := tx.OauthToken.Where(tx.OauthToken.RefreshTokenJti.Eq(req.Token)).Delete()
				if err != nil {
					return err
				}
				// refresh_token が見つからなければ access_token 側を試す (RFC7009 フォールバック仕様)
				if res.RowsAffected == 0 {
					if _, err := tx.OauthToken.Where(tx.OauthToken.AccessTokenJti.Eq(req.Token)).Delete(); err != nil {
						return err
					}
				}
			default:
				// 不明なヒントの場合は OR 条件を用いて1回のクエリで両方から探索・削除
				if _, err := tx.OauthToken.Where(tx.OauthToken.AccessTokenJti.Eq(req.Token)).Or(tx.OauthToken.RefreshTokenJti.Eq(req.Token)).Delete(); err != nil {
					return err
				}
			}
		} else {
			// ヒントがない場合は OR 条件を用いて1回のクエリで両方を探索・削除 (クエリ数を削減)
			if _, err := tx.OauthToken.Where(tx.OauthToken.AccessTokenJti.Eq(req.Token)).Or(tx.OauthToken.RefreshTokenJti.Eq(req.Token)).Delete(); err != nil {
				return err
			}
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
