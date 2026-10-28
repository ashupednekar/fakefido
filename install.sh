#!/usr/bin/env sh
set -eu

SERVICE_NAME=fakefido
INSTALL_DIR=/usr/local/bin
DATA_DIR=/var/lib/fakefido
CONFIG_DIR=/etc/fakefido
UNIT_FILE=/etc/systemd/system/fakefido.service
INSTALL_USER=${SUDO_USER:-root}
INSTALL_GROUP=$(id -gn "$INSTALL_USER")
R2_PUBLIC_URL="https://pub-7e86b3749663480cb15de05614bda9e4.r2.dev"

say() { printf '%s\n' "$*"; }
fail() { say "Error: $*" >&2; exit 1; }

[ "$(id -u)" -eq 0 ] || fail "Run this installer as root (for example: curl ... | sudo sh)."
[ "$(uname -s)" = Linux ] || fail "The system service installer currently supports Linux only."
case "$R2_PUBLIC_URL" in
  https://*.r2.dev) ;;
  *) fail "Use the installer from a published FakeFido release." ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) fail "Unsupported architecture: $(uname -m)" ;;
esac

install_usbip() {
  if command -v usbip >/dev/null 2>&1; then return; fi
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update
    if ! apt-get install -y usbip; then
      apt-get install -y linux-tools-common "linux-tools-$(uname -r)"
    fi
  elif command -v dnf >/dev/null 2>&1; then
    dnf install -y usbip
  elif command -v pacman >/dev/null 2>&1; then
    pacman -Sy --noconfirm usbip
  elif command -v zypper >/dev/null 2>&1; then
    zypper --non-interactive install usbip
  else
    fail "Could not identify a supported package manager. Install the usbip command and rerun."
  fi
  command -v usbip >/dev/null 2>&1 || fail "USB/IP package installed, but the usbip command is unavailable for this kernel."
}

install_usbip
mkdir -p "$INSTALL_DIR" "$DATA_DIR" "$CONFIG_DIR"
chown "$INSTALL_USER:$INSTALL_GROUP" "$DATA_DIR" "$CONFIG_DIR"
chmod 2770 "$DATA_DIR"
chmod 700 "$CONFIG_DIR"
archive="fakefido-linux-$arch.tar.gz"
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT HUP INT TERM
say "Downloading FakeFido for linux/$arch"
curl -fsSL "$R2_PUBLIC_URL/latest/$archive" -o "$tmpdir/$archive"
curl -fsSL "$R2_PUBLIC_URL/latest/$archive.sha256" -o "$tmpdir/$archive.sha256"
(cd "$tmpdir" && sha256sum -c "$archive.sha256")
tar -xzf "$tmpdir/$archive" -C "$tmpdir"
install -m 0755 "$tmpdir/fakefido" "$INSTALL_DIR/fakefido"

if [ ! -f "$CONFIG_DIR/fakefido.env" ]; then
  passphrase=$(od -An -N24 -tx1 /dev/urandom | tr -d ' \n')
  umask 077
  printf 'FAKEFIDO_PASSPHRASE=%s\n' "$passphrase" > "$CONFIG_DIR/fakefido.env"
  say "Created a random vault passphrase in $CONFIG_DIR/fakefido.env (readable only by the installing user and root). Back it up securely."
fi
chown "$INSTALL_USER:$INSTALL_GROUP" "$CONFIG_DIR/fakefido.env"
chmod 600 "$CONFIG_DIR/fakefido.env"
if [ -f "$DATA_DIR/vault.json" ]; then
  chown "$INSTALL_USER:$INSTALL_GROUP" "$DATA_DIR/vault.json"
  chmod 660 "$DATA_DIR/vault.json"
fi

cat > "$UNIT_FILE" <<'UNIT'
[Unit]
Description=FakeFido virtual FIDO authenticator demo
After=network.target

[Service]
Type=simple
EnvironmentFile=/etc/fakefido/fakefido.env
ExecStart=/bin/sh -c 'exec /usr/local/bin/fakefido --vault /var/lib/fakefido/vault.json --passphrase "$FAKEFIDO_PASSPHRASE" start'
Restart=on-failure
RestartSec=3
WorkingDirectory=/var/lib/fakefido
UMask=0007

[Install]
WantedBy=multi-user.target
UNIT

if command -v modprobe >/dev/null 2>&1; then
  modprobe vhci-hcd 2>/dev/null || say "Note: vhci-hcd could not be loaded now; ensure your kernel supports USB/IP."
fi
systemctl daemon-reload
systemctl enable --now "$SERVICE_NAME.service"
say "FakeFido is installed and running as a systemd service."
say "Check status with: systemctl status fakefido"
say "Logs are available with: journalctl -u fakefido -f"
