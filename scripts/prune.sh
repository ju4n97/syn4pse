#!/usr/bin/env bash
set -e

echo "[+] Pruning workspace with artifacts and ignore list..."

# Directories to ignore (relative to repo root)
IGNORE_DIRS=(relic)

# Remove Go workspace files
rm -f go.work go.work.sum
echo "[+] Removed go.work and go.work.sum"

# Build the prune arguments
PRUNE_ARGS=()
for dir in "${IGNORE_DIRS[@]}"; do
  PRUNE_ARGS+=(-path "./$dir" -prune -o)
done

# Run find safely with dynamic prune and delete patterns
find . \
  "${PRUNE_ARGS[@]}" \
  \( \
    -type d -empty \
    -o -type d -name 'node_modules' \
    -o -type d -name 'dist' \
    -o -type d -name 'build' \
    -o -type d -name 'cache' \
    -o -type d -name 'coverage' \
    -o -type d -name 'bin' \
    -o -type d -name 'pkg' \
    -o -type d -name 'pb' \
    -o -type d -name '.task' \
    -o -type f -name '*.log' \
    -o -type f -name '*.tmp' \
    -o -type f -name '*~' \
    -o -type f -name '*.exe' \
    -o -type f -name '*.out' \
    -o -type f -name '*.o' \
    -o -type f -name '*.so' \
    -o -type f -name '*.a' \
    -o -type f -name '*.dll' \
    -o -name 'uv.lock' \
    -o -name 'bun.lock' \
    -o -name 'yarn.lock' \
    -o -name 'package-lock.json' \
    -o -name 'pnpm-lock.yaml' \
    -o -name 'go.sum' \
  \) \
  -exec rm -rf {} +

echo "[+] Workspace fully pruned."
