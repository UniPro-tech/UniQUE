package templates

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"time"
)

//go:embed registration_verification_email.html
var registrationVerificationEmailHTML string

//go:embed email_change_verification.html
var emailChangeVerificationHTML string

//go:embed password_reset_verification.html
var passwordResetVerificationHTML string

// RenderRegistrationVerificationHTML renders the registration verification email HTML and a plain-text fallback.
func RenderRegistrationVerificationHTML(appName, baseURL, code, name, copyrightName string) (html string, text string, err error) {
	tmpl, err := template.New("registration_verification_email").Parse(registrationVerificationEmailHTML)
	if err != nil {
		return "", "", err
	}

	// VerifyURL を組み立ててテンプレート側に渡す
	verifyURL := fmt.Sprintf("%s/email-verify?code=%s", baseURL, code)
	data := map[string]interface{}{
		"AppName":       appName,
		"CopyrightName": copyrightName,
		"VerifyURL":     verifyURL,
		"Year":          time.Now().Year(),
		"Name":          name,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", err
	}

	text = fmt.Sprintf("メールアドレスの確認をお願いします。\n\nご登録ありがとうございます。以下のリンクをクリックしてください。\n【重要】Discordアカウントの連携を事前に済ませてください。\n\n%s/email-verify?code=%s", baseURL, code)
	return buf.String(), text, nil
}

// RenderEmailChangeVerificationHTML renders the email-change verification HTML and a plain-text fallback.
func RenderEmailChangeVerificationHTML(appName, baseURL, code, copyrightName string) (html string, text string, err error) {
	tmpl, err := template.New("email_change_verification").Parse(emailChangeVerificationHTML)
	if err != nil {
		return "", "", err
	}

	data := map[string]interface{}{
		"AppName":       appName,
		"BaseURL":       baseURL,
		"Code":          code,
		"Year":          time.Now().Year(),
		"CopyrightName": copyrightName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", err
	}

	text = fmt.Sprintf("メールアドレスの確認をお願いします:\n%s/email-verify?code=%s", baseURL, code)
	return buf.String(), text, nil
}

// RenderPasswordResetVerificationHTML renders the password reset verification HTML and a plain-text fallback.
func RenderPasswordResetVerificationHTML(appName, baseURL, code, name, copyrightName string) (html string, text string, err error) {
	tmpl, err := template.New("password_reset_verification").Parse(passwordResetVerificationHTML)
	if err != nil {
		return "", "", err
	}

	resetURL := fmt.Sprintf("%s/password-reset?code=%s", baseURL, code)
	data := map[string]interface{}{
		"AppName":       appName,
		"CopyrightName": copyrightName,
		"ResetURL":      resetURL,
		"Year":          time.Now().Year(),
		"Name":          name,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", err
	}

	text = fmt.Sprintf("パスワードをリセットするには、以下のリンクをクリックしてください。\n\n%s/password-reset?code=%s\n\nこのリンクの有効期限は10分です。", baseURL, code)
	return buf.String(), text, nil
}
