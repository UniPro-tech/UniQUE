package discord

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/config"
	"github.com/UniPro-tech/UniQUE-API/internal/logging"
)

func AddRoleToUser(ctx context.Context, externalUserID string, roleID string, cfg *config.Config) error {
	// Discord APIを使用してロールを追加
	botToken := cfg.DiscordConfig.BotToken
	guildId := cfg.DiscordConfig.Guild.ID

	if botToken == "" || guildId == "" {
		return fmt.Errorf("discord bot token or guild id not configured")
	}

	discordAPIEndpoint := fmt.Sprintf("https://discord.com/api/%s/guilds/%s/members/%s/roles/%s", cfg.DiscordApiVersion, guildId, externalUserID, roleID)
	httpClient := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "PUT", discordAPIEndpoint, nil)
	if err != nil {
		return err
	}
	logging.Inject(ctx, req.Header)
	req.Header.Set("Authorization", "Bot "+botToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body := readDiscordBody(resp.Body)
	logDiscordResponse("add_guild_role", resp.StatusCode, body,
		slog.String("guild_id", guildId),
		slog.String("discord_user_id", strings.ReplaceAll(strings.ReplaceAll(externalUserID, "\n", ""), "\r", "")),
		slog.String("role_id", strings.ReplaceAll(strings.ReplaceAll(roleID, "\n", ""), "\r", "")),
	)

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add role, status code: %d", resp.StatusCode)
	}
	return nil
}
