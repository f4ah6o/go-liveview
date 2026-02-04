package liveview

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"strings"
	"sync"

	"github.com/a-h/templ"
	"github.com/fu2hito/go-liveview/internal/protocol"
	"github.com/fu2hito/go-liveview/internal/render"
	"github.com/fu2hito/go-liveview/internal/socket"
)

// session holds LiveView instance and context together
type session struct {
	lv  LiveView
	ctx *Context
}

// Manager manages LiveView instances
type Manager struct {
	liveViews   map[string]func() LiveView
	sessions    map[string]*session
	broadcaster *Broadcaster
	mu          sync.RWMutex
	server      *socket.Server
}

// SetBroadcaster sets the broadcaster for the manager
func (m *Manager) SetBroadcaster(b *Broadcaster) {
	m.broadcaster = b
}

// GetBroadcaster returns the broadcaster
func (m *Manager) GetBroadcaster() *Broadcaster {
	return m.broadcaster
}

// NewManager creates a new LiveView manager
func NewManager(server *socket.Server) *Manager {
	return &Manager{
		liveViews: make(map[string]func() LiveView),
		sessions:  make(map[string]*session),
		server:    server,
	}
}

// Register registers a LiveView for a topic
func (m *Manager) Register(topic string, factory func() LiveView) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.liveViews[topic] = factory

	// Register WebSocket handler
	m.server.RegisterHandler(topic, m.handleMessage)
}

func (m *Manager) handleMessage(ctx context.Context, conn *socket.Conn, msg *protocol.Message) {
	switch msg.Event {
	case "phx_join":
		m.handleJoin(ctx, conn, msg)
	case "event":
		m.handleEvent(ctx, conn, msg)
	case "phx_leave":
		m.handleLeave(ctx, conn, msg)
	}
}

func (m *Manager) handleJoin(ctx context.Context, conn *socket.Conn, msg *protocol.Message) {
	var joinPayload protocol.JoinPayload
	if err := unmarshalPayload(msg.Payload, &joinPayload); err != nil {
		log.Printf("Failed to unmarshal join payload: %v", err)
		return
	}

	m.mu.RLock()
	factory, ok := m.liveViews[msg.Topic]
	m.mu.RUnlock()

	if !ok {
		log.Printf("Unknown topic: %s", msg.Topic)
		return
	}

	// Create new LiveView instance
	lv := factory()

	// Create context
	lvCtx := NewContext(ctx, &socketAdapter{conn: conn}, conn.ID())

	// Set broadcaster if available
	if m.broadcaster != nil {
		lvCtx.SetBroadcaster(m.broadcaster)
	}

	// Mount the LiveView
	params := url.Values{}
	for k, v := range joinPayload.Params {
		if s, ok := v.(string); ok {
			params.Set(k, s)
		}
	}

	if err := lv.Mount(lvCtx, params); err != nil {
		log.Printf("Failed to mount LiveView: %v", err)
		return
	}

	// Render initial view
	comp := lv.Render(lvCtx)

	// Convert templ component to Rendered
	html := renderComponent(comp)
	r := render.ParseTemplOutput(html)
	lvCtx.SetRenderedValue(&BaseRendered{
		Static:  r.Static,
		Dynamic: r.Dynamic,
	})

	// Store session with LiveView instance
	m.mu.Lock()
	m.sessions[conn.ID()] = &session{lv: lv, ctx: lvCtx}
	m.mu.Unlock()

	// Send join reply
	reply := protocol.NewJoinReply(msg.Topic, *msg.Ref, r)
	conn.Send(reply)
}

func (m *Manager) handleEvent(ctx context.Context, conn *socket.Conn, msg *protocol.Message) {
	var eventPayload protocol.EventPayload
	if err := unmarshalPayload(msg.Payload, &eventPayload); err != nil {
		log.Printf("Failed to unmarshal event payload: %v", err)
		return
	}

	m.mu.RLock()
	sess, ok := m.sessions[conn.ID()]
	m.mu.RUnlock()

	if !ok {
		log.Printf("No session found for connection: %s", conn.ID())
		return
	}

	// Use existing LiveView instance from session
	lv := sess.lv
	lvCtx := sess.ctx

	if err := lv.HandleEvent(lvCtx, eventPayload.Event, eventPayload.Value); err != nil {
		log.Printf("Failed to handle event: %v", err)
		return
	}

	// Re-render
	comp := lv.Render(lvCtx)
	html := renderComponent(comp)
	newRendered := render.ParseTemplOutput(html)

	// Calculate diff
	var prevRendered *render.Rendered
	if r := lvCtx.RenderedValue(); r != nil {
		if br, ok := r.(*BaseRendered); ok {
			prevRendered = &render.Rendered{
				Static:  br.Static,
				Dynamic: br.Dynamic,
			}
		}
	}
	diff := render.Diff(prevRendered, newRendered)

	// Update stored rendered
	lvCtx.SetRenderedValue(&BaseRendered{
		Static:  newRendered.Static,
		Dynamic: newRendered.Dynamic,
	})

	// Send diff
	diffMsg, err := protocol.NewDiffMessage(msg.Topic, protocol.DiffPayload{
		Static:  convertToInterfaceSlice(diff.Static),
		Dynamic: diff.Dynamic,
	})
	if err != nil {
		log.Printf("Failed to create diff message: %v", err)
		return
	}
	conn.Send(diffMsg)
}

func (m *Manager) handleLeave(ctx context.Context, conn *socket.Conn, msg *protocol.Message) {
	m.mu.Lock()
	delete(m.sessions, conn.ID())
	m.mu.Unlock()
}

func renderComponent(comp templ.Component) string {
	var buf strings.Builder
	if err := comp.Render(context.Background(), &buf); err != nil {
		return ""
	}
	return buf.String()
}

func unmarshalPayload(data []byte, target interface{}) error {
	if err := json.Unmarshal(data, target); err == nil {
		return nil
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	return json.Unmarshal([]byte(encoded), target)
}

func convertToInterfaceSlice(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}

// socketAdapter adapts socket.Conn to LiveView Socket interface
type socketAdapter struct {
	conn *socket.Conn
}

func (s *socketAdapter) PushEvent(event string, payload map[string]interface{}) {
	// Implementation
}

func (s *socketAdapter) PutFlash(kind, message string) {
	// Implementation
}

func (s *socketAdapter) AllowUpload(name string, options UploadConfig) {
	// Implementation
}

func (s *socketAdapter) CancelUpload(name string) {
	// Implementation
}

func (s *socketAdapter) ConsumeUploadedEntries(entries []UploadEntry, fn func(meta map[string]interface{}, entry UploadEntry)) {
	// Implementation
}
