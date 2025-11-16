#!/bin/bash
set -e

rm -f go.work go.work.sum

echo "[+] Initializing Go workspace..."
go work init

echo "[+] Adding all local modules to the workspace..."
while IFS= read -r mod; do
    dir=$(dirname "$mod")
    # Skip submodules
    if [[ "$dir" == *"third_party"* ]]; then
        echo "[!] Skipping $dir (submodule)"
        continue
    fi
    go work use "$dir"
done < <(find . -name "go.mod" -not -path "./third_party/*")

echo "[+] Syncing workspace dependencies..."
go work sync

echo "[+] Go workspace synced successfully."
