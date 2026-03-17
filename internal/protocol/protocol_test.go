package protocol_test

import (
	"net"
	"testing"

	"github.com/victorfengming/vishare/internal/protocol"
)

func pipe() (net.Conn, net.Conn) {
	c1, c2 := net.Pipe()
	return c1, c2
}

func roundtrip(t *testing.T, m protocol.Message) protocol.Message {
	t.Helper()
	c1, c2 := pipe()
	defer c1.Close()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- protocol.WriteMsg(c1, m)
		c1.Close()
	}()

	got, err := protocol.ReadMsg(c2)
	if err != nil {
		t.Fatalf("ReadMsg: %v", err)
	}
	if werr := <-errCh; werr != nil {
		t.Fatalf("WriteMsg: %v", werr)
	}
	return got
}

func TestMouseMove(t *testing.T) {
	p := protocol.MouseMovePayload{X: 100, Y: -50}
	msg := protocol.Message{Type: protocol.MsgMouseMove, Payload: protocol.EncodeMouseMove(p)}
	got := roundtrip(t, msg)

	if got.Type != protocol.MsgMouseMove {
		t.Fatalf("type mismatch: got %d want %d", got.Type, protocol.MsgMouseMove)
	}
	dec, err := protocol.DecodeMouseMove(got.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if dec != p {
		t.Fatalf("payload mismatch: got %+v want %+v", dec, p)
	}
}

func TestMouseButton(t *testing.T) {
	p := protocol.MouseButtonPayload{Button: 1, Down: 1}
	msg := protocol.Message{Type: protocol.MsgMouseButton, Payload: protocol.EncodeMouseButton(p)}
	got := roundtrip(t, msg)

	dec, err := protocol.DecodeMouseButton(got.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if dec != p {
		t.Fatalf("payload mismatch: got %+v want %+v", dec, p)
	}
}

func TestMouseWheel(t *testing.T) {
	p := protocol.MouseWheelPayload{DeltaX: 3, DeltaY: -5}
	msg := protocol.Message{Type: protocol.MsgMouseWheel, Payload: protocol.EncodeMouseWheel(p)}
	got := roundtrip(t, msg)

	dec, err := protocol.DecodeMouseWheel(got.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if dec != p {
		t.Fatalf("payload mismatch: got %+v want %+v", dec, p)
	}
}

func TestKey(t *testing.T) {
	for _, msgType := range []protocol.MsgType{protocol.MsgKeyDown, protocol.MsgKeyUp} {
		p := protocol.KeyPayload{Keycode: 0xFF51, Rawcode: 0x71, Mask: 0x0004}
		msg := protocol.Message{Type: msgType, Payload: protocol.EncodeKey(p)}
		got := roundtrip(t, msg)

		if got.Type != msgType {
			t.Fatalf("type mismatch")
		}
		dec, err := protocol.DecodeKey(got.Payload)
		if err != nil {
			t.Fatal(err)
		}
		if dec != p {
			t.Fatalf("payload mismatch: got %+v want %+v", dec, p)
		}
	}
}

func TestSwitchTo(t *testing.T) {
	name := "windows-pc"
	msg := protocol.Message{Type: protocol.MsgSwitchTo, Payload: protocol.EncodeSwitchTo(name)}
	got := roundtrip(t, msg)

	dec, err := protocol.DecodeSwitchTo(got.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if dec != name {
		t.Fatalf("name mismatch: got %q want %q", dec, name)
	}
}

func TestHandshake(t *testing.T) {
	p := protocol.HandshakePayload{ScreenName: "ubuntu-main", ScreenW: 2560, ScreenH: 1440}
	msg := protocol.Message{Type: protocol.MsgHandshake, Payload: protocol.EncodeHandshake(p)}
	got := roundtrip(t, msg)

	dec, err := protocol.DecodeHandshake(got.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if dec != p {
		t.Fatalf("payload mismatch: got %+v want %+v", dec, p)
	}
}

func TestPingPong(t *testing.T) {
	for _, msgType := range []protocol.MsgType{protocol.MsgPing, protocol.MsgPong} {
		msg := protocol.Message{Type: msgType}
		got := roundtrip(t, msg)
		if got.Type != msgType {
			t.Fatalf("type mismatch: got %d want %d", got.Type, msgType)
		}
		if len(got.Payload) != 0 {
			t.Fatalf("expected empty payload, got %d bytes", len(got.Payload))
		}
	}
}

func TestSwitchBack(t *testing.T) {
	msg := protocol.Message{Type: protocol.MsgSwitchBack}
	got := roundtrip(t, msg)
	if got.Type != protocol.MsgSwitchBack {
		t.Fatalf("type mismatch")
	}
}
