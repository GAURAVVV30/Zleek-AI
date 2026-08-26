package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisBus struct {
	client   *redis.Client
	handlers map[string][]Handler
	mu       sync.RWMutex
}

func NewRedisBus(client *redis.Client) *RedisBus {
	return &RedisBus{
		client:   client,
		handlers: make(map[string][]Handler),
	}
}

func (b *RedisBus) Publish(ctx context.Context, e Event) error {
	payloadBytes, err := json.Marshal(e.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	wrapper := struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}{
		Type:    e.Type,
		Payload: payloadBytes,
	}

	data, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to Redis channel based on event type
	return b.client.Publish(ctx, "domain_events:"+e.Type, data).Err()
}

func (b *RedisBus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
	b.mu.Unlock()
}

func (b *RedisBus) StartListening(ctx context.Context, eventTypes []string) error {
	channels := make([]string, len(eventTypes))
	for i, et := range eventTypes {
		channels[i] = "domain_events:" + et
	}

	pubsub := b.client.Subscribe(ctx, channels...)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-ch:
			var wrapper struct {
				Type    string          `json:"type"`
				Payload json.RawMessage `json:"payload"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &wrapper); err != nil {
				continue
			}

			event := Event{
				Type:    wrapper.Type,
				Payload: wrapper.Payload, // As raw message
			}

			b.mu.RLock()
			handlers := b.handlers[wrapper.Type]
			b.mu.RUnlock()

			// In a real system, we'd use a worker pool here
			for _, h := range handlers {
				// We create a new context with a timeout for the handler
				hCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_ = h(hCtx, event)
				cancel()
			}
		}
	}
}
