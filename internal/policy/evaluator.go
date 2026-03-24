package policy

import (
	"context"
	"easymail/internal/domain/mailpipeline"
	"easymail/internal/model"
	"strings"
)

type Evaluator struct{}

func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate(_ context.Context, sender, recipient string) (mailpipeline.PolicyDecision, error) {
	if sender == "" || recipient == "" {
		return mailpipeline.PolicyReject, nil
	}
	if _, err := model.FindAccountByName(sender); err != nil {
		return mailpipeline.PolicyReject, nil
	}

	recipient = strings.ToLower(recipient)
	d := strings.SplitN(recipient, "@", 2)
	if len(d) != 2 {
		return mailpipeline.PolicyReject, nil
	}
	if _, err := model.FindDomainByName(d[1]); err == nil {
		if _, err = model.FindAccountByName(recipient); err != nil {
			return mailpipeline.PolicyReject, nil
		}
	}
	return mailpipeline.PolicyDunno, nil
}

