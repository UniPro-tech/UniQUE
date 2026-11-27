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

func AddToGuild(ctx context.Context, externalUserID string, db *gorm.DB, cfg *config.Config) error {
	// ExternalIdentityを取得
	q := query.Use(db)
	externalIdentity, err := q.ExternalIdentity.Where(q.ExternalIdentity.Provider.Eq("discord"), q.ExternalIdentity.ExternalUserID.Eq(externalUserID)).First()
	if err != nil {
		return err
	}

	// Discord APIを使用してギルドに追加
	accessToken := externalIdentity.AccessToken
	botToken := cfg.DiscordConfig.BotToken
	guildId := cfg.DiscordConfig.Guild.ID
	if botToken == "" || guildId == "" {
		return fmt.Errorf("discord bot token or guild id not configured")
	}

	discordAPIEndpoint := fmt.Sprintf("https://discord.com/api/%s/guilds/%s/members/%s", cfg.DiscordApiVersion, guildId, externalUserID)
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// body: { "access_token": "<user access token>" }
	bodyObj := map[string]string{"access_token": accessToken}
	bodyBytes, err := json.Marshal(bodyObj)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", discordAPIEndpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	logging.Inject(ctx, req.Header)
	req.Header.Set("Authorization", "Bot "+botToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body := readDiscordBody(resp.Body)
	logDiscordResponse("add_guild_member", resp.StatusCode, body,
		slog.String("guild_id", guildId),
		slog.String("discord_user_id", strings.ReplaceAll(strings.ReplaceAll(externalUserID, "\n", ""), "\r", "")),
	)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to add to guild, status code: %d", resp.StatusCode)
	}
	return nil
}
