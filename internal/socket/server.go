package socket

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/fu2hito/go-liveview/internal/protocol"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Server manages WebSocket connections
type Server struct {
	connections map[string]*Conn
	mu          sync.RWMutex
	handlers    map[string]HandlerFunc
}

// HandlerFunc is a function that handles LiveView connections
type HandlerFunc func(ctx context.Context, conn *Conn, msg *protocol.Message)

// NewServer creates a new WebSocket server
func NewServer() *Server {
	return &Server{
		connections: make(map[string]*Conn),
		handlers:    make(map[string]HandlerFunc),
	}
}

// RegisterHandler registers a handler for a specific topic/event
func (s *Server) RegisterHandler(topic string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[topic] = handler
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	c := s.newConnection(conn)
	go c.readLoop()
	go c.writeLoop()
}

// Conn represents a WebSocket connection
type Conn struct {
	server *Server
	ws     *websocket.Conn
	send   chan *protocol.Message
	id     string
	topics map[string]bool
	mu     sync.RWMutex
}

func (s *Server) newConnection(ws *websocket.Conn) *Conn {
	c := &Conn{
		server: s,
		ws:     ws,
		send:   make(chan *protocol.Message, 256),
		id:     generateID(),
		topics: make(map[string]bool),
	}
	s.mu.Lock()
	s.connections[c.id] = c
	s.mu.Unlock()
	return c
}

func generateID() string {
	return time.Now().Format("20060102150405") + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func (c *Conn) readLoop() {
	defer func() {
		c.server.mu.Lock()
		delete(c.server.connections, c.id)
		c.server.mu.Unlock()
		c.ws.Close()
	}()

	c.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		msg, err := protocol.DecodeMessage(data)
		if err != nil {
			log.Printf("Failed to decode message: %v", err)
			continue
		}

		// Handle heartbeat
		if msg.Event == "heartbeat" {
			c.send <- &protocol.Message{
				Ref:     msg.Ref,
				Topic:   "phoenix",
				Event:   "phx_reply",
				Payload: []byte(`{"status":"ok","response":{}}`),
			}
			continue
		}

		// Dispatch to handler
		c.server.mu.RLock()
		handler, ok := c.server.handlers[msg.Topic]
		c.server.mu.RUnlock()

		if ok {
			go handler(context.Background(), c, msg)
		}
	}
}

func (c *Conn) writeLoop() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			data, err := msg.Encode()
			if err != nil {
				log.Printf("Failed to encode message: %v", err)
				continue
			}
			c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.ws.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Send sends a message to the client
func (c *Conn) Send(msg *protocol.Message) {
	select {
	case c.send <- msg:
	default:
		// Channel full, drop message
	}
}

// ID returns the connection ID
func (c *Conn) ID() string {
	return c.id
}
