# ViShare

> Share your mouse and keyboard between Ubuntu/macOS and Windows over a direct LAN cable — no KVM switch needed.

ViShare is a lightweight, open-source software KVM tool that lets you control multiple computers with a single keyboard and mouse. Move your cursor to the edge of the screen and it seamlessly switches to the next machine. Configuration is file-based with no GUI required.

**Inspired by [Barrier](https://github.com/debauchee/barrier) and [Synergy](https://symless.com/synergy).**

---

## Features

- **Cross-platform**: Ubuntu (X11) / macOS (Apple Silicon & Intel) ↔ Windows 10/11
- **Direct LAN connection**: Works over a simple network cable, no router or internet needed
- **Low latency**: TCP over local network, input feels native
- **Keyboard & mouse**: Move, click, scroll, and all key events
- **Clipboard sync**: *(planned)*
- **System tray**: Connection status indicator, quit from tray icon
- **Simple config**: One TOML file per machine, no GUI needed

---

## How It Works

```
┌──────────────────────────┐        Direct LAN Cable        ┌─────────────────────┐
│  Ubuntu / macOS (Server) │ ◄────────────────────────────► │   Windows (Client)  │
│                          │       192.168.100.0/24          │                     │
│   Physical keyboard      │   TCP :24800                    │   Receives events   │
│   Physical mouse         │───────────────────────────────► │   Injects input     │
│                          │                                 │                     │
└──────────────────────────┘                                 └─────────────────────┘

  Move cursor to right edge  ─────────────────────────────►  Cursor appears on Windows
  Move cursor to left edge   ◄─────────────────────────────  Cursor appears on server
```

One machine runs as **server** (owns the physical keyboard and mouse), the other runs as **client** (receives and replays input events).

---

## Platform Support

| Role   | Platform              | Status |
|--------|-----------------------|--------|
| Server | Ubuntu (X11)          | ✅ Supported |
| Server | macOS (Apple Silicon) | ✅ Supported |
| Server | macOS (Intel)         | ✅ Supported |
| Client | Windows 10/11         | ✅ Supported |
| Client | Ubuntu (X11)          | ✅ Supported |
| Client | macOS                 | ✅ Supported |

---

## Installation

### Prerequisites

**Ubuntu (Server)**
```bash
sudo apt install libx11-dev libxtst-dev libx11-xcb-dev libxkbcommon-dev libxkbcommon-x11-dev gcc
```

**macOS (Server)**
```bash
# Xcode Command Line Tools only — no Homebrew packages needed
xcode-select --install
```

> macOS requires **Accessibility** and **Input Monitoring** permissions on first run.
> Grant them in: **System Settings → Privacy & Security**

**Windows (Client)**
- MinGW-w64 or Visual Studio Build Tools (for CGO)
- No additional system libraries needed

---

## Build from Source

```bash
git clone https://github.com/victorfengming/vishare.git
cd vishare
go mod tidy
```

### Linux (Ubuntu)

```bash
go build -o build/vishare ./cmd/vishare
```

### macOS (Apple Silicon / ARM64)

```bash
GOOS=darwin GOARCH=arm64 go build -o build/vishare-darwin-arm64 ./cmd/vishare
```

Or use the provided script:

```bash
./build_mac.sh
```

### Windows

```bash
# Run on Windows
go build -o build/vishare.exe ./cmd/vishare
```

Cross-compile from Linux (requires `mingw-w64`):

```bash
sudo apt install mingw-w64
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
  go build -o build/vishare.exe ./cmd/vishare
```

> **Note**: CGO is required (used by robotgo). macOS binaries must be built on a Mac.

---

## Quick Start

### Step 1 — Connect the machines

Plug a network cable directly between the two computers and assign static IPs:

| Machine | IP Address |
|---------|-----------|
| Ubuntu / macOS (server) | `192.168.100.1/24` |
| Windows (client) | `192.168.100.2/24` |

**Ubuntu**
```bash
nmcli con add type ethernet ifname eth0 ip4 192.168.100.1/24
```

**macOS**: System Settings → Network → Ethernet → TCP/IP → Manual → `192.168.100.1` / `255.255.255.0`

**Windows**: Network Settings → Ethernet → IPv4 → Manual → `192.168.100.2` / `255.255.255.0`

### Step 2 — Configure

**Server — `config.server.toml`**
```toml
role        = "server"
listen_addr = "192.168.100.1:24800"   # replace with your server IP

[[screens]]
name       = "mac-main"       # or "ubuntu-main"
edge_right = "windows-pc"

[[screens]]
name      = "windows-pc"
edge_left = "mac-main"        # or "ubuntu-main"
```

**Windows — `config.client.toml`**
```toml
role        = "client"
server_addr = "192.168.100.1:24800"
screen_name = "windows-pc"
```

### Step 3 — Run

```bash
# On Ubuntu/macOS (start server first)
./vishare --config config.server.toml

# On Windows
vishare.exe --config config.client.toml
```

A system tray icon will appear. 🟢 Green = connected, 🔴 Red = disconnected.

---

## macOS Notes

### Permissions

On first launch, macOS will prompt for two permissions — both are required:

| Permission | Purpose |
|-----------|---------|
| **Accessibility** | Inject mouse and keyboard events |
| **Input Monitoring** | Capture global keyboard/mouse hook |

Grant them in **System Settings → Privacy & Security → Accessibility / Input Monitoring**.

### Apple Silicon (arm64) Build Script

```bash
./build_mac.sh
# Output: build/vishare-darwin-arm64
```

---

## Configuration Reference

```toml
# ── Common ────────────────────────────────────────────────
role = "server"          # "server" or "client"

# ── Server only ───────────────────────────────────────────
listen_addr = "0.0.0.0:24800"

[[screens]]
name        = "mac-main"
edge_right  = "windows-pc"   # screen to switch to when cursor exits right edge
edge_left   = ""             # empty = no screen on this side
edge_top    = ""
edge_bottom = ""

[[screens]]
name      = "windows-pc"
edge_left = "mac-main"

# ── Client only ───────────────────────────────────────────
server_addr = "192.168.100.1:24800"
screen_name = "windows-pc"
```

---

## Roadmap

- [x] Mouse move & click
- [x] Keyboard events
- [x] Mouse scroll
- [x] System tray (connect status, quit)
- [x] macOS (Apple Silicon) server support
- [ ] Clipboard sync
- [ ] Wayland support (Ubuntu 22.04+)
- [ ] Multiple clients (3+ machines)
- [ ] Screen layout: vertical arrangement
- [ ] TLS encryption
- [ ] Auto-discovery on LAN (mDNS)

---

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | [Go](https://golang.org) 1.22+ |
| Input capture & injection | [robotgo](https://github.com/go-vgo/robotgo) |
| Global hook | [gohook](https://github.com/robotn/gohook) |
| System tray | [fyne.io/systray](https://github.com/fyne-io/systray) |
| Config | [BurntSushi/toml](https://github.com/BurntSushi/toml) |
| Logging | [zerolog](https://github.com/rs/zerolog) |
| Networking | Go standard library (`net`) |

---

## Comparison with Barrier

| Feature | ViShare | Barrier |
|---------|---------|---------|
| Protocol | Custom TCP | Barrier protocol |
| GUI | None (config file) | Qt GUI |
| macOS server | Yes | Yes |
| Clipboard sync | Planned | Yes |
| Wayland | Planned | Limited |
| Ubuntu ↔ Windows | Yes | Yes |
| macOS ↔ Windows | Yes | Yes |
| Dependencies | Go + CGO | C++ + Qt |
| Binary size | ~11 MB | ~30 MB |

---

## Contributing

Pull requests are welcome. For major changes, please open an issue first.

```bash
git clone https://github.com/victorfengming/vishare.git
cd vishare
go test ./...
```

---

## License

[MIT](LICENSE)
