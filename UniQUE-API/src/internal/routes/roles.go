package routes

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/constants"
	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/UniPro-tech/UniQUE-API/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

func RegisterRoleRoutes(r *gin.Engine) {
	g := r.Group("/roles")
	{
		// ロール一覧・詳細の取得はROLE_MANAGE権限が必要
		g.GET("", middleware.RequirePermission(constants.ROLE_MANAGE), listRoles)
		g.GET(":id", middleware.RequirePermission(constants.ROLE_MANAGE), getRole)
		g.GET(":id/users", middleware.RequirePermission(constants.ROLE_MANAGE), listUsersForRole)

		// ロールの作成・更新・削除はROLE_MANAGE権限が必要
		g.POST("", middleware.RequirePermission(constants.ROLE_MANAGE), createRole)
		g.PUT(":id", middleware.RequirePermission(constants.ROLE_MANAGE), updateRole)
		g.DELETE(":id", middleware.RequirePermission(constants.ROLE_MANAGE), deleteRole)
		g.PATCH(":id", middleware.RequirePermission(constants.ROLE_MANAGE), patchRole)
		g.POST(":id/assign_all", middleware.RequirePermission(constants.ROLE_MANAGE), assignRoleToAll)
		g.PUT(":id/users/:user_id", middleware.RequirePermission(constants.ROLE_MANAGE), addUserToRole)
		g.DELETE(":id/users/:user_id", middleware.RequirePermission(constants.ROLE_MANAGE), removeUserFromRole)
	}
}

// assignRoleToAll godoc
// @Summary Assign role to all existing users
// @Description Assign the role to all users with status active or established
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} map[string]interface{}
// @Router /roles/{id}/assign_all [post]
func assignRoleToAll(c *gin.Context) {
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
	// ensure role exists
	if _, err := q.Role.Where(query.Role.ID.Eq(id)).First(); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	users, err := q.User.Where(query.User.Status.In("active", "established")).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	assigned := 0
	for _, usr := range users {
		// skip if already assigned
		if _, ferr := q.UserRole.Where(query.UserRole.UserID.Eq(usr.ID), query.UserRole.RoleID.Eq(id)).First(); ferr == nil {
			continue
		} else if ferr != gorm.ErrRecordNotFound {
			log.Printf("failed checking existing user_role for user %s role %s: %v", usr.ID, id, ferr)
			continue
		}
		ur := &model.UserRole{UserID: usr.ID, RoleID: id}
		if cerr := q.UserRole.Create(ur); cerr != nil {
			log.Printf("failed to assign role %s to user %s: %v", id, usr.ID, cerr)
			continue
		}
		assigned++
	}
	c.JSON(http.StatusOK, gin.H{"assigned": assigned})
}

// listUsersForRole godoc
// @Summary List users for a role
// @Description Get users assigned to a role
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} routes.UserListResponse
// @Router /roles/{id}/users [get]
func listUsersForRole(c *gin.Context) {
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
	// Use JOIN to fetch users with optional profile in a single query.
	type row struct {
		ID                string     `json:"id"`
		CustomID          string     `json:"custom_id"`
		Email             string     `json:"email"`
		ExternalEmail     string     `json:"external_email"`
		EmailVerified     bool       `json:"email_verified"`
		AffiliationPeriod string     `json:"affiliation_period"`
		Status            string     `json:"status"`
		CreatedAt         time.Time  `json:"created_at"`
		UpdatedAt         time.Time  `json:"updated_at"`
		DisplayName       *string    `json:"display_name"`
		Bio               *string    `json:"bio"`
		WebsiteURL        *string    `json:"website_url"`
		JoinedAt          *time.Time `json:"joined_at"`
	}

	// Batch fetch via user_roles -> collect user IDs -> IN queries to avoid N+1
	urs, err := q.UserRole.Where(query.UserRole.RoleID.Eq(id)).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(urs) == 0 {
		c.JSON(http.StatusOK, UserListResponse{Data: []UserDTO{}})
		return
	}
	ids := make([]string, 0, len(urs))
	for _, ur := range urs {
		ids = append(ids, ur.UserID)
	}
	users, err := q.User.Where(query.User.ID.In(ids...)).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	profiles, _ := q.Profile.Where(query.Profile.UserID.In(ids...)).Find()
	profileMap := make(map[string]*model.Profile)
	for _, p := range profiles {
		profileMap[p.UserID] = p
	}
	var out []UserDTO
	for _, u := range users {
		dto := UserDTO{
			ID:                u.ID,
			CustomID:          u.CustomID,
			Email:             u.Email,
			ExternalEmail:     u.ExternalEmail,
			EmailVerified:     u.EmailVerified,
			AffiliationPeriod: ptrToString(u.AffiliationPeriod),
			Status:            u.Status,
			CreatedAt:         u.CreatedAt,
			UpdatedAt:         u.UpdatedAt,
		}
		if p, ok := profileMap[u.ID]; ok {
			dto.Profile = &ProfileDTO{
				UserID:           p.UserID,
				DisplayName:      p.DisplayName,
				Bio:              ptrToString(p.Bio),
				WebsiteURL:       ptrToString(p.WebsiteURL),
				TwitterHandle:    ptrToString(p.TwitterHandle),
				Birthdate:        formatDate(p.Birthdate),
				BirthdateVisible: &p.BirthdateVisible,
				JoinedAt:         formatDate(p.JoinedAt),
			}
		}
		out = append(out, dto)
	}
	c.JSON(http.StatusOK, UserListResponse{Data: out})
}

// listRoles godoc
// @Summary List roles
// @Description List roles
// @Tags roles
// @Produce json
// @Success 200 {object} routes.RoleListResponse
// @Router /roles [get]
func listRoles(c *gin.Context) {
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
	roles, err := q.Role.Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var out []RoleDTO
	for _, r := range roles {
		out = append(out, RoleDTO{
			ID:                r.ID,
			CustomID:          r.CustomID,
			Name:              r.Name,
			Description:       ptrToString(r.Description),
			PermissionBitmask: r.PermissionBitmask,
			IsDefault:         r.IsDefault,
			CreatedAt:         r.CreatedAt,
			UpdatedAt:         r.UpdatedAt,
			DeletedAt:         utils.DeletedAtPtr(r.DeletedAt),
		})
	}
	c.JSON(http.StatusOK, RoleListResponse{Data: out})
}

// createRole godoc
// @Summary Create a role
// @Description Create a new role
// @Tags roles
// @Accept json
// @Produce json
// @Param role body routes.CreateRoleRequest true "Create role"
// @Success 201 {object} routes.RoleDTO
// @Router /roles [post]
func createRole(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to list applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	var input CreateRoleRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := model.Role{
		ID:                ulid.Make().String(),
		CustomID:          input.CustomID,
		Name:              input.Name,
		Description:       stringToPtr(input.Description),
		PermissionBitmask: input.PermissionBitmask,
		IsDefault:         false,
	}
	if input.IsDefault != nil {
		role.IsDefault = *input.IsDefault
	}
	q := query.Use(db)
	if err := q.Role.Create(&role); err != nil {
		// MySQLの重複エラーをチェック
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// エラーメッセージから重複したキーを判定
			errMsg := mysqlErr.Message
			if strings.Contains(errMsg, "custom_id") || strings.Contains(errMsg, "name") {
				c.JSON(http.StatusConflict, gin.H{"error": "role already exists", "code": "R0002"})
				return
			}
			// その他の重複エラー
			c.JSON(http.StatusConflict, gin.H{"error": "duplicate entry", "code": "R0002"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If requested, assign the newly created role to all existing users with status active or established
	if input.AssignToExisting != nil && *input.AssignToExisting {
		users, uerr := q.User.Where(query.User.Status.In("active", "established")).Find()
		if uerr != nil {
			// Log and continue
			log.Printf("failed to fetch users for assign_to_existing: %v", uerr)
		} else {
			for _, usr := range users {
				// skip if already assigned
				if _, ferr := q.UserRole.Where(query.UserRole.UserID.Eq(usr.ID), query.UserRole.RoleID.Eq(role.ID)).First(); ferr == nil {
					continue
				} else if ferr != gorm.ErrRecordNotFound {
					log.Printf("failed checking existing user_role for user %s role %s: %v", usr.ID, role.ID, ferr)
					continue
				}
				ur := &model.UserRole{UserID: usr.ID, RoleID: role.ID}
				if cerr := q.UserRole.Create(ur); cerr != nil {
					log.Printf("failed to assign role %s to user %s: %v", role.ID, usr.ID, cerr)
				}
			}
		}
	}
	resp := RoleDTO{
		ID:                role.ID,
		CustomID:          role.CustomID,
		Name:              role.Name,
		Description:       ptrToString(role.Description),
		PermissionBitmask: role.PermissionBitmask,
		IsDefault:         role.IsDefault,
		CreatedAt:         role.CreatedAt,
		UpdatedAt:         role.UpdatedAt,
		DeletedAt:         utils.DeletedAtPtr(role.DeletedAt),
	}
	c.JSON(http.StatusCreated, resp)
}

// getRole godoc
// @Summary Get a role
// @Description Get a single role
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} routes.RoleDTO
// @Router /roles/{id} [get]
func getRole(c *gin.Context) {
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
	r, err := q.Role.Where(query.Role.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	resp := RoleDTO{
		ID:                r.ID,
		CustomID:          r.CustomID,
		Name:              r.Name,
		Description:       ptrToString(r.Description),
		PermissionBitmask: r.PermissionBitmask,
		IsDefault:         r.IsDefault,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
		DeletedAt:         utils.DeletedAtPtr(r.DeletedAt),
	}
	c.JSON(http.StatusOK, resp)
}

// updateRole godoc
// @Summary Update a role
// @Description Update role fields
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body routes.UpdateRoleRequest true "Update role"
// @Success 200 {object} routes.RoleDTO
// @Router /roles/{id} [put]
func updateRole(c *gin.Context) {
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
	var input UpdateRoleRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.PermissionBitmask != nil {
		updates["permission_bitmask"] = *input.PermissionBitmask
	}
	if input.IsDefault != nil {
		updates["is_default"] = *input.IsDefault
	}
	q := query.Use(db)
	if len(updates) > 0 {
		if _, err := q.Role.Where(query.Role.ID.Eq(id)).Updates(updates); err != nil {
			// MySQLの重複エラーをチェック
			if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
				c.JSON(http.StatusConflict, gin.H{"error": "role name already exists", "code": "R0002"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	updated, _ := q.Role.Where(query.Role.ID.Eq(id)).First()
	resp := RoleDTO{
		ID:                updated.ID,
		CustomID:          updated.CustomID,
		Name:              updated.Name,
		Description:       ptrToString(updated.Description),
		PermissionBitmask: updated.PermissionBitmask,
		IsDefault:         updated.IsDefault,
		CreatedAt:         updated.CreatedAt,
		UpdatedAt:         updated.UpdatedAt,
		DeletedAt:         utils.DeletedAtPtr(updated.DeletedAt),
	}
	c.JSON(http.StatusOK, resp)
}

// patchRole godoc
// @Summary Patch a role
// @Description Partially update role fields
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body routes.PatchRoleRequest true "Patch role"
// @Success 200 {object} routes.RoleDTO
// @Router /roles/{id} [patch]
func patchRole(c *gin.Context) {
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
	var input PatchRoleRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q := query.Use(db)
	role, err := q.Role.Where(query.Role.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if input.Name != nil {
		role.Name = *input.Name
	}
	if input.Description != nil {
		role.Description = stringToPtr(*input.Description)
	}
	if input.PermissionBitmask != nil {
		role.PermissionBitmask = *input.PermissionBitmask
	}
	if input.IsDefault != nil {
		role.IsDefault = *input.IsDefault
	}
	if err := q.Role.Save(role); err != nil {
		// MySQLの重複エラーをチェック
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			c.JSON(http.StatusConflict, gin.H{"error": "role name already exists", "code": "R0002"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := RoleDTO{
		ID:                role.ID,
		CustomID:          role.CustomID,
		Name:              role.Name,
		Description:       ptrToString(role.Description),
		PermissionBitmask: role.PermissionBitmask,
		IsDefault:         role.IsDefault,
		CreatedAt:         role.CreatedAt,
		UpdatedAt:         role.UpdatedAt,
		DeletedAt:         utils.DeletedAtPtr(role.DeletedAt),
	}
	c.JSON(http.StatusOK, resp)
}

// deleteRole godoc
// @Summary Delete a role
// @Description Soft delete a role by ID
// @Tags roles
// @Param id path string true "Role ID"
// @Success 204 "No Content"
// @Router /roles/{id} [delete]
func deleteRole(c *gin.Context) {
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
	if _, err := q.Role.Delete(&model.Role{ID: id}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// addUserToRole godoc
// @Summary Add a user to a role
// @Description Assign a role to a user
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param user_id path string true "User ID"
// @Success 204 "No Content"
// @Router /roles/{id}/users/{user_id} [put]
func addUserToRole(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to list applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	roleID := c.Param("id")
	userID := c.Param("user_id")
	q := query.Use(db)
	// Check if role exists
	if _, err := q.Role.Where(query.Role.ID.Eq(roleID)).First(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	// Check if user exists
	if _, err := q.User.Where(query.User.ID.Eq(userID)).First(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	// Check if already assigned
	if _, err := q.UserRole.Where(query.UserRole.UserID.Eq(userID), query.UserRole.RoleID.Eq(roleID)).First(); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already has this role"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Assign role to user
	ur := &model.UserRole{UserID: userID, RoleID: roleID}
	if err := q.UserRole.Create(ur); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// removeUserFromRole godoc
// @Summary Remove a user from a role
// @Description Unassign a role from a user
// @Tags roles
// @Param id path string true "Role ID"
// @Param user_id path string true "User ID"
// @Success 204 "No Content"
// @Router /roles/{id}/users/{user_id} [delete]
func removeUserFromRole(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to list applications with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	roleID := c.Param("id")
	userID := c.Param("user_id")
	q := query.Use(db)
	_, err := q.UserRole.Where(query.UserRole.UserID.Eq(userID), query.UserRole.RoleID.Eq(roleID)).Delete(&model.UserRole{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
