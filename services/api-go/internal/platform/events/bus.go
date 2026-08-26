package events

import (
	"context"
)

type Event struct {
	Type    string
	Payload any
}

type Handler func(ctx context.Context, e Event) error

type Bus interface {
	Publish(ctx context.Context, e Event) error
	Subscribe(eventType string, handler Handler)
}

// MemoryBus provides a simple in-memory implementation of Bus
type MemoryBus struct {
	handlers map[string][]Handler
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		handlers: make(map[string][]Handler),
	}
}

func (b *MemoryBus) Publish(ctx context.Context, e Event) error {
	if handlers, ok := b.handlers[e.Type]; ok {
		for _, h := range handlers {
			if err := h(ctx, e); err != nil {
				// In a real system, we'd log this or handle retries
				return err
			}
		}
	}
	return nil
}

func (b *MemoryBus) Subscribe(eventType string, handler Handler) {
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}
