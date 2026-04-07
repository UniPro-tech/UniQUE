package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type auditDetails struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		action := methodToAuditAction(c.Request.Method)
		if action == "" || shouldSkipAudit(c) {
			return
		}

		userAny, ok := c.Get("user")
		if !ok {
			return
		}
		user, ok := userAny.(*model.User)
		if !ok || user == nil {
			return
		}

		dbAny, ok := c.Get("db")
		if !ok {
			return
		}
		db, ok := dbAny.(*gorm.DB)
		if !ok || db == nil {
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		detailsBytes, _ := json.Marshal(auditDetails{
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Status:    c.Writer.Status(),
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})
		details := string(detailsBytes)
		userID := user.ID

		auditLog := &model.AuditLog{
			ID:             ulid.Make().String(),
			UserID:         &userID,
			Action:         action,
			TargetResource: path,
			Trusted:        false,
			Details:        &details,
		}

		if err := query.Use(db).AuditLog.Create(auditLog); err != nil {
			log.Printf("failed to insert audit log: %v", err)
		}
	}
}

func methodToAuditAction(method string) string {
	switch method {
	case http.MethodPost:
		return "CREATE"
	case http.MethodPut, http.MethodPatch:
		return "UPDATE"
	case http.MethodDelete:
		return "DELETE"
	default:
		return ""
	}
}

func shouldSkipAudit(c *gin.Context) bool {
	path := c.Request.URL.Path
	if path == "/health" {
		return true
	}
	if strings.HasPrefix(path, "/swagger") || path == "/swagger.json" {
		return true
	}
	return false
}
