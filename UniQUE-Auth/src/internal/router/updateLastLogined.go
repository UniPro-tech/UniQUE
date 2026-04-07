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
	if err := q.Session.Save(session); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Status(201)
}
