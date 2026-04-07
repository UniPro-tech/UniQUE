package routes

import (
	"github.com/UniPro-tech/UniQUE-MailServer/internal/config"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/templates"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/utils"
	"github.com/gin-gonic/gin"
)

type RegistrationRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Code  string `json:"code"`
}

// RegistrationEmail godoc
// @Summary Send Registration Email
// @Description send a registration email to a new user
// @Tags registration
// @Accept  json
// @Produce  json
// @Param registration body RegistrationRequest true "Registration Info"
// @Success 200 {object} map[string]string
// @Router /register [post]
func RegistrationEmail(c *gin.Context) {
	// リクエストのbodyからemail/name等を受け取る（最低限のバリデーション）
	var req RegistrationRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	config := c.MustGet("config").(*config.Config)

	// ベースURLをリクエストから生成
	baseURL := config.FrontendURL

	// テンプレートをレンダリング
	htmlStr, textStr, err := templates.RenderRegistrationVerificationHTML(config.AppName, baseURL, req.Code, req.Name, config.CopyrightName)
	if err != nil {
		c.JSON(500, gin.H{"error": "template render error"})
		return
	}

	// メール送信
	if err := utils.SendMail(htmlStr, textStr, "メールアドレスの確認", req.Email); err != nil {
		c.JSON(500, gin.H{"error": "mail send error: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"ok": true})
}
