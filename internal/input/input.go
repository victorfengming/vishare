package input

// Injector is the platform-agnostic interface for replaying input events
// received from the server.
type Injector interface {
	// MoveMouse moves the cursor to the given absolute screen coordinates.
	MoveMouse(x, y int)

	// MouseButton presses (down=true) or releases (down=false) the given button.
	// button: 1=left, 2=middle, 3=right (matches gohook Button values).
	MouseButton(button uint16, down bool)

	// MouseWheel scrolls by the given deltas (screen units).
	MouseWheel(deltaX, deltaY int)

	// KeyDown presses the key identified by keycode/rawcode/mask.
	KeyDown(keycode, rawcode, mask uint16)

	// KeyUp releases the key identified by keycode/rawcode/mask.
	KeyUp(keycode, rawcode, mask uint16)
}
