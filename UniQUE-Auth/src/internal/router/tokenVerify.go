package router

import (
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TokenVerifyRequest struct {
	Jti string `form:"jti" binding:"required"`
}

type TokenVerifyResponse struct {
	Valid bool `json:"valid"`
}

// TokenVerifyGet godoc
// @Summary token verify endpoint
// @Schemes
// @Description 内部用のトークン検証エンドポイントです。Kubernetes / Istio の認証ポリシーにより外部からのアクセスは制限されています。
// @Tags internal
// @Success 200 {object} TokenVerifyResponse "OK"
// @Param jti query string true "JTI"
// @Accept json
// @Router /internal/token_verify [get]
func TokenVerifyGet(c *gin.Context) {
	req := TokenVerifyRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(500, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	tokenset, err := q.OauthToken.Where(q.OauthToken.AccessTokenJti.Eq(req.Jti), q.OauthToken.DeletedAt.IsNull()).First()

	if err != nil || tokenset == nil {
		c.JSON(400, gin.H{"error": "invalid token"})
		return
	}

	c.JSON(200, TokenVerifyResponse{Valid: true})
}
