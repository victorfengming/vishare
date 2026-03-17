# ViShare

> Share your mouse and keyboard between Ubuntu and Windows over a direct LAN cable — no KVM switch needed.

ViShare is a lightweight, open-source software KVM tool that lets you control multiple computers with a single keyboard and mouse. Move your cursor to the edge of the screen and it seamlessly switches to the next machine. Configuration is file-based with no GUI required.

**Inspired by [Barrier](https://github.com/debauchee/barrier) and [Synergy](https://symless.com/synergy).**

---

## Features

- **Cross-platform**: Ubuntu (X11) ↔ Windows 10/11
- **Direct LAN connection**: Works over a simple network cable, no router or internet needed
- **Low latency**: TCP over local network, input feels native
- **Keyboard & mouse**: Move, click, scroll, and all key events
- **Clipboard sync**: *(planned)*
- **System tray**: Connection status indicator, reconnect/quit from tray icon
- **Simple config**: One TOML file per machine, no GUI needed

---

## How It Works

```
┌─────────────────────┐          Direct LAN Cable         ┌─────────────────────┐
│   Ubuntu (Server)   │ ◄──────────────────────────────► │   Windows (Client)  │
│                     │        192.168.100.0/24            │                     │
│  Physical keyboard  │   TCP :24800                       │   Receives events   │
│  Physical mouse     │──────────────────────────────────► │   Injects input     │
│                     │                                    │                     │
└─────────────────────┘                                    └─────────────────────┘

  Move cursor to right edge  ──────────────────────────►  Cursor appears on Windows
  Move cursor to left edge   ◄──────────────────────────  Cursor appears on Ubuntu
```

One machine runs as **server** (owns the physical keyboard and mouse), the other runs as **client** (receives and replays input events).

---

## Installation

### Prerequisites

**Ubuntu (Server)**
```bash
sudo apt install libx11-dev libxtst-dev libpng-dev gcc
```

**Windows (Client)**
- Visual Studio Build Tools or MinGW (for CGO)
- No additional system libraries needed

### Build from source

```bash
git clone https://github.com/your-username/vishare.git
cd vishare
go mod tidy

# Build for Linux
go build -o vishare ./cmd/vishare

# Build for Windows (run on Windows)
go build -o vishare.exe ./cmd/vishare
```

> **Note**: CGO is required (used by robotgo). Cross-compilation is not supported — build on each target platform separately.

---

## Quick Start

### Step 1 — Connect the machines

Plug a network cable directly between the two computers and assign static IPs:

| Machine | IP Address |
|---------|-----------|
| Ubuntu (server) | `192.168.100.1/24` |
| Windows (client) | `192.168.100.2/24` |

**Ubuntu**
```bash
# Using nmcli
nmcli con add type ethernet ifname eth0 ip4 192.168.100.1/24
```

**Windows**: Network Settings → Ethernet → IPv4 → Manual → `192.168.100.2` / `255.255.255.0`

### Step 2 — Configure

**Ubuntu — `config.toml`**
```toml
role = "server"
listen_addr = "0.0.0.0:24800"

[[screens]]
name = "ubuntu-main"
edge_right = "windows-pc"

[[screens]]
name = "windows-pc"
edge_left = "ubuntu-main"
```

**Windows — `config.toml`**
```toml
role = "client"
server_addr = "192.168.100.1:24800"
screen_name = "windows-pc"
```

### Step 3 — Run

```bash
# On Ubuntu (start server first)
./vishare --config config.toml

# On Windows
vishare.exe --config config.toml
```

A system tray icon will appear. Green = connected, Red = disconnected.

---

## Configuration Reference

```toml
# ── Common ────────────────────────────────────────────────
role = "server"          # "server" or "client"

# ── Server only ───────────────────────────────────────────
listen_addr = "0.0.0.0:24800"

[[screens]]
name      = "ubuntu-main"
edge_right = "windows-pc"   # screen name to switch to when cursor exits right edge
edge_left  = ""             # empty = no screen on this side
edge_top   = ""
edge_bottom = ""

[[screens]]
name      = "windows-pc"
edge_left = "ubuntu-main"

# ── Client only ───────────────────────────────────────────
server_addr = "192.168.100.1:24800"
screen_name = "windows-pc"
```

---

## Roadmap

- [x] Mouse move & click
- [x] Keyboard events
- [x] Mouse scroll
- [x] System tray (connect status, reconnect, quit)
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
| System tray | [systray](https://github.com/getlantern/systray) |
| Config | [BurntSushi/toml](https://github.com/BurntSushi/toml) |
| Logging | [zerolog](https://github.com/rs/zerolog) |
| Networking | Go standard library (`net`) |

---

## Comparison with Barrier

| Feature | ViShare | Barrier |
|---------|---------|---------|
| Protocol | Custom TCP | Barrier protocol |
| GUI | None (config file) | Qt GUI |
| Clipboard sync | Planned | Yes |
| Wayland | Planned | Limited |
| Ubuntu ↔ Windows | Yes | Yes |
| Dependencies | Go + CGO | C++ + Qt |
| Binary size | ~10 MB | ~30 MB |

---

## Contributing

Pull requests are welcome. For major changes, please open an issue first.

```bash
git clone https://github.com/your-username/vishare.git
cd vishare
go test ./...
```

---

## License

[MIT](LICENSE)
