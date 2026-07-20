package router

import (
	"errors"
	"net/http"
	"strings"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	userId := c.Param("uid")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	var secret, uri string

	// トランザクション内で読み込みから更新までを一貫して実行
	err := q.Transaction(func(tx *query.Query) error {
		// 行ロック（Clauses(clause.Locking{Strength: "UPDATE"})）を必要に応じて追記しても良いですが、
		// 通常の分離レベルでも同一トランザクション内でクエリを完結させることで安全性を高めます
		user, err := tx.User.Where(tx.User.ID.Eq(userId)).First()
		if err != nil {
			return err
		}

		// トランザクションのコンテキスト（tx）を引き継いでパスワード認証を行う
		authUser, authErr, reason := passwordAuthentication(tx, user.CustomID, req.Password)
		if authErr != nil {
			return authErr
		}
		if authUser == nil {
			// 独自の文字列エラーを返すことで、外側で401としてハンドリングする
			if reason != nil {
				return errors.New("auth_failed:" + *reason)
			}
			return errors.New("auth_failed")
		}

		if authUser.IsTotpEnabled {
			return errors.New("totp_already_enabled")
		}

		const issuer = "UniQUE"
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      issuer,
			AccountName: authUser.CustomID,
		})
		if err != nil {
			return errors.New("failed_to_generate_totp")
		}

		_, err = tx.User.Where(tx.User.ID.Eq(authUser.ID)).Updates(map[string]any{
			"totp_secret": key.Secret(),
			"updated_at":  time.Now().UTC(),
		})
		if err != nil {
			return err
		}

		secret = key.Secret()
		uri = key.URL()
		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if strings.HasPrefix(err.Error(), "auth_failed") {
			res := gin.H{"error": "invalid credentials"}
			if parts := strings.Split(err.Error(), ":"); len(parts) > 1 {
				res["reason"] = parts[1]
			}
			c.JSON(http.StatusUnauthorized, res)
			return
		}
		if err.Error() == "totp_already_enabled" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "TOTP is already enabled"})
			return
		}
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, GenerateTOTPResponse{
		Secret: secret,
		URI:    uri,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	userID := c.Param("uid")
	var valid bool

	err := q.Transaction(func(tx *query.Query) error {
		user, err := tx.User.Where(tx.User.ID.Eq(userID)).First()
		if err != nil {
			return err
		}

		if user.TotpSecret == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "totp secret is not set"})
			return errors.New("totp secret is not set")
		}

		valid = totp.Validate(req.Code, *user.TotpSecret)
		if valid {
			_, err = tx.User.Where(tx.User.ID.Eq(user.ID)).Updates(map[string]any{
				"is_totp_enabled": true,
				"updated_at":      time.Now().UTC(),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	userId := c.Param("uid")

	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Database is not available"))
		return
	}
	q := query.Use(db)

	err := q.Transaction(func(tx *query.Query) error {
		user, err := tx.User.Where(tx.User.ID.Eq(userId)).First()
		if err != nil {
			return err
		}

		authUser, authErr, reason := passwordAuthentication(tx, user.CustomID, req.Password)
		if authErr != nil {
			return authErr
		}
		if authUser == nil {
			if reason != nil {
				return errors.New("auth_failed:" + *reason)
			}
			return errors.New("auth_failed")
		}

		if !authUser.IsTotpEnabled {
			return errors.New("totp_not_enabled")
		}

		_, err = tx.User.Where(tx.User.ID.Eq(authUser.ID)).Updates(map[string]any{
			"totp_secret":     "",
			"is_totp_enabled": false,
			"updated_at":      time.Now().UTC(),
		})
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if strings.HasPrefix(err.Error(), "auth_failed") {
			res := gin.H{"error": "invalid credentials"}
			if parts := strings.Split(err.Error(), ":"); len(parts) > 1 {
				res["reason"] = parts[1]
			}
			c.JSON(http.StatusUnauthorized, res)
			return
		}
		if err.Error() == "totp_not_enabled" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "TOTP is not enabled"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, DisableTOTPResponse{
		Message: "TOTP disabled successfully",
	})
}
