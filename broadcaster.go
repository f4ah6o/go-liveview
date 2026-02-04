package liveview

import (
	"context"
	"sync"
)

// Broadcaster handles broadcasting messages to multiple LiveView instances
type Broadcaster struct {
	pubsub      PubSub
	subscribers map[string]map[string]func(msg BroadcastMessage)
	mu          sync.RWMutex
}

// NewBroadcaster creates a new broadcaster
func NewBroadcaster(pubsub PubSub) *Broadcaster {
	return &Broadcaster{
		pubsub:      pubsub,
		subscribers: make(map[string]map[string]func(msg BroadcastMessage)),
	}
}

// Subscribe adds a subscriber to a topic
func (b *Broadcaster) Subscribe(topic, subscriberID string, handler func(msg BroadcastMessage)) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscribers[topic]; !ok {
		b.subscribers[topic] = make(map[string]func(msg BroadcastMessage))

		// Subscribe to PubSub topic
		b.pubsub.Subscribe(topic, func(msg interface{}) {
			if broadcastMsg, ok := msg.(BroadcastMessage); ok {
				b.broadcastToSubscribers(topic, broadcastMsg)
			}
		})
	}

	b.subscribers[topic][subscriberID] = handler
	return nil
}

// Unsubscribe removes a subscriber from a topic
func (b *Broadcaster) Unsubscribe(topic, subscriberID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if topicSubs, ok := b.subscribers[topic]; ok {
		delete(topicSubs, subscriberID)

		if len(topicSubs) == 0 {
			delete(b.subscribers, topic)
			b.pubsub.Unsubscribe(topic)
		}
	}

	return nil
}

// Broadcast sends a message to all subscribers of a topic
func (b *Broadcaster) Broadcast(topic, event string, payload interface{}) error {
	msg := BroadcastMessage{
		Topic:   topic,
		Event:   event,
		Payload: payload,
	}

	return b.pubsub.Publish(topic, msg)
}

func (b *Broadcaster) broadcastToSubscribers(topic string, msg BroadcastMessage) {
	b.mu.RLock()
	topicSubs, ok := b.subscribers[topic]
	if !ok {
		b.mu.RUnlock()
		return
	}

	// Copy handlers to avoid holding lock during execution
	handlers := make([]func(msg BroadcastMessage), 0, len(topicSubs))
	for _, handler := range topicSubs {
		handlers = append(handlers, handler)
	}
	b.mu.RUnlock()

	// Call handlers
	for _, handler := range handlers {
		handler(msg)
	}
}

// SubscribeContext adds a subscriber that automatically unsubscribes when context is done
func (b *Broadcaster) SubscribeContext(ctx context.Context, topic, subscriberID string, handler func(msg BroadcastMessage)) error {
	err := b.Subscribe(topic, subscriberID, handler)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		b.Unsubscribe(topic, subscriberID)
	}()

	return nil
}
