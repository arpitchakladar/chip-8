#!/usr/bin/env bash
set -e

cd "$(dirname "$0")/.."

go build -o ./chip-8 ./cmd/chip-8

echo "Built chip-8 binary in project root"

