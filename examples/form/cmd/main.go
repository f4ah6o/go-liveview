package main

import (
	"log"
	"net/http"

	"github.com/fu2hito/go-liveview"
	"github.com/fu2hito/go-liveview/examples/form"
	"github.com/fu2hito/go-liveview/internal/socket"
)

func main() {
	// Create WebSocket server
	wsServer := socket.NewServer()

	// Create LiveView manager
	manager := liveview.NewManager(wsServer)

	// Register LiveViews
	manager.Register("form", func() liveview.LiveView {
		return form.New()
	})

	// Create HTTP handler
	handler := liveview.NewHandler(manager, wsServer, liveview.HandlerOptions{
		Template: liveview.DefaultTemplate("Form Example", "csrf-token-here"),
	})

	// Serve static files
	fs := http.FileServer(http.Dir("./js/dist"))
	http.Handle("/liveview.js", fs)
	http.Handle("/", handler)

	log.Println("Form server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
