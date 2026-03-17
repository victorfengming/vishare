package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/rs/zerolog/log"
	"github.com/victorfengming/vishare/internal/config"
	"github.com/victorfengming/vishare/internal/input"
	"github.com/victorfengming/vishare/internal/protocol"
)

const (
	sendBufSize   = 256
	pingInterval  = 5 * time.Second
	edgePollMs    = 10 * time.Millisecond
	edgeHysteresis = 3 // frames
)

// StatusMsg is sent on the status channel to update the tray.
type StatusMsg struct {
	Connected bool
	ClientName string
}

type clientState struct {
	conn       net.Conn
	sendCh     chan protocol.Message
	screenName string
	screenW    uint16
	screenH    uint16
}

type Server struct {
	cfg      *config.Config
	statusCh chan<- StatusMsg

	mu         sync.Mutex
	activeClient *clientState  // currently controlled client (nil = local)
	clients    map[string]*clientState

	localW, localH int
}

func New(cfg *config.Config, statusCh chan<- StatusMsg) *Server {
	w, h := robotgo.GetScreenSize()
	return &Server{
		cfg:      cfg,
		statusCh: statusCh,
		clients:  make(map[string]*clientState),
		localW:   w,
		localH:   h,
	}
}

func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.cfg.ListenAddr, err)
	}
	log.Info().Str("addr", s.cfg.ListenAddr).Msg("server listening")

	// Start input capture
	eventCh, err := input.StartCapture()
	if err != nil {
		ln.Close()
		return fmt.Errorf("start capture: %w", err)
	}
	defer input.StopCapture()

	// Accept loop
	go s.acceptLoop(ctx, ln)

	// Edge detection + event dispatcher
	go s.dispatcher(ctx, eventCh)

	<-ctx.Done()
	ln.Close()
	return nil
}

func (s *Server) acceptLoop(ctx context.Context, ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Error().Err(err).Msg("accept error")
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		go s.handleClient(ctx, conn)
	}
}

func (s *Server) handleClient(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	log.Info().Str("addr", addr).Msg("client connected")

	// Expect handshake first
	msg, err := protocol.ReadMsg(conn)
	if err != nil || msg.Type != protocol.MsgHandshake {
		log.Error().Str("addr", addr).Msg("expected handshake")
		return
	}
	hs, err := protocol.DecodeHandshake(msg.Payload)
	if err != nil {
		log.Error().Err(err).Msg("decode handshake")
		return
	}
	log.Info().Str("name", hs.ScreenName).Uint16("w", hs.ScreenW).Uint16("h", hs.ScreenH).Msg("client handshake")

	cs := &clientState{
		conn:       conn,
		sendCh:     make(chan protocol.Message, sendBufSize),
		screenName: hs.ScreenName,
		screenW:    hs.ScreenW,
		screenH:    hs.ScreenH,
	}

	s.mu.Lock()
	s.clients[hs.ScreenName] = cs
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, hs.ScreenName)
		if s.activeClient == cs {
			s.activeClient = nil
		}
		s.mu.Unlock()
		if s.statusCh != nil {
			s.statusCh <- StatusMsg{Connected: false}
		}
		log.Info().Str("name", hs.ScreenName).Msg("client disconnected")
	}()

	if s.statusCh != nil {
		s.statusCh <- StatusMsg{Connected: true, ClientName: hs.ScreenName}
	}

	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	go s.writeLoop(ctx2, cs)
	go s.keepalive(ctx2, cs)
	s.readLoop(ctx2, cs)
}

func (s *Server) readLoop(ctx context.Context, cs *clientState) {
	for {
		msg, err := protocol.ReadMsg(cs.conn)
		if err != nil {
			return
		}
		switch msg.Type {
		case protocol.MsgSwitchBack:
			log.Info().Str("client", cs.screenName).Msg("client returned control")
			s.mu.Lock()
			if s.activeClient == cs {
				s.activeClient = nil
			}
			s.mu.Unlock()
		case protocol.MsgPong:
			// keepalive response, ignore
		default:
			log.Warn().Uint8("type", uint8(msg.Type)).Msg("unexpected msg from client")
		}
	}
}

func (s *Server) writeLoop(ctx context.Context, cs *clientState) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-cs.sendCh:
			if !ok {
				return
			}
			if err := protocol.WriteMsg(cs.conn, msg); err != nil {
				return
			}
		}
	}
}

func (s *Server) keepalive(ctx context.Context, cs *clientState) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			select {
			case cs.sendCh <- protocol.Message{Type: protocol.MsgPing}:
			default:
			}
		}
	}
}

// dispatcher handles input events and edge detection.
func (s *Server) dispatcher(ctx context.Context, eventCh <-chan input.HookEvent) {
	edgeTicker := time.NewTicker(edgePollMs)
	defer edgeTicker.Stop()

	var edgeCount atomic.Int32

	for {
		select {
		case <-ctx.Done():
			return

		case ev, ok := <-eventCh:
			if !ok {
				return
			}
			s.mu.Lock()
			ac := s.activeClient
			s.mu.Unlock()

			if ac == nil {
				// Local control — forward to active client if switching
				// (edge detection handles the switch; no forwarding needed here
				//  except if we want to suppress events — we don't for now)
				continue
			}
			// Forward event to active client
			s.forwardEvent(ac, ev)

		case <-edgeTicker.C:
			s.mu.Lock()
			ac := s.activeClient
			s.mu.Unlock()

			if ac != nil {
				// Client has control; edge detection on client side
				continue
			}

			// Check if cursor is at an edge of the local screen
			cx, cy := robotgo.GetMousePos()
			targetScreen := s.edgeTarget(cx, cy)
			if targetScreen == "" {
				edgeCount.Store(0)
				continue
			}
			cnt := edgeCount.Add(1)
			if cnt < edgeHysteresis {
				continue
			}
			edgeCount.Store(0)

			s.mu.Lock()
			target, ok := s.clients[targetScreen]
			s.mu.Unlock()
			if !ok {
				log.Warn().Str("screen", targetScreen).Msg("target screen not connected")
				continue
			}

			// Warp cursor to center
			robotgo.Move(s.localW/2, s.localH/2)

			s.mu.Lock()
			s.activeClient = target
			s.mu.Unlock()

			log.Info().Str("target", targetScreen).Msg("switching to client")
			select {
			case target.sendCh <- protocol.Message{
				Type:    protocol.MsgSwitchTo,
				Payload: protocol.EncodeSwitchTo(targetScreen),
			}:
			default:
			}
		}
	}
}

// edgeTarget returns the screen name to switch to if cursor is at an edge.
func (s *Server) edgeTarget(cx, cy int) string {
	const margin = 2
	// Find server's own screen config
	var srv *config.ScreenConfig
	for i := range s.cfg.Screens {
		if s.cfg.Screens[i].Name == "" || s.cfg.Screens[i].EdgeLeft != "" ||
			s.cfg.Screens[i].EdgeRight != "" || s.cfg.Screens[i].EdgeTop != "" ||
			s.cfg.Screens[i].EdgeBottom != "" {
			// heuristic: find the screen with edges pointing to others (server screen)
		}
		// Just iterate all screens and use first that has edges defined
		sc := &s.cfg.Screens[i]
		if sc.EdgeLeft != "" || sc.EdgeRight != "" || sc.EdgeTop != "" || sc.EdgeBottom != "" {
			srv = sc
			break
		}
	}
	if srv == nil && len(s.cfg.Screens) > 0 {
		srv = &s.cfg.Screens[0]
	}
	if srv == nil {
		return ""
	}

	if cx <= margin && srv.EdgeLeft != "" {
		return srv.EdgeLeft
	}
	if cx >= s.localW-1-margin && srv.EdgeRight != "" {
		return srv.EdgeRight
	}
	if cy <= margin && srv.EdgeTop != "" {
		return srv.EdgeTop
	}
	if cy >= s.localH-1-margin && srv.EdgeBottom != "" {
		return srv.EdgeBottom
	}
	return ""
}

// forwardEvent translates and sends an input event to the client.
func (s *Server) forwardEvent(cs *clientState, ev input.HookEvent) {
	var msg protocol.Message
	switch ev.Kind {
	case input.EventMouseMove:
		// Translate server coords → client coords
		clientX := int16(float64(ev.X) / float64(s.localW) * float64(cs.screenW))
		clientY := int16(float64(ev.Y) / float64(s.localH) * float64(cs.screenH))
		msg = protocol.Message{
			Type: protocol.MsgMouseMove,
			Payload: protocol.EncodeMouseMove(protocol.MouseMovePayload{
				X: clientX, Y: clientY,
				ClientW: int16(cs.screenW), ClientH: int16(cs.screenH),
			}),
		}
	case input.EventMouseDown:
		msg = protocol.Message{
			Type:    protocol.MsgMouseButton,
			Payload: protocol.EncodeMouseButton(protocol.MouseButtonPayload{Button: ev.Button, Down: 1}),
		}
	case input.EventMouseUp:
		msg = protocol.Message{
			Type:    protocol.MsgMouseButton,
			Payload: protocol.EncodeMouseButton(protocol.MouseButtonPayload{Button: ev.Button, Down: 0}),
		}
	case input.EventMouseWheel:
		msg = protocol.Message{
			Type:    protocol.MsgMouseWheel,
			Payload: protocol.EncodeMouseWheel(protocol.MouseWheelPayload{DeltaX: int16(ev.X), DeltaY: int16(ev.Y)}),
		}
	case input.EventKeyDown:
		msg = protocol.Message{
			Type:    protocol.MsgKeyDown,
			Payload: protocol.EncodeKey(protocol.KeyPayload{Keycode: ev.Keycode, Rawcode: ev.Rawcode, Mask: ev.Mask}),
		}
	case input.EventKeyUp:
		msg = protocol.Message{
			Type:    protocol.MsgKeyUp,
			Payload: protocol.EncodeKey(protocol.KeyPayload{Keycode: ev.Keycode, Rawcode: ev.Rawcode, Mask: ev.Mask}),
		}
	default:
		return
	}

	select {
	case cs.sendCh <- msg:
	default:
		log.Warn().Msg("client send buffer full, dropping event")
	}
}
