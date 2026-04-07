package middleware

import (
	"net/http"
	"strings"

	"github.com/UniPro-tech/UniQUE-API/internal/constants"
	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserPermissions retrieves all permissions for a user based on their roles
func GetUserPermissions(userID string, db *gorm.DB) (constants.Permission, error) {
	q := query.Use(db)

	// ユーザーのロールを取得
	userRoles, err := q.UserRole.Where(q.UserRole.UserID.Eq(userID)).Find()
	if err != nil {
		return 0, err
	}

	if len(userRoles) == 0 {
		// ロールがない場合は権限なし
		return 0, nil
	}

	// ロールIDを収集
	roleIDs := make([]string, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}

	// ロール情報を取得
	roles, err := q.Role.Where(q.Role.ID.In(roleIDs...)).Find()
	if err != nil {
		return 0, err
	}

	// すべてのロールの権限をOR演算で合成
	var combinedPermissions constants.Permission = 0
	for _, role := range roles {
		combinedPermissions |= constants.Permission(role.PermissionBitmask)
	}

	return combinedPermissions, nil
}

// RequirePermission is a middleware that checks if the authenticated user has the required permission
func RequirePermission(required constants.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
			return
		}

		userModel, ok := user.(*model.User)
		if !ok || userModel == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報が取得できませんでした"})
			return
		}

		db, exists := c.Get("db")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "データベース接続エラー"})
			return
		}

		dbConn, ok := db.(*gorm.DB)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "データベース接続エラー"})
			return
		}

		// OAuth トークンの場合は権限ベースの操作は許可しない
		if isOAuthI, _ := c.Get("isOAuth"); isOAuthI != nil {
			if isOAuth, ok := isOAuthI.(bool); ok && isOAuth {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "OAuth トークンではこの操作は許可されていません"})
				return
			}
		}

		permissions, err := GetUserPermissions(userModel.ID, dbConn)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "権限の取得に失敗しました"})
			return
		}

		if !permissions.HasPermission(required) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "この操作を実行する権限がありません"})
			return
		}

		// 権限情報をコンテキストに保存（後続の処理で使用可能）
		c.Set("permissions", permissions)
		c.Next()
	}
}

// RequirePermissionOrSelf checks if the user has the required permission OR is accessing their own resource
func RequirePermissionOrSelf(required constants.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "認証が必要です"})
			return
		}

		userModel, ok := user.(*model.User)
		if !ok || userModel == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報が取得できませんでした"})
			return
		}

		// URLパラメータからターゲットユーザーIDを取得
		targetUserID := c.Param("id")
		if targetUserID == "" {
			targetUserID = c.Param("uid")
		}

		db, exists := c.Get("db")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
			return
		}

		dbConn, ok := db.(*gorm.DB)
		if !ok || dbConn == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
			return
		}

		// 自分自身のリソースへのアクセスの場合は許可（ただし OAuth トークンの場合は scope を確認）
		if targetUserID == userModel.ID {
			if isOAuthI, _ := c.Get("isOAuth"); isOAuthI != nil {
				if isOAuth, ok := isOAuthI.(bool); ok && isOAuth {
					// OAuth の場合は scope に email/profile のいずれかがあれば自身情報取得を許可
					scopeI, _ := c.Get("scope")
					scopeStr, _ := scopeI.(string)
					if strings.Contains(scopeStr, "email") || strings.Contains(scopeStr, "profile") {
						c.Next()
						return
					}
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "OAuth トークンに必要なスコープがありません"})
					return
				}
			}
			permissions, err := GetUserPermissions(userModel.ID, dbConn)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "権限の取得に失敗しました"})
				return
			}
			c.Set("permissions", permissions)
			c.Next()
			return
		}

		// 自分以外のリソースへのアクセスの場合は権限チェック
		// OAuth トークンは自分以外への権限を持たない
		if isOAuthI, _ := c.Get("isOAuth"); isOAuthI != nil {
			if isOAuth, ok := isOAuthI.(bool); ok && isOAuth {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "OAuth トークンではこの操作は許可されていません"})
				return
			}
		}

		permissions, err := GetUserPermissions(userModel.ID, dbConn)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "権限の取得に失敗しました"})
			return
		}

		if !permissions.HasPermission(required) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "この操作を実行する権限がありません"})
			return
		}

		// 権限情報をコンテキストに保存
		c.Set("permissions", permissions)
		c.Next()
	}
}

// CheckPermission is a helper function to check permission without middleware
func CheckPermission(userID string, required constants.Permission, db *gorm.DB) (bool, error) {
	permissions, err := GetUserPermissions(userID, db)
	if err != nil {
		return false, err
	}
	return permissions.HasPermission(required), nil
}

// GetPermissionsText returns human-readable permission list for a user
func GetPermissionsText(userID string, db *gorm.DB) ([]string, error) {
	permissions, err := GetUserPermissions(userID, db)
	if err != nil {
		return nil, err
	}
	return constants.GetPermissionsText(int64(permissions)), nil
}
