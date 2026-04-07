#!/bin/sh

if [ ! -f ./cmd/server/main.go ]; then
  echo "Please run this script from the src/ directory."
  exit 1
fi

export COMMIT=$(git rev-parse --short HEAD)
export BRANCH=$(git branch --show-current)

swag i -g cmd/server/main.go

# if --dev flag is provided, set gin to debug mode
if [ "$1" = "--dev" ]; then
  export DB_DSN="root:rootpass123@tcp(localhost:3306)/devdb?parseTime=true"
  export GIN_MODE=debug
  go run -ldflags "\
  -X unibot/internal/config.GitCommit=$COMMIT \
  -X unibot/internal/config.GitBranch=$BRANCH" \
  cmd/server/main.go
else
  export VERSION=$(git describe --tags --abbrev=0)
  export GIN_MODE=release
  go build -ldflags "\
  -X unibot/internal/config.Version=$VERSION \
  -X unibot/internal/config.GitCommit=$COMMIT \
  -X unibot/internal/config.GitBranch=$BRANCH" \
  cmd/server/main.go
fi
