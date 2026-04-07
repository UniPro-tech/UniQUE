package router

import (
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
// @Router /internal/totp/{user_id} [POST]
func GenerateTOTP(c *gin.Context) {
	// Passwordを検証してユーザーを認証する
	req := GenerateTOTPRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userId := c.Param("uid")

	// dbの取得
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(500, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	// useridからcustomidを取得する
	// そのcustomidとpasswordでユーザーを認証する
	user, err := q.User.Where(q.User.ID.Eq(userId)).First()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	user, err, reason := passwordAuthentication(q, user.CustomID, req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(401, gin.H{"error": "invalid credentials", "reason": reason})
		return
	}

	if user.IsTotpEnabled {
		c.JSON(400, gin.H{"error": "TOTP is already enabled"})
		return
	}

	const issuer = "UniQUE"

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: user.CustomID,
	})

	_, err = q.User.Where(q.User.ID.Eq(user.ID)).Updates(map[string]interface{}{
		"totp_secret": key.Secret(),
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Not implemented yet
	c.JSON(200, gin.H{
		"secret": key.Secret(), // TOTPシークレット
		"uri":    key.URL(),    // QRコード用URI (otpauth://...)
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
// @Router /internal/totp/{user_id}/verify [POST]
func VerifyTOTP(c *gin.Context) {
	// TOTPコードを検証する
	// 検証に成功したら、ユーザーのis_totp_enabledをtrueに更新する
	// dbの取得
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(500, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	req := VerifyTOTPRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID := c.Param("uid")

	user, err := q.User.Where(q.User.ID.Eq(userID)).First()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	valid := totp.Validate(req.Code, user.TotpSecret)

	if valid {
		_, err = q.User.Where(q.User.ID.Eq(user.ID)).Updates(map[string]interface{}{
			"is_totp_enabled": true,
		})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, VerifyTOTPResponse{
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
// @Router /internal/totp/{user_id}/disable [POST]
func DisableTOTP(c *gin.Context) {
	// Passwordを検証してユーザーを認証する
	req := DisableTOTPRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userId := c.Param("uid")

	// dbの取得
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.JSON(500, gin.H{"error": "database not available"})
		return
	}
	q := query.Use(db)

	// useridからcustomidを取得する
	// そのcustomidとpasswordでユーザーを認証する
	user, err := q.User.Where(q.User.ID.Eq(userId)).First()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	user, err, reason := passwordAuthentication(q, user.CustomID, req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		c.JSON(401, gin.H{"error": "invalid credentials", "reason": reason})
		return
	}

	if !user.IsTotpEnabled {
		c.JSON(400, gin.H{"error": "TOTP is not enabled"})
		return
	}

	_, err = q.User.Where(q.User.ID.Eq(user.ID)).Updates(map[string]interface{}{
		"totp_secret":     "",
		"is_totp_enabled": false,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, DisableTOTPResponse{
		Message: "TOTP disabled successfully",
	})
}
