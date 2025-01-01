default:
  just --list

test:
  go test -cover ./...

coverage:
  go test -coverprofile=coverage.txt

generate:
  go generate ./...

build:
  go build ./cmd/fakeapi

install-linter:
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: install-linter
  golangci-lint run
