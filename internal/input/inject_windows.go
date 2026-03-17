//go:build windows

package input

import (
	"syscall"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procSendInput        = user32.NewProc("SendInput")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
)

const (
	smCxScreen = 0
	smCyScreen = 1

	inputMouse    = 0
	inputKeyboard = 1

	// Mouse event flags
	mouseMoveAbs    uint32 = 0x8001 // MOUSEEVENTF_MOVE | MOUSEEVENTF_ABSOLUTE
	mouseLeftDown   uint32 = 0x0002
	mouseLeftUp     uint32 = 0x0004
	mouseRightDown  uint32 = 0x0008
	mouseRightUp    uint32 = 0x0010
	mouseMiddleDown uint32 = 0x0020
	mouseMiddleUp   uint32 = 0x0040
	mouseWheel      uint32 = 0x0800
	mouseHWheel     uint32 = 0x1000

	// Keyboard event flags
	keyEventKeyUp   uint32 = 0x0002
	keyEventUnicode uint32 = 0x0004

	wheelDelta = 120
)

// mouseINPUT mirrors the Win32 INPUT struct (type=INPUT_MOUSE) on amd64.
// Verified size: 4+4+4+4+4+4+4+4+8 = 40 bytes.
type mouseINPUT struct {
	typ       uint32
	_         uint32 // padding: aligns union to 8 bytes on x64
	dx        int32
	dy        int32
	mouseData uint32
	dwFlags   uint32
	time      uint32
	_         uint32 // padding: aligns extraInfo to 8 bytes
	extraInfo uintptr
}

// kbdINPUT mirrors the Win32 INPUT struct (type=INPUT_KEYBOARD) on amd64.
// Verified size: 4+4+2+2+4+4+4+8+8 = 40 bytes.
type kbdINPUT struct {
	typ       uint32
	_         [4]byte
	wVk       uint16
	wScan     uint16
	dwFlags   uint32
	time      uint32
	_         [4]byte
	extraInfo uintptr
	_         [8]byte // fill union to match mouseINPUT size (32 bytes)
}

func screenSize() (w, h int) {
	cx, _, _ := procGetSystemMetrics.Call(smCxScreen)
	cy, _, _ := procGetSystemMetrics.Call(smCyScreen)
	return int(cx), int(cy)
}

func sendMouseInput(mi mouseINPUT) {
	procSendInput.Call(1, uintptr(unsafe.Pointer(&mi)), unsafe.Sizeof(mi))
}

func sendKbdInput(ki kbdINPUT) {
	procSendInput.Call(1, uintptr(unsafe.Pointer(&ki)), unsafe.Sizeof(ki))
}

// WindowsInjector implements Injector using Win32 SendInput.
type WindowsInjector struct{}

func NewInjector() Injector {
	return &WindowsInjector{}
}

func (w *WindowsInjector) MoveMouse(x, y int) {
	sw, sh := screenSize()
	if sw == 0 || sh == 0 {
		return
	}
	// MOUSEEVENTF_ABSOLUTE uses normalized coordinates 0..65535
	nx := int32(x * 65535 / sw)
	ny := int32(y * 65535 / sh)
	sendMouseInput(mouseINPUT{
		typ:     inputMouse,
		dx:      nx,
		dy:      ny,
		dwFlags: mouseMoveAbs,
	})
}

func (w *WindowsInjector) MouseButton(button uint16, down bool) {
	var flags uint32
	switch button {
	case 1:
		if down {
			flags = mouseLeftDown
		} else {
			flags = mouseLeftUp
		}
	case 3:
		if down {
			flags = mouseRightDown
		} else {
			flags = mouseRightUp
		}
	case 2:
		if down {
			flags = mouseMiddleDown
		} else {
			flags = mouseMiddleUp
		}
	default:
		return
	}
	sendMouseInput(mouseINPUT{typ: inputMouse, dwFlags: flags})
}

func (w *WindowsInjector) MouseWheel(deltaX, deltaY int) {
	if deltaY != 0 {
		sendMouseInput(mouseINPUT{
			typ:       inputMouse,
			mouseData: uint32(int32(deltaY) * wheelDelta),
			dwFlags:   mouseWheel,
		})
	}
	if deltaX != 0 {
		sendMouseInput(mouseINPUT{
			typ:       inputMouse,
			mouseData: uint32(int32(deltaX) * wheelDelta),
			dwFlags:   mouseHWheel,
		})
	}
}

func (w *WindowsInjector) KeyDown(keycode, rawcode, mask uint16) {
	sendKey(keycode, false)
}

func (w *WindowsInjector) KeyUp(keycode, rawcode, mask uint16) {
	sendKey(keycode, true)
}

func sendKey(keysym uint16, up bool) {
	var flags uint32
	if up {
		flags = keyEventKeyUp
	}
	// Known special keys: use Windows Virtual Key code
	if vk, ok := keysymToVK[keysym]; ok {
		sendKbdInput(kbdINPUT{typ: inputKeyboard, wVk: vk, dwFlags: flags})
		return
	}
	// Printable ASCII (0x20–0x7e): inject as Unicode character
	if keysym >= 0x20 && keysym <= 0x7e {
		sendKbdInput(kbdINPUT{
			typ:     inputKeyboard,
			wScan:   keysym,
			dwFlags: flags | keyEventUnicode,
		})
	}
}

// keysymToVK maps X11 keysyms to Windows Virtual Key codes.
var keysymToVK = map[uint16]uint16{
	// Function keys
	0xFFBE: 0x70, 0xFFBF: 0x71, 0xFFC0: 0x72, 0xFFC1: 0x73, // F1–F4
	0xFFC2: 0x74, 0xFFC3: 0x75, 0xFFC4: 0x76, 0xFFC5: 0x77, // F5–F8
	0xFFC6: 0x78, 0xFFC7: 0x79, 0xFFC8: 0x7A, 0xFFC9: 0x7B, // F9–F12

	// Modifiers
	0xFFE1: 0x10, 0xFFE2: 0x10, // Shift → VK_SHIFT
	0xFFE3: 0x11, 0xFFE4: 0x11, // Ctrl  → VK_CONTROL
	0xFFE9: 0x12, 0xFFEA: 0x12, // Alt   → VK_MENU
	0xFFEB: 0x5B,               // Super_L → VK_LWIN
	0xFFEC: 0x5C,               // Super_R → VK_RWIN

	// Navigation
	0xFF51: 0x25, // Left  → VK_LEFT
	0xFF52: 0x26, // Up    → VK_UP
	0xFF53: 0x27, // Right → VK_RIGHT
	0xFF54: 0x28, // Down  → VK_DOWN
	0xFF50: 0x24, // Home  → VK_HOME
	0xFF57: 0x23, // End   → VK_END
	0xFF55: 0x21, // PgUp  → VK_PRIOR
	0xFF56: 0x22, // PgDn  → VK_NEXT
	0xFF63: 0x2D, // Ins   → VK_INSERT
	0xFFFF: 0x2E, // Del   → VK_DELETE

	// Editing / control
	0xFF08: 0x08, // BackSpace → VK_BACK
	0xFF09: 0x09, // Tab       → VK_TAB
	0xFF0D: 0x0D, // Return    → VK_RETURN
	0xFF1B: 0x1B, // Escape    → VK_ESCAPE
	0xFF20: 0x14, // CapsLock  → VK_CAPITAL
	0xFF7F: 0x90, // NumLock   → VK_NUMLOCK
	0xFF14: 0x91, // ScrollLck → VK_SCROLL
	0xFF61: 0x2C, // Print     → VK_SNAPSHOT

	// Numpad digits
	0xFFB0: 0x60, 0xFFB1: 0x61, 0xFFB2: 0x62, 0xFFB3: 0x63,
	0xFFB4: 0x64, 0xFFB5: 0x65, 0xFFB6: 0x66, 0xFFB7: 0x67,
	0xFFB8: 0x68, 0xFFB9: 0x69,

	// Numpad operators
	0xFFAA: 0x6A, // * → VK_MULTIPLY
	0xFFAB: 0x6B, // + → VK_ADD
	0xFFAD: 0x6D, // - → VK_SUBTRACT
	0xFFAE: 0x6E, // . → VK_DECIMAL
	0xFFAF: 0x6F, // / → VK_DIVIDE
}
