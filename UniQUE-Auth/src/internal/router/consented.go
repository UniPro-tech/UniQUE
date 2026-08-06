package router

import (
	"errors"
	"net/http"
	"net/url"

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
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	// エラーチェックを先に行うように順序を修正
	authReq, err := q.AuthorizationRequest.Where(q.AuthorizationRequest.ID.Eq(req.AuthorizationID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(400, gin.H{"error": "invalid auth_request_id"})
			return
		}
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}
	if authReq == nil || !authReq.IsConsented {
		c.JSON(400, gin.H{"error": "invalid auth_request_id"})
		return
	}

	// Generate authorization code or token based on response_type
	if authReq.ResponseType == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device flow is not supported for this endpoint"})
		return
	}
	switch *authReq.ResponseType {
	case "code":
		if authReq.Code == nil || *authReq.Code == "" {
			c.JSON(400, gin.H{"error": "not check consented"})
			return
		}

		// 文字列結合ではなく、net/url を使用して安全にURLとクエリを構築する
		parsedURL, err := url.Parse(*authReq.RedirectURI)
		if err != nil {
			c.JSON(500, gin.H{"error": "invalid redirect_uri format"})
			return
		}

		urlQuery := parsedURL.Query()
		urlQuery.Set("code", *authReq.Code)
		if authReq.State != nil && *authReq.State != "" {
			urlQuery.Set("state", *authReq.State)
		}

		parsedURL.RawQuery = urlQuery.Encode()

		// 構築したURLへリダイレクト
		c.Redirect(301, parsedURL.String())
		return
	default:
		c.JSON(400, gin.H{"error": "unsupported response_type"})
		return
	}
}
