package mailpipeline

import "context"

type PolicyDecision string

const (
	PolicyDunno  PolicyDecision = "dunno"
	PolicyReject PolicyDecision = "reject"
)

type PolicyEvaluator interface {
	Evaluate(ctx context.Context, sender, recipient string) (PolicyDecision, error)
}

type Storage interface {
	SaveRaw(ctx context.Context, recipient string, raw []byte) (string, error)
}

