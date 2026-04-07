#!/bin/sh

set -eu

# Auth の RSA 鍵を生成し、権限を安全な状態に整える

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
KEYS_DIR="$SCRIPT_DIR/../keys"
PRIVATE_DIR="$KEYS_DIR/private"
PUBLIC_DIR="$KEYS_DIR/public"
PRIVATE_KEY="$PRIVATE_DIR/rsa_private.pem"
PUBLIC_KEY="$PUBLIC_DIR/rsa_public.pem"

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl が見つからないため鍵を生成できません。" >&2
  exit 1
fi

umask 077

mkdir -p "$PRIVATE_DIR" "$PUBLIC_DIR"
chmod 700 "$PRIVATE_DIR"
chmod 755 "$PUBLIC_DIR"

if [ ! -f "$PRIVATE_KEY" ] || [ ! -f "$PUBLIC_KEY" ]; then
  echo "RSA 鍵を生成します。"
  openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out "$PRIVATE_KEY"
  openssl rsa -pubout -in "$PRIVATE_KEY" -out "$PUBLIC_KEY"
fi

# 権限を固定
chmod 600 "$PRIVATE_KEY"
chmod 644 "$PUBLIC_KEY"
