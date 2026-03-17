//go:build linux || darwin

package input

// keysymTable maps X11/macOS keysyms to robotgo key name strings.
var keysymTable = map[uint16]string{
	// Function keys
	0xFFBE: "f1", 0xFFBF: "f2", 0xFFC0: "f3", 0xFFC1: "f4",
	0xFFC2: "f5", 0xFFC3: "f6", 0xFFC4: "f7", 0xFFC5: "f8",
	0xFFC6: "f9", 0xFFC7: "f10", 0xFFC8: "f11", 0xFFC9: "f12",

	// Modifiers
	0xFFE1: "shift", 0xFFE2: "shift",
	0xFFE3: "ctrl", 0xFFE4: "ctrl",
	0xFFE9: "alt", 0xFFEA: "alt",
	0xFFEB: "cmd", 0xFFEC: "cmd",

	// Navigation
	0xFF51: "left", 0xFF52: "up", 0xFF53: "right", 0xFF54: "down",
	0xFF50: "home", 0xFF57: "end",
	0xFF55: "pageup", 0xFF56: "pagedown",
	0xFF63: "insert", 0xFFFF: "delete",

	// Editing
	0xFF08: "backspace",
	0xFF09: "tab",
	0xFF0D: "enter",
	0xFF1B: "escape",
	0xFF20: "capslock",

	// Numpad
	0xFFB0: "num0", 0xFFB1: "num1", 0xFFB2: "num2", 0xFFB3: "num3",
	0xFFB4: "num4", 0xFFB5: "num5", 0xFFB6: "num6", 0xFFB7: "num7",
	0xFFB8: "num8", 0xFFB9: "num9",
	0xFFAA: "multiply", 0xFFAB: "add", 0xFFAD: "subtract",
	0xFFAE: "decimal", 0xFFAF: "divide",
	0xFF7F: "numlock", 0xFF14: "scrolllock",
	0xFF61: "print",

	// Space
	0xFF80: "space", 0x0020: "space",
}
