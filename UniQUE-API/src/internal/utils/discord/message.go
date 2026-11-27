package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/logging"
	"github.com/UniPro-tech/UniQUE-API/internal/query"
	"gorm.io/gorm"
)

func SendDirectMessage(ctx context.Context, userID string, content string, db *gorm.DB, config *config.Config) error {
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

	httpClient := &http.Client{Timeout: 10 * time.Second}

	// 1) DMチャンネルを作成 (or fetch)
	channelReqBody := map[string]string{"recipient_id": externalIdentity.ExternalUserID}
	channelBody, err := json.Marshal(channelReqBody)
	if err != nil {
		return err
	}
	channelReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://discord.com/api/%s/users/@me/channels", config.DiscordApiVersion), bytes.NewReader(channelBody))
	if err != nil {
		return err
	}
	logging.Inject(ctx, channelReq.Header)
	channelReq.Header.Set("Authorization", "Bot "+botToken)
	channelReq.Header.Set("Content-Type", "application/json")

	channelResp, err := httpClient.Do(channelReq)
	if err != nil {
		return err
	}
	defer channelResp.Body.Close()

	channelResponseBody := readDiscordBody(channelResp.Body)
	logDiscordResponse("create_dm_channel", channelResp.StatusCode, channelResponseBody,
		slog.String("discord_user_id", strings.ReplaceAll(strings.ReplaceAll(externalIdentity.ExternalUserID, "\n", ""), "\r", "")),
	)

	if channelResp.StatusCode != http.StatusOK && channelResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create DM channel, status code: %d", channelResp.StatusCode)
	}

	var channelRes struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(channelResponseBody, &channelRes); err != nil {
		return err
	}

	// 2) メッセージを送信
	msgReqBody := map[string]string{"content": content}
	msgBody, err := json.Marshal(msgReqBody)
	if err != nil {
		return err
	}
	msgURL := fmt.Sprintf("https://discord.com/api/%s/channels/%s/messages", config.DiscordApiVersion, channelRes.ID)
	msgReq, err := http.NewRequestWithContext(ctx, "POST", msgURL, bytes.NewReader(msgBody))
	if err != nil {
		return err
	}
	logging.Inject(ctx, msgReq.Header)
	msgReq.Header.Set("Authorization", "Bot "+botToken)
	msgReq.Header.Set("Content-Type", "application/json")

	msgResp, err := httpClient.Do(msgReq)
	if err != nil {
		return err
	}
	defer msgResp.Body.Close()

	msgResponseBody := readDiscordBody(msgResp.Body)
	logDiscordResponse("send_dm", msgResp.StatusCode, msgResponseBody,
		slog.String("discord_user_id", strings.ReplaceAll(strings.ReplaceAll(externalIdentity.ExternalUserID, "\n", ""), "\r", "")),
		slog.String("channel_id", strings.ReplaceAll(strings.ReplaceAll(channelRes.ID, "\n", ""), "\r", "")),
	)

	if msgResp.StatusCode != http.StatusOK && msgResp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send DM, status code: %d", msgResp.StatusCode)
	}
	return nil
}

// SendWelcomeMessage は承認時に送るウェルカムメッセージを組み立てて送信するヘルパー
func SendWelcomeMessage(ctx context.Context, userID string, email string, password string, displayName string, db *gorm.DB, config *config.Config) error {
	if displayName == "" {
		displayName = "メンバー"
	}
	message := fmt.Sprintf("# 🎉 %sさん、UniProjectへようこそ！\n\nメンバー登録が承認されました。\n## Welcomeページについて\nサークルに参加いただきありがとうございます！\nまずはこちらのページをお読みください。\nhttps://wiki.uniproject.jp/6a9d1fde55037e9c944c61a5\n## メールアドレスについて\n自由にお使いいただけるメールです。詳しくは[こちらのWiki](https://wiki.uniproject.jp/ツール/すべてのメンバー向け/メール)をご覧ください。\nメールアドレス: %s\nパスワード: %s\n\n今後ともよろしくお願いします！", displayName, email, password)
	return SendDirectMessage(ctx, userID, message, db, config)
}
