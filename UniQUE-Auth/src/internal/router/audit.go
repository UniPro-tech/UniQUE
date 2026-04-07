package router

import (
	"encoding/json"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/model"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid"
	"gorm.io/gorm"
)

func writeAuditLog(c *gin.Context, action, target string, userID, applicationID, sessionID *string, details map[string]interface{}) {
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		return
	}

	var detailStr string
	if details != nil {
		if detailsBytes, err := json.Marshal(details); err == nil {
			detailStr = string(detailsBytes)
		}
	}

	entry := &model.AuditLog{
		ID:             generateAuditID(),
		UserID:         userID,
		ApplicationID:  applicationID,
		SessionID:      sessionID,
		Action:         normalizeAuditAction(action),
		TargetResource: target,
		Trusted:        false,
		Details:        &detailStr,
	}

	if err := query.Use(db).AuditLog.Create(entry); err != nil {
		log.Printf("failed to insert audit log: %v", err)
	}
}

// normalizeAuditAction maps arbitrary action strings to the allowed ENUM values
// in the database. If the provided action is not recognized, fall back to
// 'AUTHORIZATION'.
func normalizeAuditAction(a string) string {
	switch strings.ToUpper(a) {
	case "CREATE":
		return "CREATE"
	case "READ":
		return "READ"
	case "UPDATE":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	case "LOGIN":
		return "LOGIN"
	case "LOGOUT":
		return "LOGOUT"
	case "AUTHORIZATION":
		return "AUTHORIZATION"
	default:
		return "AUTHORIZATION"
	}
}

func generateAuditID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}
