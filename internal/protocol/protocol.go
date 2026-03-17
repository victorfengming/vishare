package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MsgType uint8

const (
	MsgMouseMove   MsgType = 0x01
	MsgMouseButton MsgType = 0x02
	MsgMouseWheel  MsgType = 0x03
	MsgKeyDown     MsgType = 0x04
	MsgKeyUp       MsgType = 0x05
	MsgSwitchTo    MsgType = 0x10
	MsgSwitchBack  MsgType = 0x11
	MsgHandshake   MsgType = 0x20
	MsgPing        MsgType = 0x21
	MsgPong        MsgType = 0x22
)

type Message struct {
	Type    MsgType
	Payload []byte
}

// WriteMsg writes a framed message: [type:1][payloadLen:2][payload:N]
func WriteMsg(w io.Writer, m Message) error {
	if len(m.Payload) > 0xFFFF {
		return fmt.Errorf("payload too large: %d bytes", len(m.Payload))
	}
	hdr := [3]byte{byte(m.Type), byte(len(m.Payload) >> 8), byte(len(m.Payload))}
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	if len(m.Payload) > 0 {
		_, err := w.Write(m.Payload)
		return err
	}
	return nil
}

// ReadMsg reads a framed message from r.
func ReadMsg(r io.Reader) (Message, error) {
	var hdr [3]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return Message{}, err
	}
	msgType := MsgType(hdr[0])
	payloadLen := uint16(hdr[1])<<8 | uint16(hdr[2])
	var payload []byte
	if payloadLen > 0 {
		payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(r, payload); err != nil {
			return Message{}, err
		}
	}
	return Message{Type: msgType, Payload: payload}, nil
}

// ---- Payload encode/decode helpers ----

// MouseMovePayload: X int16, Y int16, ClientW int16, ClientH int16 — 8 bytes
type MouseMovePayload struct {
	X, Y           int16
	ClientW, ClientH int16
}

func EncodeMouseMove(p MouseMovePayload) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint16(b[0:], uint16(p.X))
	binary.BigEndian.PutUint16(b[2:], uint16(p.Y))
	binary.BigEndian.PutUint16(b[4:], uint16(p.ClientW))
	binary.BigEndian.PutUint16(b[6:], uint16(p.ClientH))
	return b
}

func DecodeMouseMove(b []byte) (MouseMovePayload, error) {
	if len(b) < 8 {
		return MouseMovePayload{}, fmt.Errorf("MouseMove payload too short")
	}
	return MouseMovePayload{
		X:       int16(binary.BigEndian.Uint16(b[0:])),
		Y:       int16(binary.BigEndian.Uint16(b[2:])),
		ClientW: int16(binary.BigEndian.Uint16(b[4:])),
		ClientH: int16(binary.BigEndian.Uint16(b[6:])),
	}, nil
}

// MouseButtonPayload: Button uint16, Down uint8 — 3 bytes
type MouseButtonPayload struct {
	Button uint16
	Down   uint8
}

func EncodeMouseButton(p MouseButtonPayload) []byte {
	return []byte{byte(p.Button >> 8), byte(p.Button), p.Down}
}

func DecodeMouseButton(b []byte) (MouseButtonPayload, error) {
	if len(b) < 3 {
		return MouseButtonPayload{}, fmt.Errorf("MouseButton payload too short")
	}
	return MouseButtonPayload{Button: uint16(b[0])<<8 | uint16(b[1]), Down: b[2]}, nil
}

// MouseWheelPayload: DeltaX int16, DeltaY int16 — 4 bytes
type MouseWheelPayload struct {
	DeltaX, DeltaY int16
}

func EncodeMouseWheel(p MouseWheelPayload) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b[0:], uint16(p.DeltaX))
	binary.BigEndian.PutUint16(b[2:], uint16(p.DeltaY))
	return b
}

func DecodeMouseWheel(b []byte) (MouseWheelPayload, error) {
	if len(b) < 4 {
		return MouseWheelPayload{}, fmt.Errorf("MouseWheel payload too short")
	}
	return MouseWheelPayload{
		DeltaX: int16(binary.BigEndian.Uint16(b[0:])),
		DeltaY: int16(binary.BigEndian.Uint16(b[2:])),
	}, nil
}

// KeyPayload: Keycode uint16, Rawcode uint16, Mask uint16 — 6 bytes
type KeyPayload struct {
	Keycode, Rawcode, Mask uint16
}

func EncodeKey(p KeyPayload) []byte {
	b := make([]byte, 6)
	binary.BigEndian.PutUint16(b[0:], p.Keycode)
	binary.BigEndian.PutUint16(b[2:], p.Rawcode)
	binary.BigEndian.PutUint16(b[4:], p.Mask)
	return b
}

func DecodeKey(b []byte) (KeyPayload, error) {
	if len(b) < 6 {
		return KeyPayload{}, fmt.Errorf("Key payload too short")
	}
	return KeyPayload{
		Keycode: binary.BigEndian.Uint16(b[0:]),
		Rawcode: binary.BigEndian.Uint16(b[2:]),
		Mask:    binary.BigEndian.Uint16(b[4:]),
	}, nil
}

// SwitchToPayload: ScreenName [32]byte — 32 bytes
func EncodeSwitchTo(screenName string) []byte {
	b := make([]byte, 32)
	copy(b, screenName)
	return b
}

func DecodeSwitchTo(b []byte) (string, error) {
	if len(b) < 32 {
		return "", fmt.Errorf("SwitchTo payload too short")
	}
	n := 32
	for i, c := range b[:32] {
		if c == 0 {
			n = i
			break
		}
	}
	return string(b[:n]), nil
}

// HandshakePayload: ScreenName [32]byte, ScreenW uint16, ScreenH uint16 — 36 bytes
type HandshakePayload struct {
	ScreenName      string
	ScreenW, ScreenH uint16
}

func EncodeHandshake(p HandshakePayload) []byte {
	b := make([]byte, 36)
	copy(b, p.ScreenName)
	binary.BigEndian.PutUint16(b[32:], p.ScreenW)
	binary.BigEndian.PutUint16(b[34:], p.ScreenH)
	return b
}

func DecodeHandshake(b []byte) (HandshakePayload, error) {
	if len(b) < 36 {
		return HandshakePayload{}, fmt.Errorf("Handshake payload too short")
	}
	n := 32
	for i, c := range b[:32] {
		if c == 0 {
			n = i
			break
		}
	}
	return HandshakePayload{
		ScreenName: string(b[:n]),
		ScreenW:    binary.BigEndian.Uint16(b[32:]),
		ScreenH:    binary.BigEndian.Uint16(b[34:]),
	}, nil
}
