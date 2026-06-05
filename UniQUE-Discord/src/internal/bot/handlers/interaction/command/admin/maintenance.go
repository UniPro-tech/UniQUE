package admin

import (
	"github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/interaction/command/admin/maintenance"

	"github.com/disgoorg/disgo/discord"
)

func LoadMaintenanceCommandContext() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "maintenance",
		Description: "メンテナンス用コマンド",
		Options: []discord.ApplicationCommandOption{
			maintenance.LoadShutdownCommandContext(),
		},
	}
}
