package sessiontrace

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type countingSink struct {
	n atomic.Int64
}

func (s *countingSink) Write(ctx context.Context, e Event) error {
	s.n.Add(1)
	return nil
}
func (s *countingSink) Close() error { return nil }

func TestAsyncTracerWritesEvents(t *testing.T) {
	sink := &countingSink{}
	tr := NewAsync(Config{
		Enabled:   true,
		QueueSize: 16,
	}, sink)
	defer tr.Close()

	sess := tr.NewSession(context.Background(), SessionMeta{Protocol: ProtocolLMTP, Remote: "r", Local: "l"})
	sess.Event("a", map[string]any{"k": "v"})
	sess.End("end", nil)

	// best-effort: allow worker to drain
	time.Sleep(50 * time.Millisecond)
	if sink.n.Load() == 0 {
		t.Fatalf("expected events written")
	}
}

func TestAsyncSessionEventLimit(t *testing.T) {
	sink := &countingSink{}
	tr := NewAsync(Config{
		Enabled:          true,
		QueueSize:        128,
		MaxEventsPerSess: 3,
	}, sink)
	defer tr.Close()

	sess := tr.NewSession(context.Background(), SessionMeta{Protocol: ProtocolLMTP})
	for i := 0; i < 10; i++ {
		sess.Event("x", nil)
	}
	sess.End("end", nil)
	time.Sleep(50 * time.Millisecond)
	if sink.n.Load() > 10 {
		t.Fatalf("unexpected count: %d", sink.n.Load())
	}
}
