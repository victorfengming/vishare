//go:build windows

package input

// Windows 上没有全局 hook 捕获（server 不在 Windows 运行），提供空实现供编译通过。

type HookEvent struct {
	Kind    uint8
	Keycode uint16
	Rawcode uint16
	Mask    uint16
	Button  uint16
	X, Y    int16
}

const (
	EventKeyDown    = uint8(3)
	EventKeyUp      = uint8(5)
	EventMouseMove  = uint8(9)
	EventMouseDown  = uint8(8)
	EventMouseUp    = uint8(6)
	EventMouseWheel = uint8(11)
)

func StartCapture() (<-chan HookEvent, error) {
	ch := make(chan HookEvent)
	return ch, nil
}

func StopCapture() {}
