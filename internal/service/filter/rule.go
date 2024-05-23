package filter

import (
	"easymail/internal/model"
	"errors"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"strings"
)

func loadRules() (k *ast.KnowledgeLibrary, kb *ast.KnowledgeBase, rb *builder.RuleBuilder, re *engine.GruleEngine, err error) {
	drlList := make([]string, 0)
	rules, err := model.GetFilterRules()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for _, rule := range rules {
		if s, err := rule.Convert2DRL(); err == nil {
			drlList = append(drlList, s)
		}
	}

	// must have one rule at least
	if len(drlList) == 0 {
		return nil, nil, nil, nil, errors.New("no rule loaded")
	}
	//fmt.Println(strings.Join(drlList, "\n"))
	bs := pkg.NewBytesResource([]byte(strings.Join(drlList, "\n")))
	k = ast.NewKnowledgeLibrary()
	rb = builder.NewRuleBuilder(k)
	err = rb.BuildRuleFromResource("EasymailRules", "1.0.0", bs)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	kb, err = k.NewKnowledgeBaseInstance("EasymailRules", "1.0.0")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	re = engine.NewGruleEngine()
	return
}
