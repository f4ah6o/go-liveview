package liveview

// Options configures the LiveView server
type Options struct {
	// Secret is the secret key for signing sessions
	Secret string

	// ReconnectStrategy determines how to handle reconnections
	ReconnectStrategy ReconnectStrategy

	// PubSub is the PubSub adapter for distributed LiveViews
	PubSub PubSub
}

// ReconnectStrategy determines how to handle reconnections
type ReconnectStrategy int

const (
	// ReconnectReset resets the LiveView state on reconnection
	ReconnectReset ReconnectStrategy = iota
	// ReconnectRestore restores the LiveView state on reconnection
	ReconnectRestore
)
