package routes

import (
	"github.com/UniPro-tech/UniQUE-MailServer/internal/config"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/templates"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/utils"
	"github.com/gin-gonic/gin"
)

type EmailChangeRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Code  string `json:"code"`
}

// EmailChangeEmail godoc
// @Summary Send Email Change Verification Email
// @Description send an email-change verification email to a user
// @Tags email
// @Accept  json
// @Produce  json
// @Param emailChange body EmailChangeRequest true "Email Change Info"
// @Success 200 {object} map[string]string
// @Router /email-change [post]
func EmailChangeEmail(c *gin.Context) {
	// リクエストのbodyからemail/name等を受け取る（最低限のバリデーション）
	var req EmailChangeRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	config := c.MustGet("config").(*config.Config)

	// ベースURLをリクエストから生成
	baseURL := config.FrontendURL

	// テンプレートをレンダリング
	htmlStr, textStr, err := templates.RenderEmailChangeVerificationHTML(config.AppName, baseURL, req.Code, config.CopyrightName)
	if err != nil {
		c.JSON(500, gin.H{"error": "template render error"})
		return
	}

	// メール送信
	if err := utils.SendMail(htmlStr, textStr, "メールアドレス変更の確認", req.Email); err != nil {
		c.JSON(500, gin.H{"error": "mail send error: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"ok": true})
}
