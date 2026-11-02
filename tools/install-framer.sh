#!/usr/bin/env bash
set -euo pipefail
# Ozon Tech Framer (Apache-2.0), pinned for reproducible request formats.
revision=ae45a6eb6a20f744f5fcc7d606e7a8c66c591d2c
destination="$1"
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
git clone --quiet https://github.com/ozontech/framer.git "$work/source"
git -C "$work/source" checkout --quiet "$revision"
# Course adapter: configurable request timeout, no unsolicited pprof listener,
# accurate phout timestamps and immediate flush for interrupted runs.
script_dir=$(cd "$(dirname "$0")" && pwd)
git -C "$work/source" apply "$script_dir/framer.patch"
(cd "$work/source" && GOWORK=off go build -trimpath -o "$work/framer" ./cmd/framer)
mkdir -p "$(dirname "$destination")"
cp "$work/framer" "$destination.tmp"
chmod 755 "$destination.tmp"
mv "$destination.tmp" "$destination"
