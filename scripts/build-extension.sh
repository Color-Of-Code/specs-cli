#!/usr/bin/env bash
# Build a per-platform .vsix that bundles the matching specs binary.
#
# Usage: scripts/build-extension.sh <vsce-target>
#   <vsce-target>: linux-x64 | linux-arm64 | darwin-x64 | darwin-arm64 | win32-x64
#
# Expects the matching specs binary to be present at:
#   dist/specs_<goos>_<goarch>/specs[.exe]
# (this matches GoReleaser's default layout).
set -euo pipefail

target="${1:?usage: $0 <vsce-target>}"
case "$target" in
  linux-x64)    goos=linux  goarch=amd64 exe=specs ;;
  linux-arm64)  goos=linux  goarch=arm64 exe=specs ;;
  darwin-x64)   goos=darwin goarch=amd64 exe=specs ;;
  darwin-arm64) goos=darwin goarch=arm64 exe=specs ;;
  win32-x64)    goos=windows goarch=amd64 exe=specs.exe ;;
  *) echo "unknown target: $target" >&2; exit 2 ;;
esac

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
ext_dir="$repo_root/extension"
bin_dir="$ext_dir/bin"

# Locate the binary. First try GoReleaser dist; fall back to a host build
# (for local testing of the current platform only).
src=""
for candidate in \
  "$repo_root/dist/specs_${goos}_${goarch}/$exe" \
  "$repo_root/dist/specs_${goos}_${goarch}_v1/$exe" \
  "$repo_root/cli/dist/specs_${goos}_${goarch}/$exe" \
  "$repo_root/cli/dist/specs_${goos}_${goarch}_v1/$exe"; do
  if [[ -f "$candidate" ]]; then
    src="$candidate"
    break
  fi
done
if [[ -z "$src" ]]; then
  echo "no GoReleaser binary found for $target; falling back to 'go build'" >&2
  GOOS="$goos" GOARCH="$goarch" go -C "$repo_root/cli" build -o "$bin_dir/$exe" ./cmd/specs
else
  mkdir -p "$bin_dir"
  cp "$src" "$bin_dir/$exe"
fi
chmod +x "$bin_dir/$exe" 2>/dev/null || true

# Vendor mermaid.min.js for the webview (kept out of git; downloaded each build).
mermaid_dst="$ext_dir/media/mermaid.min.js"
if [[ ! -f "$mermaid_dst" ]]; then
  echo "Fetching mermaid.min.js" >&2
  curl -sSLf https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js -o "$mermaid_dst"
fi

# Ensure node deps + compile.
( cd "$ext_dir" && npm ci --silent && npm run compile )

# Sync extension version to SPECS_VERSION (if set; CI passes the tag-derived value).
if [[ -n "${SPECS_VERSION:-}" ]]; then
  ( cd "$ext_dir" && npm version --no-git-tag-version --allow-same-version "$SPECS_VERSION" >/dev/null )
fi

# Package.
mkdir -p "$repo_root/dist"
( cd "$ext_dir" && npx --yes @vscode/vsce package --target "$target" --no-dependencies -o "$repo_root/dist/specs-${target}.vsix" )

echo "built: dist/specs-${target}.vsix"
