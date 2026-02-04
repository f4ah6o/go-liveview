package form

import (
	"context"
	"io"
	"net/url"

	"github.com/a-h/templ"
	"github.com/fu2hito/go-liveview"
)

// User represents a user in the form
type User struct {
	Name  string
	Email string
	Age   int
}

// Form is a form validation LiveView
type Form struct {
	user   User
	errors map[string]string
}

// New creates a new Form LiveView
func New() liveview.LiveView {
	return &Form{
		errors: make(map[string]string),
	}
}

// Mount is called when the LiveView first mounts
func (f *Form) Mount(ctx *liveview.Context, params url.Values) error {
	f.user = User{}
	f.errors = make(map[string]string)
	ctx.Assign("user", f.user)
	ctx.Assign("errors", f.errors)
	ctx.Assign("submitted", false)
	return nil
}

// HandleEvent handles events sent from the client
func (f *Form) HandleEvent(ctx *liveview.Context, event string, payload map[string]interface{}) error {
	switch event {
	case "validate":
		f.validate(payload)
	case "save":
		if f.validate(payload) {
			// Save the user (in production, this would save to DB)
			ctx.Assign("submitted", true)
		}
	case "reset":
		f.user = User{}
		f.errors = make(map[string]string)
		ctx.Assign("user", f.user)
		ctx.Assign("errors", f.errors)
		ctx.Assign("submitted", false)
	}
	return nil
}

// HandleParams handles URL parameter changes
func (f *Form) HandleParams(ctx *liveview.Context, params url.Values) error {
	return nil
}

// Render returns the component to render
func (f *Form) Render(ctx *liveview.Context) templ.Component {
	user, _ := ctx.Get("user")
	errors, _ := ctx.Get("errors")
	submitted, _ := ctx.Get("submitted")
	return formTemplate(user.(User), errors.(map[string]string), submitted.(bool))
}

func (f *Form) validate(payload map[string]interface{}) bool {
	f.errors = make(map[string]string)

	if name, ok := payload["name"].(string); ok {
		f.user.Name = name
		if name == "" {
			f.errors["name"] = "Name is required"
		}
	}

	if email, ok := payload["email"].(string); ok {
		f.user.Email = email
		if email == "" {
			f.errors["email"] = "Email is required"
		}
	}

	return len(f.errors) == 0
}

func formTemplate(user User, errors map[string]string, submitted bool) templ.Component {
	html := `
		<div>
			<h1>User Form</h1>
			` + renderForm(user, errors, submitted) + `
		</div>
	`
	return &simpleComponent{html: html}
}

func renderForm(user User, errors map[string]string, submitted bool) string {
	if submitted {
		return `<p>Form submitted successfully!</p>
			<button phx-click="reset">Reset</button>`
	}

	html := `<form phx-change="validate" phx-submit="save">`

	// Name field
	html += `<div>
		<label>Name:</label>
		<input type="text" name="name" value="` + user.Name + `" />
		` + renderError(errors["name"]) + `
	</div>`

	// Email field
	html += `<div>
		<label>Email:</label>
		<input type="email" name="email" value="` + user.Email + `" />
		` + renderError(errors["email"]) + `
	</div>`

	html += `<button type="submit">Save</button>
		<button type="button" phx-click="reset">Reset</button>
	</form>`

	return html
}

func renderError(msg string) string {
	if msg == "" {
		return ""
	}
	return `<span style="color: red;">` + msg + `</span>`
}

type simpleComponent struct {
	html string
}

func (s *simpleComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(s.html))
	return err
}
