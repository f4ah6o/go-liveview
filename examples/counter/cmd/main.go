package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fu2hito/go-liveview"
	"github.com/fu2hito/go-liveview/examples/counter"
	"github.com/fu2hito/go-liveview/internal/socket"
)

func main() {
	// Create WebSocket server
	wsServer := socket.NewServer()

	// Create LiveView manager
	manager := liveview.NewManager(wsServer)

	// Register LiveViews
	manager.Register("counter", func() liveview.LiveView {
		return counter.New()
	})

	// Create HTTP handler
	handler := liveview.NewHandler(manager, wsServer, liveview.HandlerOptions{
		Template: defaultTemplate(),
	})

	// Get the absolute path to the JS file
	// When running with `go run`, we need to find the file relative to the source
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	// Go up from examples/counter/cmd to repo root
	repoRoot := filepath.Join(dir, "..", "..", "..")
	jsPath := filepath.Join(repoRoot, "js", "dist", "liveview.global.js")

	// Debug: print the resolved path
	fmt.Printf("JS file path: %s\n", jsPath)

	// Serve static files (for the JS client)
	// We serve the IIFE build (liveview.global.js) as /liveview.js for browser compatibility
	http.HandleFunc("/liveview.js", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(jsPath); os.IsNotExist(err) {
			log.Printf("ERROR: JavaScript file not found at %s", jsPath)
			http.Error(w, "JavaScript file not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, jsPath)
	})
	http.Handle("/", handler)

	// Try to find an available port
	startPort := 8080
	var listener net.Listener
	var err error
	var port int

	for i := 0; i < 100; i++ {
		port = startPort + i
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Fatal("Could not find an available port")
	}

	fmt.Println("=====================================")
	fmt.Println("  Go LiveView Server")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Printf("Server starting on http://localhost:%d\n", port)
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  - Counter LiveView with +/- buttons")
	fmt.Println("  - Real-time WebSocket communication")
	fmt.Println("  - DOM diff patching")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	log.Fatal(http.Serve(listener, nil))
}

func defaultTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go LiveView - Counter</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            margin-top: 0;
        }
        .counter {
            font-size: 48px;
            font-weight: bold;
            color: #6366f1;
            margin: 20px 0;
            text-align: center;
        }
        .buttons {
            display: flex;
            gap: 10px;
            justify-content: center;
        }
        button {
            padding: 12px 24px;
            font-size: 18px;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            transition: background 0.2s;
        }
        button:hover {
            opacity: 0.9;
        }
        .dec {
            background: #ef4444;
            color: white;
        }
        .inc {
            background: #22c55e;
            color: white;
        }
        .status {
            margin-top: 20px;
            padding: 10px;
            border-radius: 4px;
            text-align: center;
        }
        .status.connected {
            background: #dcfce7;
            color: #166534;
        }
        .status.disconnected {
            background: #fee2e2;
            color: #991b1b;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔥 Go LiveView Counter</h1>
        <div id="live-view-root">
            <div class="counter">0</div>
            <div class="buttons">
                <button class="dec" phx-click="dec">-</button>
                <button class="inc" phx-click="inc">+</button>
            </div>
        </div>
        <div id="status" class="status disconnected">Disconnected</div>
    </div>

    <script src="/liveview.js" onload="console.log('liveview.js loaded successfully')" onerror="console.error('Failed to load liveview.js')"></script>
    <script>
        console.log('Inline script starting...');
        console.log('LiveSocket exists:', typeof LiveSocket !== 'undefined');
        console.log('LiveViewRenderer exists:', typeof LiveViewRenderer !== 'undefined');
        
        // Debug helper
        function logError(msg) {
            const status = document.getElementById('status');
            status.textContent = 'Error: ' + msg;
            status.style.backgroundColor = '#fee2e2';
            status.style.color = '#991b1b';
            console.error(msg);
        }

        try {
            if (typeof LiveSocket === 'undefined') {
                throw new Error("LiveSocket is not defined. Check if /liveview.js is loading correctly.");
            }

            // Initialize the socket
            const liveSocket = new LiveSocket("/");
            liveSocket.connect();
            console.log("LiveSocket connected initiated");

            // Connect to the counter channel
            const channel = liveSocket.channel("counter", {});
            
            // Initialize the renderer
            const renderer = new LiveViewRenderer("live-view-root");
            
            // Setup event delegation
            renderer.setupEventDelegation((event, payload) => {
                channel.push(event, payload);
            });

            // Handle channel events
            channel.on("join", (payload) => {
                console.log("Joined:", payload);
                updateStatus(true);
                if (payload && payload.rendered) {
                    renderer.render(payload.rendered);
                }
            });
            
            channel.on("diff", (diff) => {
                console.log("Diff:", diff);
                renderer.render(diff);
            });
            
            channel.on("error", (payload) => {
                 console.error("Channel error:", payload);
                 logError("Channel error");
            });

            channel.join();
        } catch (e) {
            logError(e.message);
        }

        // Status UI helpers
        function updateStatus(connected) {
            const status = document.getElementById('status');
            if (connected) {
                status.textContent = 'Connected';
                status.className = 'status connected';
            } else {
                status.textContent = 'Disconnected';
                status.className = 'status disconnected';
            }
        }
    </script>
</body>
</html>`
}
