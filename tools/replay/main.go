package main

import (
	"context"
	"encoding/json"
	"easymail/internal/domain/mailpipeline"
	"easymail/internal/policy"
	"flag"
	"fmt"
	"os"
)

type policyCase struct {
	Name      string `json:"name"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Expect    string `json:"expect"`
}

func main() {
	fixtures := flag.String("fixtures", "", "fixture json path")
	flag.Parse()
	if *fixtures == "" {
		fmt.Println("missing -fixtures")
		os.Exit(2)
	}
	bs, err := os.ReadFile(*fixtures)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	var cases []policyCase
	if err = json.Unmarshal(bs, &cases); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	eval := policy.NewEvaluator()
	for _, c := range cases {
		d, err := eval.Evaluate(context.Background(), c.Sender, c.Recipient)
		if err != nil {
			fmt.Printf("case=%s err=%v\n", c.Name, err)
			os.Exit(1)
		}
		if string(d) != c.Expect {
			fmt.Printf("case=%s mismatch want=%s got=%s\n", c.Name, c.Expect, d)
			os.Exit(1)
		}
	}
	fmt.Printf("replay ok, cases=%d, decisions=[%s|%s]\n", len(cases), mailpipeline.PolicyDunno, mailpipeline.PolicyReject)
}

