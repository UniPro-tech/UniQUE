#!/bin/bash
set -e

# 1. git clone
git clone https://github.com/UniPro-tech/UniQUE.git
cd UniQUE

# 2. 最新タグ取得
latest_tag=$(git describe --tags --abbrev=0)
git checkout "$latest_tag"

echo "Checked out to latest tag: $latest_tag"

cd UniQUE-DB

# 3. migration up
export MIGRATIONS_DIR="./migrations"
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up