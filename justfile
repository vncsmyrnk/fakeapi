default:
  just --list

run-server *args:
  go run cmd/server/main.go {{args}}

run-cli *args:
  go run cmd/cli/main.go {{args}}

debug *args:
  dlv debug --headless --listen=:2345 --api-version=2 cmd/server/main.go -- {{args}}

generate:
  go generate ./...

build-server:
  @mkdir -p dist
  go build -o dist/fakeapi ./cmd/server/main.go

docker-build:
  nix build .#docker

install-linter:
  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

lint: install-linter lint-only

lint-only:
  golangci-lint run
