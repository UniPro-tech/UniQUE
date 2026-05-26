package router

import (
	"errors"
	"net/http"
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

// SessionVerifyGet godoc
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
	// GETリクエストのため、ShouldBindQuery を使用して明示的にクエリからバインドする
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	session, err := q.Session.Where(
		q.Session.ID.Eq(req.JTI),
		q.Session.DeletedAt.IsNull(),
		q.Session.ExpiresAt.Gt(time.Now().UTC()),
	).First()

	if err != nil {
		// 対象が見つからない（または削除済み・期限切れ）場合は、正常な処理として Valid: false を返す
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, SessionVerifyResponse{Valid: false})
			return
		}
		// データベース接続エラーなど、予期せぬシステムエラーの場合は500を返す
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, SessionVerifyResponse{
		Valid:     true,
		UserID:    session.UserID,
		ExpiresAt: session.ExpiresAt,
	})
}
