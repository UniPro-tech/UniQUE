package contextmenu

import (
	usercommand "github.com/UniPro-tech/UniQUE/Discord/internal/bot/handlers/interaction/contextMenu/userCommand"
	"github.com/disgoorg/disgo/discord"
)

var GeneralContextMenus = []discord.ApplicationCommandCreate{
	usercommand.LoadGetUniQUEMenuContext(),
}
