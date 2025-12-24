default:
  just --list

test:
  go test -cover ./...

run *args:
  go run cmd/fakeapi/main.go {{args}}

coverage:
  go test -coverprofile=coverage.txt ./...

generate:
  go generate ./...

build:
  go build ./cmd/fakeapi

install-linter:
  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

lint: install-linter
  golangci-lint run
