package liveview

import (
	"context"
	"net/url"

	"github.com/a-h/templ"
)

// LiveView is the interface that all LiveViews must implement
type LiveView interface {
	// Mount is called when the LiveView first mounts
	Mount(ctx *Context, params url.Values) error

	// HandleEvent handles events sent from the client
	HandleEvent(ctx *Context, event string, payload map[string]interface{}) error

	// HandleParams handles URL parameter changes
	HandleParams(ctx *Context, params url.Values) error

	// Render returns the component to render
	Render(ctx *Context) templ.Component
}

// Context provides access to LiveView context
type Context struct {
	context.Context
	Socket      Socket
	ID          string
	Assigns     map[string]interface{}
	Changed     map[string]bool
	broadcaster *Broadcaster
}

// Socket provides socket operations
type Socket interface {
	// PushEvent sends an event to the client
	PushEvent(event string, payload map[string]interface{})

	// PutFlash sets a flash message
	PutFlash(kind, message string)

	// AllowUpload configures file uploads
	AllowUpload(name string, options UploadConfig)

	// CancelUpload cancels an upload
	CancelUpload(name string)

	// ConsumeUploadedEntries processes uploaded files
	ConsumeUploadedEntries(entries []UploadEntry, fn func(meta map[string]interface{}, entry UploadEntry))
}

// UploadConfig configures file uploads
type UploadConfig struct {
	Accept      []string
	MaxEntries  int
	MaxFileSize int64
}

// UploadEntry represents an uploaded file
type UploadEntry struct {
	Name        string
	ClientName  string
	ClientType  string
	ClientSize  int64
	UUID        string
	Progress    int
	Preflighted bool
	Done        bool
	Cancelled   bool
	Valid       bool
}

// Assign assigns a value to the context
func (c *Context) Assign(key string, value interface{}) {
	c.Assigns[key] = value
	c.Changed[key] = true
}

// Get retrieves a value from the context
func (c *Context) Get(key string) (interface{}, bool) {
	val, ok := c.Assigns[key]
	return val, ok
}

// SetBroadcaster sets the broadcaster for this context
func (c *Context) SetBroadcaster(b *Broadcaster) {
	c.broadcaster = b
}

// GetBroadcaster returns the broadcaster
func (c *Context) GetBroadcaster() *Broadcaster {
	return c.broadcaster
}

// NewContext creates a new LiveView context
func NewContext(ctx context.Context, socket Socket, id string) *Context {
	return &Context{
		Context: ctx,
		Socket:  socket,
		ID:      id,
		Assigns: make(map[string]interface{}),
		Changed: make(map[string]bool),
	}
}

// Rendered represents the rendered state interface
type Rendered interface {
	GetStatic() []string
	GetDynamic() []interface{}
}

// BaseRendered is a base implementation of Rendered
type BaseRendered struct {
	Static  []string      `json:"s"`
	Dynamic []interface{} `json:"d"`
}

func (r *BaseRendered) GetStatic() []string {
	return r.Static
}

func (r *BaseRendered) GetDynamic() []interface{} {
	return r.Dynamic
}

// RenderedValue returns the current rendered state
func (c *Context) RenderedValue() Rendered {
	if r, ok := c.Assigns["__rendered__"].(Rendered); ok {
		return r
	}
	return nil
}

// SetRenderedValue sets the rendered state
func (c *Context) SetRenderedValue(r Rendered) {
	c.Assigns["__rendered__"] = r
}
