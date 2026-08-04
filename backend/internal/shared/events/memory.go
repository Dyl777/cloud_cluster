package events

import (
	"context"
	"sync"
)

type MemoryBus struct {
	mu   sync.Mutex
	subs map[Topic]map[uint64]Handler
	seq  uint64
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{subs: make(map[Topic]map[uint64]Handler)}
}

func (b *MemoryBus) nextID() string {
	b.seq++
	return "evt-" + string(rune('a'+b.seq%26)) + "000" + uintToString(b.seq)
}

func uintToString(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func (b *MemoryBus) Publish(ctx context.Context, topic Topic, payload any) error {
	b.mu.Lock()
	handlers := make([]Handler, 0, len(b.subs[topic]))
	for _, h := range b.subs[topic] {
		handlers = append(handlers, h)
	}
	evt := Event{ID: b.nextID(), Topic: topic, Version: "1.0", Payload: payload}
	b.mu.Unlock()
	for _, h := range handlers {
		go h(ctx, evt)
	}
	return nil
}

func (b *MemoryBus) Subscribe(topic Topic, handler Handler) func() {
	b.mu.Lock()
	if b.subs[topic] == nil {
		b.subs[topic] = make(map[uint64]Handler)
	}
	b.seq++
	id := b.seq
	b.subs[topic][id] = handler
	b.mu.Unlock()
	return func() {
		b.mu.Lock()
		delete(b.subs[topic], id)
		b.mu.Unlock()
	}
}

func (b *MemoryBus) Close() error { return nil }
