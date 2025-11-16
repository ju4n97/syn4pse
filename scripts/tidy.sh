#!/usr/bin/env bash
set -e

echo "[+] Tidying all Go modules in the workspace..."

for d in $(go list -f '{{.Dir}}' -m); do
  (
    cd "$d"
    echo "[+] Tidying module in $d..."
    go mod tidy
  )
done

echo "[+] All modules tidied successfully."
