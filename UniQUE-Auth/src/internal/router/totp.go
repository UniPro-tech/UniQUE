package router

import (
	"errors"
	"net/http"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

type GenerateTOTPRequest struct {
	Password string `json:"password" binding:"required"`
}

type GenerateTOTPResponse struct {
	Secret string `json:"secret"` // TOTPシークレット
	URI    string `json:"uri"`    // QRコード用URI (otpauth://...)
}

// GenerateTOTP godoc
// @Summary Generate TOTP secret and QR code
// @Schemes
// @Description TOTPのシークレットとQRコード用URIを生成します。ユーザーがTOTPをセットアップする際に使用します。
// @Tags internal
// @Success 200 {object} GenerateTOTPResponse "OK"
// @Param request body GenerateTOTPRequest true "Generate TOTP Request"
// @Accept json
// @Router /internal/totp/{uid} [POST]
func GenerateTOTP(c *gin.Context) {
	req := GenerateTOTPRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.Param("uid")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	// useridからユーザーを取得する
	user, err := q.User.Where(q.User.ID.Eq(userId)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// customidとpasswordでユーザーを認証する
	authUser, authErr, reason := passwordAuthentication(q, user.CustomID, req.Password)
	if authErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if authUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials", "reason": reason})
		return
	}

	if authUser.IsTotpEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TOTP is already enabled"})
		return
	}

	const issuer = "UniQUE"
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: authUser.CustomID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate totp key"})
		return
	}

	_, err = q.User.Where(q.User.ID.Eq(authUser.ID)).Updates(map[string]any{
		"totp_secret": key.Secret(),
		"updated_at":  time.Now().UTC(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, GenerateTOTPResponse{
		Secret: key.Secret(),
		URI:    key.URL(),
	})
}

type VerifyTOTPRequest struct {
	Code string `json:"code" binding:"required"`
}

type VerifyTOTPResponse struct {
	Valid bool `json:"valid"` // TOTPコードが有効かどうか
}

// VerifyTOTP godoc
// @Summary Verify TOTP code
// @Schemes
// @Description TOTPコードを検証し、is_totp_enabledをtrueに設定します。ユーザーがTOTPセットアップを完了する際に使用します。
// @Tags internal
// @Success 200 {object} VerifyTOTPResponse "OK"
// @Param request body VerifyTOTPRequest true "Verify TOTP Request"
// @Accept json
// @Router /internal/totp/{uid}/verify [POST]
func VerifyTOTP(c *gin.Context) {
	req := VerifyTOTPRequest{}
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

	userID := c.Param("uid")

	user, err := q.User.Where(q.User.ID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// TOTPコードの検証
	valid := totp.Validate(req.Code, user.TotpSecret)

	if valid {
		// 検証成功時にフラグを有効化
		_, err = q.User.Where(q.User.ID.Eq(user.ID)).Updates(map[string]any{
			"is_totp_enabled": true,
			"updated_at":      time.Now().UTC(),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}

	c.JSON(http.StatusOK, VerifyTOTPResponse{
		Valid: valid,
	})
}

type DisableTOTPRequest struct {
	Password string `json:"password" binding:"required"`
}

type DisableTOTPResponse struct {
	Message string `json:"message"` // 結果メッセージ
}

// DisableTOTP godoc
// @Summary Disable TOTP
// @Schemes
// @Description TOTPを無効化します。ユーザーがTOTPを無効にする際に使用します。
// @Tags internal
// @Success 200 {object} DisableTOTPResponse "OK"
// @Param request body DisableTOTPRequest true "Disable TOTP Request"
// @Accept json
// @Router /internal/totp/{uid}/disable [POST]
func DisableTOTP(c *gin.Context) {
	req := DisableTOTPRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.Param("uid")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	user, err := q.User.Where(q.User.ID.Eq(userId)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	authUser, authErr, reason := passwordAuthentication(q, user.CustomID, req.Password)
	if authErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if authUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials", "reason": reason})
		return
	}

	if !authUser.IsTotpEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "TOTP is not enabled"})
		return
	}

	_, err = q.User.Where(q.User.ID.Eq(authUser.ID)).Updates(map[string]any{
		"totp_secret":     "",
		"is_totp_enabled": false,
		"updated_at":      time.Now().UTC(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, DisableTOTPResponse{
		Message: "TOTP disabled successfully",
	})
}
