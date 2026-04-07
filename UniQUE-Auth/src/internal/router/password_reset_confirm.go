package router

import (
	"net/http"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PasswordResetConfirmRequest struct {
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type PasswordResetConfirmResponse struct {
	Message string `json:"message"`
}

// PasswordResetConfirmPost godoc
// @Summary Confirm password reset
// @Description Confirm password reset by providing the verification code and new password.
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body PasswordResetConfirmRequest true "Password Reset Confirmation"
// @Success 200 {object} PasswordResetConfirmResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /internal/password_reset/confirm [post]
func PasswordResetConfirmPost(c *gin.Context) {
	req := PasswordResetConfirmRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}

	q := query.Use(db)

	// 検証コードを検索
	verifyCode, err := q.EmailVerificationCode.Where(
		q.EmailVerificationCode.Code.Eq(req.Code),
		q.EmailVerificationCode.RequestType.Eq("password_reset"),
	).First()

	if err != nil || verifyCode == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired code"})
		return
	}

	// コードの有効期限チェック
	if time.Now().After(verifyCode.ExpiresAt) {
		// 期限切れコードを削除
		_, _ = q.EmailVerificationCode.Where(q.EmailVerificationCode.ID.Eq(verifyCode.ID)).Delete()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "code expired"})
		return
	}

	// ユーザーを検索
	user, err := q.User.Where(q.User.ID.Eq(verifyCode.UserID)).First()
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// 新しいパスワードハッシュを生成
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// パスワードハッシュを更新
	if _, err := q.User.Where(q.User.ID.Eq(user.ID)).Update(q.User.PasswordHash, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	// 使用済みの検証コードを削除
	if _, err := q.EmailVerificationCode.Where(q.EmailVerificationCode.ID.Eq(verifyCode.ID)).Delete(); err != nil {
		// 削除失敗時はログするが、ユーザーへはパスワード更新成功と返す
		// （既にパスワードは更新されているため）
	}

	c.JSON(http.StatusOK, PasswordResetConfirmResponse{
		Message: "パスワードをリセットしました。新しいパスワードでログインしてください。",
	})
}
