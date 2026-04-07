package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/UniPro-tech/UniQUE-Auth/internal/model"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid"
	"gorm.io/gorm"
)

type PasswordResetRequestRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetRequestResponse struct {
	Message string `json:"message"`
}

// PasswordResetRequestPost godoc
// @Summary Request password reset
// @Description Request password reset by providing external email. This sends a password reset email with a code.
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body PasswordResetRequestRequest true "Password Reset Request"
// @Success 200 {object} PasswordResetRequestResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /internal/password_reset/request [post]
func PasswordResetRequestPost(c *gin.Context) {
	req := PasswordResetRequestRequest{}
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

	configAny := c.MustGet("config")
	cfg, ok := configAny.(*config.Config)
	if !ok || cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config not available"})
		return
	}

	q := query.Use(db)

	// ユーザーをexternal_emailで検索
	user, err := q.User.Where(q.User.ExternalEmail.Eq(req.Email)).First()
	if err != nil {
		// ユーザーが見つからない場合でも、情報漏洩防止のため成功応答を返す
		c.JSON(http.StatusOK, PasswordResetRequestResponse{
			Message: "パスワードリセットメールを送信しました。メールを確認してください。",
		})
		return
	}

	if user == nil {
		// ユーザーが見つからない場合
		c.JSON(http.StatusOK, PasswordResetRequestResponse{
			Message: "パスワードリセットメールを送信しました。メールを確認してください。",
		})
		return
	}

	// 検証コードを生成
	code := generatePasswordResetCode()

	// email_verification_codesテーブルに保存
	expiresAt := time.Now().Add(10 * time.Minute)
	verificationCode := &model.EmailVerificationCode{
		ID:          generateID(),
		UserID:      user.ID,
		Code:        code,
		ExpiresAt:   expiresAt,
		RequestType: "password_reset",
		CreatedAt:   time.Now(),
	}

	if err := q.EmailVerificationCode.Create(verificationCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate verification code"})
		return
	}

	// MailServer APIを呼び出してメール送信
	mailErr := sendPasswordResetEmail(c, user, code, cfg)
	if mailErr != nil {
		// メール送信失敗時はログするが、ユーザーへは成功応答を返す（情報漏洩防止）
		fmt.Printf("Failed to send password reset email: %v\n", mailErr)
	}

	c.JSON(http.StatusOK, PasswordResetRequestResponse{
		Message: "パスワードリセットメールを送信しました。メールを確認してください。",
	})
}

// generatePasswordResetCode 数字とアルファベットを含む検証コード（6文字）を生成
func generatePasswordResetCode() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	code := make([]byte, 6)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

// generateID ULIDを使用してユニークなIDを生成
func generateID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}

// sendPasswordResetEmail MailServer APIを呼び出してパスワードリセットメールを送信
func sendPasswordResetEmail(c *gin.Context, user *model.User, code string, config *config.Config) error {
	// MailServer APIのURLを構築
	mailServerURL := fmt.Sprintf("%s/password-reset", config.MailServerURL)

	// ユーザープロフィール取得（名前を取得するため）
	dbAny := c.MustGet("db")
	db, ok := dbAny.(*gorm.DB)
	if !ok || db == nil {
		return fmt.Errorf("database not available")
	}

	q := query.Use(db)
	profile, err := q.Profile.Where(q.Profile.UserID.Eq(user.ID)).First()
	if err != nil {
		return err
	}

	displayName := ""
	if profile != nil {
		displayName = profile.DisplayName
	}

	// リクエストボディを作成
	payload := map[string]string{
		"email": user.ExternalEmail,
		"name":  displayName,
		"code":  code,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// HTTPリクエストを作成
	resp, err := http.Post(mailServerURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send request to mail server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mail server returned status %d", resp.StatusCode)
	}

	return nil
}
