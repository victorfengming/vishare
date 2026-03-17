# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2024-03-18

### Added

- **Protocol**: custom TCP framing with 3-byte header (type + uint16 payload length)
  - Messages: `MouseMove`, `MouseButton`, `MouseWheel`, `KeyDown`, `KeyUp`
  - Control: `SwitchTo`, `SwitchBack`, `Handshake`, `Ping`, `Pong`
- **Server** (Ubuntu / macOS):
  - Global keyboard & mouse capture via `gohook`
  - Screen edge detection at 100 Hz with 3-frame hysteresis (30 ms dwell)
  - Cursor warp to center on edge crossing
  - Coordinate translation: server resolution → client resolution
  - Multi-client accept loop with per-client send/read/keepalive goroutines
  - Exponential backoff reconnect on client drop
- **Client** (Windows / Ubuntu / macOS):
  - Input injection via `robotgo` (Linux/macOS) and Win32 `SendInput` (Windows)
  - Edge-based `SwitchBack` when cursor exits client screen border
  - Exponential backoff reconnect: 1 s → 2 s → … → 30 s cap
- **System tray** via `fyne.io/systray`:
  - Green icon = connected, red icon = disconnected
  - Quit menu item
- **Config**: TOML file with full validation
  - Server: `listen_addr`, `[[screens]]` with edge references
  - Client: `server_addr`, `screen_name`
- **Platform support**:
  - Server: Ubuntu (X11), macOS (Apple Silicon & Intel)
  - Client: Windows 10/11, Ubuntu, macOS
- **Windows injection** via Win32 `SendInput` (no CGO runtime dependency):
  - Absolute mouse movement (normalized 0–65535)
  - Mouse buttons (left / right / middle)
  - Mouse wheel (vertical & horizontal)
  - Keyboard: X11 keysym → Windows Virtual Key table + Unicode fallback
- **Tests**: protocol round-trip (net.Pipe) and config validation
- **CI**: GitHub Actions — build & test on Ubuntu, cross-compile Windows `.exe`
- **macOS build script**: `build_mac.sh` for Apple Silicon one-liner build

[0.1.0]: https://github.com/victorfengming/vishare/releases/tag/v0.1.0
