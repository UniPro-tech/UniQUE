package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"gorm.io/gorm"
)

func SendDirectMessage(userID string, content string, db *gorm.DB, config *config.Config) error {
	// ExternalIdentityを取得
	q := query.Use(db)
	externalIdentity, err := q.ExternalIdentity.Where(q.ExternalIdentity.Provider.Eq("discord"), q.ExternalIdentity.ExternalUserID.Eq(userID)).First()
	if err != nil {
		return err
	}
	// Botトークンを使ってDMを送信
	botToken := config.DiscordConfig.BotToken
	if botToken == "" {
		return fmt.Errorf("discord bot token not configured")
	}

	httpClient := &http.Client{}

	// 1) DMチャンネルを作成 (or fetch)
	channelReqBody := map[string]string{"recipient_id": externalIdentity.ExternalUserID}
	channelBody, err := json.Marshal(channelReqBody)
	if err != nil {
		return err
	}
	channelReq, err := http.NewRequest("POST", "https://discord.com/api/v10/users/@me/channels", bytes.NewReader(channelBody))
	if err != nil {
		return err
	}
	channelReq.Header.Set("Authorization", "Bot "+botToken)
	channelReq.Header.Set("Content-Type", "application/json")

	channelResp, err := httpClient.Do(channelReq)
	if err != nil {
		return err
	}
	defer channelResp.Body.Close()
	if channelResp.StatusCode != http.StatusOK && channelResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create DM channel, status code: %d", channelResp.StatusCode)
	}

	var channelRes struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(channelResp.Body).Decode(&channelRes); err != nil {
		return err
	}

	// 2) メッセージを送信
	msgReqBody := map[string]string{"content": content}
	msgBody, err := json.Marshal(msgReqBody)
	if err != nil {
		return err
	}
	msgURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelRes.ID)
	msgReq, err := http.NewRequest("POST", msgURL, bytes.NewReader(msgBody))
	if err != nil {
		return err
	}
	msgReq.Header.Set("Authorization", "Bot "+botToken)
	msgReq.Header.Set("Content-Type", "application/json")

	msgResp, err := httpClient.Do(msgReq)
	if err != nil {
		return err
	}
	defer msgResp.Body.Close()
	if msgResp.StatusCode != http.StatusOK && msgResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send DM, status code: %d", msgResp.StatusCode)
	}
	return nil
}

// SendWelcomeMessage は承認時に送るウェルカムメッセージを組み立てて送信するヘルパー
func SendWelcomeMessage(userID string, email string, password string, displayName string, db *gorm.DB, config *config.Config) error {
	if displayName == "" {
		displayName = "メンバー"
	}
	message := fmt.Sprintf("# 🎉 %sさん、UniProjectへようこそ！\n\nメンバー登録が承認されました。\n## メールアドレスについて\n自由にお使いいただけるメールです。詳しくは[こちらのWiki](https://wiki.uniproject.jp/Tools/メール)をご覧ください。\nメールアドレス: %s\nパスワード: %s\n\n今後ともよろしくお願いします！", displayName, email, password)
	return SendDirectMessage(userID, message, db, config)
}
