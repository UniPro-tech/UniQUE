package router

import (
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
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(500, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	// トークンの失効処理
	// token_type_hint が指定されている場合はその種別を優先して探索し、見つからなければフォールバックしてもう一方を探索する
	// RFC7009 に従い、トークンが見つからなくてもエラーは返さず成功(200)を返す
	if req.TokenTypeHint != nil {
		switch *req.TokenTypeHint {
		case "access_token":
			if res, err := q.OauthToken.Where(q.OauthToken.AccessTokenJti.Eq(req.Token)).Delete(); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			} else if res.RowsAffected == 0 {
				// access_token が見つからなければ refresh_token 側を試す
				if _, err := q.OauthToken.Where(q.OauthToken.RefreshTokenJti.Eq(req.Token)).Delete(); err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
			}
		case "refresh_token":
			if res, err := q.OauthToken.Where(q.OauthToken.RefreshTokenJti.Eq(req.Token)).Delete(); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			} else if res.RowsAffected == 0 {
				// refresh_token が見つからなければ access_token 側を試す
				if _, err := q.OauthToken.Where(q.OauthToken.AccessTokenJti.Eq(req.Token)).Delete(); err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
			}
		default:
			// 不明なヒントは両方試す
			if _, err := q.OauthToken.Where(q.OauthToken.AccessTokenJti.Eq(req.Token)).Delete(); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			if _, err := q.OauthToken.Where(q.OauthToken.RefreshTokenJti.Eq(req.Token)).Delete(); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}
	} else {
		// ヒントがない場合は両方を探索する
		if _, err := q.OauthToken.Where(q.OauthToken.AccessTokenJti.Eq(req.Token)).Delete(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if _, err := q.OauthToken.Where(q.OauthToken.RefreshTokenJti.Eq(req.Token)).Delete(); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	// RFC7009: トークンが存在しない場合でも 200 を返す
	c.Status(200)
}
