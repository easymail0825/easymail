package policy

import postfixpolicy "easymail/internal/postfix/policy"

// NewServer keeps protocol behavior compatible while moving
// policy into the new modular package namespace.
func NewServer(family, listen string) *postfixpolicy.Server {
	return postfixpolicy.New(family, listen)
}

