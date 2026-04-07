package router

import (
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ConsentedGetRequest struct {
	AuthorizationID string `form:"authorization_id" binding:"required"`
}

// ConsentedGet godoc
// @Summary Redirect for Authorization Request
// @Schemes
// @Description 認可リクエストが完了したのちにリダイレクトするためのエンドポイントです。
// @Tags authorization
// @Accept json
// @Produce json
// @Param authorization_id query string true "Authorization Request ID"
// @Success 301 {string} string "Redirect to client application with authorization code"
// @Router /consented [get]
func ConsentedGet(c *gin.Context) {
	req := ConsentedGetRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
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

	authReq, err := q.AuthorizationRequest.Where(q.AuthorizationRequest.ID.Eq(req.AuthorizationID)).First()
	if authReq == nil || !authReq.IsConsented {
		c.JSON(400, gin.H{"error": "invalid auth_request_id"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	// Generate authorization code or token based on response_type
	switch authReq.ResponseType {
	case "code":
		if authReq.Code == nil || *authReq.Code == "" {
			c.JSON(400, gin.H{"error": "not check consented"})
			return
		}
		// Redirect to client's redirect_uri with authorization code and state
		redirectURL := authReq.RedirectURI + "?code=" + *authReq.Code
		if authReq.State != nil && *authReq.State != "" {
			redirectURL += "&state=" + *authReq.State
		}
		c.Redirect(301, redirectURL)
		return
	default:
		c.JSON(400, gin.H{"error": "unsupported response_type"})
		return
	}
}
