#!/usr/bin/env bash
set -e

cd "$(dirname "$0")/.."

./scripts/build-wasm.sh

cp chip-8.wasm examples/chip-8.wasm
cp chip-8.wasm docs/static/assets/chip-8.wasm
cp wasm_exec.js examples/wasm_exec.js
cp wasm_exec.js docs/static/scripts/wasm_exec.js

rm chip-8.wasm
rm wasm_exec.js

echo "Copied chip-8.wasm and wasm_exec.js to examples/ and docs/static/"

