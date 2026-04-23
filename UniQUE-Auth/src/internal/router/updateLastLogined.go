package router

import (
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateLastLoginedRequest struct {
	SID string `json:"sid" binding:"required"`
}

// UpdateLastLogined godoc
// @Summary      Update Last Logined
// @Description  内部のLastLogined更新エンドポイント
// @Tags         internal
// @Param request body UpdateLastLoginedRequest true "Session Data"
// @Success      201  {null}  ""
// @Failure      400  {object}  map[string]string
// @Router       /internal/update_last_logined [post]
func UpdateLastLogined(c *gin.Context) {
	req := UpdateLastLoginedRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(500, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	session, err := q.Session.Where(q.Session.ID.Eq(req.SID), q.Session.DeletedAt.IsNull()).First()
	if err != nil || session == nil {
		c.JSON(400, gin.H{"error": "invalid session id"})
		return
	}

	session.LastLoginAt = time.Now()

	// ExpiresAt と CreatedAt の差分（time.Duration）を取得
	duration := session.ExpiresAt.Sub(session.CreatedAt)

	// 差分が30日（30 * 24時間）未満かどうかを比較
	if duration < 30*24*time.Hour {
		// rememberなし
		session.ExpiresAt = time.Now().AddDate(0, 0, 7)
	} else {
		// rememberあり
		session.ExpiresAt = time.Now().AddDate(0, 1, 0)
	}
	if err := q.Session.Save(session); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Status(201)
}
