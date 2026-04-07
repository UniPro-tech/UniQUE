package routes

import (
	"github.com/UniPro-tech/UniQUE-MailServer/internal/config"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/templates"
	"github.com/UniPro-tech/UniQUE-MailServer/internal/utils"
	"github.com/gin-gonic/gin"
)

type PasswordResetRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Code  string `json:"code"`
}

// PasswordResetEmail godoc
// @Summary Send Password Reset Email
// @Description send a password reset email to a user
// @Tags password
// @Accept  json
// @Produce  json
// @Param passwordReset body PasswordResetRequest true "Password Reset Info"
// @Success 200 {object} map[string]string
// @Router /password-reset [post]
func PasswordResetEmail(c *gin.Context) {
	var req PasswordResetRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	config := c.MustGet("config").(*config.Config)
	baseURL := config.FrontendURL

	// テンプレートをレンダリング
	htmlStr, textStr, err := templates.RenderPasswordResetVerificationHTML(config.AppName, baseURL, req.Code, req.Name, config.CopyrightName)
	if err != nil {
		c.JSON(500, gin.H{"error": "template render error"})
		return
	}

	// メール送信
	if err := utils.SendMail(htmlStr, textStr, "パスワードリセット", req.Email); err != nil {
		c.JSON(500, gin.H{"error": "mail send error: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"ok": true})
}
