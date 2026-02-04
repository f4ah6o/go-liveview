package liveview

import (
	"net/http"
	"strings"

	"github.com/fu2hito/go-liveview/internal/socket"
)

// Handler is an HTTP handler for LiveViews
type Handler struct {
	manager  *Manager
	server   *socket.Server
	template string
}

// HandlerOptions configures the handler
type HandlerOptions struct {
	// Template is the HTML template for the initial page load
	Template string
}

// NewHandler creates a new LiveView HTTP handler
func NewHandler(manager *Manager, server *socket.Server, opts HandlerOptions) *Handler {
	return &Handler{
		manager:  manager,
		server:   server,
		template: opts.Template,
	}
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if this is a WebSocket upgrade request
	if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
		h.server.ServeHTTP(w, r)
		return
	}

	// Serve the initial HTML page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(h.template))
}

// DefaultTemplate returns a default HTML template with LiveView client
func DefaultTemplate(title, csrfToken string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + title + `</title>
    <meta name="csrf-token" content="` + csrfToken + `">
</head>
<body>
    <div id="live-view-root"></div>
    <script src="/liveview.js"></script>
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            const liveSocket = new LiveSocket('/live', {
                params: { _csrf_token: document.querySelector('meta[name="csrf-token"]').content }
            });
            liveSocket.connect();
        });
    </script>
</body>
</html>`
}
