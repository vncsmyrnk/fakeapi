default:
  just --list

test:
  CGO_ENABLED=1 go test -cover -race ./...

build:
  go build -buildvcs=false ./cmd/fakeapi

install-linter:
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: install-linter
  golangci-lint run -buildvcs=false
