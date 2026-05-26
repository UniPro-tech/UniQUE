package router

import (
	"errors"
	"net/http"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/model"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"gorm.io/gen"
	"gorm.io/gorm"
)

type ConsentCreateRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	ApplicationID string `json:"application_id" binding:"required"`
	Scope         string `json:"scope" binding:"required"`
}

// ConsentResponse is used for Swagger responses to avoid depending on gorm.DeletedAt
type ConsentResponse struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	ApplicationID string     `json:"application_id"`
	Scope         string     `json:"scope"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}

// ConsentList godoc
// @Summary list or search consents
// @Description 一覧取得およびクエリ条件による検索を行う。
// @Tags internal
// @Accept json
// @Produce json
// @Param user_id query string false "User ID"
// @Param application_id query string false "Application ID"
// @Param scope query string false "Scope"
// @Success 200 {array} router.ConsentResponse
// @Router /internal/consents [get]
func ConsentList(c *gin.Context) {
	userID := c.Query("user_id")
	appID := c.Query("application_id")
	scope := c.Query("scope")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	// build conditions dynamically
	conds := []gen.Condition{}
	if userID != "" {
		conds = append(conds, q.Consent.UserID.Eq(userID))
	}
	if appID != "" {
		conds = append(conds, q.Consent.ApplicationID.Eq(appID))
	}
	if scope != "" {
		conds = append(conds, q.Consent.Scope.Eq(scope))
	}

	var results []*model.Consent
	var err error

	// perform query
	if len(conds) > 0 {
		results, err = q.Consent.Where(conds...).Find()
	} else {
		results, err = q.Consent.Find()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// ConsentCreate godoc
// @Summary create consent
// @Tags internal
// @Accept json
// @Produce json
// @Param request body ConsentCreateRequest true "Create Consent"
// @Success 201 {object} router.ConsentResponse
// @Router /internal/consents [post]
func ConsentCreate(c *gin.Context) {
	req := ConsentCreateRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	now := time.Now().UTC()
	newConsent := &model.Consent{
		ID:            ulid.Make().String(),
		UserID:        req.UserID,
		ApplicationID: req.ApplicationID,
		Scope:         req.Scope,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := q.Consent.Create(newConsent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	writeAuditLog(c, "CREATE", "consents/"+newConsent.ID, &newConsent.UserID, &newConsent.ApplicationID, nil, map[string]interface{}{
		"method":          c.Request.Method,
		"path":            c.Request.URL.Path,
		"status":          http.StatusCreated,
		"ip":              c.ClientIP(),
		"user_agent":      c.Request.UserAgent(),
		"consent_user_id": newConsent.UserID,
		"application_id":  newConsent.ApplicationID,
		"scope":           newConsent.Scope,
	})

	c.JSON(http.StatusCreated, newConsent)
}

// ConsentDeleteByID godoc
// @Summary delete consent by id
// @Tags internal
// @Accept json
// @Produce json
// @Param id path string true "Consent ID"
// @Success 200 {object} map[string]string
// @Router /internal/consents/{id} [delete]
func ConsentDeleteByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	var consent *model.Consent

	// 紐づくトークンとコンセント自体の削除をトランザクション化
	err := q.Transaction(func(tx *query.Query) error {
		var fetchErr error
		consent, fetchErr = tx.Consent.Where(tx.Consent.ID.Eq(id)).First()
		if fetchErr != nil {
			return fetchErr
		}
		if consent == nil {
			return gorm.ErrRecordNotFound
		}

		// ループ処理を廃止し、紐づくトークンを一括削除(バッチデリート)
		if _, err := tx.OauthToken.Where(tx.OauthToken.ConsentID.Eq(consent.ID)).Delete(); err != nil {
			return err
		}

		// コンセントの削除
		if _, err := tx.Consent.Delete(consent); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "consent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	writeAuditLog(c, "DELETE", "consents/"+id, &consent.UserID, &consent.ApplicationID, nil, map[string]interface{}{
		"method":          c.Request.Method,
		"path":            c.Request.URL.Path,
		"status":          http.StatusOK,
		"ip":              c.ClientIP(),
		"user_agent":      c.Request.UserAgent(),
		"consent_user_id": consent.UserID,
		"application_id":  consent.ApplicationID,
		"scope":           consent.Scope,
	})

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
