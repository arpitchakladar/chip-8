#!/usr/bin/env bash
set -e

cd "$(dirname "$0")/.."

GOARCH=wasm GOOS=js go build -o ./chip-8.wasm ./wasm/chip-8

./scripts/get-wasm-exec.sh ./

echo "Built main.wasm in project root"

