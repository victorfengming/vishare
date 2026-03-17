//go:build linux

package input

import (
	hook "github.com/robotn/gohook"
)

// HookEvent mirrors the relevant fields from hook.Event.
type HookEvent struct {
	Kind    uint8
	Keycode uint16
	Rawcode uint16
	Mask    uint16
	Button  uint16
	X, Y    int16
}

const (
	EventKeyDown    = hook.KeyDown
	EventKeyUp      = hook.KeyUp
	EventMouseMove  = hook.MouseMove
	EventMouseDown  = hook.MouseDown
	EventMouseUp    = hook.MouseUp
	EventMouseWheel = hook.MouseWheel
)

// StartCapture starts the global input hook and sends events on the returned channel.
// Call StopCapture to stop.
func StartCapture() (<-chan HookEvent, error) {
	ch := make(chan HookEvent, 256)
	evCh := hook.Start()
	go func() {
		for ev := range evCh {
			ch <- HookEvent{
				Kind:    ev.Kind,
				Keycode: ev.Keycode,
				Rawcode: ev.Rawcode,
				Mask:    ev.Mask,
				Button:  ev.Button,
				X:       ev.X,
				Y:       ev.Y,
			}
		}
		close(ch)
	}()
	return ch, nil
}

// StopCapture tears down the global hook.
func StopCapture() {
	hook.End()
}
