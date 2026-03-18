//go:build linux

package input

import (
	"fmt"
	"sync"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

var (
	grabMu      sync.Mutex
	grabConn    *xgb.Conn
	inputGrabOn bool
)

func GrabLocalInput() error {
	grabMu.Lock()
	defer grabMu.Unlock()

	if inputGrabOn {
		return nil
	}

	conn, err := xgb.NewConn()
	if err != nil {
		return fmt.Errorf("connect X server: %w", err)
	}

	root := xproto.Setup(conn).DefaultScreen(conn).Root

	ptr, err := xproto.GrabPointer(
		conn,
		false,
		root,
		xproto.EventMaskButtonPress|
			xproto.EventMaskButtonRelease|
			xproto.EventMaskPointerMotion,
		xproto.GrabModeAsync,
		xproto.GrabModeAsync,
		0,
		0,
		xproto.TimeCurrentTime,
	).Reply()
	if err != nil {
		conn.Close()
		return fmt.Errorf("grab pointer: %w", err)
	}
	if ptr == nil || ptr.Status != xproto.GrabStatusSuccess {
		conn.Close()
		return fmt.Errorf("grab pointer status=%d", ptr.Status)
	}

	key, err := xproto.GrabKeyboard(
		conn,
		false,
		root,
		xproto.TimeCurrentTime,
		xproto.GrabModeAsync,
		xproto.GrabModeAsync,
	).Reply()
	if err != nil {
		_ = xproto.UngrabPointerChecked(conn, xproto.TimeCurrentTime).Check()
		conn.Close()
		return fmt.Errorf("grab keyboard: %w", err)
	}
	if key == nil || key.Status != xproto.GrabStatusSuccess {
		_ = xproto.UngrabPointerChecked(conn, xproto.TimeCurrentTime).Check()
		conn.Close()
		return fmt.Errorf("grab keyboard status=%d", key.Status)
	}

	conn.Sync()

	grabConn = conn
	inputGrabOn = true
	return nil
}

func ReleaseLocalInput() {
	grabMu.Lock()
	defer grabMu.Unlock()

	if !inputGrabOn || grabConn == nil {
		return
	}

	_ = xproto.UngrabKeyboardChecked(grabConn, xproto.TimeCurrentTime).Check()
	_ = xproto.UngrabPointerChecked(grabConn, xproto.TimeCurrentTime).Check()
	grabConn.Sync()
	grabConn.Close()
	grabConn = nil
	inputGrabOn = false
}
