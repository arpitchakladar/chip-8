#!/usr/bin/env bash

set -euo pipefail

OUT_DIR="${1:-examples}"
OUT_FILE="${OUT_DIR}/wasm_exec.js"

if ! command -v go >/dev/null 2>&1; then
	echo "Error: Go is not installed or not in PATH"
	exit 1
fi

GOROOT=$(go env GOROOT)
SOURCE_FILE="${GOROOT}/lib/wasm/wasm_exec.js"

if [[ ! -f "$SOURCE_FILE" ]]; then
	SOURCE_FILE="${GOROOT}/misc/wasm/wasm_exec.js"
fi

if [[ ! -f "$SOURCE_FILE" ]]; then
	echo "Error: could not find wasm_exec.js in GOROOT ($GOROOT)"
	exit 1
fi

echo "Creating directory: ${OUT_DIR}..."
mkdir -p "$OUT_DIR"

echo "Copying wasm_exec.js from GOROOT..."
cp "$SOURCE_FILE" "$OUT_FILE"

echo "Successfully saved to ${OUT_FILE}"

