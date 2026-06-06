package general

import (
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/UniPro-tech/UniQUE/Discord/internal"
)

func LoadAboutCommandContext() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "about",
		Description: "ボットの情報を表示します",
	}
}

func About(ctx *internal.BotContext) func(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return func(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
		config := ctx.Config

		responseEmbed := discord.Embed{
			Title:       "About " + config.BotName + " 🤖",
			Description: config.Description,
			Color:       config.Colors.Primary,
			Fields: []discord.EmbedField{
				{
					Name:  "Version",
					Value: config.BotVersion,
				},
				{
					Name:  "GitHub",
					Value: config.GitHub,
				},
			},
			Footer: &discord.EmbedFooter{
				Text:    "Requested by " + e.User().Username,
				IconURL: e.User().EffectiveAvatarURL(),
			},
			Timestamp: func() *time.Time {
				t := time.Now()
				return &t
			}(),
		}

		_, err := e.Client().Rest.CreateFollowupMessage(e.ApplicationID(), e.Token(), discord.NewMessageCreate().WithEmbeds(responseEmbed))
		return err
	}
}
