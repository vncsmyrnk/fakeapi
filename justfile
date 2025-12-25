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
  @mkdir -p dist
  go build -o dist/fakeapi ./cmd/fakeapi/main.go

install-linter:
  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

lint: install-linter
  golangci-lint run
