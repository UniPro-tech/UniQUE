package usercommand

import (
	"fmt"

	"github.com/UniPro-tech/UniQUE/Discord/internal"
	uniqueapi "github.com/UniPro-tech/UniQUE/Discord/internal/uniqueAPI"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func LoadGetUniQUEMenuContext() discord.UserCommandCreate {
	return discord.UserCommandCreate{
		Name: "UniQUEユーザーを取得",
	}
}

func GetUniQUE(ctx *internal.BotContext) func(data discord.UserCommandInteractionData, e *handler.CommandEvent) error {
	return func(data discord.UserCommandInteractionData, e *handler.CommandEvent) error {
		targetUser := data.TargetUser()
		uniqueUserInfo, err := uniqueapi.GetUserInfo(ctx, targetUser.ID)
		if err != nil {
			return err
		}
		if uniqueUserInfo == nil {
			responseEmbed := discord.Embed{
				Title:       "UniQUEユーザーが見つかりませんでした",
				Description: fmt.Sprintf("%s さんのUniQUEユーザーは見つかりませんでした。\n非メンバーもしくはまだ連携されていないかもしれません。", targetUser.Username),
				Color:       ctx.Config.Colors.Error,
				Footer: &discord.EmbedFooter{
					Text:    fmt.Sprintf("Requested by %s", e.User().Username),
					IconURL: e.User().EffectiveAvatarURL(),
				},
			}
			_, err = e.Client().Rest.CreateFollowupMessage(e.ApplicationID(), e.Token(),
				discord.NewMessageCreate().WithEmbeds(responseEmbed))
			return err
		} else {
			responseEmbed := discord.Embed{
				Title:       "UniQUEユーザーの取得に成功",
				Description: fmt.Sprintf("%s さんのUniQUEユーザーを取得しました。", targetUser.Username),
				Color:       ctx.Config.Colors.Success,
				Fields: []discord.EmbedField{
					{
						Name:  "ユーザーID",
						Value: fmt.Sprintf("`%s`", uniqueUserInfo.ID),
					},
					{
						Name:  "表示名",
						Value: uniqueUserInfo.Profile.DisplayName,
					},
					{
						Name:  "カスタムID",
						Value: fmt.Sprintf("`%s`", uniqueUserInfo.CustomID),
					},
					{
						Name:  "URL",
						Value: fmt.Sprintf("[UniQUE Frontend](%s/dashboard/members/%s)", ctx.Config.UniqueFrontendURL, uniqueUserInfo.ID),
					},
				},
				Footer: &discord.EmbedFooter{
					Text:    fmt.Sprintf("Requested by %s", e.User().Username),
					IconURL: e.User().EffectiveAvatarURL(),
				},
			}
			_, err = e.Client().Rest.CreateFollowupMessage(e.ApplicationID(), e.Token(),
				discord.NewMessageCreate().WithEmbeds(responseEmbed))
			return err
		}
	}
}
