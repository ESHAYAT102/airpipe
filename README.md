# AirPipe

Encrypted file transfer. WebRTC peer-to-peer with relay fallback. No accounts. No apps.

```
$ airpipe send config.yaml

  ╔══════════════════════════════════════════╗
  ║  RIVER FALCON MARBLE 42                  ║
  ╚══════════════════════════════════════════╝

  Tell them: airpipe.sanyamgarg.com
  They type the code, they get the file.
```

**Try it:** [airpipe.sanyamgarg.com/send](https://airpipe.sanyamgarg.com/send)

## Install

```bash
curl -sSL https://airpipe.sanyamgarg.com/install.sh | sh
```

Or:
```bash
go install github.com/Sanyam-G/Airpipe/cmd/airpipe@latest
```

Update an existing installation:
```bash
airpipe update
```

## Usage

**Send a file (passphrase flow):**
```bash
airpipe send photo.jpg
```
Shows a passphrase and a QR code. The receiver either:
- Types the passphrase at [airpipe.sanyamgarg.com](https://airpipe.sanyamgarg.com)
- Scans the QR code
- Runs `airpipe download RIVER FALCON MARBLE 42` from their terminal

**Download with a passphrase:**
```bash
airpipe download RIVER FALCON MARBLE 42
```

**Send multiple files (auto-zipped):**
```bash
airpipe send file1.txt file2.txt photos/
```

**Receive a file from a browser (P2P):**
```bash
airpipe receive ./downloads
```
Prints a QR. Phone scans it, drops a file. The browser and your CLI negotiate WebRTC and the file flows directly between them, falling back to relay-streaming if NAT punching fails.

## How it works

**Two transports, one passphrase-derived key:**

- **Passphrase flow** (`airpipe send` / `airpipe download`): file is encrypted locally with NaCl secretbox and uploaded to the relay as ciphertext. The receiver derives the same key from the passphrase and decrypts locally. The relay only ever sees the ciphertext and a 16-char hex token — never the passphrase or the key.

- **WebRTC P2P flow** (browser sender → CLI receiver via QR): both peers connect to the relay's WebSocket room. They exchange SDP offers/answers and ICE candidates through the room (encrypted with the passphrase-derived key, so the relay can't impersonate either side). When the WebRTC DataChannel opens, file bytes flow directly between the peers. If NAT punching fails after 15 seconds, both sides fall back to streaming through the relay's WebSocket — the relay sees ciphertext only in either case.

Two layers of encryption: DTLS (built into WebRTC) protects the wire; NaCl secretbox sits on top so even a malicious relay can't substitute peers in the SDP exchange.

Files expire after 10 minutes.

## Self-host

```bash
docker run -p 8080:8080 ghcr.io/sanyam-g/airpipe-relay
airpipe --relay https://your-server:8080 send file.txt
```

Configurable via env vars:
- `AIRPIPE_ALLOWED_ORIGINS` (comma-separated, or `*`) — CORS allowlist for browser clients
- `AIRPIPE_RATE_LIMIT_PER_MIN` — per-IP token bucket (default 60)
- `AIRPIPE_LOG_FORMAT` — `json` (default) or `text`

`docker-compose.yml` includes an opt-in Watchtower sidecar that auto-pulls new images.

## License

MIT
