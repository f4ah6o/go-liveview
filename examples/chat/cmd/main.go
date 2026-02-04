package main

import (
	"log"
	"net/http"

	"github.com/fu2hito/go-liveview"
	"github.com/fu2hito/go-liveview/examples/chat"
	"github.com/fu2hito/go-liveview/internal/socket"
)

func main() {
	// Create WebSocket server
	wsServer := socket.NewServer()

	// Create PubSub (local for single-node, use RedisPubSub for distributed)
	pubsub := liveview.NewLocalPubSub()

	// Create broadcaster
	broadcaster := liveview.NewBroadcaster(pubsub)

	// Create LiveView manager with broadcaster
	manager := liveview.NewManager(wsServer)
	manager.SetBroadcaster(broadcaster)

	// Register LiveViews
	manager.Register("chat", func() liveview.LiveView {
		return chat.New()
	})

	// Create HTTP handler
	handler := liveview.NewHandler(manager, wsServer, liveview.HandlerOptions{
		Template: liveview.DefaultTemplate("Chat Example", "csrf-token-here"),
	})

	// Serve static files
	fs := http.FileServer(http.Dir("./js/dist"))
	http.Handle("/liveview.js", fs)
	http.Handle("/", handler)

	log.Println("Chat server starting on :8080")
	log.Println("Open http://localhost:8080 in multiple browsers to test real-time chat")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
