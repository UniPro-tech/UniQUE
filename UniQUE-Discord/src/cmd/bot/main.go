package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"github.com/UniPro-tech/UniQUE/Discord/internal"
	customHandlers "github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/event"
	interaction_handler "github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/interaction"
	"github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/interaction/command"
	contextmenu "github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/interaction/contextMenu"
	"github.com/UniPro-tech/UniQUE/Discord/internal/settings"
)

func main() {
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(logHandler).With(slog.String("component", "discord")))

	databaseSettings, err := settings.LoadValues()
	if err != nil {
		slog.Warn("settings_load_failed", slog.Any("error", err))
		databaseSettings = map[string]string{}
	}
	token := settings.Resolve(databaseSettings, "DISCORD_TOKEN", "discord.service_token", "")
	if token == "" {
		slog.Error("discord_token_missing")
		os.Exit(1)
	}

	ctxData := &internal.BotContext{
		Config: internal.LoadConfig(databaseSettings),
	}

	r := handler.New()

	interaction_handler.RegistHandler(r, ctxData)

	// disgo クライアントの構築
	client, err := disgo.New(token,
		//bot.WithDefaultGateway(),
		bot.WithGatewayConfigOpts(
			// Intents
			gateway.WithIntents(
				gateway.IntentsNonPrivileged,
				gateway.IntentMessageContent,
			),
		),
		// Event Listener
		bot.WithEventListenerFunc(func(e *events.Ready) {
			customHandlers.Ready(ctxData, e)
		}),
		// Cache
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagVoiceStates),
			cache.WithCaches(cache.FlagChannels),
			cache.WithCaches(cache.FlagMessages),
			cache.WithCaches(cache.FlagRoles),
			cache.WithCaches(cache.FlagMembers),
			cache.WithCaches(cache.FlagGuilds),
		),
		// Handler
		bot.WithEventListeners(r),
	)
	if err != nil {
		slog.Error("bot_build_failed", slog.Any("error", err))
		os.Exit(1)
	}

	defer client.Close(context.TODO())

	// 接続開始
	if err = client.OpenGateway(context.TODO()); err != nil {
		slog.Error("gateway_connect_failed", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("bot_running")

	// Slash Command の登録
	var generalCommands []discord.ApplicationCommandCreate
	for _, cmd := range command.GeneralCommands {
		generalCommands = append(generalCommands, cmd)
	}
	for _, cmd := range contextmenu.GeneralContextMenus {
		generalCommands = append(generalCommands, cmd)
	}

	var adminCommands []discord.ApplicationCommandCreate
	for _, cmd := range command.AdminCommands {
		adminCommands = append(adminCommands, cmd)
	}

	if _, err := client.Rest.SetGlobalCommands(client.ApplicationID, generalCommands); err != nil {
		slog.Error("global_commands_register_failed", slog.Any("error", err))
		os.Exit(1)
	}
	guildID, parseErr := snowflake.Parse(ctxData.Config.AdminGuildID)
	if parseErr != nil {
		slog.Error("admin_guild_id_invalid", slog.Any("error", parseErr))
		os.Exit(1)
	}
	if _, err := client.Rest.SetGuildCommands(client.ApplicationID, guildID, adminCommands); err != nil {
		slog.Error("admin_commands_register_failed", slog.Any("error", err))
		os.Exit(1)
	}

	// 終了待機
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
