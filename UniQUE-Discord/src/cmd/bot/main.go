package main

import (
	"context"
	"log"
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
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is not set")
	}

	ctxData := &internal.BotContext{
		Config: internal.LoadConfig(),
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
		log.Fatal("error while building disgo instance: ", err)
	}

	defer client.Close(context.TODO())

	// 接続開始
	if err = client.OpenGateway(context.TODO()); err != nil {
		log.Fatal("error while connecting to gateway: ", err)
	}

	log.Println("Bot is running...")

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
		log.Fatal("failed to register global commands: ", err)
	}
	guildID, parseErr := snowflake.Parse(ctxData.Config.AdminGuildID)
	if parseErr != nil {
		log.Fatal("invalid CONFIG_ADMIN_GUILD_ID: ", parseErr)
	}
	if _, err := client.Rest.SetGuildCommands(client.ApplicationID, guildID, adminCommands); err != nil {
		log.Fatal("failed to register admin commands: ", err)
	}

	// 終了待機
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
