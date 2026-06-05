COMMIT=$(git rev-parse --short HEAD)
BRANCH=$(git branch --show-current)

if [[ "$*" == *"--dev"* ]]; then
  export DISCORD_TOKEN="your_token_here"
  export CONFIG_ADMIN_GUILD_ID="your_guild_id_here"
  export CONFIG_ADMIN_ROLE_ID="your_role_id_here"

  go run -ldflags "\
-X github.com/UniPro-tech/UniQUE/Discord/internal.GitCommit=$COMMIT \
-X github.com/UniPro-tech/UniQUE/Discord/internal.GitBranch=$BRANCH" \
cmd/bot/main.go
else
  VERSION=$(git describe --tags --abbrev=0)

  go build -ldflags "\
-X github.com/UniPro-tech/UniQUE/Discord/internal.Version=$VERSION \
-X github.com/UniPro-tech/UniQUE/Discord/internal.GitCommit=$COMMIT \
-X github.com/UniPro-tech/UniQUE/Discord/internal.GitBranch=$BRANCH" \
cmd/bot/main.go
fi