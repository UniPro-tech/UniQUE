package router

import (
	"errors"
	"net/http"
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
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	session, err := q.Session.Where(q.Session.ID.Eq(req.SID), q.Session.DeletedAt.IsNull()).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": "session not found"})
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if session == nil {
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}

	session.LastLoginAt = time.Now().UTC()
	session.UpdatedAt = time.Now().UTC()

	// CreatedAt 基準の当初期間を使うことで、更新を重ねても remember/非 remember の区分が保たれる
	isRemember := session.IsRemember
	if isRemember {
		session.ExpiresAt = time.Now().Add(30 * 24 * time.Hour) // remember: 30日
	} else {
		session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 非 remember: 7日
	}
	if err := q.Session.Save(session); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Status(201)
}
