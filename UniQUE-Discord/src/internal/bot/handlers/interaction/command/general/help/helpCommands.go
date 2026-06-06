package help

type HelpCommand struct {
	Name        string
	Description string
	Usage       string
}

var HelpCommands = []HelpCommand{
	{
		Name:        "/help",
		Description: "コマンドのヘルプを表示します。",
		Usage:       "テキストチャンネルで/help <command>",
	},
	{
		Name:        "/ping",
		Description: "スピードテストを行います。",
		Usage:       "テキストチャンネルで/ping",
	},
	{
		Name:        "/about",
		Description: "このボットの情報を表示します。",
		Usage:       "テキストチャンネルで/about",
	},
}
