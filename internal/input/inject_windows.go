//go:build windows

package input

// WindowsInjector implements Injector using Win32 SendInput on Windows.
// This is a stub — the SendInput implementation will be added later.
type WindowsInjector struct{}

func NewInjector() Injector {
	return &WindowsInjector{}
}

func (w *WindowsInjector) MoveMouse(x, y int) {
	// TODO: implement via SendInput MOUSEEVENTF_MOVE | MOUSEEVENTF_ABSOLUTE
}

func (w *WindowsInjector) MouseButton(button uint16, down bool) {
	// TODO: implement via SendInput MOUSEEVENTF_LEFTDOWN / LEFTUP etc.
}

func (w *WindowsInjector) MouseWheel(deltaX, deltaY int) {
	// TODO: implement via SendInput MOUSEEVENTF_WHEEL / MOUSEEVENTF_HWHEEL
}

func (w *WindowsInjector) KeyDown(keycode, rawcode, mask uint16) {
	// TODO: implement via SendInput KEYEVENTF_SCANCODE
}

func (w *WindowsInjector) KeyUp(keycode, rawcode, mask uint16) {
	// TODO: implement via SendInput KEYEVENTF_SCANCODE | KEYEVENTF_KEYUP
}
