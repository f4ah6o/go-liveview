package liveview

import (
	"sync"
)

// PubSub is the interface for PubSub adapters
type PubSub interface {
	// Subscribe subscribes to a topic and calls handler when messages are received
	Subscribe(topic string, handler func(msg interface{})) error

	// Unsubscribe unsubscribes from a topic
	Unsubscribe(topic string) error

	// Publish publishes a message to a topic
	Publish(topic string, msg interface{}) error
}

// LocalPubSub is an in-memory PubSub implementation for single-node deployments
type LocalPubSub struct {
	topics map[string][]func(interface{})
	mu     sync.RWMutex
}

// NewLocalPubSub creates a new local PubSub instance
func NewLocalPubSub() *LocalPubSub {
	return &LocalPubSub{
		topics: make(map[string][]func(interface{})),
	}
}

// Subscribe implements PubSub.Subscribe
func (p *LocalPubSub) Subscribe(topic string, handler func(msg interface{})) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.topics[topic] = append(p.topics[topic], handler)
	return nil
}

// Unsubscribe implements PubSub.Unsubscribe
func (p *LocalPubSub) Unsubscribe(topic string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.topics, topic)
	return nil
}

// Publish implements PubSub.Publish
func (p *LocalPubSub) Publish(topic string, msg interface{}) error {
	p.mu.RLock()
	handlers := p.topics[topic]
	p.mu.RUnlock()

	// Call handlers asynchronously
	for _, handler := range handlers {
		go handler(msg)
	}

	return nil
}

// BroadcastMessage represents a message broadcast to all subscribers
type BroadcastMessage struct {
	Topic   string
	Event   string
	Payload interface{}
}
