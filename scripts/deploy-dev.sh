#!/usr/bin/env bash
# Deploy the extension for live development.
#
# 1. Removes any previously installed .vsix-based extension (after confirmation).
# 2. Builds the specs CLI binary into extension/bin/.
# 3. Compiles the TypeScript extension source.
# 4. Symlinks the extension folder into ~/.vscode/extensions/ so that changes
#    in this repo are picked up immediately (after a window reload).
#
# Usage: scripts/deploy-dev.sh
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
ext_dir="$repo_root/extension"
ext_id="jdehaan.specs"
vscode_ext_dir="$HOME/.vscode/extensions"

# ---------------------------------------------------------------------------
# 1. Remove old extension installations (ask before each delete)
# ---------------------------------------------------------------------------
echo "==> Checking for old extension installations…"

found_old=false
for dir in "$vscode_ext_dir/$ext_id"* ; do
  [[ -e "$dir" ]] || continue
  # Skip if it is already our dev symlink
  if [[ -L "$dir" ]]; then
    echo "  (skipping symlink: $dir)"
    continue
  fi
  found_old=true
  read -r -p "  Remove $dir ? [y/N] " ans
  case "$ans" in
    [yY]|[yY][eE][sS])
      rm -rf "$dir"
      echo "  Removed."
      ;;
    *)
      echo "  Kept."
      ;;
  esac
done

if ! $found_old; then
  echo "  No old installations found."
fi

# ---------------------------------------------------------------------------
# 2. Build the specs CLI binary
# ---------------------------------------------------------------------------
echo "==> Building specs binary…"
bin_dir="$ext_dir/bin"
mkdir -p "$bin_dir"
(cd "$repo_root/cli" && go build -o "$bin_dir/specs" ./cmd/specs)
echo "  Built: $bin_dir/specs"

# ---------------------------------------------------------------------------
# 3. Compile the extension TypeScript
# ---------------------------------------------------------------------------
echo "==> Compiling extension…"
(cd "$ext_dir" && npm ci --silent && npm run compile)
echo "  Done."

# ---------------------------------------------------------------------------
# 4. Symlink into VS Code extensions folder
# ---------------------------------------------------------------------------
link_target="$vscode_ext_dir/$ext_id"

if [[ -L "$link_target" ]]; then
  current="$(readlink -f "$link_target")"
  if [[ "$current" == "$ext_dir" ]]; then
    echo "==> Symlink already in place: $link_target -> $ext_dir"
  else
    echo "==> Updating symlink (was: $current)"
    ln -sfn "$ext_dir" "$link_target"
    echo "  $link_target -> $ext_dir"
  fi
elif [[ -e "$link_target" ]]; then
  read -r -p "  $link_target exists and is not a symlink. Remove and replace? [y/N] " ans
  case "$ans" in
    [yY]|[yY][eE][sS])
      rm -rf "$link_target"
      ln -sfn "$ext_dir" "$link_target"
      echo "  $link_target -> $ext_dir"
      ;;
    *)
      echo "  Aborted. Symlink not created."
      exit 1
      ;;
  esac
else
  mkdir -p "$vscode_ext_dir"
  ln -sfn "$ext_dir" "$link_target"
  echo "==> Symlinked: $link_target -> $ext_dir"
fi

echo ""
echo "Done. Reload your VS Code window (Ctrl+Shift+P → 'Developer: Reload Window')"
echo "to activate the development extension."
