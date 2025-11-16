#!/usr/bin/env bash
set -e

echo "[+] Pruning cache..."
go clean -cache -modcache -testcache
bun pm cache rm
echo "[+] Cache pruned successfully."