default:
  just --list

test:
  go test -cover -race ./...

install-linter:
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: install-linter
  golangci-lint run
