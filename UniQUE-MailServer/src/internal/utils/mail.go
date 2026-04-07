package utils

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"strings"

	"github.com/UniPro-tech/UniQUE-MailServer/internal/config"
)

var smtpConfig *config.SmtpConfig

// InitMailer はSMTP設定を保持する。main起動時に呼び出すこと。
func InitMailer(cfg *config.SmtpConfig) {
	smtpConfig = cfg
}

// SendMail はHTML/プレーンテキスト両方を含むメールを送信する。
func SendMail(html, text, subject, to string) error {
	if smtpConfig == nil {
		return fmt.Errorf("mailer not initialized: call InitMailer first")
	}

	addr := fmt.Sprintf("%s:%d", smtpConfig.Host, smtpConfig.Port)
	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)

	// 【App.name】を含める
	subject = fmt.Sprintf("【%s】 %s", "UniQUE", subject)

	from := fmt.Sprintf("%s <%s>", smtpConfig.FromName, smtpConfig.From)

	msg := buildMIMEMessage(from, to, subject, html, text)

	if smtpConfig.Secure {
		return sendWithImplicitTLS(addr, auth, to, msg)
	}
	return smtp.SendMail(addr, auth, smtpConfig.From, []string{to}, msg)
}

// buildMIMEMessage はmultipart/alternative形式のMIMEメッセージを構築する。
func buildMIMEMessage(from, to, subject, html, text string) []byte {
	const boundary = "==UniQUE-MailServer-Boundary=="

	var msg strings.Builder

	// ヘッダー
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject)))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")

	// プレーンテキストパート
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	msg.WriteString(encodeBase64Wrapped([]byte(text)))
	msg.WriteString("\r\n")

	// HTMLパート
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	msg.WriteString(encodeBase64Wrapped([]byte(html)))
	msg.WriteString("\r\n")

	// 終端バウンダリ
	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return []byte(msg.String())
}

// encodeBase64Wrapped はBase64エンコード後に76文字ごとに改行を挿入する (RFC 2045準拠)。
func encodeBase64Wrapped(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	var wrapped strings.Builder
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		wrapped.WriteString(encoded[i:end])
		wrapped.WriteString("\r\n")
	}
	return wrapped.String()
}

// sendWithImplicitTLS は暗黙的TLS (ポート465等) でメールを送信する。
func sendWithImplicitTLS(addr string, auth smtp.Auth, to string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName: smtpConfig.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpConfig.Host)
	if err != nil {
		return fmt.Errorf("SMTP new client: %w", err)
	}
	defer client.Close()

	if err = client.Hello("localhost"); err != nil {
		return fmt.Errorf("SMTP hello: %w", err)
	}
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth: %w", err)
	}
	if err = client.Mail("<" + smtpConfig.From + ">"); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("SMTP write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("SMTP data close: %w", err)
	}

	return client.Quit()
}
