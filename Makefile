GO ?= go

.PHONY: build test fmt vet replay ci

build:
	$(GO) build ./...

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

replay:
	$(GO) run ./tools/replay -fixtures ./tools/replay/fixtures/policy_cases.json

ci: fmt vet build test replay
