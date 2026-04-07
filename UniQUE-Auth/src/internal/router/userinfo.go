package router

import (
	"strings"

	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserInfoResponse struct {
	Sub               string   `json:"sub"`
	Name              string   `json:"name,omitempty"`
	Bio               string   `json:"bio,omitempty"`
	PreferredUsername string   `json:"preferred_username,omitempty"`
	Email             string   `json:"email,omitempty"`
	EmailVerified     bool     `json:"email_verified,omitempty"`
	Birthdate         string   `json:"birthdate,omitempty"`
	Website           string   `json:"website,omitempty"`
	Twitter           string   `json:"twitter,omitempty"`
	AffiliationPeriod string   `json:"affiliation_period,omitempty"`
	Roles             []string `json:"roles,omitempty"`
	UpdatedAt         int64    `json:"updated_at,omitempty"`
}

// UserInfoGet godoc
// @Summary      UserInfo Endpoint
// @Description  OAuth2 UserInfo Endpoint
// @Tags         userinfo
// @Accept       json
// @Produce      json
// @Success      200  {object}  UserInfoResponse
// @Failure      400  {object}  map[string]string
// @Router       /userinfo [get]
func UserInfoGet(c *gin.Context) {
	// Authorization Headerからアクセストークンを取得
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(400, gin.H{"error": "missing access token"})
		return
	}
	// Accept both "Bearer <token>" and raw token in header
	if strings.HasPrefix(accessToken, "Bearer ") || strings.HasPrefix(accessToken, "bearer ") {
		parts := strings.SplitN(accessToken, " ", 2)
		if len(parts) == 2 {
			accessToken = parts[1]
		}
	}

	// アクセストークンの検証
	_, userID, scope, err := util.ValidateAccessToken(accessToken, c)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid access token"})
		return
	}

	// ユーザー情報の取得
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(500, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	user, err := q.User.Where(q.User.ID.Eq(userID), q.User.Status.Eq("active")).First()
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to fetch user"})
		return
	}
	if user == nil {
		c.JSON(400, gin.H{"error": "user not found"})
		return
	}

	profile, err := q.Profile.Where(q.Profile.UserID.Eq(userID)).First()
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to fetch profile"})
		return
	}
	if profile == nil {
		c.JSON(400, gin.H{"error": "profile not found"})
		return
	}

	rolesBinding, err := q.UserRole.Where(q.UserRole.UserID.Eq(userID)).Find()
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to fetch user roles"})
		return
	}
	rolesCusomIDs := make([]string, 0, len(rolesBinding))
	for _, roleBinding := range rolesBinding {
		role, err := q.Role.Where(q.Role.ID.Eq(roleBinding.RoleID)).First()
		if err != nil || role == nil {
			continue
		}
		rolesCusomIDs = append(rolesCusomIDs, role.CustomID)
	}

	c.JSON(200, UserInfoResponse{
		Sub:               user.ID,
		Name:              profile.DisplayName,
		Bio:               profile.Bio,
		PreferredUsername: user.CustomID,
		Email: func() string {
			if util.ContainsScope(scope, "email") {
				return user.Email
			}
			return ""
		}(),
		EmailVerified: func() bool {
			if util.ContainsScope(scope, "email") {
				return user.EmailVerified
			}
			return false
		}(),
		Birthdate: func() string {
			if util.ContainsScope(scope, "profile") {
				return profile.Birthdate.Format("2006-01-02")
			}
			return ""
		}(),
		Website: func() string {
			if util.ContainsScope(scope, "profile") {
				return profile.WebsiteURL
			}
			return ""
		}(),
		Twitter: func() string {
			if util.ContainsScope(scope, "profile") {
				return profile.TwitterHandle
			}
			return ""
		}(),
		AffiliationPeriod: user.AffiliationPeriod,
		Roles:             rolesCusomIDs,
		UpdatedAt:         profile.UpdatedAt.Unix(),
	})
}
