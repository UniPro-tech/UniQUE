package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/constants"
	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/UniPro-tech/UniQUE-API/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// RegisterAnnouncementRoutes registers announcement routes
func RegisterAnnouncementRoutes(r *gin.Engine) {
	g := r.Group("/announcements")
	{
		g.GET("", listAnnouncements)
		g.GET(":id", getAnnouncement)

		// 管理系: OAuth スコープ (announcements.write / announcements.delete) を許可する
		g.POST("", middleware.RequirePermissionOrScope(constants.ANNOUNCEMENT_CREATE, "announcements.write"), createAnnouncement)
		g.PUT(":id", middleware.RequirePermissionOrScope(constants.ANNOUNCEMENT_UPDATE, "announcements.write"), updateAnnouncement)
		g.DELETE(":id", middleware.RequirePermissionOrScope(constants.ANNOUNCEMENT_DELETE, "announcements.delete"), deleteAnnouncement)
		g.PATCH(":id", middleware.RequirePermissionOrScope(constants.ANNOUNCEMENT_UPDATE, "announcements.write"), patchAnnouncement)
		g.POST(":id/pin", middleware.RequirePermissionOrScope(constants.ANNOUNCEMENT_PIN, "announcements.write"), pinAnnouncement)
	}
}

// listAnnouncements godoc
// @Summary List announcements
// @Description List announcements, pinned first. Use `limit` query to limit results and `deleted` to include deleted records.
// @Tags announcements
// @Produce json
// @Param limit query int false "Limit number of announcements"
// @Param deleted query bool false "Include deleted announcements"
// @Success 200 {object} routes.AnnouncementListResponse
// @Router /announcements [get]
func listAnnouncements(c *gin.Context) {
	db := getDB(c)
	if db == nil {
		return
	}
	q := query.Use(db)
	limitStr := c.Query("limit")
	if limitStr == "" {
		limitStr = "100"
	}
	deletedStr := c.Query("deleted")
	var dao = q.Announcement.Order(query.Announcement.IsPinned.Desc(), query.Announcement.CreatedAt.Desc())

	if deletedStr != "" {
		deleted, errp := strconv.ParseBool(deletedStr)
		if errp != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deleted parameter"})
			return
		}
		if deleted {
			ui, exists := c.Get("user")
			if !exists {
				c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
				return
			}
			su, ok := ui.(*model.User)
			if !ok || su == nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
				return
			}
			perms, _ := middleware.GetUserPermissions(su.ID, db)
			if !perms.HasPermission(constants.ANNOUNCEMENT_UPDATE) {
				c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
				return
			}
		} else {
			dao = dao.Where(query.Announcement.DeletedAt.IsNull())
		}
	} else {
		dao = dao.Where(query.Announcement.DeletedAt.IsNull())
	}
	if limitStr != "" {
		limit, errp := strconv.Atoi(limitStr)
		if errp != nil || limit < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		if limit > 0 {
			dao = dao.Limit(limit)
		}
	}
	anns, err := dao.Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userIDs := make([]string, 0, len(anns))
	for _, a := range anns {
		if a.CreatedBy != "" {
			userIDs = append(userIDs, a.CreatedBy)
		}
	}
	userMap := make(map[string]UserDTO)
	if len(userIDs) > 0 {
		users, _ := q.User.Where(query.User.ID.In(userIDs...)).Find()
		profiles, _ := q.Profile.Where(query.Profile.UserID.In(userIDs...)).Find()
		profileMap := make(map[string]*model.Profile)
		for _, p := range profiles {
			profileMap[p.UserID] = p
		}
		for _, u := range users {
			dto := UserDTO{
				ID:       u.ID,
				CustomID: u.CustomID,
			}
			if p, ok := profileMap[u.ID]; ok {
				dto.Profile = &ProfileDTO{
					UserID:      p.UserID,
					DisplayName: p.DisplayName,
					JoinedAt:    formatDate(p.JoinedAt),
				}
			}
			userMap[u.ID] = dto
		}
	}

	out := make([]AnnouncementDTO, 0, len(anns))
	for _, a := range anns {
		createdBy := UserDTO{ID: "", CustomID: ""}
		if a.CreatedBy != "" {
			if u, ok := userMap[a.CreatedBy]; ok {
				createdBy = u
			} else {
				createdBy = UserDTO{ID: a.CreatedBy}
			}
		}
		out = append(out, AnnouncementDTO{
			ID:        a.ID,
			Title:     a.Title,
			Content:   a.Content,
			CreatedBy: createdBy,
			IsPinned:  a.IsPinned,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			DeletedAt: utils.DeletedAtPtr(a.DeletedAt),
		})
	}
	c.JSON(http.StatusOK, AnnouncementListResponse{Data: out})
}

// getAnnouncement godoc
// @Summary Get an announcement
// @Description Get a single announcement by ID
// @Tags announcements
// @Produce json
// @Param id path string true "Announcement ID"
// @Success 200 {object} routes.AnnouncementDTO
// @Router /announcements/{id} [get]
func getAnnouncement(c *gin.Context) {
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)
	a, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	createdBy := UserDTO{ID: "", CustomID: ""}
	if a.CreatedBy != "" {
		if u, err := q.User.Where(query.User.ID.Eq(a.CreatedBy)).First(); err == nil {
			dtoUser := UserDTO{ID: u.ID, CustomID: u.CustomID}
			if p, err := q.Profile.Where(query.Profile.UserID.Eq(u.ID)).First(); err == nil {
				dtoUser.Profile = &ProfileDTO{UserID: p.UserID, DisplayName: p.DisplayName, JoinedAt: formatDate(p.JoinedAt)}
			}
			createdBy = dtoUser
		} else {
			createdBy = UserDTO{ID: a.CreatedBy}
		}
	}

	dto := AnnouncementDTO{
		ID:        a.ID,
		Title:     a.Title,
		Content:   a.Content,
		CreatedBy: createdBy,
		IsPinned:  a.IsPinned,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		DeletedAt: utils.DeletedAtPtr(a.DeletedAt),
	}
	c.JSON(http.StatusOK, dto)
}

// createAnnouncement godoc
// @Summary Create an announcement
// @Description Create a new announcement
// @Tags announcements
// @Accept json
// @Produce json
// @Param announcement body routes.CreateAnnouncementRequest true "Create announcement"
// @Success 201 {object} routes.AnnouncementDTO
// @Router /announcements [post]
func createAnnouncement(c *gin.Context) {
	db := getDB(c)
	if db == nil {
		return
	}
	var input CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userObj, _ := c.Get("user")
	createdByID := ""
	if um, ok := userObj.(*model.User); ok && um != nil {
		createdByID = um.ID
	}
	ann := &model.Announcement{
		ID:        ulid.Make().String(),
		Title:     input.Title,
		Content:   input.Content,
		CreatedBy: createdByID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if input.IsPinned != nil {
		ann.IsPinned = *input.IsPinned
	}

	var dto AnnouncementDTO

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)

		if err := q.Announcement.Create(ann); err != nil {
			return err
		}

		createdByObj := UserDTO{ID: "", CustomID: ""}
		if ann.CreatedBy != "" {
			if u, err := q.User.Where(query.User.ID.Eq(ann.CreatedBy)).First(); err == nil {
				createdByObj = UserDTO{ID: u.ID, CustomID: u.CustomID}
				if p, err := q.Profile.Where(query.Profile.UserID.Eq(u.ID)).First(); err == nil {
					createdByObj.Profile = &ProfileDTO{UserID: p.UserID, DisplayName: p.DisplayName, JoinedAt: formatDate(p.JoinedAt)}
				}
			} else {
				createdByObj = UserDTO{ID: ann.CreatedBy}
			}
		}

		dto = AnnouncementDTO{
			ID:        ann.ID,
			Title:     ann.Title,
			Content:   ann.Content,
			CreatedBy: createdByObj,
			IsPinned:  ann.IsPinned,
			CreatedAt: ann.CreatedAt,
			UpdatedAt: ann.UpdatedAt,
			DeletedAt: utils.DeletedAtPtr(ann.DeletedAt),
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto)
}

// updateAnnouncement godoc
// @Summary Update an announcement
// @Description Update announcement fields
// @Tags announcements
// @Accept json
// @Produce json
// @Param id path string true "Announcement ID"
// @Param announcement body routes.UpdateAnnouncementRequest true "Update announcement"
// @Success 200 {object} routes.AnnouncementDTO
// @Router /announcements/{id} [put]
func updateAnnouncement(c *gin.Context) {
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	var input UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. モデル構造体で更新用データを準備
	updates := model.Announcement{
		UpdatedAt: time.Now().UTC(),
	}
	selectColumns := []field.Expr{query.Announcement.UpdatedAt}

	if input.Title != nil {
		updates.Title = *input.Title
		selectColumns = append(selectColumns, query.Announcement.Title)
	}
	if input.Content != nil {
		updates.Content = *input.Content
		selectColumns = append(selectColumns, query.Announcement.Content)
	}
	if input.IsPinned != nil {
		updates.IsPinned = *input.IsPinned
		selectColumns = append(selectColumns, query.Announcement.IsPinned)
	}

	var dto AnnouncementDTO
	var isNotFound bool

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)

		// 2. Selectでカラムを絞り込んでモデル更新を実行
		if _, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).
			Select(selectColumns...).
			Updates(&updates); err != nil {
			return err
		}

		a, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).First()
		if err != nil {
			isNotFound = true
			return err
		}

		createdBy := UserDTO{ID: "", CustomID: ""}
		if a.CreatedBy != "" {
			if u, err := q.User.Where(query.User.ID.Eq(a.CreatedBy)).First(); err == nil {
				createdBy = UserDTO{ID: u.ID, CustomID: u.CustomID}
				if p, err := q.Profile.Where(query.Profile.UserID.Eq(u.ID)).First(); err == nil {
					createdBy.Profile = &ProfileDTO{UserID: p.UserID, DisplayName: p.DisplayName, JoinedAt: formatDate(p.JoinedAt)}
				}
			} else {
				createdBy = UserDTO{ID: a.CreatedBy}
			}
		}

		dto = AnnouncementDTO{
			ID:        a.ID,
			Title:     a.Title,
			Content:   a.Content,
			CreatedBy: createdBy,
			IsPinned:  a.IsPinned,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			DeletedAt: utils.DeletedAtPtr(a.DeletedAt),
		}
		return nil
	})

	if err != nil {
		if isNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto)
}

// patchAnnouncement godoc
// @Summary Partially update an announcement
// @Description Partially update announcement fields
// @Tags announcements
// @Accept json
// @Produce json
// @Param id path string true "Announcement ID"
// @Param announcement body routes.PatchAnnouncementRequest true "Patch announcement"
// @Success 200 {object} routes.AnnouncementDTO
// @Router /announcements/{id} [patch]
func patchAnnouncement(c *gin.Context) {
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	var input PatchAnnouncementRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. モデル構造体で更新用データを準備
	updates := model.Announcement{
		UpdatedAt: time.Now().UTC(),
	}
	selectColumns := []field.Expr{query.Announcement.UpdatedAt}

	if input.Title != nil {
		updates.Title = *input.Title
		selectColumns = append(selectColumns, query.Announcement.Title)
	}
	if input.Content != nil {
		updates.Content = *input.Content
		selectColumns = append(selectColumns, query.Announcement.Content)
	}
	if input.IsPinned != nil {
		updates.IsPinned = *input.IsPinned
		selectColumns = append(selectColumns, query.Announcement.IsPinned)
	}

	var dto AnnouncementDTO
	var isNotFound bool

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)

		// 2. Selectでカラムを絞り込んでモデル更新を実行
		if _, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).
			Select(selectColumns...).
			Updates(&updates); err != nil {
			return err
		}

		a, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).First()
		if err != nil {
			isNotFound = true
			return err
		}

		createdBy := UserDTO{ID: "", CustomID: ""}
		if a.CreatedBy != "" {
			if u, err := q.User.Where(query.User.ID.Eq(a.CreatedBy)).First(); err == nil {
				createdBy = UserDTO{ID: u.ID, CustomID: u.CustomID}
				if p, err := q.Profile.Where(query.Profile.UserID.Eq(u.ID)).First(); err == nil {
					createdBy.Profile = &ProfileDTO{UserID: p.UserID, DisplayName: p.DisplayName, JoinedAt: formatDate(p.JoinedAt)}
				}
			} else {
				createdBy = UserDTO{ID: a.CreatedBy}
			}
		}

		dto = AnnouncementDTO{
			ID:        a.ID,
			Title:     a.Title,
			Content:   a.Content,
			CreatedBy: createdBy,
			IsPinned:  a.IsPinned,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			DeletedAt: utils.DeletedAtPtr(a.DeletedAt),
		}
		return nil
	})

	if err != nil {
		if isNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto)
}

// deleteAnnouncement godoc
// @Summary Delete an announcement
// @Description Delete an announcement by ID
// @Tags announcements
// @Produce json
// @Param id path string true "Announcement ID"
// @Success 204 {string} string
// @Router /announcements/{id} [delete]
func deleteAnnouncement(c *gin.Context) {
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)
		if _, err := q.Announcement.Delete(&model.Announcement{ID: id}); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// pinAnnouncement godoc
// @Summary Pin or unpin an announcement
// @Description Toggle pin state for an announcement
// @Tags announcements
// @Accept json
// @Produce json
// @Param id path string true "Announcement ID"
// @Param body body map[string]bool true "{\"pin\": true}"
// @Success 200 {object} routes.AnnouncementDTO
// @Router /announcements/{id}/pin [post]
func pinAnnouncement(c *gin.Context) {
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	var input struct {
		Pin bool `json:"pin"`
	}
	_ = c.ShouldBindJSON(&input)

	// 1. モデル構造体でピン留め用データを準備
	updates := model.Announcement{
		IsPinned:  input.Pin,
		UpdatedAt: time.Now().UTC(),
	}
	// 明示的にIsPinnedとUpdatedAtを更新対象にする
	selectColumns := []field.Expr{
		query.Announcement.IsPinned,
		query.Announcement.UpdatedAt,
	}

	var dto AnnouncementDTO
	var isNotFound bool

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)

		// 2. 指定したモデルのフィールドのみを対象にUpdatesを実行
		if _, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).
			Select(selectColumns...).
			Updates(&updates); err != nil {
			return err
		}

		a, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).First()
		if err != nil {
			isNotFound = true
			return err
		}

		createdBy := UserDTO{ID: "", CustomID: ""}
		if a.CreatedBy != "" {
			if u, err := q.User.Where(query.User.ID.Eq(a.CreatedBy)).First(); err == nil {
				createdBy = UserDTO{ID: u.ID, CustomID: u.CustomID}
				if p, err := q.Profile.Where(query.Profile.UserID.Eq(u.ID)).First(); err == nil {
					createdBy.Profile = &ProfileDTO{UserID: p.UserID, DisplayName: p.DisplayName, JoinedAt: formatDate(p.JoinedAt)}
				}
			} else {
				createdBy = UserDTO{ID: a.CreatedBy}
			}
		}

		dto = AnnouncementDTO{
			ID:        a.ID,
			Title:     a.Title,
			Content:   a.Content,
			CreatedBy: createdBy,
			IsPinned:  a.IsPinned,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			DeletedAt: utils.DeletedAtPtr(a.DeletedAt),
		}
		return nil
	})

	if err != nil {
		if isNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto)
}
