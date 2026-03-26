package sessiontrace

import (
	"context"
	"time"
)

type Protocol string

const (
	ProtocolLMTP   Protocol = "lmtp"
	ProtocolMilter Protocol = "milter"
	ProtocolPolicy Protocol = "policy"
	ProtocolAuth   Protocol = "dovecot_auth"
)

type Event struct {
	Timestamp time.Time         `json:"ts"`
	SessionID string            `json:"session_id"`
	Protocol  Protocol          `json:"protocol"`
	Stage     string            `json:"stage"`
	Remote    string            `json:"remote,omitempty"`
	Local     string            `json:"local,omitempty"`
	Duration  time.Duration     `json:"duration,omitempty"`
	Err       string            `json:"err,omitempty"`
	Fields    map[string]any    `json:"fields,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type SessionMeta struct {
	Protocol Protocol
	Remote   string
	Local    string
	Tags     map[string]string
}

type Sink interface {
	Write(ctx context.Context, e Event) error
	Close() error
}

type Tracer interface {
	NewSession(ctx context.Context, meta SessionMeta) Session
	Close() error
}

type Session interface {
	ID() string
	Event(stage string, fields map[string]any)
	Error(stage string, err error, fields map[string]any)
	End(stage string, fields map[string]any)
}

