# FakeFido

FakeFido is a small command-line demo that connects [BulwarkID's Virtual FIDO](https://github.com/bulwarkid/virtual-fido) library to a local credential vault and virtual USB device.

> [!WARNING]
> **FakeFido is not a real or secure WebAuthn authenticator.** Credentials (often called passkeys) are stored in a local JSON vault file. There is no secure hardware protecting the keys, device-bound storage, or hardware-backed user verification. This is different from passkeys protected by Apple or Android devices and physical security keys such as YubiKeys. Do not use FakeFido to protect real accounts or sensitive data.

## Install on Linux

The installer supports Linux on amd64 and arm64. It installs the system USB/IP dependency, downloads the matching binary from the configured public R2 URL, creates a systemd unit, and enables the service.

```sh
export FAKEFIDO_R2_PUBLIC_URL="https://your-public-r2-domain"
curl -fsSL "$FAKEFIDO_R2_PUBLIC_URL/latest/install.sh" \
  | sudo env FAKEFIDO_R2_PUBLIC_URL="$FAKEFIDO_R2_PUBLIC_URL" sh
```

The installer creates a random vault passphrase in `/etc/fakefido/fakefido.env` and stores the vault at `/var/lib/fakefido/vault.json`. Back up both securely. The service runs as root so it can attach USB/IP devices.

```sh
sudo systemctl status fakefido
sudo journalctl -u fakefido -f
sudo systemctl stop fakefido
```

The installer supports apt, dnf, pacman, and zypper package managers. Linux kernel support for USB/IP (`vhci-hcd`) is required.

## Build from source

```sh
go build .
./fakefido start
```

The CLI also provides `list`, `delete`, and `pin` commands. Run `./fakefido --help` for details.

## Releases and deployment

Pushing a `v*` tag builds Linux amd64 and arm64 archives and publishes versioned and `latest/` objects to R2 with Wrangler. The workflow expects these GitHub repository secrets:

- `CLOUDFLARE_ACCOUNT_ID`
- `CLOUDFLARE_API_TOKEN` (R2 object write access)
- `CLOUDFLARE_R2_BUCKET` (bucket name)

Configure a public domain for that R2 bucket; the installer needs its base URL. The landing page deploy workflow uses the same account ID and API token and deploys `site/` to the Cloudflare Pages project named `fakefido`. Create that Pages project and add the same two Cloudflare secrets in the repository settings.

## Underlying project

FakeFido builds on [BulwarkID's Virtual FIDO repository](https://github.com/bulwarkid/virtual-fido), a virtual FIDO2/U2F USB device implementation. See that project for the underlying protocol implementation and its documentation.
