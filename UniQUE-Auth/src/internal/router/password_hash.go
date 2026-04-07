package router

import (
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/gin-gonic/gin"
)

type PasswordHashRequest struct {
	Password string `json:"password" binding:"required"`
}

type PasswordHashResponse struct {
	PasswordHash string `json:"password_hash"`
}

// PasswordHashPost godoc
// @Summary Generate password hash
// @Description パスワードのハッシュ（bcrypt）を生成する内部エンドポイント
// @Tags internal
// @Accept json
// @Produce json
// @Param request body PasswordHashRequest true "Password to hash"
// @Success 200 {object} PasswordHashResponse
// @Router /internal/password_hash [post]
func PasswordHashPost(c *gin.Context) {
	req := PasswordHashRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hashed, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed_to_hash_password"})
		return
	}

	c.JSON(200, PasswordHashResponse{PasswordHash: hashed})
}
