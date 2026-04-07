#!/bin/sh
set -e

echo "=== build start ==="
bun run build

echo "=== build done, start runtime ==="
exec bun run start