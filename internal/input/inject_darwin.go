//go:build darwin

package input

import (
	"github.com/go-vgo/robotgo"
)

type DarwinInjector struct{}

func NewInjector() Injector {
	return &DarwinInjector{}
}

func (d *DarwinInjector) MoveMouse(x, y int) {
	robotgo.Move(x, y)
}

func (d *DarwinInjector) MouseButton(button uint16, down bool) {
	b := darwinButton(button)
	if down {
		robotgo.MouseDown(b)
	} else {
		robotgo.MouseUp(b)
	}
}

func (d *DarwinInjector) MouseWheel(deltaX, deltaY int) {
	if deltaX != 0 || deltaY != 0 {
		robotgo.Scroll(deltaX, deltaY)
	}
}

func (d *DarwinInjector) KeyDown(keycode, rawcode, mask uint16) {
	name := keycodeToName(keycode)
	if name == "" {
		return
	}
	robotgo.KeyDown(name)
}

func (d *DarwinInjector) KeyUp(keycode, rawcode, mask uint16) {
	name := keycodeToName(keycode)
	if name == "" {
		return
	}
	robotgo.KeyUp(name)
}

func darwinButton(button uint16) string {
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
