COMMIT=$(git rev-parse --short HEAD)
BRANCH=$(git branch --show-current)

if [[ "$*" == *"--dev"* ]]; then
  go run -ldflags "\
-X github.com/UniPro-tech/UniQUE/Discord/internal.GitCommit=$COMMIT \
-X github.com/UniPro-tech/UniQUE/Discord/internal.GitBranch=$BRANCH" \
cmd/bot/main.go
else
 VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "latest")

  go build -ldflags "\
-X github.com/UniPro-tech/UniQUE/Discord/internal.Version=$VERSION \
-X github.com/UniPro-tech/UniQUE/Discord/internal.GitCommit=$COMMIT \
-X github.com/UniPro-tech/UniQUE/Discord/internal.GitBranch=$BRANCH" \
cmd/bot/main.go
fi