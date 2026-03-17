//go:build linux

package input

import (
	"github.com/go-vgo/robotgo"
)

// LinuxInjector implements Injector using robotgo on Linux/X11.
type LinuxInjector struct{}

func NewInjector() Injector {
	return &LinuxInjector{}
}

func (l *LinuxInjector) MoveMouse(x, y int) {
	robotgo.Move(x, y)
}

func (l *LinuxInjector) MouseButton(button uint16, down bool) {
	b := robotgoButton(button)
	if down {
		robotgo.MouseDown(b)
	} else {
		robotgo.MouseUp(b)
	}
}

func (l *LinuxInjector) MouseWheel(deltaX, deltaY int) {
	if deltaX != 0 || deltaY != 0 {
		robotgo.Scroll(deltaX, deltaY)
	}
}

func (l *LinuxInjector) KeyDown(keycode, rawcode, mask uint16) {
	name := keycodeToName(keycode)
	if name == "" {
		return
	}
	robotgo.KeyDown(name)
}

func (l *LinuxInjector) KeyUp(keycode, rawcode, mask uint16) {
	name := keycodeToName(keycode)
	if name == "" {
		return
	}
	robotgo.KeyUp(name)
}

func robotgoButton(button uint16) string {
	switch button {
	case 2:
		return "center"
	case 3:
		return "right"
	default:
		return "left"
	}
}

func keycodeToName(keycode uint16) string {
	if name, ok := keysymTable[keycode]; ok {
		return name
	}
	if keycode >= 0x20 && keycode <= 0x7e {
		return string(rune(keycode))
	}
	return ""
}
