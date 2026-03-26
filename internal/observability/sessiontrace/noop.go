package sessiontrace

import "context"

type noopTracer struct{}
type noopSession struct{}

func NewNoop() Tracer { return noopTracer{} }

func (noopTracer) NewSession(ctx context.Context, meta SessionMeta) Session { return noopSession{} }
func (noopTracer) Close() error                                            { return nil }

func (noopSession) ID() string                                          { return "" }
func (noopSession) Event(stage string, fields map[string]any)           {}
func (noopSession) Error(stage string, err error, fields map[string]any) {}
func (noopSession) End(stage string, fields map[string]any)             {}

