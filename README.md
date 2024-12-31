[![CI workflow](https://github.com/vncsmyrnk/fakeapi/actions/workflows/ci.yml/badge.svg)](https://github.com/vncsmyrnk/fakeapi/actions/workflows/ci.yml)
[![Release workflow](https://github.com/vncsmyrnk/fakeapi/actions/workflows/release.yml/badge.svg)](https://github.com/vncsmyrnk/fakeapi/actions/workflows/release.yml)
[![contributions](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/vncsmyrnk/fakeapi/issues)
[![Issue count](https://img.shields.io/github/issues-search?query=repo%3Avncsmyrnk%2Ffakeapi%20is%3Aopen&label=open%20issues)](https://github.com/vncsmyrnk/fakeapi/issues)

# Fake API

CLI for dummy REST API testing other software.

Fake API allows you to quickly set up and run a RESTful API server with customizable endpoints and responses. This makes it an ideal tool for developers who need a reliable and easy-to-use mock server for testing their software without relying on external services.

## Examples

You are building a client interface for an app that does not yet have its backend services properly set up. You only have a contract of what the backend will be in the future. How can you mock this API in a way that behaves as closely as possible to the real one? Fake API was designed to address this issue.

It is useful not only for building frontend interfaces but also whenever you need to mock a REST service with customizable controls.

```sh
cat <<EOF > server-config.json
[
  {
    "path": "/my-route",
    "method": "POST",
    "status": 200,
    "content": {
      "name": "value",
      "other-name": {
        "some-other-name": "another-value"
      }
    }
  }
]
EOF
fakeapi --port 8080 server-config.json
```

```sh
curl -X POST localhost:8080/my-route
# {"name":"value","other-name":{"some-other-name":"another-value"}}
```

## Install

### Directly with Go

```sh
go install github.com/vncsmyrnk/fakeapi@latest
```

### Manually

Check the [releases section](https://github.com/vncsmyrnk/fakeapi/releases) for the binaries.

Example:

```sh
curl -L https://github.com/vncsmyrnk/fakeapi/releases/latest/download/fakeapi-linux-amd64 -o fakeapi # Download the correct binary according to your settings
mv fakeapi ~/.local/bin # Move it to an executable path
```

### Via package managers

Work in progress...

## Development

This project is still in development. Feel free to open a issue or a PR!
