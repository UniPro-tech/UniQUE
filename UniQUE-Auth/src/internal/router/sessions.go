package router

import (
	"errors"
	"net/http"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SessionResponse is used for Swagger documentation to avoid gorm.DeletedAt parsing issues
type SessionResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	IPAddress   string     `json:"ip_address"`
	UserAgent   string     `json:"user_agent"`
	IsRemember  bool       `json:"is_remember"`
	ExpiresAt   time.Time  `json:"expires_at"`
	LastLoginAt time.Time  `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type SessionListResponse struct {
	Data []SessionResponse `json:"data"`
}

// SessionsGet godoc
// @Summary Get sessions for a user
// @Description 内部用: 指定ユーザーのセッション一覧を取得する。Kubernetes / Istio の認証ポリシーにより外部からのアクセスは制限されています。
// @Tags internal
// @Param user_id query string true "User ID"
// @Success 200 {object} SessionListResponse "OK"
// @Router /internal/sessions/ [get]
func SessionsGet(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter is required"})
		return
	}

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	sessions, err := q.Session.Where(q.Session.UserID.Eq(userID), q.Session.DeletedAt.IsNull()).Order(q.Session.CreatedAt.Desc()).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := make([]SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		if s == nil {
			continue
		}
		resp = append(resp, SessionResponse{
			ID:          s.ID,
			UserID:      s.UserID,
			IPAddress:   s.IPAddress,
			UserAgent:   s.UserAgent,
			IsRemember:  s.IsRemember,
			ExpiresAt:   s.ExpiresAt,
			LastLoginAt: s.LastLoginAt,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
			DeletedAt:   util.DeletedAtPtr(s.DeletedAt),
		})
	}
	c.JSON(http.StatusOK, SessionListResponse{Data: resp})
}

// SessionsDelete godoc
// @Summary Delete sessions for a user
// @Description 内部用: 指定ユーザーのセッションを削除（ソフトデリート）する。Kubernetes / Istio の認証ポリシーにより外部からのアクセスは制限されています。
// @Tags internal
// @Success 204 {string} string "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /internal/sessions/:sid [delete]
func SessionsDelete(c *gin.Context) {
	sid := c.Param("sid")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	var userID string

	// レコードの取得と削除をトランザクション化
	err := q.Transaction(func(tx *query.Query) error {
		session, err := tx.Session.Where(tx.Session.ID.Eq(sid)).First()
		if err != nil {
			return err
		}
		userID = session.UserID

		if _, err := tx.Session.Where(tx.Session.ID.Eq(sid)).Delete(); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 既に削除されている、または存在しない場合は冪等性を保つために204を返す
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	writeAuditLog(c, "DELETE", "sessions/"+sid, &userID, nil, &sid, map[string]interface{}{
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"status":     http.StatusNoContent,
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	})

	c.Status(http.StatusNoContent)
}

// GetSessionById godoc
// @Summary Get session by ID
// @Description 内部用: セッションIDからセッション情報を取得する。Kubernetes / Istio の認証ポリシーにより外部からのアクセスは制限されています。
// @Tags internal
// @Param sid path string true "Session ID"
// @Success 200 {object} SessionResponse "OK"
// @Router /internal/sessions/{sid} [get]
func GetSessionById(c *gin.Context) {
	sid := c.Param("sid")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	session, err := q.Session.Where(q.Session.ID.Eq(sid)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := SessionResponse{
		ID:          session.ID,
		UserID:      session.UserID,
		IPAddress:   session.IPAddress,
		UserAgent:   session.UserAgent,
		IsRemember:  session.IsRemember,
		ExpiresAt:   session.ExpiresAt,
		LastLoginAt: session.LastLoginAt,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
		DeletedAt:   util.DeletedAtPtr(session.DeletedAt),
	}
	c.JSON(http.StatusOK, resp)
}
