package eventbus
package eventbus

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHandler is a test EventHandler that records calls
type mockHandler struct {
	name       string
	eventTypes []string
	handleFunc func(ctx context.Context, event Event) error
	callCount  int32
}

func newMockHandler(name string, eventTypes []string, fn func(ctx context.Context, event Event) error) *mockHandler {
	if fn == nil {
		fn = func(ctx context.Context, event Event) error { return nil }
	}
	return &mockHandler{name: name, eventTypes: eventTypes, handleFunc: fn}
}

func (h *mockHandler) Handle(ctx context.Context, event Event) error {
	atomic.AddInt32(&h.callCount, 1)
	return h.handleFunc(ctx, event)
}

func (h *mockHandler) GetEventTypes() []string { return h.eventTypes }
func (h *mockHandler) GetName() string         { return h.name }
func (h *mockHandler) Calls() int              { return int(atomic.LoadInt32(&h.callCount)) }

func TestMemoryEventBus_Subscribe(t *testing.T) {
	bus := NewMemoryEventBus()
	defer bus.Shutdown(context.Background())

	t.Run("subscribe handler successfully", func(t *testing.T) {
		h := newMockHandler("h1", []string{"test.event"}, nil)
		err := bus.Subscribe(h)
		require.NoError(t, err)
		subs := bus.GetSubscribers("test.event")
		assert.Len(t, subs, 1)
	})

	t.Run("nil handler returns error", func(t *testing.T) {
		err := bus.Subscribe(nil)
		require.Error(t, err)
	})

	t.Run("handler with no event types returns error", func(t *testing.T) {
		h := newMockHandler("empty", []string{}, nil)
		err := bus.Subscribe(h)
		require.Error(t, err)
	})

	t.Run("duplicate subscription returns error", func(t *testing.T) {
		h := newMockHandler("dup", []string{"dup.event"}, nil)
		require.NoError(t, bus.Subscribe(h))
		err := bus.Subscribe(h)
		require.Error(t, err)
	})
}

func TestMemoryEventBus_Publish(t *testing.T) {
	bus := NewMemoryEventBus()
	defer bus.Shutdown(context.Background())

	t.Run("handler called on publish", func(t *testing.T) {
		h := newMockHandler("receiver", []string{"order.created"}, nil)
		require.NoError(t, bus.Subscribe(h))
		event := NewBaseEvent("order.created", "evt-1", map[string]string{"id": "42"})
		err := bus.Publish(context.Background(), event)
		require.NoError(t, err)
		assert.Equal(t, 1, h.Calls())
	})

	t.Run("no handlers returns nil", func(t *testing.T) {
		event := NewBaseEvent("unknown.event", "evt-2", nil)
		err := bus.Publish(context.Background(), event)
		require.NoError(t, err)
	})

	t.Run("handler error is returned", func(t *testing.T) {
		errH := newMockHandler("errhandler", []string{"fail.event"}, func(ctx context.Context, event Event) error {
			return errors.New("handler failed")
		})
		require.NoError(t, bus.Subscribe(errH))
		event := NewBaseEvent("fail.event", "evt-3", nil)
		err := bus.Publish(context.Background(), event)
		require.Error(t, err)
	})

	t.Run("multiple handlers all called", func(t *testing.T) {
		h1 := newMockHandler("multi1", []string{"multi.event"}, nil)
		h2 := newMockHandler("multi2", []string{"multi.event"}, nil)
		require.NoError(t, bus.Subscribe(h1))
		require.NoError(t, bus.Subscribe(h2))
		event := NewBaseEvent("multi.event", "evt-4", nil)
		err := bus.Publish(context.Background(), event)
		require.NoError(t, err)
		assert.Equal(t, 1, h1.Calls())
		assert.Equal(t, 1, h2.Calls())
	})
}

func TestMemoryEventBus_PublishAsync(t *testing.T) {
	bus := NewMemoryEventBus()
	defer bus.Shutdown(context.Background())

	var called int32
	h := newMockHandler("async-h", []string{"async.event"}, func(ctx context.Context, event Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})
	require.NoError(t, bus.Subscribe(h))

	event := NewBaseEvent("async.event", "evt-async", nil)
	bus.PublishAsync(context.Background(), event)

	// Wait up to 500ms for async processing
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&called) > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	assert.Equal(t, int32(1), atomic.LoadInt32(&called))
}

func TestMemoryEventBus_Unsubscribe(t *testing.T) {
	bus := NewMemoryEventBus()
	defer bus.Shutdown(context.Background())

	t.Run("unsubscribe existing handler", func(t *testing.T) {
		h := newMockHandler("removable", []string{"remove.event"}, nil)
		require.NoError(t, bus.Subscribe(h))
		err := bus.Unsubscribe("removable")
		require.NoError(t, err)
		subs := bus.GetSubscribers("remove.event")
		assert.Empty(t, subs)
	})

	t.Run("unsubscribe non-existent handler returns error", func(t *testing.T) {
		err := bus.Unsubscribe("ghost-handler")
		require.Error(t, err)
	})
}

func TestMemoryEventBus_GetSubscribers(t *testing.T) {
	bus := NewMemoryEventBus()
	defer bus.Shutdown(context.Background())

	t.Run("returns empty slice for unknown event type", func(t *testing.T) {
		subs := bus.GetSubscribers("no.such.event")
		assert.Nil(t, subs)
	})

	t.Run("returns all subscribers for event type", func(t *testing.T) {
		h1 := newMockHandler("sub-a", []string{"sub.test"}, nil)
		h2 := newMockHandler("sub-b", []string{"sub.test"}, nil)
		require.NoError(t, bus.Subscribe(h1))
		require.NoError(t, bus.Subscribe(h2))
		subs := bus.GetSubscribers("sub.test")
		assert.Len(t, subs, 2)
	})
}

func TestMemoryEventBus_Shutdown(t *testing.T) {
	bus := NewMemoryEventBus()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := bus.Shutdown(ctx)
	require.NoError(t, err)

	// Double shutdown should be a no-op
	err = bus.Shutdown(ctx)
	require.NoError(t, err)
}

func TestBaseEvent_Methods(t *testing.T) {
	event := NewBaseEvent("test.type", "event-id-123", map[string]string{"key": "val"})
	assert.Equal(t, "test.type", event.GetType())
	assert.Equal(t, "event-id-123", event.GetID())
	assert.NotZero(t, event.GetTimestamp())
	assert.NotNil(t, event.GetPayload())
	assert.NotNil(t, event.GetMetadata())
}
