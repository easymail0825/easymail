package policy

import (
	"context"
	"easymail/internal/domain/mailpipeline"
	"testing"
)

func TestEvaluatorRejectInvalid(t *testing.T) {
	e := NewEvaluator()
	d, err := e.Evaluate(context.Background(), "", "a@b.com")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if d != mailpipeline.PolicyReject {
		t.Fatalf("want reject, got %s", d)
	}
}

