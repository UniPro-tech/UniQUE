package routes

import (
	"net/http"

	"github.com/UniPro-tech/UniQUE-API/internal/constants"
	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/UniPro-tech/UniQUE-API/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

func RegisterApplicationRoutes(r *gin.Engine) {
	// 公開ルート（読み取り専用）
	g := r.Group("/applications")
	{
		g.GET("", listApplications)
		g.POST("", createApplication)
		g.GET(":id", getApplication)
		g.PUT(":id", updateApplication)
		g.PATCH(":id", patchApplication)
		g.DELETE(":id", deleteApplication)
		g.GET(":id/redirect_uris", listRedirectURIsForApplication)
		g.POST(":id/redirect_uris", createRedirectURIForApplication)
		g.DELETE(":id/redirect_uris", deleteRedirectURIForApplication)
	}
}

// listApplications godoc
// @Summary List applications
// @Description List third-party applications
// @Tags applications
// @Produce json
// @Success 200 {object} routes.ApplicationListResponse
// @Router /applications [get]
func listApplications(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to list applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	q := query.Use(db)

	// Check if the requester has APP_READ permission. If not,
	// only return applications owned by the authenticated user.
	hasAppReadPermission := false
	var authUser *model.User
	if ui, exists := c.Get("user"); exists {
		if su, ok := ui.(*model.User); ok && su != nil {
			authUser = su
			perms, _ := middleware.GetUserPermissions(su.ID, db)
			hasAppReadPermission = perms.HasPermission(constants.APP_READ)
		}
	}

	var apps []*model.Application
	var err error
	if hasAppReadPermission {
		apps, err = q.Application.Find()
	} else if authUser != nil {
		apps, err = q.Application.Where(query.Application.UserID.Eq(authUser.ID)).Find()
	} else {
		// unauthenticated and no APP_READ -> return empty list
		c.JSON(http.StatusOK, ApplicationListResponse{Data: []ApplicationDTO{}})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var out []ApplicationDTO
	for _, a := range apps {
		out = append(out, ApplicationDTO{
			ID:               a.ID,
			Name:             a.Name,
			Description:      ptrToString(a.Description),
			WebsiteURL:       ptrToString(a.WebsiteURL),
			PrivacyPolicyURL: ptrToString(a.PrivacyPolicyURL),
			UserID:           a.UserID,
			PublicClient:     a.PublicClient,
			CreatedAt:        a.CreatedAt,
			UpdatedAt:        a.UpdatedAt,
			DeletedAt:        utils.DeletedAtPtr(a.DeletedAt),
		})
	}
	if out == nil {
		out = []ApplicationDTO{}
	}
	c.JSON(http.StatusOK, ApplicationListResponse{Data: out})
}

// createApplication godoc
// @Summary Create an application
// @Description Register a new third-party application
// @Tags applications
// @Accept json
// @Produce json
// @Param app body routes.CreateApplicationRequest true "Create application"
// @Success 201 {object} routes.ApplicationDTO
// @Router /applications [post]
func createApplication(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to create applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}

	// Authentication
	_, exists := c.Get("user")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input CreateApplicationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Validate client secret requirement for confidential clients
	if !input.PublicClient && input.ClientSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_secret is required for confidential clients"})
		return
	}
	app := model.Application{
		ID:               ulid.Make().String(),
		Name:             input.Name,
		Description:      stringToPtr(input.Description),
		WebsiteURL:       stringToPtr(input.WebsiteURL),
		PrivacyPolicyURL: stringToPtr(input.PrivacyPolicyURL),
		ClientSecret:     input.ClientSecret,
		PublicClient:     input.PublicClient,
		UserID:           input.UserID,
	}
	q := query.Use(db)
	// If a user is present in the context (session), use it as the owner.
	if ui, exists := c.Get("user"); exists {
		if su, ok := ui.(*model.User); ok && su != nil {
			app.UserID = su.ID
		}
	}
	if err := q.Application.Create(&app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := ApplicationDTO{
		ID:               app.ID,
		Name:             app.Name,
		Description:      ptrToString(app.Description),
		WebsiteURL:       ptrToString(app.WebsiteURL),
		PrivacyPolicyURL: ptrToString(app.PrivacyPolicyURL),
		UserID:           app.UserID,
		PublicClient:     app.PublicClient,
		CreatedAt:        app.CreatedAt,
		UpdatedAt:        app.UpdatedAt,
		DeletedAt:        utils.DeletedAtPtr(app.DeletedAt),
	}
	c.JSON(http.StatusCreated, resp)
}

// getApplication godoc
// @Summary Get an application
// @Description Get a single application
// @Tags applications
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {object} routes.ApplicationDTO
// @Router /applications/{id} [get]
func getApplication(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to get applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)
	a, err := q.Application.Where(query.Application.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	resp := ApplicationDTO{
		ID:               a.ID,
		Name:             a.Name,
		Description:      ptrToString(a.Description),
		WebsiteURL:       ptrToString(a.WebsiteURL),
		PrivacyPolicyURL: ptrToString(a.PrivacyPolicyURL),
		UserID:           a.UserID,
		PublicClient:     a.PublicClient,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
		DeletedAt:        utils.DeletedAtPtr(a.DeletedAt),
	}
	c.JSON(http.StatusOK, resp)
}

// updateApplication godoc
// @Summary Update an application
// @Description Update application fields
// @Tags applications
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param app body routes.UpdateApplicationRequest true "Update application"
// @Success 200 {object} routes.ApplicationDTO
// @Router /applications/{id} [put]
func updateApplication(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to update applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	var input UpdateApplicationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if input.Name.Set {
		if input.Name.Value == nil {
			updates["name"] = nil
		} else {
			updates["name"] = *input.Name.Value
		}
	}
	if input.Description.Set {
		if input.Description.Value == nil {
			updates["description"] = nil
		} else {
			updates["description"] = *input.Description.Value
		}
	}
	if input.WebsiteURL.Set {
		if input.WebsiteURL.Value == nil {
			updates["website_url"] = nil
		} else {
			updates["website_url"] = *input.WebsiteURL.Value
		}
	}
	if input.PrivacyPolicyURL.Set {
		if input.PrivacyPolicyURL.Value == nil {
			updates["privacy_policy_url"] = nil
		} else {
			updates["privacy_policy_url"] = *input.PrivacyPolicyURL.Value
		}
	}
	if input.ClientSecret.Set {
		if input.ClientSecret.Value == nil {
			updates["client_secret"] = nil
		} else {
			updates["client_secret"] = *input.ClientSecret.Value
		}
	}
	if input.PublicClient.Set {
		if input.PublicClient.Value == nil {
			updates["public_client"] = nil
		} else {
			updates["public_client"] = *input.PublicClient.Value
		}
	}
	q := query.Use(db)

	// fetch application to check ownership
	appModel, err := q.Application.Where(query.Application.ID.Eq(id)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// auth user required for owner check
	ui, exists := c.Get("user")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	authUser, ok := ui.(*model.User)
	if !ok || authUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Could not retrieve user information"})
		return
	}

	// allow if user has APP_UPDATE permission or is owner
	perms, _ := middleware.GetUserPermissions(authUser.ID, db)
	if !perms.HasPermission(constants.APP_UPDATE) && appModel.UserID != authUser.ID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have permission to perform this operation"})
		return
	}

	if len(updates) > 0 {
		if _, err := q.Application.Where(query.Application.ID.Eq(id)).Updates(updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	updated, _ := q.Application.Where(query.Application.ID.Eq(id)).First()
	resp := ApplicationDTO{
		ID:               updated.ID,
		Name:             updated.Name,
		Description:      ptrToString(updated.Description),
		WebsiteURL:       ptrToString(updated.WebsiteURL),
		PrivacyPolicyURL: ptrToString(updated.PrivacyPolicyURL),
		UserID:           updated.UserID,
		PublicClient:     updated.PublicClient,
		CreatedAt:        updated.CreatedAt,
		UpdatedAt:        updated.UpdatedAt,
		DeletedAt:        utils.DeletedAtPtr(updated.DeletedAt),
	}
	c.JSON(http.StatusOK, resp)
}

// patchApplication godoc
// @Summary Partially update an application
// @Description パッチ更新。指定されたフィールドのみ更新（null を送ると NULL に）
// @Tags applications
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param app body routes.PatchApplicationRequest true "Patch application"
// @Success 200 {object} routes.ApplicationDTO
// @Router /applications/{id} [patch]
func patchApplication(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to list applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	var body PatchApplicationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if body.Name.Set {
		if body.Name.Value == nil {
			updates["name"] = nil
		} else {
			updates["name"] = *body.Name.Value
		}
	}
	if body.Description.Set {
		if body.Description.Value == nil {
			updates["description"] = nil
		} else {
			updates["description"] = *body.Description.Value
		}
	}
	if body.WebsiteURL.Set {
		if body.WebsiteURL.Value == nil {
			updates["website_url"] = nil
		} else {
			updates["website_url"] = *body.WebsiteURL.Value
		}
	}
	if body.PrivacyPolicyURL.Set {
		if body.PrivacyPolicyURL.Value == nil {
			updates["privacy_policy_url"] = nil
		} else {
			updates["privacy_policy_url"] = *body.PrivacyPolicyURL.Value
		}
	}
	if body.ClientSecret.Set {
		if body.ClientSecret.Value == nil {
			updates["client_secret"] = nil
		} else {
			updates["client_secret"] = *body.ClientSecret.Value
		}
	}
	if body.PublicClient.Set {
		if body.PublicClient.Value == nil {
			updates["public_client"] = nil
		} else {
			updates["public_client"] = *body.PublicClient.Value
		}
	}

	q := query.Use(db)

	// fetch application to check ownership
	appModel, err := q.Application.Where(query.Application.ID.Eq(id)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// auth user required for owner check
	ui, exists := c.Get("user")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}
	authUser, ok := ui.(*model.User)
	if !ok || authUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報が取得できませんでした"})
		return
	}

	// allow if user has APP_UPDATE permission or is owner
	perms, _ := middleware.GetUserPermissions(authUser.ID, db)
	if !perms.HasPermission(constants.APP_UPDATE) && appModel.UserID != authUser.ID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "この操作を実行する権限がありません"})
		return
	}

	if len(updates) > 0 {
		if _, err := q.Application.Where(query.Application.ID.Eq(id)).Updates(updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	updated, err := q.Application.Where(query.Application.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := ApplicationDTO{
		ID:               updated.ID,
		Name:             updated.Name,
		Description:      ptrToString(updated.Description),
		WebsiteURL:       ptrToString(updated.WebsiteURL),
		PrivacyPolicyURL: ptrToString(updated.PrivacyPolicyURL),
		UserID:           updated.UserID,
		PublicClient:     updated.PublicClient,
		CreatedAt:        updated.CreatedAt,
		UpdatedAt:        updated.UpdatedAt,
		DeletedAt:        utils.DeletedAtPtr(updated.DeletedAt),
	}
	c.JSON(http.StatusOK, resp)
}

func deleteApplication(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to list applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)

	// fetch application to check ownership
	appModel, err := q.Application.Where(query.Application.ID.Eq(id)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// auth user required for owner check
	ui, exists := c.Get("user")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
		return
	}
	authUser, ok := ui.(*model.User)
	if !ok || authUser == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報が取得できませんでした"})
		return
	}

	// allow if user has APP_DELETE permission or is owner
	perms, _ := middleware.GetUserPermissions(authUser.ID, db)
	if !perms.HasPermission(constants.APP_DELETE) && appModel.UserID != authUser.ID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "この操作を実行する権限がありません"})
		return
	}

	if _, err := q.Application.Delete(&model.Application{ID: id}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// listRedirectURIsForApplication godoc
// @Summary List redirect URIs for an application
// @Description Get redirect URIs registered for a given application
// @Tags applications
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {object} RedirectURIListResponse
// @Router /applications/{id}/redirect_uris [get]
func listRedirectURIsForApplication(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to list applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)
	results, err := q.RedirectURI.Where(q.RedirectURI.ApplicationID.Eq(id), q.RedirectURI.DeletedAt.IsNull()).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]RedirectURIDTO, len(results))
	for i, r := range results {
		response[i] = RedirectURIDTO{
			URI:       r.URI,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}
	c.JSON(http.StatusOK, RedirectURIListResponse{Data: response})
}

// createRedirectURIForApplication godoc
// @Summary Create redirect URI for an application
// @Description Register a new redirect URI for the application
// @Tags applications
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param body body CreateRedirectURIRequest true "Create redirect URI"
// @Success 201 {object} RedirectURIDTO
// @Router /applications/{id}/redirect_uris [post]
func createRedirectURIForApplication(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to list applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	var body CreateRedirectURIRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.URI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uri required"})
		return
	}
	q := query.Use(db)
	// 重複チェック: 同じURIが既に登録されている場合は409返す
	if existing, err := q.RedirectURI.Where(q.RedirectURI.ApplicationID.Eq(id), q.RedirectURI.URI.Eq(body.URI), q.RedirectURI.DeletedAt.IsNull()).First(); err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "redirect uri already exists"})
		return
	} else if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	r := &model.RedirectURI{ApplicationID: id, URI: body.URI}
	if err := q.RedirectURI.Create(r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := RedirectURIDTO{
		URI:       r.URI,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	c.JSON(http.StatusCreated, response)
}

// deleteRedirectURIForApplication godoc
// @Summary Delete redirect URI for an application
// @Description Delete a registered redirect URI by application id and uri (query param)
// @Tags applications
// @Produce json
// @Param id path string true "Application ID"
// @Param uri query string true "Redirect URI"
// @Success 200 {object} map[string]string
// @Router /applications/{id}/redirect_uris [delete]
func deleteRedirectURIForApplication(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to list applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	uri := c.Query("uri")
	if uri == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uri required"})
		return
	}
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application id required"})
		return
	}
	q := query.Use(db)
	r, err := q.RedirectURI.Where(q.RedirectURI.ApplicationID.Eq(id), q.RedirectURI.URI.Eq(uri), q.RedirectURI.DeletedAt.IsNull()).First()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if r == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "redirect uri not found"})
		return
	}
	if _, err := q.RedirectURI.Where(q.RedirectURI.ApplicationID.Eq(r.ApplicationID), q.RedirectURI.URI.Eq(r.URI), q.RedirectURI.DeletedAt.IsNull()).Delete(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
