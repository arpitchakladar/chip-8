#!/usr/bin/env bash

set -euo pipefail

# --- Configuration ---
# Use the first argument as the directory, default to "examples"
OUT_DIR="${1:-examples}"
OUT_FILE="${OUT_DIR}/wasm_exec.js"

# --- Validations ---

# Ensure Go is installed to use 'go env'
if ! command -v go >/dev/null 2>&1; then
	echo "Error: Go is not installed or not in PATH"
	exit 1
fi

# Determine the source path from GOROOT
# In newer Go versions, it's in lib/wasm/; in older ones, it was in misc/wasm/
GOROOT=$(go env GOROOT)
SOURCE_FILE="${GOROOT}/lib/wasm/wasm_exec.js"

# Fallback check for older Go versions (pre-1.24ish) where it lived in misc/
if [[ ! -f "$SOURCE_FILE" ]]; then
	SOURCE_FILE="${GOROOT}/misc/wasm/wasm_exec.js"
fi

# Final check if the source actually exists
if [[ ! -f "$SOURCE_FILE" ]]; then
	echo "Error: could not find wasm_exec.js in GOROOT ($GOROOT)"
	exit 1
fi

# --- Execution ---

echo "Creating directory: ${OUT_DIR}..."
mkdir -p "$OUT_DIR"

echo "Copying wasm_exec.js from GOROOT..."
cp "$SOURCE_FILE" "$OUT_FILE"

echo "Successfully saved to ${OUT_FILE}"

