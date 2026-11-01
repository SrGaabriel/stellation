#!/usr/bin/env bash
# Seeds a per-user Freesm Launcher data dir with the pack's instance on first run, then starts it.
#   STELLATION_LAUNCHER  Freesm Launcher binary
#   STELLATION_TEMPLATE  dir holding freesmlauncher.cfg and instances/
#   STELLATION_INSTANCE  instance name (its folder under instances/)
#   STELLATION_DATA      per-user data dir
#   STELLATION_JAVA      optional: pin this Java instead of the launcher's auto-download (NixOS can't run downloaded JREs)
#   STELLATION_PACK_URL  optional: install from this pack.toml instead (testing; persists in this data dir)
# Usage: run.sh [--launcher | launcher args...]
set -euo pipefail

inst="$STELLATION_DATA/instances/$STELLATION_INSTANCE"
mkdir -p "$STELLATION_DATA/instances"

if [ ! -e "$STELLATION_DATA/freesmlauncher.cfg" ]; then
  cp "$STELLATION_TEMPLATE/freesmlauncher.cfg" "$STELLATION_DATA/"
  chmod u+w "$STELLATION_DATA/freesmlauncher.cfg"
fi
if [ ! -d "$inst" ]; then
  cp -R "$STELLATION_TEMPLATE/instances/$STELLATION_INSTANCE" "$inst"
  chmod -R u+w "$inst"
fi
# Always refresh the bootstrap jar so a new launcher build can ship a new one
cp -f "$STELLATION_TEMPLATE/instances/$STELLATION_INSTANCE/minecraft/packwiz-installer-bootstrap.jar" "$inst/minecraft/"
chmod u+w "$inst/minecraft/packwiz-installer-bootstrap.jar"

# set_key FILE KEY VALUE: replace KEY in [General], or append it to that section.
# Values go through ENVIRON because awk -v would eat backslashes.
set_key() {
  local tmp
  tmp="$(mktemp)"
  K="$2" V="$3" awk '
    BEGIN { k = ENVIRON["K"]; v = ENVIRON["V"] }
    /^\[/ { if (sec == "[General]" && !done) { print k "=" v; done = 1 } sec = $0 }
    sec == "[General]" && index($0, k "=") == 1 { if (!done) print k "=" v; done = 1; next }
    { print }
    END { if (!done) { if (sec != "[General]") print "[General]"; print k "=" v } }
  ' "$1" > "$tmp"
  cat "$tmp" > "$1"
  rm -f "$tmp"
}

if [ -n "${STELLATION_JAVA:-}" ]; then
  set_key "$STELLATION_DATA/freesmlauncher.cfg" AutomaticJavaSwitch false
  set_key "$STELLATION_DATA/freesmlauncher.cfg" AutomaticJavaDownload false
  set_key "$inst/instance.cfg" AutomaticJava false
  set_key "$inst/instance.cfg" OverrideJavaLocation true
  set_key "$inst/instance.cfg" JavaPath "$STELLATION_JAVA"
fi

if [ -n "${STELLATION_PACK_URL:-}" ]; then
  # Same quoting the launcher writes: "\"$INST_JAVA\" -jar ..."
  set_key "$inst/instance.cfg" PreLaunchCommand "\"\\\"\$INST_JAVA\\\" -jar packwiz-installer-bootstrap.jar $STELLATION_PACK_URL\""
fi

# packwiz-installer's window is Java AWT, which paints solid white under non-reparenting
# window managers (niri, sway, Hyprland via Xwayland). Harmless elsewhere.
if [ "$(uname -s)" = Linux ]; then
  export _JAVA_AWT_WM_NONREPARENTING="${_JAVA_AWT_WM_NONREPARENTING:-1}"
fi

# No arguments: straight into the game, which joins the server by itself (JoinServerOnLaunch).
# --launcher opens the launcher window instead, for settings and the like.
if [ "${1:-}" = "--launcher" ]; then
  shift
elif [ "$#" -eq 0 ]; then
  set -- --launch "$STELLATION_INSTANCE"
fi

exec "$STELLATION_LAUNCHER" --dir "$STELLATION_DATA" "$@"
