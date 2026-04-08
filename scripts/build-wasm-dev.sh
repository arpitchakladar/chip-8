#!/usr/bin/env bash
set -e

cd "$(dirname "$0")/.."

./scripts/build-wasm.sh

cp main.wasm examples/main.wasm
cp main.wasm docs/static/assets/main.wasm
cp wasm_exec.js examples/wasm_exec.js
cp wasm_exec.js docs/static/scripts/wasm_exec.js

rm main.wasm
rm wasm_exec.js

echo "Copied main.wasm and wasm_exec.js to examples/ and docs/static/"

