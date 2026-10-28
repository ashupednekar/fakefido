# FakeFido

FakeFido is a small command-line demo that connects [BulwarkID’s Virtual FIDO](https://github.com/bulwarkid/virtual-fido) library to a virtual USB device and a local credential vault. It is intended for learning and experiments with authenticator flows.

## Important security limitation

> **FakeFido is not a real or secure WebAuthn authenticator.** Credentials (often called passkeys) are stored in a JSON vault file on disk. The authenticator is virtual and has no secure hardware, device-bound key storage, or hardware-backed user verification. This is different from passkeys protected by Apple or Android devices and physical security keys such as YubiKeys. Do not use FakeFido to protect real accounts or sensitive data.

## Install on Linux

Prebuilt Linux binaries are published for `amd64` and `arm64`. The installer detects the machine architecture, installs USB/IP using the system package manager, installs the binary, creates a systemd service, and starts it. It requires root access and a Linux kernel with USB/IP support.

```sh
export FAKEFIDO_R2_PUBLIC_URL="__FAKEFIDO_R2_PUBLIC_URL__"
curl -fsSL "$FAKEFIDO_R2_PUBLIC_URL/latest/install.sh" \
  | sudo env FAKEFIDO_R2_PUBLIC_URL="$FAKEFIDO_R2_PUBLIC_URL" sh
```

The service runs as root so it can attach the virtual USB device. The installer creates a random passphrase in `/etc/fakefido/fakefido.env` and stores the vault at `/var/lib/fakefido/vault.json`. Keep a backup of the passphrase and vault.

```sh
sudo systemctl status fakefido
sudo journalctl -u fakefido -f
sudo systemctl stop fakefido
```

The installer supports apt, dnf, pacman, and zypper. USB/IP’s `vhci-hcd` kernel module must be available. The current demo attaches to a configured USB/IP bus/device, so your system’s device setup must match the project configuration.

## Build from source

Install Go, clone the repository, then build and run:

```sh
go build -o fakefido .
./fakefido start
```

Run `./fakefido --help` for the available commands. The demo also has commands to list and delete vault identities and configure PIN behavior.

## Releases

Pushing a `v*` tag builds Linux `amd64` and `arm64` archives, attaches them and SHA-256 checksums to a GitHub release, and uploads versioned and `latest/` copies to R2. The installer downloads the matching archive from the public R2 bucket.

The release workflow uses these repository secrets:

- `CLOUDFLARE_ACCOUNT_ID`
- `CLOUDFLARE_API_TOKEN` with Cloudflare Pages and R2 permissions
- `CLOUDFLARE_R2_BUCKET` with the release bucket name

The landing page is in [`site/`](site/) and deploys to the Cloudflare Pages project named `fakefido` when it changes on `main`.

## Underlying project

FakeFido builds on [BulwarkID’s Virtual FIDO repository](https://github.com/bulwarkid/virtual-fido), which implements the virtual FIDO2/U2F USB device and protocol handling. See the upstream repository for its documentation and implementation details.
