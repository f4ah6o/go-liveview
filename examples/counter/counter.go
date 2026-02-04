package counter

import (
	"context"
	"io"
	"net/url"
	"strconv"

	"github.com/a-h/templ"
	"github.com/fu2hito/go-liveview"
)

// Counter is a simple counter LiveView
type Counter struct {
	count int
}

// New creates a new Counter LiveView
func New() liveview.LiveView {
	return &Counter{}
}

// Mount is called when the LiveView first mounts
func (c *Counter) Mount(ctx *liveview.Context, params url.Values) error {
	c.count = 0
	ctx.Assign("count", c.count)
	return nil
}

// HandleEvent handles events sent from the client
func (c *Counter) HandleEvent(ctx *liveview.Context, event string, payload map[string]interface{}) error {
	switch event {
	case "inc":
		c.count++
	case "dec":
		c.count--
	case "set":
		if val, ok := payload["value"].(string); ok {
			if n, err := strconv.Atoi(val); err == nil {
				c.count = n
			}
		}
	}
	ctx.Assign("count", c.count)
	return nil
}

// HandleParams handles URL parameter changes
func (c *Counter) HandleParams(ctx *liveview.Context, params url.Values) error {
	return nil
}

// Render returns the component to render
func (c *Counter) Render(ctx *liveview.Context) templ.Component {
	count, _ := ctx.Get("count")
	return counterTemplate(count.(int))
}

// counterTemplate renders the counter HTML
func counterTemplate(count int) templ.Component {
	// This would be a templ component in production
	// For now, returning a simple component
	return &simpleComponent{html: `
		<div>
			<h1>Count: <!--$0-->` + strconv.Itoa(count) + `<!--/$0--></h1>
			<button phx-click="dec">-</button>
			<button phx-click="inc">+</button>
			<form phx-submit="set">
				<input type="number" name="value" value="` + strconv.Itoa(count) + `" />
				<button type="submit">Set</button>
			</form>
		</div>
	`}
}

type simpleComponent struct {
	html string
}

func (s *simpleComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(s.html))
	return err
}
