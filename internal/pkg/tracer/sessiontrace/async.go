package sessiontrace

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	Enabled            bool
	QueueSize          int
	MaxFieldValueBytes int
	MaxEventsPerSess   int
}

type asyncTracer struct {
	cfg   Config
	sinks []Sink

	ch     chan Event
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewAsync(cfg Config, sinks ...Sink) Tracer {
	if !cfg.Enabled || len(sinks) == 0 {
		return NewNoop()
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 4096
	}
	if cfg.MaxFieldValueBytes <= 0 {
		cfg.MaxFieldValueBytes = 2048
	}
	if cfg.MaxEventsPerSess <= 0 {
		cfg.MaxEventsPerSess = 500
	}
	t := &asyncTracer{
		cfg:    cfg,
		sinks:  sinks,
		ch:     make(chan Event, cfg.QueueSize),
		stopCh: make(chan struct{}),
	}
	t.wg.Add(1)
	go t.run()
	return t
}

func (t *asyncTracer) run() {
	defer t.wg.Done()
	ctx := context.Background()
	for {
		select {
		case <-t.stopCh:
			// drain best-effort
			for {
				select {
				case e := <-t.ch:
					for _, s := range t.sinks {
						_ = s.Write(ctx, e)
					}
				default:
					return
				}
			}
		case e := <-t.ch:
			for _, s := range t.sinks {
				_ = s.Write(ctx, e)
			}
		}
	}
}

func (t *asyncTracer) NewSession(ctx context.Context, meta SessionMeta) Session {
	return &asyncSession{
		tracer:  t,
		id:      uuid.NewString(),
		meta:    meta,
		started: time.Now(),
	}
}

func (t *asyncTracer) Close() error {
	close(t.stopCh)
	t.wg.Wait()
	var firstErr error
	for _, s := range t.sinks {
		if err := s.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type asyncSession struct {
	tracer  *asyncTracer
	id      string
	meta    SessionMeta
	started time.Time

	mu       sync.Mutex
	eventCnt int
	ended    bool
}

func (s *asyncSession) ID() string { return s.id }

func (s *asyncSession) Event(stage string, fields map[string]any) {
	s.emit(stage, "", fields, 0)
}

func (s *asyncSession) Error(stage string, err error, fields map[string]any) {
	msg := ""
	if err != nil {
		msg = Trunc(err.Error(), 2048)
	}
	s.emit(stage, msg, fields, 0)
}

func (s *asyncSession) End(stage string, fields map[string]any) {
	s.emit(stage, "", fields, time.Since(s.started))
	s.mu.Lock()
	s.ended = true
	s.mu.Unlock()
}

func (s *asyncSession) emit(stage, errMsg string, fields map[string]any, dur time.Duration) {
	s.mu.Lock()
	if s.ended || s.eventCnt >= s.tracer.cfg.MaxEventsPerSess {
		s.mu.Unlock()
		return
	}
	s.eventCnt++
	s.mu.Unlock()

	e := Event{
		Timestamp: time.Now(),
		SessionID: s.id,
		Protocol:  s.meta.Protocol,
		Stage:     stage,
		Remote:    s.meta.Remote,
		Local:     s.meta.Local,
		Duration:  dur,
		Err:       errMsg,
		Fields:    fields,
		Tags:      s.meta.Tags,
	}
	select {
	case s.tracer.ch <- e:
	default:
		// drop on backpressure
	}
}
