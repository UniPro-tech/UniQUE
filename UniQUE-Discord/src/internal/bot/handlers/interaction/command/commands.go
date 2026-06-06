package command

import (
	"github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/interaction/command/admin"
	"github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/interaction/command/general"

	"github.com/disgoorg/disgo/discord"
)

var GeneralCommands = []discord.ApplicationCommandCreate{
	general.LoadPingCommandContext(),
	general.LoadAboutCommandContext(),
	general.LoadHelpCommandContext(),
}

var AdminCommands = []discord.ApplicationCommandCreate{
	admin.LoadMaintenanceCommandContext(),
}
