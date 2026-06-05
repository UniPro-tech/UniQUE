package interaction_handler

import (
	"slices"
	"time"

	"github.com/UniPro-tech/UniQUE/Discord/internal"
	"github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/interaction/command/admin/maintenance"
	"github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/interaction/command/general"
	usercommand "github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/interaction/contextMenu/userCommand"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

func RegistHandler(r *handler.Mux, ctxData *internal.BotContext) {
	r.Route("/ping", func(r handler.Router) {
		r.Use(DeferReplyMiddleware(ctxData, false, false))
		r.SlashCommand("/", general.Ping(ctxData))
	})
	r.Route("/about", func(r handler.Router) {
		r.Use(DeferReplyMiddleware(ctxData, false, false))
		r.SlashCommand("/", general.About(ctxData))
	})
	r.Route("/help", func(r handler.Router) {
		r.Use(DeferReplyMiddleware(ctxData, false, false))
		r.SlashCommand("/", general.Help(ctxData))
	})
	r.Route("/maintenance", func(r handler.Router) {
		r.Use(DeferReplyMiddleware(ctxData, true, false))
		r.Use(AdminOnlyMiddleware(ctxData))
		r.SlashCommand("/shutdown", maintenance.Shutdown(ctxData))
	})
	r.Route("/UniQUEユーザーを取得",
		func(r handler.Router) {
			r.Use(DeferReplyMiddleware(ctxData, true, false))
			r.UserCommand("/", usercommand.GetUniQUE(ctxData))
		},
	)
}

func IsOwner(member discord.Member) bool {
	config := internal.LoadConfig()
	adminRoleID := config.AdminRoleID
	return slices.Contains(member.RoleIDs, snowflake.MustParse(adminRoleID))
}

func AdminOnlyMiddleware(ctx *internal.BotContext) func(next handler.Handler) handler.Handler {
	return func(next handler.Handler) handler.Handler {
		return func(e *handler.InteractionEvent) error {
			config := ctx.Config
			if !IsOwner(e.Member().Member) {
				errorEmbed := discord.Embed{
					Title:       "権限エラー",
					Description: "権限がありません。",
					Color:       config.Colors.Error,
					Footer: &discord.EmbedFooter{
						Text:    "Requested by " + *e.Member().Nick,
						IconURL: e.User().EffectiveAvatarURL(),
					},
					Timestamp: func() *time.Time {
						t := time.Now()
						return &t
					}(),
				}
				_, err := e.Client().Rest.CreateFollowupMessage(e.ApplicationID(), e.Token(), discord.NewMessageCreate().WithEmbeds(errorEmbed).WithEphemeral(true))
				return err
			}

			return next(e)
		}
	}
}

func DeferReplyMiddleware(ctx *internal.BotContext, ephemeral bool, update bool) func(next handler.Handler) handler.Handler {
	return func(next handler.Handler) handler.Handler {
		if !update {
			return func(e *handler.InteractionEvent) error {
				e.DeferCreateMessage(ephemeral)
				return next(e)
			}
		} else {
			return func(e *handler.InteractionEvent) error {
				e.DeferUpdateMessage()
				return next(e)
			}
		}
	}
}
