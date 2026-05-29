package routes

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/constants"
	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/UniPro-tech/UniQUE-API/internal/utils"
	discordutil "github.com/UniPro-tech/UniQUE-API/internal/utils/discord"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid/v2"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

func RegisterUserRoutes(r *gin.Engine) {
	// 公開ルート
	g := r.Group("/users")
	{
		// ユーザー一覧の取得は認証不要（基本情報のみ公開）
		g.GET("", listUsers)

		// ユーザー情報の取得は認証不要（基本情報のみ公開、詳細は自分自身のみ）
		g.GET(":id", getUser)

		// ユーザーのアプリ一覧は自分自身 OR APP_READ権限
		g.GET(":id/apps", middleware.RequirePermissionOrSelf(constants.APP_READ), listAppsForUser)

		// ロールの追加・削除はPERMISSION_MANAGE権限が必要
		g.POST(":id/roles", middleware.RequirePermission(constants.PERMISSION_MANAGE), addRoleForUser)
		g.DELETE(":id/roles/:roleId", middleware.RequirePermission(constants.PERMISSION_MANAGE), removeRoleForUser)

		// ロール一覧の取得は自分自身 OR PERMISSION_MANAGE権限
		g.GET(":id/roles", middleware.RequirePermissionOrSelf(constants.PERMISSION_MANAGE), listRolesForUser)

		// 権限一覧の取得は自分自身のみ
		g.GET(":id/permissions", middleware.RequirePermissionOrSelf(constants.USER_READ), getUserPermissions)

		// 外部ID連携の閲覧は自分自身 OR USER_READ権限(EXTERNAL_IDENTITY_READと同等)
		g.GET(":id/external_identities", middleware.RequirePermissionOrSelf(constants.USER_READ), listExternalIdentities)

		// 外部ID連携の追加は自分自身 OR USER_UPDATE権限(EXTERNAL_IDENTITY_WRITEと同等)
		g.POST(":id/external_identities", middleware.RequirePermissionOrSelf(constants.USER_UPDATE), addExternalIdentity)

		// 外部ID連携の削除は自分自身 OR USER_UPDATE権限(EXTERNAL_IDENTITY_DELETEと同等)
		g.DELETE(":id/external_identities/:eid", middleware.RequirePermissionOrSelf(constants.USER_UPDATE), removeExternalIdentity)

		// ユーザー情報の更新は自分自身 OR USER_UPDATE権限
		g.PUT(":id", middleware.RequirePermissionOrSelf(constants.USER_UPDATE), updateUser)
		g.PATCH(":id", middleware.RequirePermissionOrSelf(constants.USER_UPDATE), patchUser)

		// パスワード変更: 自分自身またはUSER_UPDATE権限
		g.PUT(":id/password/change", middleware.RequirePermissionOrSelf(constants.USER_UPDATE), changePassword)

		// ユーザーの削除はUSER_DELETE権限が必要
		g.DELETE(":id", middleware.RequirePermission(constants.USER_DELETE), deleteUser)

		// ユーザー登録の承認はフロントと同じくUSER_CREATE権限が必要
		g.POST(":id/approve", middleware.RequirePermission(constants.USER_CREATE), approveUserRegist)

		// ユーザー登録の却下もUSER_CREATE権限が必要（フロントと整合）
		g.POST(":id/reject", middleware.RequirePermission(constants.USER_CREATE), rejectUserRegist)

		// メール認証の再送は自分自身 OR USER_UPDATE権限
		g.POST(":id/resend_email_verification", middleware.RequirePermissionOrSelf(constants.USER_UPDATE), resendEmailVerification)
	}

	// 内部用ルート（作成系）
	ig := r.Group("/internal/users")
	{
		ig.POST("", createUser)
		ig.POST("email_verify/discord_link", linkDiscordByEmailCode)
		ig.POST("email_verify", emailCodeCheck)
	}
}

// parseDateFlexible attempts to parse date/time strings in multiple
// common layouts. Frontend sends "YYYY-MM-DD" only, but accept
// RFC3339 as well for robustness.
func parseDateFlexible(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	layouts := []string{time.DateOnly, "2006-01-02", time.RFC3339, time.RFC3339Nano}
	var parseErr error
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		} else {
			parseErr = err
		}
	}
	return time.Time{}, parseErr
}

func getDB(c *gin.Context) *gorm.DB {
	dbi, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return nil
	}
	return dbi
}

// getPendingEmail は認証待ちのメールアドレスを取得する
func getPendingEmail(userID string, q *query.Query) string {
	evc, err := q.EmailVerificationCode.Where(
		query.EmailVerificationCode.UserID.Eq(userID),
		query.EmailVerificationCode.RequestType.Eq("email_change"),
	).Order(query.EmailVerificationCode.CreatedAt.Desc()).First()
	if err != nil {
		return ""
	}
	return ptrToString(evc.NewEmail)
}

// listUsers godoc
// @Summary List users
// @Description List users with embedded profile. Returns all data if USER_READ permission, otherwise basic info only
// @Tags users
// @Produce json
// @Success 200 {object} routes.UserListResponse
// @Router /users [get]
func listUsers(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	q := query.Use(db)
	users, err := q.User.Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(users) == 0 {
		c.JSON(http.StatusOK, UserListResponse{Data: []UserDTO{}})
		return
	}
	ids := make([]string, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	profiles, _ := q.Profile.Where(query.Profile.UserID.In(ids...)).Find()
	profileMap := make(map[string]*model.Profile)
	for _, p := range profiles {
		profileMap[p.UserID] = p
	}

	// USER_READ権限があるかチェック
	hasUserReadPermission := false
	if user, exists := c.Get("user"); exists {
		if userModel, ok := user.(*model.User); ok && userModel != nil {
			permissions, _ := middleware.GetUserPermissions(userModel.ID, db)
			log.Printf("permission:%s", string(rune(permissions)))
			hasUserReadPermission = permissions.HasPermission(constants.USER_READ)
		}
	} else {
		log.Printf("no authenticated user found in context")
	}

	var out []UserDTO
	for _, u := range users {
		dto := UserDTO{
			ID:                u.ID,
			CustomID:          u.CustomID,
			Email:             u.Email,
			EmailVerified:     u.EmailVerified,
			AffiliationPeriod: ptrToString(u.AffiliationPeriod),
			Status:            u.Status,
			CreatedAt:         u.CreatedAt,
			UpdatedAt:         u.UpdatedAt,
		}

		if p, ok := profileMap[u.ID]; ok {
			profileDTO := &ProfileDTO{
				UserID:           p.UserID,
				DisplayName:      p.DisplayName,
				Bio:              ptrToString(p.Bio),
				WebsiteURL:       ptrToString(p.WebsiteURL),
				TwitterHandle:    ptrToString(p.TwitterHandle),
				JoinedAt:         formatDate(p.JoinedAt),
				BirthdateVisible: &p.BirthdateVisible,
				IsAdult:          isAdult(p.Birthdate),
			}
			// USER_READ権限があればExternalEmailとIsTOTPEnabledを返す
			if hasUserReadPermission {
				dto.ExternalEmail = u.ExternalEmail
				dto.IsTOTPEnabled = u.IsTotpEnabled
			}
			// birthdateはUSER_READ権限があるか、birthdateVisibleがtrueの場合のみ返す
			if hasUserReadPermission || p.BirthdateVisible {
				profileDTO.Birthdate = formatDate(p.Birthdate)
			}
			dto.Profile = profileDTO
		}
		out = append(out, dto)
	}
	c.JSON(http.StatusOK, UserListResponse{Data: out})
}

// createUser godoc
// @Summary Create a user
// @Description Create a new user with optional profile
// @Tags users
// @Accept json
// @Produce json
// @Param user body routes.CreateUserRequest true "Create user"
// @Success 201 {object} routes.UserDTO
// @Router /internal/users [post]
func createUser(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	config := c.MustGet("config").(config.Config)
	var input CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// カスタムIDの検証
	if !utils.IsValidCustomID(input.CustomID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid custom_id format"})
		return
	}
	// auth-server/internal/password_hash
	req := map[string]string{
		"password": input.Password,
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp, err := http.Post(config.IssuerInternalURL+"/internal/password_hash", "application/json", strings.NewReader(string(reqBody)))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	defer resp.Body.Close()
	var respData struct {
		PasswordHash string `json:"password_hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse password hash response"})
		return
	}

	status := "established"
	if input.Status != "" {
		status = input.Status
	}

	now := time.Now().UTC()
	user := model.User{
		ID:                ulid.Make().String(),
		CustomID:          input.CustomID,
		Email:             input.Email,
		PasswordHash:      respData.PasswordHash,
		ExternalEmail:     input.ExternalEmail,
		Status:            status,
		AffiliationPeriod: stringToPtr(input.AffiliationPeriod),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	var isConflict bool
	var isCustomIDConflict bool
	var isEmailConflict bool
	var profileDTO *ProfileDTO

	err = db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)
		if err := q.User.Create(&user); err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
				isConflict = true
				if strings.Contains(mysqlErr.Message, "custom_id") {
					isCustomIDConflict = true
				} else if strings.Contains(mysqlErr.Message, "email") {
					isEmailConflict = true
				}
			}
			return err
		}

		// 新規作成ユーザーに対して is_default=true のロールを付与
		if defaultRoles, derr := q.Role.Where(query.Role.IsDefault.Is(true)).Find(); derr == nil {
			for _, dr := range defaultRoles {
				ur := &model.UserRole{
					UserID:    user.ID,
					RoleID:    dr.ID,
					CreatedAt: now,
					UpdatedAt: now,
				}
				if err := q.UserRole.Create(ur); err != nil {
					log.Printf("failed to assign default role %s to user %s: %v", dr.ID, user.ID, err)
				}
			}
		} else {
			log.Printf("failed to fetch default roles: %v", derr)
		}

		err = sendRegistrationEmailVerification(user.ID, user.ExternalEmail, "", q, &config)
		if err != nil {
			return err
		}

		if input.Profile != nil {
			p := &model.Profile{
				UserID:      user.ID,
				DisplayName: input.Profile.DisplayName,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if input.Profile.Bio != "" {
				p.Bio = &input.Profile.Bio
			}
			if input.Profile.WebsiteURL != "" {
				p.WebsiteURL = &input.Profile.WebsiteURL
			}
			if err := q.Profile.Create(p); err != nil {
				return err
			}
			profileDTO = &ProfileDTO{
				UserID:           p.UserID,
				DisplayName:      p.DisplayName,
				Bio:              ptrToString(p.Bio),
				WebsiteURL:       ptrToString(p.WebsiteURL),
				TwitterHandle:    ptrToString(p.TwitterHandle),
				Birthdate:        formatDate(p.Birthdate),
				BirthdateVisible: &p.BirthdateVisible,
				JoinedAt:         formatDate(p.JoinedAt),
				IsAdult:          isAdult(p.Birthdate),
			}
		}
		return nil
	})

	if err != nil {
		if isConflict {
			if isCustomIDConflict {
				c.JSON(http.StatusConflict, gin.H{"error": "username already exists", "code": "R0006"})
			} else if isEmailConflict {
				c.JSON(http.StatusConflict, gin.H{"error": "email already exists", "code": "R0007"})
			} else {
				c.JSON(http.StatusConflict, gin.H{"error": "duplicate entry", "code": "R0002"})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	dbResp := UserDTO{
		ID:                user.ID,
		CustomID:          user.CustomID,
		Email:             user.Email,
		ExternalEmail:     user.ExternalEmail,
		EmailVerified:     user.EmailVerified,
		AffiliationPeriod: ptrToString(user.AffiliationPeriod),
		Status:            user.Status,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
		IsTOTPEnabled:     user.IsTotpEnabled,
		Profile:           profileDTO,
	}
	c.JSON(http.StatusCreated, dbResp)
}

// getUser godoc
// @Summary Get a user
// @Description Get a single user with optional profile
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} routes.UserDTO
// @Router /users/{id} [get]
func getUser(c *gin.Context) {
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")

	// 認証ユーザーを取得して自分かどうかを判定
	user, exists := c.Get("user")
	isSelf := false
	hasUserReadPermission := false
	isOAuth := IsOAuth(c)
	scopeStr := GetScope(c)
	if exists {
		userModel, ok := user.(*model.User)
		if ok && userModel != nil {
			if userModel.ID == id {
				isSelf = true
			}
			// OAuth トークンの場合は scope に基づく自己情報取得のみ許可
			if !isOAuth {
				permissions, _ := middleware.GetUserPermissions(userModel.ID, db)
				hasUserReadPermission = permissions.HasPermission(constants.USER_READ)
			} else {
				// OAuthかつisSelf=falseの場合は情報取得不可
				if !isSelf {
					c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to access other users' information with an access token"})
					return
				}
			}
		}
	}

	q := query.Use(db)
	u, err := q.User.Where(query.User.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// 基本情報は常に返す
	dto := UserDTO{
		ID:                u.ID,
		CustomID:          u.CustomID,
		EmailVerified:     u.EmailVerified,
		AffiliationPeriod: ptrToString(u.AffiliationPeriod),
		Status:            u.Status,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}

	// 自分自身またはUSER_READ権限がある場合はセンシティブ情報を含める
	sensitiveAllowed := false
	if isOAuth {
		// OAuth の場合は自身かつ scope に email が含まれている必要がある
		if isSelf && strings.Contains(scopeStr, "email") {
			sensitiveAllowed = true
		}
	} else {
		if isSelf || hasUserReadPermission {
			sensitiveAllowed = true
		}
	}
	if sensitiveAllowed {
		dto.Email = u.Email
		dto.ExternalEmail = u.ExternalEmail
		dto.IsTOTPEnabled = u.IsTotpEnabled
	}

	// PendingEmailは自分自身のみ（OAuth の場合は scope に email が必要）
	if isSelf {
		if isOAuth {
			if strings.Contains(scopeStr, "email") {
				dto.PendingEmail = getPendingEmail(u.ID, q)
			}
		} else {
			dto.PendingEmail = getPendingEmail(u.ID, q)
		}
	}

	// プロフィールは常に返す（birthdateはbirthdateVisibleで制御）
	if p, err := q.Profile.Where(query.Profile.UserID.Eq(u.ID)).First(); err == nil {
		profileDTO := &ProfileDTO{
			UserID:           p.UserID,
			DisplayName:      p.DisplayName,
			Bio:              ptrToString(p.Bio),
			WebsiteURL:       ptrToString(p.WebsiteURL),
			TwitterHandle:    ptrToString(p.TwitterHandle),
			JoinedAt:         formatDate(p.JoinedAt),
			BirthdateVisible: &p.BirthdateVisible,
			IsAdult:          isAdult(p.Birthdate),
		}
		// birthdateはUSER_READ権限があるか、自分自身、またはbirthdateVisible、OAuth で scope に profile があれば返す
		if hasUserReadPermission || isSelf || p.BirthdateVisible || (isOAuth && strings.Contains(scopeStr, "profile")) {
			profileDTO.Birthdate = formatDate(p.Birthdate)
		}
		dto.Profile = profileDTO
	}
	c.JSON(http.StatusOK, dto)
}

// updateUser godoc
// @Summary Update a user
// @Description Update a user's fields and optional profile
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body routes.UpdateUserRequest true "Update user"
// @Success 200 {object} routes.UserDTO
// @Router /users/{id} [put]
func updateUser(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)
	user, err := q.User.Where(query.User.ID.Eq(id)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var input UpdateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// カスタムIDの検証
	if input.CustomID != nil && !utils.IsValidCustomID(*input.CustomID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid custom_id format"})
		return
	}

	// トランザクションの外側で、あらかじめパースエラーが発生し得る処理を検証・実施
	var birthdatePtr *time.Time
	if input.Profile != nil && input.Profile.Birthdate.Set {
		if input.Profile.Birthdate.Value != nil && *input.Profile.Birthdate.Value != "" {
			t, err := time.Parse("2006-01-02", *input.Profile.Birthdate.Value)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birthdate format, expected YYYY-MM-DD"})
				return
			}
			birthdatePtr = &t
		}
	}

	var joinedAtPtr *time.Time
	if input.Profile != nil && input.Profile.JoinedAt.Set {
		if input.Profile.JoinedAt.Value != nil && *input.Profile.JoinedAt.Value != "" {
			t, err := parseDateFlexible(*input.Profile.JoinedAt.Value)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid joined_at format"})
				return
			}
			joinedAtPtr = &t
		}
	}

	now := time.Now().UTC()

	// ユーザー更新用モデルの構築（構造体 + Select方式）
	updatesUser := model.User{
		UpdatedAt: now,
	}
	selectUserColumns := []field.Expr{query.User.UpdatedAt}

	if input.Email != nil && *input.Email != user.Email {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permissions not found"})
			return
		}
		perms, ok := permissions.(constants.Permission)
		if !ok || !perms.HasPermission(constants.USER_UPDATE) {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "email change not implemented"})
			return
		}
		updatesUser.Email = *input.Email
		selectUserColumns = append(selectUserColumns, query.User.Email)
	}

	if input.ExternalEmail != nil && *input.ExternalEmail != user.ExternalEmail {
		updatesUser.EmailVerified = false
		selectUserColumns = append(selectUserColumns, query.User.EmailVerified)
	}
	if input.AffiliationPeriod != nil {
		updatesUser.AffiliationPeriod = input.AffiliationPeriod // pointer copy
		selectUserColumns = append(selectUserColumns, query.User.AffiliationPeriod)
	}
	if input.Status != nil {
		updatesUser.Status = *input.Status
		selectUserColumns = append(selectUserColumns, query.User.Status)
	}

	// プロフィール更新用モデルの構築（構造体 + Select方式）
	updatesProfile := model.Profile{
		UpdatedAt: now,
	}
	selectProfileColumns := []field.Expr{query.Profile.UpdatedAt}

	if input.Profile != nil {
		if input.Profile.DisplayName.Set {
			if input.Profile.DisplayName.Value != nil {
				updatesProfile.DisplayName = *input.Profile.DisplayName.Value
			} else {
				updatesProfile.DisplayName = ""
			}
			selectProfileColumns = append(selectProfileColumns, query.Profile.DisplayName)
		}
		if input.Profile.Bio.Set {
			updatesProfile.Bio = input.Profile.Bio.Value
			selectProfileColumns = append(selectProfileColumns, query.Profile.Bio)
		}
		if input.Profile.WebsiteURL.Set {
			updatesProfile.WebsiteURL = input.Profile.WebsiteURL.Value
			selectProfileColumns = append(selectProfileColumns, query.Profile.WebsiteURL)
		}
		if input.Profile.TwitterHandle.Set {
			updatesProfile.TwitterHandle = input.Profile.TwitterHandle.Value
			selectProfileColumns = append(selectProfileColumns, query.Profile.TwitterHandle)
		}
		if input.Profile.BirthdateVisible.Set {
			if input.Profile.BirthdateVisible.Value != nil {
				updatesProfile.BirthdateVisible = *input.Profile.BirthdateVisible.Value
			} else {
				updatesProfile.BirthdateVisible = false
			}
			selectProfileColumns = append(selectProfileColumns, query.Profile.BirthdateVisible)
		}
		if input.Profile.Birthdate.Set {
			updatesProfile.Birthdate = birthdatePtr
			selectProfileColumns = append(selectProfileColumns, query.Profile.Birthdate)
		}
		if input.Profile.JoinedAt.Set {
			updatesProfile.JoinedAt = joinedAtPtr
			selectProfileColumns = append(selectProfileColumns, query.Profile.JoinedAt)
		}
	}

	var updated *model.User

	err = db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)

		if input.ExternalEmail != nil && *input.ExternalEmail != user.ExternalEmail {
			// 既存の未使用コードを削除
			_, err := q.EmailVerificationCode.Where(
				query.EmailVerificationCode.UserID.Eq(id),
				query.EmailVerificationCode.RequestType.Eq("email_change"),
			).Delete()
			if err != nil {
				return err
			}
			// 認証コードの生成と送信
			if err := sendEmailChangeVerification(id, *input.ExternalEmail, "", q, config.LoadConfig()); err != nil {
				return err
			}
		}

		// Userの更新
		if len(selectUserColumns) > 1 {
			if _, err := q.User.Where(query.User.ID.Eq(id)).Select(selectUserColumns...).Updates(&updatesUser); err != nil {
				return err
			}
		}

		// statusが変更された場合
		// もしstatusがsuspend / archiveならsessionまるまる消す
		if updatesUser.Status == "suspended" || updatesUser.Status == "archived" {
			if _, err := q.Session.Where(query.Session.UserID.Eq(id)).Delete(); err != nil {
				return err
			}
		}

		// Profileの処理
		if input.Profile != nil {
			existing, err := q.Profile.Where(query.Profile.UserID.Eq(user.ID)).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			if err == gorm.ErrRecordNotFound || existing == nil {
				// 新規作成
				newProfile := &model.Profile{
					UserID:    user.ID,
					CreatedAt: now,
					UpdatedAt: now,
				}
				if input.Profile.DisplayName.Set && input.Profile.DisplayName.Value != nil {
					newProfile.DisplayName = *input.Profile.DisplayName.Value
				}
				if input.Profile.Bio.Set {
					newProfile.Bio = input.Profile.Bio.Value
				}
				if input.Profile.WebsiteURL.Set {
					newProfile.WebsiteURL = input.Profile.WebsiteURL.Value
				}
				if input.Profile.TwitterHandle.Set {
					newProfile.TwitterHandle = input.Profile.TwitterHandle.Value
				}
				if input.Profile.BirthdateVisible.Set && input.Profile.BirthdateVisible.Value != nil {
					newProfile.BirthdateVisible = *input.Profile.BirthdateVisible.Value
				}
				if input.Profile.Birthdate.Set {
					newProfile.Birthdate = birthdatePtr
				}
				if input.Profile.JoinedAt.Set {
					newProfile.JoinedAt = joinedAtPtr
				} else {
					newProfile.JoinedAt = &now
				}

				if err := q.Profile.Create(newProfile); err != nil {
					return err
				}
			} else {
				// 構造体とSelect指定による安全な更新（NullフィールドのNULL更新もサポート）
				if len(selectProfileColumns) > 1 {
					if _, err := q.Profile.Where(query.Profile.UserID.Eq(user.ID)).Select(selectProfileColumns...).Updates(&updatesProfile); err != nil {
						return err
					}
				}
			}
		}

		// 最新データの再取得
		var errFirst error
		updated, errFirst = q.User.Where(query.User.ID.Eq(id)).First()
		return errFirst
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンスDTOのビルド
	dto := UserDTO{
		ID:                updated.ID,
		CustomID:          updated.CustomID,
		Email:             updated.Email,
		ExternalEmail:     updated.ExternalEmail,
		PendingEmail:      getPendingEmail(updated.ID, q),
		EmailVerified:     updated.EmailVerified,
		AffiliationPeriod: ptrToString(updated.AffiliationPeriod),
		Status:            updated.Status,
		CreatedAt:         updated.CreatedAt,
		UpdatedAt:         updated.UpdatedAt,
		IsTOTPEnabled:     updated.IsTotpEnabled,
	}
	if p, err := q.Profile.Where(query.Profile.UserID.Eq(updated.ID)).First(); err == nil {
		dto.Profile = &ProfileDTO{
			UserID:           p.UserID,
			DisplayName:      p.DisplayName,
			Bio:              ptrToString(p.Bio),
			WebsiteURL:       ptrToString(p.WebsiteURL),
			TwitterHandle:    ptrToString(p.TwitterHandle),
			Birthdate:        formatDate(p.Birthdate),
			BirthdateVisible: &p.BirthdateVisible,
			JoinedAt:         formatDate(p.JoinedAt),
			IsAdult:          isAdult(p.Birthdate),
		}
	}
	c.JSON(http.StatusOK, dto)
}

// patchUser godoc
// @Summary Partially update a user
// @Description パッチ更新。メールアドレスの変更は管理者のみ可能。プロフィールのフィールドは指定されたもののみ更新される（例: display_nameのみ更新など）
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body routes.PatchUserRequest true "Update user"
// @Success 200 {object} routes.UserDTO
// @Router /users/{id} [patch]
func patchUser(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)
	user, err := q.User.Where(query.User.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body PatchUserRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// トランザクション処理の開始
	err = q.Transaction(func(tx *query.Query) error {
		now := time.Now().UTC()

		// カスタムIDの検証
		if body.CustomID.Set {
			if body.CustomID.Value != nil {
				if !utils.IsValidCustomID(*body.CustomID.Value) {
					return fmt.Errorf("invalid_custom_id_format")
				}
				// custom_idを更新
				if _, err := tx.User.Where(tx.User.ID.Eq(id)).Updates(&model.User{
					CustomID:  *body.CustomID.Value,
					UpdatedAt: now,
				}); err != nil {
					return err
				}
			}
		}

		if body.Email.Set {
			// メールアドレスの変更は管理者のみ可能
			permissions, exists := c.Get("permissions")
			if !exists {
				return fmt.Errorf("permissions_not_found")
			}

			perms, ok := permissions.(constants.Permission)
			if !ok || !perms.HasPermission(constants.USER_UPDATE) {
				return fmt.Errorf("email_change_not_implemented")
			}

			// 管理者ならメールアドレスを更新（nullならNULLに、値ありなら更新）
			if body.Email.Value == nil {
				if _, err := tx.User.Where(tx.User.ID.Eq(id)).Updates(map[string]interface{}{
					"email":      nil,
					"updated_at": now,
				}); err != nil {
					return err
				}
			} else if *body.Email.Value != user.Email {
				if _, err := tx.User.Where(tx.User.ID.Eq(id)).Updates(&model.User{
					Email:     *body.Email.Value,
					UpdatedAt: now,
				}); err != nil {
					return err
				}
			}
		}

		if body.ExternalEmail.Set {
			// 既存の未使用コードを削除
			_, err = tx.EmailVerificationCode.Where(
				tx.EmailVerificationCode.UserID.Eq(id),
				tx.EmailVerificationCode.RequestType.Eq("email_change"),
			).Delete()
			if err != nil {
				return err
			}
			if body.ExternalEmail.Value == nil {
				// 明示的に null を送られた -> external_email を NULL に
				// ★ if を追加して修正
				if _, err = tx.User.Where(tx.User.ID.Eq(id)).Updates(map[string]interface{}{
					"external_email": nil,
					"email_verified": false,
					"updated_at":     now,
				}); err != nil {
					return err
				}
			} else if *body.ExternalEmail.Value != user.ExternalEmail {
				// external_emailは直接更新せず、認証コードのnew_emailに保存
				if err := sendEmailChangeVerification(id, *body.ExternalEmail.Value, "", tx, config.LoadConfig()); err != nil {
					return err
				}
				// ★ if を追加して修正
				if _, err = tx.User.Where(tx.User.ID.Eq(id)).Updates(map[string]interface{}{
					"email_verified": false,
					"updated_at":     now,
				}); err != nil {
					return err
				}
			}
		}

		if body.AffiliationPeriod.Set {
			if body.AffiliationPeriod.Value == nil {
				if _, err := tx.User.Where(tx.User.ID.Eq(id)).Updates(map[string]interface{}{
					"affiliation_period": nil,
					"updated_at":         now,
				}); err != nil {
					return err
				}
			} else {
				if _, err := tx.User.Where(tx.User.ID.Eq(id)).Updates(&model.User{
					AffiliationPeriod: body.AffiliationPeriod.Value,
					UpdatedAt:         now,
				}); err != nil {
					return err
				}
			}
		}

		if body.Status.Set {
			if body.Status.Value == nil {
				if _, err := tx.User.Where(tx.User.ID.Eq(id)).Updates(map[string]interface{}{
					"status":     nil,
					"updated_at": now,
				}); err != nil {
					return err
				}
			} else {
				if _, err := tx.User.Where(tx.User.ID.Eq(id)).Updates(&model.User{
					Status:    *body.Status.Value,
					UpdatedAt: now,
				}); err != nil {
					return err
				}
			}
			if *body.Status.Value == "suspended" || *body.Status.Value == "archived" {
				if _, err := tx.Session.Where(tx.Session.UserID.Eq(id)).Delete(); err != nil {
					return err
				}
			}
		}

		// プロフィールの更新
		if body.Profile != nil {
			profileUpdates := map[string]interface{}{}
			if body.Profile.DisplayName.Set {
				if body.Profile.DisplayName.Value == nil {
					profileUpdates["display_name"] = nil
				} else {
					profileUpdates["display_name"] = *body.Profile.DisplayName.Value
				}
			}
			if body.Profile.Bio.Set {
				if body.Profile.Bio.Value == nil {
					profileUpdates["bio"] = nil
				} else {
					profileUpdates["bio"] = *body.Profile.Bio.Value
				}
			}
			if body.Profile.WebsiteURL.Set {
				if body.Profile.WebsiteURL.Value == nil {
					profileUpdates["website_url"] = nil
				} else {
					profileUpdates["website_url"] = *body.Profile.WebsiteURL.Value
				}
			}
			if body.Profile.TwitterHandle.Set {
				if body.Profile.TwitterHandle.Value == nil {
					profileUpdates["twitter_handle"] = nil
				} else {
					profileUpdates["twitter_handle"] = *body.Profile.TwitterHandle.Value
				}
			}
			if body.Profile.BirthdateVisible.Set {
				if body.Profile.BirthdateVisible.Value == nil {
					profileUpdates["birthdate_visible"] = false
				} else {
					profileUpdates["birthdate_visible"] = *body.Profile.BirthdateVisible.Value
				}
			}
			if body.Profile.Birthdate.Set {
				if body.Profile.Birthdate.Value == nil || *body.Profile.Birthdate.Value == "" {
					profileUpdates["birthdate"] = nil
				} else {
					t, err := time.Parse("2006-01-02", *body.Profile.Birthdate.Value)
					if err != nil {
						return fmt.Errorf("invalid_birthdate_format")
					}
					profileUpdates["birthdate"] = t
				}
			}
			if body.Profile.JoinedAt.Set {
				if body.Profile.JoinedAt.Value == nil || *body.Profile.JoinedAt.Value == "" {
					profileUpdates["joined_at"] = nil
				} else {
					joinedAt, err := parseDateFlexible(*body.Profile.JoinedAt.Value)
					if err != nil {
						return fmt.Errorf("invalid_joined_at_format")
					}
					profileUpdates["joined_at"] = joinedAt
				}
			}
			existing, err := tx.Profile.Where(tx.Profile.UserID.Eq(user.ID)).First()
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}

			if err == gorm.ErrRecordNotFound || existing == nil {
				// 新規作成: 必須フィールドだけセット
				newProfile := &model.Profile{
					UserID:    user.ID,
					CreatedAt: now,
					UpdatedAt: now,
				}
				if v, ok := profileUpdates["display_name"]; ok {
					if v != nil {
						newProfile.DisplayName = v.(string)
					}
				}
				if v, ok := profileUpdates["bio"]; ok {
					if v == nil {
						newProfile.Bio = nil
					} else {
						bioStr := v.(string)
						newProfile.Bio = &bioStr
					}
				}
				if v, ok := profileUpdates["website_url"]; ok {
					if v == nil {
						newProfile.WebsiteURL = nil
					} else {
						urlStr := v.(string)
						newProfile.WebsiteURL = &urlStr
					}
				}
				if v, ok := profileUpdates["twitter_handle"]; ok {
					if v == nil {
						newProfile.TwitterHandle = nil
					} else {
						twitterStr := v.(string)
						newProfile.TwitterHandle = &twitterStr
					}
				}
				if v, ok := profileUpdates["birthdate_visible"]; ok {
					newProfile.BirthdateVisible = v.(bool)
				}
				if v, ok := profileUpdates["birthdate"]; ok {
					if v != nil {
						bdTime := v.(time.Time)
						newProfile.Birthdate = &bdTime
					}
				}
				if v, ok := profileUpdates["joined_at"]; ok {
					if v != nil {
						jaTime := v.(time.Time)
						newProfile.JoinedAt = &jaTime
					}
				} else {
					newProfile.JoinedAt = &now
				}
				if err := tx.Profile.Create(newProfile); err != nil {
					return err
				}
			} else if len(profileUpdates) > 0 {
				profileUpdates["updated_at"] = now
				if _, err := tx.Profile.Where(tx.Profile.UserID.Eq(user.ID)).Updates(profileUpdates); err != nil {
					return err
				}
			}
		}

		return nil
	})

	// トランザクション内のエラーハンドリング
	if err != nil {
		switch err.Error() {
		case "invalid_custom_id_format":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid custom_id format"})
		case "permissions_not_found":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permissions not found"})
		case "email_change_not_implemented":
			c.JSON(http.StatusNotImplemented, gin.H{"error": "email change not implemented"})
		case "invalid_birthdate_format":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birthdate format, expected YYYY-MM-DD"})
		case "invalid_joined_at_format":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid joined_at format"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// rebuild response dto
	updated, err := q.User.Where(query.User.ID.Eq(id)).First()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dto := UserDTO{
		ID:                updated.ID,
		CustomID:          updated.CustomID,
		Email:             updated.Email,
		ExternalEmail:     updated.ExternalEmail,
		PendingEmail:      getPendingEmail(updated.ID, q),
		EmailVerified:     updated.EmailVerified,
		AffiliationPeriod: ptrToString(updated.AffiliationPeriod),
		Status:            updated.Status,
		CreatedAt:         updated.CreatedAt,
		UpdatedAt:         updated.UpdatedAt,
		IsTOTPEnabled:     updated.IsTotpEnabled,
	}
	if p, err := q.Profile.Where(query.Profile.UserID.Eq(updated.ID)).First(); err == nil {
		dto.Profile = &ProfileDTO{
			UserID:           p.UserID,
			DisplayName:      p.DisplayName,
			Bio:              ptrToString(p.Bio),
			WebsiteURL:       ptrToString(p.WebsiteURL),
			TwitterHandle:    ptrToString(p.TwitterHandle),
			Birthdate:        formatDate(p.Birthdate),
			BirthdateVisible: &p.BirthdateVisible,
			IsAdult:          isAdult(p.Birthdate),
			JoinedAt:         formatDate(p.JoinedAt),
		}
	}
	c.JSON(http.StatusOK, dto)
}

// deleteUser godoc
// @Summary Delete user
// @Description Hard delete user data
// @Tags users
// @Param id path string true "User ID"
// @Success 204
// @Router /users/{id} [delete]
func deleteUser(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)
	if _, err := q.User.Delete(&model.User{ID: id}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// listAppsForUser godoc
// @Summary List applications for a user
// @Description Get applications owned by a user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} routes.ApplicationListResponse
// @Router /users/{id}/apps [get]
func listAppsForUser(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)
	apps, err := q.Application.Where(query.Application.UserID.Eq(id)).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(apps) == 0 {
		c.JSON(http.StatusOK, ApplicationListResponse{Data: []ApplicationDTO{}})
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
		})
	}
	c.JSON(http.StatusOK, ApplicationListResponse{Data: out})
}

// addRoleForUser godoc
// @Summary Assign a role to a user
// @Description Assign the specified role to the user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body routes.CreateUserRoleRequest true "Role assignment"
// @Success 201 {object} routes.RoleDTO
// @Router /users/{id}/roles [post]
func addRoleForUser(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	var input CreateUserRoleRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q := query.Use(db)
	if _, err := q.User.Where(query.User.ID.Eq(id)).First(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	role, err := q.Role.Where(query.Role.ID.Eq(input.RoleID)).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if _, err := q.UserRole.Where(query.UserRole.UserID.Eq(id), query.UserRole.RoleID.Eq(input.RoleID)).First(); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role already assigned"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	ur := &model.UserRole{
		UserID:    id,
		RoleID:    input.RoleID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := q.UserRole.Create(ur); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := RoleDTO{
		ID:                role.ID,
		CustomID:          role.CustomID,
		Name:              role.Name,
		Description:       ptrToString(role.Description),
		PermissionBitmask: role.PermissionBitmask,
	}
	c.JSON(http.StatusCreated, resp)
}

// removeRoleForUser godoc
// @Summary Remove a role from a user
// @Description Unassign the specified role from the user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Param roleId path string true "Role ID"
// @Success 204
// @Router /users/{id}/roles/{roleId} [delete]
func removeRoleForUser(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	roleId := c.Param("roleId")
	q := query.Use(db)
	if _, err := q.UserRole.Where(query.UserRole.UserID.Eq(id), query.UserRole.RoleID.Eq(roleId)).First(); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := q.UserRole.Delete(&model.UserRole{UserID: id, RoleID: roleId}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// listRolesForUser godoc
// @Summary List roles for a user
// @Description Get roles assigned to a user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} routes.RoleListResponse
// @Router /users/{id}/roles [get]
func listRolesForUser(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)
	urs, err := q.UserRole.Where(query.UserRole.UserID.Eq(id)).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(urs) == 0 {
		c.JSON(http.StatusOK, RoleListResponse{Data: []RoleDTO{}})
		return
	}
	ids := make([]string, 0, len(urs))
	for _, ur := range urs {
		ids = append(ids, ur.RoleID)
	}
	roles, err := q.Role.Where(query.Role.ID.In(ids...)).Find()
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
		})
	}
	c.JSON(http.StatusOK, RoleListResponse{Data: out})
}

// getUserPermissions godoc
// @Summary Get user permissions
// @Description Get the combined permissions for a user based on their roles
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} routes.PermissionsResponse
// @Router /users/{id}/permissions [get]
func getUserPermissions(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")

	q := query.Use(db)
	_, err := q.User.Where(q.User.ID.Eq(id)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	permissions, err := middleware.GetUserPermissions(id, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	permissionsText := constants.GetPermissionsText(int64(permissions))

	c.JSON(http.StatusOK, PermissionsResponse{
		PermissionBitmask: int64(permissions),
		PermissionsText:   permissionsText,
	})
}

// listExternalIdentities godoc
// @Summary List external identities for a user
// @Description Get external identities linked to a user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} routes.ExternalIdentityListResponse
// @Router /users/{id}/external_identities [get]
func listExternalIdentities(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	cfg := c.MustGet("config").(config.Config)
	id := c.Param("id")
	q := query.Use(db)
	eis, err := q.ExternalIdentity.Where(query.ExternalIdentity.UserID.Eq(id)).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(eis) == 0 {
		c.JSON(http.StatusOK, ExternalIdentityListResponse{Data: []ExternalIdentityDTO{}})
		return
	}
	var out []ExternalIdentityDTO
	for _, e := range eis {
		refreshed, _ := utils.RefreshExternalToken(e, q, &cfg)

		dto := ExternalIdentityDTO{
			ID:             refreshed.ID,
			UserID:         refreshed.UserID,
			Provider:       refreshed.Provider,
			ExternalUserID: refreshed.ExternalUserID,
			CreatedAt:      refreshed.CreatedAt,
			UpdatedAt:      refreshed.UpdatedAt,
		}

		if refreshed.IDToken != nil {
			if claims, err := utils.DecodeIDTokenClaims(*refreshed.IDToken); err == nil {
				dto.IDTokenClaims = claims
			}
		}

		if info, err := utils.FetchProviderUserInfo(refreshed); err == nil && info != nil {
			dto.Username = info.Username
			dto.DisplayName = info.DisplayName
			dto.AvatarURL = info.AvatarURL
			dto.Email = info.Email
			dto.ProviderData = info.ProviderData
		}

		out = append(out, dto)
	}
	c.JSON(http.StatusOK, ExternalIdentityListResponse{Data: out})
}

// addExternalIdentity godoc
// @Summary Link an external account
// @Description Link an external identity to the user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body routes.CreateExternalIdentityRequest true "External identity"
// @Success 201 {object} routes.ExternalIdentityDTO
// @Router /users/{id}/external_identities [post]
func addExternalIdentity(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	var input CreateExternalIdentityRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q := query.Use(db)
	if _, err := q.User.Where(query.User.ID.Eq(id)).First(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	now := time.Now().UTC()
	ei := &model.ExternalIdentity{
		ID:             ulid.Make().String(),
		UserID:         id,
		Provider:       input.Provider,
		ExternalUserID: input.ExternalUserID,
		IDToken:        stringToPtr(input.IDToken),
		AccessToken:    input.AccessToken,
		RefreshToken:   input.RefreshToken,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if input.TokenExpiresAt != nil {
		ei.TokenExpiresAt = timeToTimePtr(*input.TokenExpiresAt)
	}
	if err := q.ExternalIdentity.Create(ei); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if input.Provider == "discord" {
		config := c.MustGet("config").(config.Config)
		if err := discordutil.AddToGuild(input.ExternalUserID, db, &config); err != nil {
			log.Printf("failed to add user to discord guild: %v", err)
		}

		if user, uerr := q.User.Where(query.User.ID.Eq(id)).First(); uerr == nil {
			if user.Status == "active" {
				memberRoleId := config.DiscordConfig.Guild.MemberRoleID
				if err := discordutil.AddRoleToUser(input.ExternalUserID, memberRoleId, &config); err != nil {
					log.Printf("failed to add discord role: %v", err)
				}
			}
		}
	}

	resp := ExternalIdentityDTO{
		ID:             ei.ID,
		UserID:         ei.UserID,
		Provider:       ei.Provider,
		ExternalUserID: ei.ExternalUserID,
		CreatedAt:      ei.CreatedAt,
		UpdatedAt:      ei.UpdatedAt,
	}

	if ei.IDToken != nil {
		if claims, err := utils.DecodeIDTokenClaims(*ei.IDToken); err == nil {
			resp.IDTokenClaims = claims
		}
	}

	if info, err := utils.FetchProviderUserInfo(ei); err == nil && info != nil {
		resp.Username = info.Username
		resp.DisplayName = info.DisplayName
		resp.AvatarURL = info.AvatarURL
		resp.Email = info.Email
		resp.ProviderData = info.ProviderData
	}

	c.JSON(http.StatusCreated, resp)
}

// removeExternalIdentity godoc
// @Summary Unlink an external account
// @Description Remove an external identity linked to a user
// @Tags users
// @Param id path string true "User ID"
// @Param eid path string true "External identity ID"
// @Success 204
// @Router /users/{id}/external_identities/{eid} [delete]
func removeExternalIdentity(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	eid := c.Param("eid")
	q := query.Use(db)
	if _, err := q.ExternalIdentity.Where(query.ExternalIdentity.ID.Eq(eid), query.ExternalIdentity.UserID.Eq(id)).First(); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if _, err := q.ExternalIdentity.Delete(&model.ExternalIdentity{ID: eid}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// linkDiscordByEmailCode godoc
// @Summary Link Discord account by email verification code
// @Description メール検証コードを使ってDiscord連携を行う
// @Tags users
// @Accept json
// @Produce json
// @Param body body routes.EmailVerifyDiscordLinkRequest true "Discord link request"
// @Success 201 {object} routes.ExternalIdentityDTO
// @Router /internal/users/email_verify/discord_link [post]
func linkDiscordByEmailCode(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	var input EmailVerifyDiscordLinkRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q := query.Use(db)
	evc, err := q.EmailVerificationCode.Where(
		query.EmailVerificationCode.Code.Eq(input.Code),
	).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "code_not_found"})
		return
	}
	if time.Now().After(evc.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code_expired"})
		return
	}
	if evc.RequestType != "registration" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request_type"})
		return
	}
	if _, err := q.User.Where(query.User.ID.Eq(evc.UserID)).First(); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if _, err := q.ExternalIdentity.Where(
		query.ExternalIdentity.UserID.Eq(evc.UserID),
		query.ExternalIdentity.Provider.Eq("discord"),
	).First(); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "discord_already_linked"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing, err := q.ExternalIdentity.Where(
		query.ExternalIdentity.Provider.Eq("discord"),
		query.ExternalIdentity.ExternalUserID.Eq(input.ExternalUserID),
	).First(); err == nil {
		if existing.UserID != evc.UserID {
			c.JSON(http.StatusConflict, gin.H{"error": "discord_already_linked"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "discord_already_linked"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	ei := &model.ExternalIdentity{
		ID:             ulid.Make().String(),
		UserID:         evc.UserID,
		Provider:       "discord",
		ExternalUserID: input.ExternalUserID,
		AccessToken:    input.AccessToken,
		RefreshToken:   input.RefreshToken,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if input.TokenExpiresAt != nil {
		ei.TokenExpiresAt = timeToTimePtr(*input.TokenExpiresAt)
	}
	if err := q.ExternalIdentity.Create(ei); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := ExternalIdentityDTO{
		ID:             ei.ID,
		UserID:         ei.UserID,
		Provider:       ei.Provider,
		ExternalUserID: ei.ExternalUserID,
		CreatedAt:      ei.CreatedAt,
		UpdatedAt:      ei.UpdatedAt,
	}
	if info, err := utils.FetchProviderUserInfo(ei); err == nil && info != nil {
		resp.Username = info.Username
		resp.DisplayName = info.DisplayName
		resp.AvatarURL = info.AvatarURL
		resp.Email = info.Email
		resp.ProviderData = info.ProviderData
	}
	c.JSON(http.StatusCreated, resp)
}

// EmailCodeCheck godoc
// @Summary Verify email code
// @Description 認証コードを検証する
// @Tags users
// @Accept json
// @Produce json
// @Param body body routes.EmailCodeCheckRequest true "Email code verification"
// @Success 200 {object} routes.EmailCodeCheckResponse
// @Router /internal/users/email_verify [post]
func emailCodeCheck(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	var input EmailCodeCheckRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q := query.Use(db)
	evc, err := q.EmailVerificationCode.Where(
		query.EmailVerificationCode.Code.Eq(input.Code),
	).First()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_code"})
		return
	}
	if time.Now().After(evc.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code_expired"})
		return
	}

	user, err := q.User.Where(query.User.ID.Eq(evc.UserID)).First()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_not_found"})
		return
	}

	var verificationType string
	switch evc.RequestType {
	case "registration":
		if user.Status == "established" {
			verificationType = "signup"
		} else {
			verificationType = "migration"
		}
	case "email_change":
		if evc.NewEmail != nil && *evc.NewEmail != "" {
			verificationType = "change"
		} else {
			verificationType = "migration"
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request_type"})
		return
	}

	if verificationType == "signup" {
		_, err := q.ExternalIdentity.Where(
			query.ExternalIdentity.UserID.Eq(evc.UserID),
			query.ExternalIdentity.Provider.Eq("discord"),
		).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "discord_not_linked"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// トランザクション処理の適用（ユーザー更新とコード削除のアトミック性担保）
	err = q.Transaction(func(tx *query.Query) error {
		now := time.Now().UTC()
		switch evc.RequestType {
		case "registration":
			if _, err = tx.User.Where(tx.User.ID.Eq(evc.UserID)).Updates(map[string]interface{}{
				"email_verified": true,
				"updated_at":     now,
			}); err != nil {
				return err
			}
			externalCount, _ := tx.ExternalIdentity.Where(
				tx.ExternalIdentity.UserID.Eq(evc.UserID),
				tx.ExternalIdentity.Provider.Eq("discord"),
			).Count()
			if externalCount > 0 {
				if _, err = tx.EmailVerificationCode.Where(tx.EmailVerificationCode.ID.Eq(evc.ID)).Delete(); err != nil {
					return err
				}
			}
		case "email_change":
			if _, err = tx.User.Where(tx.User.ID.Eq(evc.UserID)).Updates(map[string]interface{}{
				"external_email": evc.NewEmail,
				"email_verified": true,
				"updated_at":     now,
			}); err != nil {
				return err
			}
			if _, err = tx.EmailVerificationCode.Where(tx.EmailVerificationCode.ID.Eq(evc.ID)).Delete(); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, EmailCodeCheckResponse{Valid: true, Type: verificationType})
}

// approveUserRegist godoc
// @Summary Approve user registration
// @Description ユーザ登録を承認する
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Param body body UserApproveDTO true "Approval details"
// @Success 200
// @Router /users/{id}/approve [post]
func approveUserRegist(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	user_id := c.Param("id")
	q := query.Use(db)

	// リクエストボディを受け取る
	var dto UserApproveDTO
	if err := c.BindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request_body"})
		return
	}

	existedUser, err := q.User.Where(q.User.ID.Eq(user_id)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "notfound"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if dto.Force == nil || !*dto.Force {
		// EmailVerifiedを確認
		if !existedUser.EmailVerified {
			c.JSON(http.StatusBadRequest, gin.H{"error": "external_email_not_verified"})
			return
		}
		// Discord連携を確認
		count, err := q.ExternalIdentity.Where(q.ExternalIdentity.UserID.Eq(existedUser.ID), q.ExternalIdentity.Provider.Eq("discord")).Count()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "discord_is_not_linked"})
			return
		}
	}

	// timeに変換（YYYY-MM-DD をフロントが送るのでそれを受け入れる）
	joinedAt, err := parseDateFlexible(dto.JoinedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_joined_at_format"})
		return
	}

	// トランザクション処理の開始
	err = q.Transaction(func(tx *query.Query) error {
		now := time.Now().UTC()

		// ユーザー情報の更新（GORMモデルを使用）
		userUpdate := &model.User{
			Status: "active",
			Email:  dto.Email,
			AffiliationPeriod: func() *string {
				period := dto.AffiliationPeriod
				return &period
			}(),
			UpdatedAt: now,
		}
		// ゼロ値や空文字の更新漏れを防ぐため Select で明示的に指定
		if _, err := tx.User.Where(tx.User.ID.Eq(user_id)).Select(
			tx.User.Status,
			tx.User.Email,
			tx.User.AffiliationPeriod,
			tx.User.UpdatedAt,
		).Updates(userUpdate); err != nil {
			return err
		}

		// プロフィール情報の更新（GORMモデルを使用）
		profileUpdate := &model.Profile{
			JoinedAt:  &joinedAt,
			UpdatedAt: now,
		}
		if _, err := tx.Profile.Where(tx.Profile.UserID.Eq(user_id)).Select(
			tx.Profile.JoinedAt,
			tx.Profile.UpdatedAt,
		).Updates(profileUpdate); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Discord連携: ロール付与とウェルカムメッセージ送信（DBコミット後に実行）
	if externalIdentity, e := q.ExternalIdentity.Where(
		query.ExternalIdentity.UserID.Eq(user_id),
		query.ExternalIdentity.Provider.Eq("discord"),
	).First(); e == nil {
		discordUserId := externalIdentity.ExternalUserID

		// ロール付与
		config := c.MustGet("config").(config.Config)
		memberRoleId := config.DiscordConfig.Guild.MemberRoleID
		if memberRoleId != "" {
			if err := discordutil.AddRoleToUser(discordUserId, memberRoleId, &config); err != nil {
				log.Printf("failed to add discord role: %v", err)
			}
		} else {
			log.Printf("DISCORD_MEMBER_ROLE_ID is not set; skipping role assignment")
		}

		// 表示名を取得
		displayName := "メンバー"
		if profile, perr := q.Profile.Where(query.Profile.UserID.Eq(user_id)).First(); perr == nil {
			if profile.DisplayName != "" {
				displayName = profile.DisplayName
			}
		} else if u, uerr := q.User.Where(query.User.ID.Eq(user_id)).First(); uerr == nil {
			if u.CustomID != "" {
				displayName = u.CustomID
			}
		}

		// ウェルカムメッセージ送信
		if err := discordutil.SendWelcomeMessage(discordUserId, dto.Email, dto.SakuraEmailPassword, displayName, db, &config); err != nil {
			log.Printf("failed to send welcome DM: %v", err)
		}
	}

	c.Status(http.StatusOK)
}

// rejectUserRegist godoc
// @Summary Reject user registration
// @Description ユーザ登録を却下し、ユーザを物理削除する
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200
// @Router /users/{id}/reject [post]
func rejectUserRegist(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	user_id := c.Param("id")
	q := query.Use(db)

	// 物理削除をトランザクション化
	err := q.Transaction(func(tx *query.Query) error {
		if _, err := tx.User.Unscoped().Delete(&model.User{ID: user_id}); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// resendEmailVerification godoc
// @Summary Resend email verification
// @Description メール認証メールを再送する
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200
// @Router /users/{id}/resend_email_verification [post]
func resendEmailVerification(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)
	user, err := q.User.Where(query.User.ID.Eq(id)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already verified"})
		return
	}
	if user.ExternalEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no external email set"})
		return
	}
	// 既存のコードを取得
	existingCodes, err := q.EmailVerificationCode.Where(
		query.EmailVerificationCode.UserID.Eq(id),
	).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(existingCodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no existing verification request"})
		return
	}
	externalEmail := existingCodes[0].NewEmail
	requestType := existingCodes[0].RequestType
	if (externalEmail == nil || *externalEmail == "") && requestType == "email_change" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no external email set in existing request"})
		return
	}

	// プロフィールから名前を取得
	name := ""
	if p, err := q.Profile.Where(query.Profile.UserID.Eq(id)).First(); err == nil {
		name = p.DisplayName
	}
	cfg := config.LoadConfig()

	// 既存コードの削除と新しいコードの生成を同一トランザクションで行う
	err = q.Transaction(func(tx *query.Query) error {
		// 既存の未使用コードを削除（モデルを明示指定）
		if _, err := tx.EmailVerificationCode.Where(
			tx.EmailVerificationCode.UserID.Eq(id),
			tx.EmailVerificationCode.RequestType.Eq(requestType),
		).Delete(&model.EmailVerificationCode{}); err != nil {
			return err
		}

		if requestType == "email_change" {
			if err = sendEmailChangeVerification(id, *externalEmail, name, tx, cfg); err != nil {
				return err
			}
		} else {
			if err = sendRegistrationEmailVerification(id, user.ExternalEmail, name, tx, cfg); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send verification email: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "verification email sent"})
}

// changePassword godoc
// @Summary Change user password
// @Description Change a user's password. The requester must be the user themself or have USER_UPDATE permission. If the requester is the user, the current password must be provided.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body UserPasswordChangeDTO true "Password change request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id}/password/change [put]
func changePassword(c *gin.Context) {
	if isOAuth := IsOAuth(c); isOAuth {
		// 403
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action with an access token"})
		return
	}
	db := getDB(c)
	if db == nil {
		return
	}
	id := c.Param("id")
	q := query.Use(db)

	userModel, err := q.User.Where(query.User.ID.Eq(id)).First()
	if err != nil || userModel == nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var input UserPasswordChangeDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine if requester is self or has USER_UPDATE
	isSelf := false
	hasPerm := false
	if u, exists := c.Get("user"); exists {
		if um, ok := u.(*model.User); ok && um != nil {
			if um.ID == id {
				isSelf = true
			}
			if perms, err := middleware.GetUserPermissions(um.ID, db); err == nil {
				hasPerm = perms.HasPermission(constants.USER_UPDATE)
			}
		}
	}

	if !isSelf && !hasPerm {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// If requester is self, verify current password
	if isSelf && !hasPerm {
		if input.CurrentPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "current_password required"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(userModel.PasswordHash), []byte(input.CurrentPassword)); err != nil {
			// If bcrypt hash is malformed, fall back to legacy SHA256 hex(password)
			if _, ok := err.(bcrypt.InvalidHashPrefixError); ok {
				sum := sha256.Sum256([]byte(input.CurrentPassword))
				if hex.EncodeToString(sum[:]) != userModel.PasswordHash {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid current password"})
					return
				}
				// matched legacy hash; continue
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid current password"})
				return
			}
		}
	}

	// Generate hash for new password via issuer internal API
	cfg := c.MustGet("config").(config.Config)
	req := map[string]string{"password": input.NewPassword}
	reqBody, jerr := json.Marshal(req)
	if jerr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": jerr.Error()})
		return
	}
	resp, err := http.Post(cfg.IssuerInternalURL+"/internal/password_hash", "application/json", strings.NewReader(string(reqBody)))
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	defer resp.Body.Close()
	var respData struct {
		PasswordHash string `json:"password_hash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse password hash response"})
		return
	}

	// パスワードハッシュの更新をトランザクション化（GORMモデルを使用）
	err = q.Transaction(func(tx *query.Query) error {
		now := time.Now().UTC()
		userUpdate := &model.User{
			PasswordHash: respData.PasswordHash,
			UpdatedAt:    now,
		}
		if _, err := tx.User.Where(tx.User.ID.Eq(id)).Select(
			tx.User.PasswordHash,
			tx.User.UpdatedAt,
		).Updates(userUpdate); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}

func sendRegistrationEmailVerification(user_id, email, name string, q *query.Query, config *config.Config) error {
	// コードを生成してDBに保存する
	// 6桁のコードを生成
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	code := fmt.Sprintf("%06X", b)

	now := time.Now().UTC()
	err := q.EmailVerificationCode.Create(&model.EmailVerificationCode{
		Code:        code,
		RequestType: "registration",
		UserID:      user_id,
		ExpiresAt:   now.Add(10 * time.Minute),
		CreatedAt:   now,
	})
	if err != nil {
		return err
	}

	// Email APIを呼び出す
	// HTTP /register { code, email, name }
	endpoint := config.EmailSenderURL + "/register"
	payload := map[string]string{
		"code":  code,
		"email": email,
		"name":  name,
	}
	client := &http.Client{Timeout: 10 * time.Second}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(reqBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to send email verification")
	}

	return nil
}

func sendEmailChangeVerification(user_id, email, name string, q *query.Query, config *config.Config) error {
	// コードを生成してDBに保存する
	// 6桁のコードを生成
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	code := fmt.Sprintf("%06X", b)

	now := time.Now().UTC()
	err := q.EmailVerificationCode.Create(&model.EmailVerificationCode{
		Code:        code,
		RequestType: "email_change",
		UserID:      user_id,
		ExpiresAt:   now.Add(10 * time.Minute),
		NewEmail:    stringToPtr(email),
		CreatedAt:   now,
	})
	if err != nil {
		return err
	}

	// Email APIを呼び出す
	// HTTP /email-change { code, email, name }
	endpoint := config.EmailSenderURL + "/email-change"
	payload := map[string]string{
		"code":  code,
		"email": email,
		"name":  name,
	}
	client := &http.Client{Timeout: 10 * time.Second}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(string(reqBody)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to send email verification")
	}

	return nil
}

func isAdult(birthdate *time.Time) *bool {
	if birthdate == nil {
		return nil
	}
	now := time.Now().UTC() // タイムゾーンを一貫させるためUTCに統一
	age := now.Year() - birthdate.Year()
	if now.Month() < birthdate.Month() || (now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
		age--
	}
	isAdult := age >= 18
	return &isAdult
}
