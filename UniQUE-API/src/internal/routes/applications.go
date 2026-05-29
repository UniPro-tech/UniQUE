package routes

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
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
		c.AbortWithError(http.StatusInternalServerError, err)
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
			TermsURL:         ptrToString(a.TermsURL),
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Validate client secret requirement for confidential clients
	if !input.PublicClient && input.ClientSecret == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "client_secret is required for confidential clients"})
		return
	}
	if input.WebsiteURL.Set && input.WebsiteURL.Value != nil {
		if err := validateExternalURL(*input.WebsiteURL.Value); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid website_url"})
			return
		}
	}
	if input.PrivacyPolicyURL.Set && input.PrivacyPolicyURL.Value != nil {
		if err := validateExternalURL(*input.PrivacyPolicyURL.Value); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid privacy_policy_url"})
			return
		}
	}
	if input.TermsURL.Set && input.TermsURL.Value != nil {
		if err := validateExternalURL(*input.TermsURL.Value); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid terms_url"})
			return
		}
	}
	// 文字数を200文字に制限
	if input.WebsiteURL.Value != nil && len(*input.WebsiteURL.Value) > 200 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "website_url must be 200 characters or less"})
		return
	}
	if input.PrivacyPolicyURL.Value != nil && len(*input.PrivacyPolicyURL.Value) > 200 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "privacy_policy_url must be 200 characters or less"})
		return
	}
	if input.TermsURL.Value != nil && len(*input.TermsURL.Value) > 200 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "terms_url must be 200 characters or less"})
		return
	}

	now := time.Now().UTC()
	app := model.Application{
		ID:               ulid.Make().String(),
		Name:             input.Name,
		Description:      input.Description.Value,
		WebsiteURL:       input.WebsiteURL.Value,
		PrivacyPolicyURL: input.PrivacyPolicyURL.Value,
		TermsURL:         input.TermsURL.Value,
		ClientSecret:     input.ClientSecret,
		PublicClient:     input.PublicClient,
		UserID:           input.UserID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// If a user is present in the context (session), use it as the owner.
	if ui, exists := c.Get("user"); exists {
		if su, ok := ui.(*model.User); ok && su != nil {
			app.UserID = su.ID
		}
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)
		if err := q.Application.Create(&app); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	resp := ApplicationDTO{
		ID:               app.ID,
		Name:             app.Name,
		Description:      ptrToString(app.Description),
		WebsiteURL:       ptrToString(app.WebsiteURL),
		PrivacyPolicyURL: ptrToString(app.PrivacyPolicyURL),
		TermsURL:         ptrToString(app.TermsURL),
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
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp := ApplicationDTO{
		ID:               a.ID,
		Name:             a.Name,
		Description:      ptrToString(a.Description),
		WebsiteURL:       ptrToString(a.WebsiteURL),
		PrivacyPolicyURL: ptrToString(a.PrivacyPolicyURL),
		TermsURL:         ptrToString(a.TermsURL),
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

	// 更新用モデルとSelectするフィールドの準備
	updates := model.Application{
		UpdatedAt: time.Now().UTC(),
	}
	selectColumns := []field.Expr{query.Application.UpdatedAt}

	if input.Name.Set {
		if input.Name.Value != nil {
			updates.Name = *input.Name.Value
		} else {
			updates.Name = ""
		}
		selectColumns = append(selectColumns, query.Application.Name)
	}
	if input.Description.Set {
		updates.Description = input.Description.Value
		selectColumns = append(selectColumns, query.Application.Description)
	}
	if input.WebsiteURL.Set {
		updates.WebsiteURL = input.WebsiteURL.Value
		selectColumns = append(selectColumns, query.Application.WebsiteURL)
	}
	if input.TermsURL.Set {
		updates.TermsURL = input.TermsURL.Value
		selectColumns = append(selectColumns, query.Application.TermsURL)
	}
	if input.PrivacyPolicyURL.Set {
		updates.PrivacyPolicyURL = input.PrivacyPolicyURL.Value
		selectColumns = append(selectColumns, query.Application.PrivacyPolicyURL)
	}
	if input.ClientSecret.Set {
		if input.ClientSecret.Value != nil {
			updates.ClientSecret = *input.ClientSecret.Value
		} else {
			updates.ClientSecret = ""
		}
		selectColumns = append(selectColumns, query.Application.ClientSecret)
	}
	if input.PublicClient.Set {
		if input.PublicClient.Value != nil {
			updates.PublicClient = *input.PublicClient.Value
		} else {
			updates.PublicClient = false
		}
		selectColumns = append(selectColumns, query.Application.PublicClient)
	}

	var resp ApplicationDTO
	var isNotFound, isUnauthorized, isForbidden bool
	var errMsg string

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)

		// fetch application to check ownership
		appModel, err := q.Application.Where(query.Application.ID.Eq(id)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				isNotFound = true
			}
			return err
		}

		// auth user required for owner check
		ui, exists := c.Get("user")
		if !exists {
			isUnauthorized = true
			errMsg = "Unauthorized"
			return errors.New(errMsg)
		}
		authUser, ok := ui.(*model.User)
		if !ok || authUser == nil {
			isUnauthorized = true
			errMsg = "Could not retrieve user information"
			return errors.New(errMsg)
		}

		// allow if user has APP_UPDATE permission or is owner
		perms, _ := middleware.GetUserPermissions(authUser.ID, tx)
		if !perms.HasPermission(constants.APP_UPDATE) && appModel.UserID != authUser.ID {
			isForbidden = true
			errMsg = "You do not have permission to perform this operation"
			return errors.New(errMsg)
		}

		if len(selectColumns) > 1 { // UpdatedAt以外に変更がある場合
			if _, err := q.Application.Where(query.Application.ID.Eq(id)).
				Select(selectColumns...).
				Updates(&updates); err != nil {
				return err
			}
		}

		updated, err := q.Application.Where(query.Application.ID.Eq(id)).First()
		if err != nil {
			return err
		}

		resp = ApplicationDTO{
			ID:               updated.ID,
			Name:             updated.Name,
			Description:      ptrToString(updated.Description),
			WebsiteURL:       ptrToString(updated.WebsiteURL),
			TermsURL:         ptrToString(updated.TermsURL),
			PrivacyPolicyURL: ptrToString(updated.PrivacyPolicyURL),
			UserID:           updated.UserID,
			PublicClient:     updated.PublicClient,
			CreatedAt:        updated.CreatedAt,
			UpdatedAt:        updated.UpdatedAt,
			DeletedAt:        utils.DeletedAtPtr(updated.DeletedAt),
		}
		return nil
	})

	if err != nil {
		if isNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else if isUnauthorized {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errMsg})
		} else if isForbidden {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
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

	updates := model.Application{
		UpdatedAt: time.Now().UTC(),
	}
	selectColumns := []field.Expr{query.Application.UpdatedAt}

	if body.Name.Set {
		if body.Name.Value != nil {
			updates.Name = *body.Name.Value
		} else {
			updates.Name = ""
		}
		selectColumns = append(selectColumns, query.Application.Name)
	}
	if body.Description.Set {
		updates.Description = body.Description.Value
		selectColumns = append(selectColumns, query.Application.Description)
	}
	if body.WebsiteURL.Set {
		updates.WebsiteURL = body.WebsiteURL.Value
		selectColumns = append(selectColumns, query.Application.WebsiteURL)
	}
	if body.TermsURL.Set {
		updates.TermsURL = body.TermsURL.Value
		selectColumns = append(selectColumns, query.Application.TermsURL)
	}
	if body.PrivacyPolicyURL.Set {
		updates.PrivacyPolicyURL = body.PrivacyPolicyURL.Value
		selectColumns = append(selectColumns, query.Application.PrivacyPolicyURL)
	}
	if body.ClientSecret.Set {
		if body.ClientSecret.Value != nil {
			updates.ClientSecret = *body.ClientSecret.Value
		} else {
			updates.ClientSecret = ""
		}
		selectColumns = append(selectColumns, query.Application.ClientSecret)
	}
	if body.PublicClient.Set {
		if body.PublicClient.Value != nil {
			updates.PublicClient = *body.PublicClient.Value
		} else {
			updates.PublicClient = false
		}
		selectColumns = append(selectColumns, query.Application.PublicClient)
	}

	var resp ApplicationDTO
	var isNotFound, isUnauthorized, isForbidden bool
	var errMsg string

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)

		// fetch application to check ownership
		appModel, err := q.Application.Where(query.Application.ID.Eq(id)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				isNotFound = true
			}
			return err
		}

		// auth user required for owner check
		ui, exists := c.Get("user")
		if !exists {
			isUnauthorized = true
			errMsg = "認証が必要です"
			return errors.New(errMsg)
		}
		authUser, ok := ui.(*model.User)
		if !ok || authUser == nil {
			isUnauthorized = true
			errMsg = "ユーザー情報が取得できませんでした"
			return errors.New(errMsg)
		}

		// allow if user has APP_UPDATE permission or is owner
		perms, _ := middleware.GetUserPermissions(authUser.ID, tx)
		if !perms.HasPermission(constants.APP_UPDATE) && appModel.UserID != authUser.ID {
			isForbidden = true
			errMsg = "この操作を実行する権限がありません"
			return errors.New(errMsg)
		}

		if len(selectColumns) > 1 { // UpdatedAt以外に変更がある場合
			if _, err := q.Application.Where(query.Application.ID.Eq(id)).
				Select(selectColumns...).
				Updates(&updates); err != nil {
				return err
			}
		}

		updated, err := q.Application.Where(query.Application.ID.Eq(id)).First()
		if err != nil {
			return err
		}

		resp = ApplicationDTO{
			ID:               updated.ID,
			Name:             updated.Name,
			Description:      ptrToString(updated.Description),
			WebsiteURL:       ptrToString(updated.WebsiteURL),
			TermsURL:         ptrToString(updated.TermsURL),
			PrivacyPolicyURL: ptrToString(updated.PrivacyPolicyURL),
			UserID:           updated.UserID,
			PublicClient:     updated.PublicClient,
			CreatedAt:        updated.CreatedAt,
			UpdatedAt:        updated.UpdatedAt,
			DeletedAt:        utils.DeletedAtPtr(updated.DeletedAt),
		}
		return nil
	})

	if err != nil {
		if isNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else if isUnauthorized {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errMsg})
		} else if isForbidden {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
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

	var isNotFound, isUnauthorized, isForbidden bool
	var errMsg string

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)

		// fetch application to check ownership
		appModel, err := q.Application.Where(query.Application.ID.Eq(id)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				isNotFound = true
			}
			return err
		}

		// auth user required for owner check
		ui, exists := c.Get("user")
		if !exists {
			isUnauthorized = true
			errMsg = "認証が必要です"
			return errors.New(errMsg)
		}
		authUser, ok := ui.(*model.User)
		if !ok || authUser == nil {
			isUnauthorized = true
			errMsg = "ユーザー情報が取得できませんでした"
			return errors.New(errMsg)
		}

		// allow if user has APP_DELETE permission or is owner
		perms, _ := middleware.GetUserPermissions(authUser.ID, tx)
		if !perms.HasPermission(constants.APP_DELETE) && appModel.UserID != authUser.ID {
			isForbidden = true
			errMsg = "この操作を実行する権限がありません"
			return errors.New(errMsg)
		}

		if _, err := q.Application.Where(query.Application.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if isNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else if isUnauthorized {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errMsg})
		} else if isForbidden {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
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
		c.AbortWithError(http.StatusInternalServerError, err)
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

	var isConflict bool
	var response RedirectURIDTO

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)
		// 重複チェック
		existing, err := q.RedirectURI.Where(q.RedirectURI.ApplicationID.Eq(id), q.RedirectURI.URI.Eq(body.URI), q.RedirectURI.DeletedAt.IsNull()).First()
		if err == nil && existing != nil {
			isConflict = true
			return errors.New("redirect uri already exists")
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		now := time.Now().UTC()
		r := &model.RedirectURI{
			ApplicationID: id,
			URI:           body.URI,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := q.RedirectURI.Create(r); err != nil {
			return err
		}
		response = RedirectURIDTO{
			URI:       r.URI,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
		return nil
	})

	if err != nil {
		if isConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "redirect uri already exists"})
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
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

	var isNotFound bool

	err := db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)
		r, err := q.RedirectURI.Where(q.RedirectURI.ApplicationID.Eq(id), q.RedirectURI.URI.Eq(uri), q.RedirectURI.DeletedAt.IsNull()).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				isNotFound = true
			}
			return err
		}
		if r == nil {
			isNotFound = true
			return errors.New("redirect uri not found")
		}

		if _, err := q.RedirectURI.Where(q.RedirectURI.ApplicationID.Eq(r.ApplicationID), q.RedirectURI.URI.Eq(r.URI), q.RedirectURI.DeletedAt.IsNull()).Delete(); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if isNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "redirect uri not found"})
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func validateExternalURL(raw string) error {
	if raw == "" {
		return nil
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return err
	}
	s := strings.ToLower(u.Scheme)
	if s != "http" && s != "https" {
		return errors.New("invalid url scheme")
	}
	return nil
}
