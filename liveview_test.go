package liveview_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/fu2hito/go-liveview"
	"github.com/fu2hito/go-liveview/internal/socket"
	"github.com/gorilla/websocket"
)

// TestLiveView is a test LiveView implementation
type TestLiveView struct {
	count int
}

func (t *TestLiveView) Mount(ctx *liveview.Context, params url.Values) error {
	t.count = 0
	ctx.Assign("count", t.count)
	return nil
}

func (t *TestLiveView) HandleEvent(ctx *liveview.Context, event string, payload map[string]interface{}) error {
	switch event {
	case "inc":
		t.count++
	case "dec":
		t.count--
	}
	ctx.Assign("count", t.count)
	return nil
}

func (t *TestLiveView) HandleParams(ctx *liveview.Context, params url.Values) error {
	return nil
}

func (t *TestLiveView) Render(ctx *liveview.Context) templ.Component {
	return &testComponent{count: t.count}
}

type testComponent struct {
	count int
}

func (t *testComponent) Render(ctx context.Context, w io.Writer) error {
	_, err := w.Write([]byte(`<div>Count: <!--$0-->` + strconv.Itoa(t.count) + `<!--/$0--></div>`))
	return err
}

func TestLiveViewIntegration(t *testing.T) {
	// Create WebSocket server
	wsServer := socket.NewServer()

	// Create LiveView manager
	manager := liveview.NewManager(wsServer)

	// Register test LiveView
	manager.Register("test", func() liveview.LiveView {
		return &TestLiveView{}
	})

	// Create HTTP handler
	handler := liveview.NewHandler(manager, wsServer, liveview.HandlerOptions{
		Template: `<!DOCTYPE html><html><body><div id="root"></div></body></html>`,
	})

	// Test HTTP endpoint
	t.Run("HTTP Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "root") {
			t.Error("Response should contain root element")
		}
	})

	// Test WebSocket upgrade
	t.Run("WebSocket Upgrade", func(t *testing.T) {
		server := httptest.NewServer(handler)
		defer server.Close()

		// Convert http:// to ws://
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/"

		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer ws.Close()

		// Send join message
		joinMsg := map[string]interface{}{
			"join_ref": "1",
			"ref":      "1",
			"topic":    "test",
			"event":    "phx_join",
			"payload": map[string]interface{}{
				"params":  map[string]interface{}{},
				"session": "",
				"static":  "",
			},
		}

		err = ws.WriteJSON(joinMsg)
		if err != nil {
			t.Fatalf("Failed to send join message: %v", err)
		}

		// Read response with timeout
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))

		var response map[string]interface{}
		err = ws.ReadJSON(&response)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		if response["event"] != "phx_reply" {
			t.Errorf("Expected phx_reply, got %v", response["event"])
		}

		// Send event
		eventMsg := map[string]interface{}{
			"topic": "test",
			"event": "event",
			"payload": map[string]interface{}{
				"type":  "click",
				"event": "inc",
				"value": map[string]interface{}{},
			},
		}

		err = ws.WriteJSON(eventMsg)
		if err != nil {
			t.Fatalf("Failed to send event: %v", err)
		}

		// Read diff response
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		err = ws.ReadJSON(&response)
		if err != nil {
			t.Fatalf("Failed to read diff response: %v", err)
		}

		if response["event"] != "diff" {
			t.Errorf("Expected diff, got %v", response["event"])
		}
	})
}

func TestPubSub(t *testing.T) {
	pubsub := liveview.NewLocalPubSub()

	received := make(chan interface{}, 1)

	// Subscribe
	err := pubsub.Subscribe("test:topic", func(msg interface{}) {
		received <- msg
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish
	testMsg := "test message"
	err = pubsub.Publish("test:topic", testMsg)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Wait for message
	select {
	case msg := <-received:
		if msg != testMsg {
			t.Errorf("Expected %v, got %v", testMsg, msg)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestBroadcaster(t *testing.T) {
	pubsub := liveview.NewLocalPubSub()
	broadcaster := liveview.NewBroadcaster(pubsub)

	received := make(chan liveview.BroadcastMessage, 1)

	// Subscribe
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := broadcaster.SubscribeContext(ctx, "test:topic", "sub-1", func(msg liveview.BroadcastMessage) {
		received <- msg
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Broadcast
	testPayload := map[string]string{"key": "value"}
	err = broadcaster.Broadcast("test:topic", "test_event", testPayload)
	if err != nil {
		t.Fatalf("Failed to broadcast: %v", err)
	}

	// Wait for message
	select {
	case msg := <-received:
		if msg.Topic != "test:topic" {
			t.Errorf("Expected topic test:topic, got %v", msg.Topic)
		}
		if msg.Event != "test_event" {
			t.Errorf("Expected event test_event, got %v", msg.Event)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for broadcast")
	}
}
