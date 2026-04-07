package router

import (
	"log"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/UniPro-tech/UniQUE-Auth/internal/model"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

type AuthenticationRequest struct {
	Username  string `json:"username" binding:"omitempty"`
	Password  string `json:"password" binding:"omitempty"`
	Code      string `json:"code" binding:"omitempty"` // for MFA/TOTP
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	Type      string `json:"type" binding:"required,oneof=password mfa totp"`
	Remember  bool   `json:"remember" default:"false"`
}

type AuthenticationResponse struct {
	SessionJWT string   `json:"session_jwt" binding:"omitempty"`
	RequireMFA bool     `json:"require_mfa" binding:"omitempty"`
	MFATypes   []string `json:"mfa_type" binding:"omitempty"` // "totp"
}

// AuthenticationPost godoc
// @Summary authentication endpoint
// @Schemes
// @Description 内部用の認証エンドポイントです。Kubernetes / Istio の認証ポリシーにより外部からのアクセスは制限されています。
// @Tags internal
// @Success 200 {object} AuthenticationResponse "OK"
// @Param request body AuthenticationRequest true "Authentication Request"
// @Accept json
// @Router /internal/authentication [post]
func AuthenticationPost(c *gin.Context) {
	req := AuthenticationRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(500, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	switch req.Type {
	case "password":
		// Handle password authentication
		user, err, reason := passwordAuthentication(q, req.Username, req.Password)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if user == nil {
			c.JSON(401, gin.H{"error": "Invalid credentials", "reason": reason})
			return
		}

		if user.IsTotpEnabled {
			c.JSON(200, AuthenticationResponse{
				RequireMFA: true,
				MFATypes:   []string{"totp"},
			})
			return
		}

		err = q.Session.Create(&model.Session{
			UserID:      user.ID,
			IPAddress:   req.IPAddress,
			UserAgent:   req.UserAgent,
			ExpiresAt:   CalculateSessionExpiry(req.Remember),
			LastLoginAt: time.Now(),
		})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		session, err := q.Session.Where(q.Session.UserID.Eq(user.ID)).Order(q.Session.CreatedAt.Desc()).First()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if session == nil {
			c.JSON(500, gin.H{"error": "Failed to create session"})
			return
		}
		// Return session ID as a token
		sessionJWT, err := util.GenerateSessionJWT(session.ID, user.ID, session.ExpiresAt, *c.MustGet("config").(*config.Config))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		ip := req.IPAddress
		if ip == "" {
			ip = c.ClientIP()
		}
		userAgent := req.UserAgent
		if userAgent == "" {
			userAgent = c.Request.UserAgent()
		}

		writeAuditLog(c, "LOGIN", "sessions/"+session.ID, &user.ID, nil, &session.ID, map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"ip":         ip,
			"user_agent": userAgent,
			"auth_type":  req.Type,
			"remember":   req.Remember,
			"session_id": session.ID,
			"user_id":    user.ID,
			"timestamp":  time.Now().UTC(),
		})

		c.JSON(200, AuthenticationResponse{
			SessionJWT: sessionJWT,
		})
	case "totp":
		// Handle TOTP authentication
		user, err := totpAuthentication(q, req.Username, req.Code)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if user == nil {
			c.JSON(401, gin.H{"error": "Invalid TOTP code", "reason": "invalid_credentials"})
			return
		}

		err = q.Session.Create(&model.Session{
			UserID:      user.ID,
			IPAddress:   req.IPAddress,
			UserAgent:   req.UserAgent,
			ExpiresAt:   CalculateSessionExpiry(req.Remember),
			LastLoginAt: time.Now(),
		})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		session, err := q.Session.Where(q.Session.UserID.Eq(user.ID)).Order(q.Session.CreatedAt.Desc()).First()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if session == nil {
			c.JSON(500, gin.H{"error": "Failed to create session"})
			return
		}
		sessionJWT, err := util.GenerateSessionJWT(session.ID, user.ID, session.ExpiresAt, *c.MustGet("config").(*config.Config))
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		ip := req.IPAddress
		if ip == "" {
			ip = c.ClientIP()
		}
		userAgent := req.UserAgent
		if userAgent == "" {
			userAgent = c.Request.UserAgent()
		}

		writeAuditLog(c, "LOGIN", "sessions/"+session.ID, &user.ID, nil, &session.ID, map[string]interface{}{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"ip":         ip,
			"user_agent": userAgent,
			"auth_type":  req.Type,
			"remember":   req.Remember,
			"session_id": session.ID,
			"user_id":    user.ID,
			"timestamp":  time.Now().UTC(),
		})

		c.JSON(200, AuthenticationResponse{
			SessionJWT: sessionJWT,
		})
	default:
		c.JSON(400, gin.H{"error": "Invalid authentication type"})
	}
}

func passwordAuthentication(q *query.Query, username, password string) (resuser *model.User, reserr error, reason *string) {
	user, err := q.User.Where(q.User.CustomID.Eq(username), q.User.DeletedAt.IsNull()).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			reason := "invalid_credentials"
			return nil, nil, &reason // user not found is not an error for this function
		}
		return nil, err, nil
	}
	if user == nil {
		reason := "invalid_credentials"
		return nil, nil, &reason
	}
	if user.Status != "active" {
		reason := "user_inactive"
		return nil, nil, &reason
	}
	if ok, err := util.VerifyPassword(password, user.PasswordHash); err != nil || !ok {
		reason := "invalid_credentials"
		sanitizedUsername := strings.ReplaceAll(strings.ReplaceAll(username, "\n", "\\n"), "\r", "\\r")
		log.Printf("Password verification failed for user %s: %v", sanitizedUsername, err)
		return nil, nil, &reason
	}
	return user, nil, nil
}

// CalculateSessionExpiry calculates the session expiry time based on the remember flag.
func CalculateSessionExpiry(remember bool) (expiryTime time.Time) {
	if remember {
		return time.Now().Add(30 * 24 * time.Hour) // 30 days
	}
	return time.Now().Add(7 * 24 * time.Hour) // 7 days
}

func totpAuthentication(q *query.Query, username string, code string) (*model.User, error) {
	user, err := q.User.Where(q.User.CustomID.Eq(username), q.User.DeletedAt.IsNull(), q.User.Status.Eq("active")).First()
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	if user.TotpSecret == "" || user.IsTotpEnabled == false {
		return nil, nil
	}
	if ok := totp.Validate(code, user.TotpSecret); !ok {
		return nil, nil
	}
	return user, nil
}
