# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.2.0] - 2026-03-18

### Added

- **Single-instance lock**: only one vishare process can run at a time
  - Linux / macOS: `syscall.Flock` on `/tmp/vishare.lock`
  - Windows: Win32 named mutex `Global\Vishare-vishare`
  - Second launch prints a clear error and exits immediately
- **Pre-shared secret authentication**: optional `secret` field in config
  - Server rejects clients whose handshake secret does not match
  - Clients send the configured secret in the handshake
  - No change in behavior when `secret` is left empty (backward-compatible)
- **macOS tray menu fix**: `NSStatusItem` menus now respond to mouse clicks
  for non-bundled binaries
  - `Info.plist` with `LSUIElement=true` embedded into the binary via
    `-sectcreate __TEXT __info_plist` linker flag (`build_mac.sh`)
  - CGO `platform_darwin.go` calls `setActivationPolicy:Accessory` +
    `activateIgnoringOtherApps:YES` before `systray.Run`
- **Shared constants package** `internal/defaults`: `SendBufSize`,
  `PingInterval`, `EdgePollMs`, `EdgeHysteresis`, `BackoffMin`, `BackoffMax`
- **Shared status package** `internal/status`: unified `Msg` type used by
  server, client, and tray — eliminates the bridge goroutines in `main.go`

### Changed

- **Server edge detection is now event-driven** (was timer-based at 100 Hz):
  edge crossings are detected directly from `EventMouseMove` hook events,
  removing the `edgeTicker` and the unnecessary `atomic.Int32` counter
- **`MouseMovePayload` reduced from 8 → 4 bytes**: removed redundant
  `ClientW`/`ClientH` fields that duplicated the handshake screen dimensions
- **`HandshakePayload` extended from 36 → 68 bytes**: added 32-byte `Secret`
  field between `ScreenName` and the screen dimensions
- **`edgeTarget()` logic corrected**: removed dead `if` block; server screen
  is now identified by `cfg.ScreenName` with a sensible fallback
- **`keycodeToName` uses `string(rune(keycode))`** instead of
  `fmt.Sprintf("%c", ...)` on Linux and macOS, removing the `fmt` import
- **`min()` helper removed** from `client.go`; uses the Go 1.21+ built-in
- `statusCh` is now closed after `srv.Run` / `cli.Run` returns, allowing
  the tray goroutine to exit cleanly on shutdown (previously leaked)

### Fixed

- macOS: system tray icon appeared but menu could not be opened — fixed by
  embedding `Info.plist` and setting the correct NSApplication activation
  policy
- All platforms: multiple vishare instances could be started silently —
  second instance now exits with an error message

[0.2.0]: https://github.com/victorfengming/vishare/compare/v0.1.0...v0.2.0

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
