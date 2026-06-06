package event_handlers

import (
	"context"
	"log"

	"github.com/UniPro-tech/UniQUE/Discord/internal"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
)

func Ready(ctx *internal.BotContext, e *events.Ready) {
	log.Println("Bot is ready 🚀")
	log.Printf("Logged in as: %v#%v", e.User.Username, e.User.Discriminator)
	resetBotStatus(e.Client())
}

func resetBotStatus(client *bot.Client) error {
	return client.SetPresence(context.Background(), gateway.PresenceOpt(func(p *gateway.MessageDataPresenceUpdate) {
		p.Activities = []discord.Activity{
			{
				Type: discord.ActivityTypeGame,
				Name: "混沌としたUniProを管理中... | /help",
			},
		}
		p.Status = discord.OnlineStatusOnline
		p.AFK = false
	}))
}
