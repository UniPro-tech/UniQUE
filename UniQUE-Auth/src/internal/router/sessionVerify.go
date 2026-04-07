package router

import (
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SessionVerifyRequest struct {
	JTI string `form:"jti" binding:"required"`
}

type SessionVerifyResponse struct {
	Valid     bool      `json:"valid"`
	UserID    string    `json:"user_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// SessionVerifyPost godoc
// @Summary session verify endpoint
// @Schemes
// @Description 内部用のセッション検証エンドポイントです。Kubernetes / Istio の認証ポリシーにより外部からのアクセスは制限されています。
// @Tags internal
// @Success 200 {object} SessionVerifyResponse "OK"
// @Param jti query string true "JTI"
// @Accept json
// @Router /internal/session_verify [get]
func SessionVerifyGet(c *gin.Context) {
	req := SessionVerifyRequest{}
	if err := c.ShouldBind(&req); err != nil {
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
	session, err := q.Session.Where(q.Session.ID.Eq(req.JTI), q.Session.DeletedAt.IsNull(), q.Session.ExpiresAt.Gt(time.Now())).First()
	if err != nil {
		c.JSON(200, SessionVerifyResponse{Valid: false})
		return
	}
	c.JSON(200, SessionVerifyResponse{Valid: true, UserID: session.UserID, ExpiresAt: session.ExpiresAt})
}
