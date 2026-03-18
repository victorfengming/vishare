package client

import (
	"context"
	"fmt"
	"math"
	"net"
	"sync/atomic"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/rs/zerolog/log"
	"github.com/victorfengming/vishare/internal/config"
	"github.com/victorfengming/vishare/internal/defaults"
	"github.com/victorfengming/vishare/internal/input"
	"github.com/victorfengming/vishare/internal/protocol"
	"github.com/victorfengming/vishare/internal/status"
)

type Client struct {
	cfg        *config.Config
	statusCh   chan<- status.Msg
	injector   input.Injector
	hasControl atomic.Bool
	skipEdge   atomic.Bool

	screenW, screenH int
}

func New(cfg *config.Config, statusCh chan<- status.Msg) *Client {
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
	backoff := defaults.BackoffMin
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
				c.statusCh <- status.Msg{Connected: false}
			}
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, defaults.BackoffMax)
			continue
		}
		backoff = defaults.BackoffMin
		log.Info().Str("addr", c.cfg.ServerAddr).Msg("connected to server")

		if err := c.session(ctx, conn); err != nil {
			log.Error().Err(err).Msg("session ended")
		}
		if c.statusCh != nil {
			c.statusCh <- status.Msg{Connected: false}
		}
		conn.Close()
	}
}

func (c *Client) session(ctx context.Context, conn net.Conn) error {
	hs := protocol.EncodeHandshake(protocol.HandshakePayload{
		ScreenName: c.cfg.ScreenName,
		Secret:     c.cfg.Secret,
		ScreenW:    uint16(c.screenW),
		ScreenH:    uint16(c.screenH),
	})
	if err := protocol.WriteMsg(conn, protocol.Message{Type: protocol.MsgHandshake, Payload: hs}); err != nil {
		return fmt.Errorf("send handshake: %w", err)
	}

	if c.statusCh != nil {
		c.statusCh <- status.Msg{Connected: true}
	}

	sendCh := make(chan protocol.Message, defaults.SendBufSize)
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
			c.skipEdge.Store(true)
			c.hasControl.Store(true)
		case protocol.MsgMouseMove:
			if !c.hasControl.Load() {
				continue
			}
			p, err := protocol.DecodeMouseMove(msg.Payload)
			if err != nil {
				continue
			}
			c.moveMouse(int(p.X), int(p.Y))
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
	ticker := time.NewTicker(defaults.PingInterval)
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
	ticker := time.NewTicker(defaults.EdgePollMs)
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
			if c.skipEdge.Load() {
				edgeCount = 0
				if !c.atEdge(cx, cy) {
					c.skipEdge.Store(false)
				}
				continue
			}
			if c.atEdge(cx, cy) {
				edgeCount++
				if edgeCount >= defaults.EdgeHysteresis {
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

func (c *Client) moveMouse(targetX, targetY int) {
	targetX = clamp(targetX, 0, c.screenW-1)
	targetY = clamp(targetY, 0, c.screenH-1)

	speed := c.cfg.MouseSpeed
	if speed <= 0 {
		speed = defaults.MouseSpeed
	}
	if speed >= 1 {
		c.injector.MoveMouse(targetX, targetY)
		return
	}

	cx, cy := robotgo.GetMousePos()
	nextX := scaledCoord(cx, targetX, speed)
	nextY := scaledCoord(cy, targetY, speed)
	c.injector.MoveMouse(
		clamp(nextX, 0, c.screenW-1),
		clamp(nextY, 0, c.screenH-1),
	)
}

func scaledCoord(current, target int, speed float64) int {
	if current == target {
		return current
	}

	step := int(math.Round(float64(target-current) * speed))
	if step == 0 {
		if target > current {
			return current + 1
		}
		return current - 1
	}

	next := current + step
	if target > current && next > target {
		return target
	}
	if target < current && next < target {
		return target
	}
	return next
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
