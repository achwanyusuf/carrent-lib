.PHONY: ci
ci:
	$(shell go env GOPATH)/bin/golangci-lint run --verbose
