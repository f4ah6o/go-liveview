package chat

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/a-h/templ"
	"github.com/fu2hito/go-liveview"
)

// Message represents a chat message
type Message struct {
	ID        string
	User      string
	Text      string
	Timestamp time.Time
}

// Chat is a real-time chat LiveView
type Chat struct {
	messages []Message
	username string
}

// New creates a new Chat LiveView
func New() liveview.LiveView {
	return &Chat{
		messages: make([]Message, 0),
	}
}

// Mount is called when the LiveView first mounts
func (c *Chat) Mount(ctx *liveview.Context, params url.Values) error {
	c.messages = make([]Message, 0)
	c.username = params.Get("username")
	if c.username == "" {
		c.username = "Anonymous"
	}

	ctx.Assign("messages", c.messages)
	ctx.Assign("username", c.username)
	ctx.Assign("message", "")

	// Subscribe to broadcast messages
	if broadcaster := ctx.GetBroadcaster(); broadcaster != nil {
		broadcaster.SubscribeContext(ctx, "chat:room", ctx.ID, func(msg liveview.BroadcastMessage) {
			if newMsg, ok := msg.Payload.(Message); ok {
				c.messages = append(c.messages, newMsg)
				ctx.Assign("messages", c.messages)
				// Trigger re-render
				ctx.Socket.PushEvent("phx:update", map[string]interface{}{})
			}
		})
	}

	return nil
}

// HandleEvent handles events sent from the client
func (c *Chat) HandleEvent(ctx *liveview.Context, event string, payload map[string]interface{}) error {
	switch event {
	case "send_message":
		if text, ok := payload["message"].(string); ok && text != "" {
			msg := Message{
				ID:        generateID(),
				User:      c.username,
				Text:      text,
				Timestamp: time.Now(),
			}

			// Broadcast to all connected clients
			if broadcaster := ctx.GetBroadcaster(); broadcaster != nil {
				broadcaster.Broadcast("chat:room", "new_message", msg)
			} else {
				// Fallback: only update current client
				c.messages = append(c.messages, msg)
				ctx.Assign("messages", c.messages)
			}

			ctx.Assign("message", "")
		}
	case "set_username":
		if username, ok := payload["username"].(string); ok {
			c.username = username
			ctx.Assign("username", c.username)
		}
	}
	return nil
}

// HandleParams handles URL parameter changes
func (c *Chat) HandleParams(ctx *liveview.Context, params url.Values) error {
	return nil
}

// Render returns the component to render
func (c *Chat) Render(ctx *liveview.Context) templ.Component {
	messages, _ := ctx.Get("messages")
	username, _ := ctx.Get("username")
	return chatTemplate(messages.([]Message), username.(string))
}

func chatTemplate(messages []Message, username string) templ.Component {
	html := `<div class="chat-container">
		<h1>Chat Room</h1>
		<div class="username">
			<input type="text" name="username" value="` + username + `" phx-change="set_username" />
		</div>
		<div class="messages">
			` + renderMessages(messages) + `
		</div>
		<form phx-submit="send_message">
			<input type="text" name="message" placeholder="Type a message..." />
			<button type="submit">Send</button>
		</form>
	</div>`
	return &simpleComponent{html: html}
}

func renderMessages(messages []Message) string {
	if len(messages) == 0 {
		return `<p class="empty">No messages yet. Be the first to say something!</p>`
	}

	html := ""
	for _, msg := range messages {
		html += `<div class="message">
			<span class="user">` + msg.User + `:</span>
			<span class="text">` + msg.Text + `</span>
			<span class="time">` + msg.Timestamp.Format("15:04") + `</span>
		</div>`
	}
	return html
}

func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

type simpleComponent struct {
	html string
}

func (s *simpleComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(s.html))
	return err
}
