#!/bin/sh
# {{.Name}} installer for macOS and Linux:
#   curl -fsSL {{.SiteURL}}/install.sh | sh
# curl doesn't quarantine what it downloads, so Gatekeeper never blocks the unsigned app.
# Running it again updates the launcher and keeps accounts, settings and worlds.
set -eu

releases={{.ReleasesURL | sh}}
name={{.Name | sh}}
slug={{.Slug | sh}}

say() { printf '\033[1;32m==>\033[0m %s\n' "$*"; }
die() { printf '\033[1;31merror:\033[0m %s\n' "$*" >&2; exit 1; }

install_macos() {
  tmp="$(mktemp -d)"
  trap 'hdiutil detach -quiet "$tmp/mnt" 2>/dev/null; rm -rf "$tmp"' EXIT
  say "Downloading $name..."
  curl -fL --progress-bar "$releases/$slug-macos.dmg" -o "$tmp/$slug.dmg"
  hdiutil attach -quiet -nobrowse -readonly -mountpoint "$tmp/mnt" "$tmp/$slug.dmg"

  dest=/Applications
  [ -w "$dest" ] || { dest="$HOME/Applications"; mkdir -p "$dest"; }
  rm -rf "$dest/$name.app"
  ditto "$tmp/mnt/$name.app" "$dest/$name.app"

  say "Installed $dest/$name.app. Opening it..."
  say "Hold Option while opening the app to get the launcher window instead of the game."
  open "$dest/$name.app"
}

install_linux() {
  if [ -e /etc/NIXOS ]; then
    say "NixOS detected, this won't run here. Use: nix run "{{.NixRef}}
    exit 0
  fi
  [ "$(uname -m)" = x86_64 ] || die "Only x86_64 Linux is supported for now."

  data="${XDG_DATA_HOME:-$HOME/.local/share}"
  app="$data/$slug-launcher" # the player's data stays in $data/$slug
  bin="$HOME/.local/bin"
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  say "Downloading $name..."
  curl -fL --progress-bar "$releases/$slug-linux.tar.gz" | tar xz -C "$tmp"
  mkdir -p "$bin" "$data/applications"
  rm -rf "$app"
  mv "$tmp/$slug" "$app"

  ln -sf "$app/launch" "$bin/$slug"
  cat > "$data/applications/$slug.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=$name
Comment=Play $name
Exec="$app/launch"
Icon=$app/freesm/share/icons/hicolor/256x256/apps/org.freesmlauncher.FreesmLauncher.png
Categories=Game;
Actions=launcher;

[Desktop Action launcher]
Name=Open the launcher
Exec="$app/launch" --launcher
EOF

  say "Installed. Find \"$name\" in your app menu, or run $slug ($slug --launcher for the launcher window)."
  case ":$PATH:" in
    *":$bin:"*) ;;
    *) say "$bin isn't on your PATH, so the $slug command needs the full path: $app/launch" ;;
  esac
  say "Starting the game..."
  nohup "$app/launch" >/dev/null 2>&1 &
}

# Everything above only defines functions, so a cut-off download runs nothing
case "$(uname -s)" in
  Darwin) install_macos ;;
  Linux) install_linux ;;
  *) die "On Windows? Use the PowerShell command from {{.SiteURL}}" ;;
esac
