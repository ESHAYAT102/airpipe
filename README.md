# AirPipe

Self-hosted file transfer with a passphrase that works anywhere. WebRTC peer-to-peer between any two devices, with an encrypted mailbox fallback for when the receiver isn't online yet.

![demo](demo.gif)

**Try it:** [airpipe.sanyamgarg.com](https://airpipe.sanyamgarg.com)

## Self-host

One container. Bundles the relay, landing page, browser sender/receiver, and the install script.

```bash
docker run -p 8080:8080 ghcr.io/sanyam-g/airpipe-relay
```

Or with the bundled `docker-compose.yml` (includes an opt-in Watchtower sidecar):

```bash
git clone https://github.com/Sanyam-G/Airpipe
cd Airpipe
docker compose up -d
```

Point the CLI at your relay:

```bash
airpipe --relay https://your-server.example send file.txt
```

### Config

| Env var | Purpose | Default |
|---|---|---|
| `PORT` | Listen port | `8080` |
| `AIRPIPE_ALLOWED_ORIGINS` | CORS allowlist for browsers, CSV or `*` | `*` |
| `AIRPIPE_RATE_LIMIT_PER_MIN` | Per-IP token bucket | `60` |
| `AIRPIPE_LOG_FORMAT` | `json` or `text` | `json` |

`/health` returns JSON: `version`, `uptime_seconds`, `active_files`, `active_bytes`, `active_ws_rooms`, `protocol_version`.

## Install the CLI

```bash
curl -sSL https://airpipe.sanyamgarg.com/install.sh | sh
```

Or via Go:
```bash
go install github.com/Sanyam-G/Airpipe/cmd/airpipe@latest
```

Self-update:
```bash
airpipe update
```

Linux + macOS, amd64 + arm64.

## Usage

**Send a file:**
```bash
airpipe send report.pdf
```
You get prompted: direct (P2P) or mailbox (relay holds it 10 min). Pick one. The CLI shows a passphrase like `RIVER FALCON MARBLE 42`, plus a QR and a link.

**Download:**
```bash
airpipe download RIVER FALCON MARBLE 42
```
The receiver doesn't need to know which mode the sender picked. Same passphrase resolves to either path.

**Send multiple files / a directory** (auto-zipped):
```bash
airpipe send file1.txt photos/
```

**Receive from a browser** (CLI side waits for a phone or laptop to drop a file in):
```bash
airpipe receive ./downloads
```
Prints a QR. Phone scans it, drops a file. WebRTC direct, fallback to relay if NAT punching fails.

**Browser to browser, no install on either side:** open `/live` on the relay. Get a passphrase + QR. Receiver types the passphrase at the homepage in their browser, both pair, sender drops a file.

## How it works

Two transports, one passphrase-derived key.

**Direct (`/live` in browser, or `airpipe send` → direct mode):**
both peers connect to the relay's WebSocket room and exchange WebRTC SDP and ICE candidates. All signaling is encrypted with the passphrase-derived key, so the relay can't impersonate either side. When the DataChannel opens, file bytes flow directly between the peers. If NAT punching fails after 15 seconds, both sides fall back to streaming through the WebSocket. The relay only ever sees ciphertext.

**Mailbox (`/send` in browser, or `airpipe send` → mailbox mode):**
file is encrypted locally with NaCl secretbox and uploaded to the relay as ciphertext. The receiver derives the same key from the passphrase and decrypts locally. The relay sees a 16-char hex token and ciphertext, nothing else. Files expire after 10 minutes.

Two layers of encryption on the live path: DTLS (built into WebRTC) protects the wire, NaCl secretbox sits on top so even a malicious relay can't substitute peers in the SDP exchange.

The passphrase derives both the relay token and the NaCl key via SHA-256 with domain separation (`airpipe:token:` and `airpipe:key:` prefixes). Same algorithm in CLI and browser.

## Stack

Go relay (gorilla/websocket, pion/webrtc), embedded HTML/CSS/JS frontend (tweetnacl.js for browser crypto), Docker, Cloudflare Tunnel optional. Single static binary, ~15 MB image.

## License

MIT
