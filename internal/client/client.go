package client

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/rs/zerolog/log"
	"github.com/uisee/vishare/internal/config"
	"github.com/uisee/vishare/internal/input"
	"github.com/uisee/vishare/internal/protocol"
)

const (
	sendBufSize    = 256
	pingInterval   = 5 * time.Second
	edgePollMs     = 10 * time.Millisecond
	edgeHysteresis = 3
	backoffMin     = 1 * time.Second
	backoffMax     = 30 * time.Second
)

// StatusMsg is sent on the status channel to update the tray.
type StatusMsg struct {
	Connected bool
}

type Client struct {
	cfg        *config.Config
	statusCh   chan<- StatusMsg
	injector   input.Injector
	hasControl atomic.Bool

	screenW, screenH int
}

func New(cfg *config.Config, statusCh chan<- StatusMsg) *Client {
	w, h := robotgo.GetScreenSize()
	return &Client{
		cfg:      cfg,
		statusCh: statusCh,
		injector: input.NewInjector(),
		screenW:  w,
		screenH:  h,
	}
}

func (c *Client) Run(ctx context.Context) error {
	backoff := backoffMin
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn, err := net.DialTimeout("tcp", c.cfg.ServerAddr, 5*time.Second)
		if err != nil {
			log.Error().Err(err).Str("addr", c.cfg.ServerAddr).Msg("connect failed, retrying")
			if c.statusCh != nil {
				c.statusCh <- StatusMsg{Connected: false}
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, backoffMax)
			continue
		}
		backoff = backoffMin
		log.Info().Str("addr", c.cfg.ServerAddr).Msg("connected to server")

		if err := c.session(ctx, conn); err != nil {
			log.Error().Err(err).Msg("session ended")
		}
		if c.statusCh != nil {
			c.statusCh <- StatusMsg{Connected: false}
		}
		conn.Close()
	}
}

func (c *Client) session(ctx context.Context, conn net.Conn) error {
	// Send handshake
	hs := protocol.EncodeHandshake(protocol.HandshakePayload{
		ScreenName: c.cfg.ScreenName,
		ScreenW:    uint16(c.screenW),
		ScreenH:    uint16(c.screenH),
	})
	if err := protocol.WriteMsg(conn, protocol.Message{Type: protocol.MsgHandshake, Payload: hs}); err != nil {
		return fmt.Errorf("send handshake: %w", err)
	}

	if c.statusCh != nil {
		c.statusCh <- StatusMsg{Connected: true}
	}

	sendCh := make(chan protocol.Message, sendBufSize)
	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	c.hasControl.Store(false)

	go c.writeLoop(ctx2, conn, sendCh)
	go c.keepalive(ctx2, sendCh)
	go c.edgePoller(ctx2, sendCh)
	return c.readLoop(ctx2, conn, sendCh)
}

func (c *Client) readLoop(ctx context.Context, conn net.Conn, sendCh chan protocol.Message) error {
	for {
		msg, err := protocol.ReadMsg(conn)
		if err != nil {
			return err
		}
		switch msg.Type {
		case protocol.MsgSwitchTo:
			log.Info().Msg("got control")
			c.hasControl.Store(true)
		case protocol.MsgMouseMove:
			if !c.hasControl.Load() {
				continue
			}
			p, err := protocol.DecodeMouseMove(msg.Payload)
			if err != nil {
				continue
			}
			c.injector.MoveMouse(int(p.X), int(p.Y))
		case protocol.MsgMouseButton:
			if !c.hasControl.Load() {
				continue
			}
			p, err := protocol.DecodeMouseButton(msg.Payload)
			if err != nil {
				continue
			}
			c.injector.MouseButton(p.Button, p.Down == 1)
		case protocol.MsgMouseWheel:
			if !c.hasControl.Load() {
				continue
			}
			p, err := protocol.DecodeMouseWheel(msg.Payload)
			if err != nil {
				continue
			}
			c.injector.MouseWheel(int(p.DeltaX), int(p.DeltaY))
		case protocol.MsgKeyDown:
			if !c.hasControl.Load() {
				continue
			}
			p, err := protocol.DecodeKey(msg.Payload)
			if err != nil {
				continue
			}
			c.injector.KeyDown(p.Keycode, p.Rawcode, p.Mask)
		case protocol.MsgKeyUp:
			if !c.hasControl.Load() {
				continue
			}
			p, err := protocol.DecodeKey(msg.Payload)
			if err != nil {
				continue
			}
			c.injector.KeyUp(p.Keycode, p.Rawcode, p.Mask)
		case protocol.MsgPing:
			select {
			case sendCh <- protocol.Message{Type: protocol.MsgPong}:
			default:
			}
		case protocol.MsgPong:
			// response to our ping, ignore
		default:
			log.Warn().Uint8("type", uint8(msg.Type)).Msg("unknown msg from server")
		}
	}
}

func (c *Client) writeLoop(ctx context.Context, conn net.Conn, sendCh <-chan protocol.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sendCh:
			if !ok {
				return
			}
			if err := protocol.WriteMsg(conn, msg); err != nil {
				return
			}
		}
	}
}

func (c *Client) keepalive(ctx context.Context, sendCh chan protocol.Message) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case sendCh <- protocol.Message{Type: protocol.MsgPing}:
			default:
			}
		}
	}
}

func (c *Client) edgePoller(ctx context.Context, sendCh chan protocol.Message) {
	ticker := time.NewTicker(edgePollMs)
	defer ticker.Stop()
	edgeCount := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !c.hasControl.Load() {
				edgeCount = 0
				continue
			}
			cx, cy := robotgo.GetMousePos()
			if c.atEdge(cx, cy) {
				edgeCount++
				if edgeCount >= edgeHysteresis {
					edgeCount = 0
					c.hasControl.Store(false)
					log.Info().Msg("returning control to server")
					select {
					case sendCh <- protocol.Message{Type: protocol.MsgSwitchBack}:
					default:
					}
				}
			} else {
				edgeCount = 0
			}
		}
	}
}

func (c *Client) atEdge(cx, cy int) bool {
	const margin = 2
	return cx <= margin || cx >= c.screenW-1-margin ||
		cy <= margin || cy >= c.screenH-1-margin
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
