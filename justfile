default:
  just --list

run-server *args:
  FAKEAPI_DB_PATH="$HOME/file.db" go run cmd/server/main.go {{args}}

run-cli *args:
  go run cmd/cli/main.go {{args}}

debug-server *args:
  FAKEAPI_DB_PATH="$HOME/file.db" dlv debug --headless --listen=:2345 --api-version=2 cmd/server/main.go -- {{args}}

debug-cli *args:
  dlv debug --headless --listen=:2345 --api-version=2 cmd/cli/main.go -- {{args}}

generate:
  go generate ./...

build-server:
  @mkdir -p dist
  go build -o dist/fakeapi ./cmd/server/main.go

docker-build:
  nix build .#docker

install-linter:
  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

install-lua-module:
  luarocks --local --lua-version=5.1 install api/lua/*.rockspec

uninstall-lua-module:
  luarocks --local --lua-version=5.1 remove api/lua/*.rockspec

lint:
  golangci-lint run
