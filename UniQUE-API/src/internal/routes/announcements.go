package routes

import (
	"net/http"
	"strconv"

	"github.com/UniPro-tech/UniQUE-API/internal/constants"
	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/UniPro-tech/UniQUE-API/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
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
	// limit パラメータの読み取り（省略時は100件）
	limitStr := c.Query("limit")
	if limitStr == "" {
		limitStr = "100"
	}
	// deletedパラメータの読み取り（省略時はfalse）
	deletedStr := c.Query("deleted")
	var (
		dao = q.Announcement.Order(query.Announcement.IsPinned.Desc(), query.Announcement.CreatedAt.Desc())
	)
	if deletedStr != "" {
		deleted, errp := strconv.ParseBool(deletedStr)
		if errp != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deleted parameter"})
			return
		}
		if deleted {
			// deleted=true は ANNOUNCEMENT_UPDATE 権限があるユーザーのみ許可
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
			// 権限あり -> 削除済み含めて返す（フィルタなし）
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
	// Build map of user IDs to UserDTOs
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
				// fallback: minimal object with ID
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
	// Build CreatedBy UserDTO
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
	// 作成者
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
	}
	if input.IsPinned != nil {
		ann.IsPinned = *input.IsPinned
	}
	q := query.Use(db)
	if err := q.Announcement.Create(ann); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// build CreatedBy object
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

	dto := AnnouncementDTO{
		ID:        ann.ID,
		Title:     ann.Title,
		Content:   ann.Content,
		CreatedBy: createdByObj,
		IsPinned:  ann.IsPinned,
		CreatedAt: ann.CreatedAt,
		UpdatedAt: ann.UpdatedAt,
		DeletedAt: utils.DeletedAtPtr(ann.DeletedAt),
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
	updates := map[string]interface{}{}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Content != nil {
		updates["content"] = *input.Content
	}
	if input.IsPinned != nil {
		updates["is_pinned"] = *input.IsPinned
	}
	q := query.Use(db)
	if len(updates) > 0 {
		if _, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).Updates(updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	a, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// Build CreatedBy object
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
	updates := map[string]interface{}{}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Content != nil {
		updates["content"] = *input.Content
	}
	if input.IsPinned != nil {
		updates["is_pinned"] = *input.IsPinned
	}
	q := query.Use(db)
	if len(updates) > 0 {
		if _, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).Updates(updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	a, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// Build CreatedBy object
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
	q := query.Use(db)
	if _, err := q.Announcement.Delete(&model.Announcement{ID: id}); err != nil {
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
	// toggle pin or set true if absent
	var input struct {
		Pin bool `json:"pin"`
	}
	_ = c.ShouldBindJSON(&input)
	q := query.Use(db)
	if _, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).Update(query.Announcement.IsPinned, input.Pin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	a, err := q.Announcement.Where(query.Announcement.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// Build CreatedBy object
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
