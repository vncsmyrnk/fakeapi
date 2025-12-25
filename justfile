default:
  just --list

test: unit-test integration-test

unit-test:
  go test -cover ./...

integration-test:
  go test -v ./test/integration/...

run *args:
  go run cmd/fakeapi/main.go {{args}}

debug *args:
  dlv debug --headless --listen=:2345 --api-version=2 cmd/fakeapi/main.go -- {{args}}

coverage:
  go test -coverprofile=coverage.txt ./...

generate:
  go generate ./...

build:
  @mkdir -p dist
  go build -o dist/fakeapi ./cmd/fakeapi/main.go

install-linter:
  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

lint: install-linter lint-only

lint-only:
  golangci-lint run
