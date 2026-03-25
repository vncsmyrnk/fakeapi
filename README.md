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
    "statusCode": 200,
    "response": {
      "name": "value",
      "other-name": {
        "some-other-name": "another-value"
      }
    }
  }
]
EOF
fakeapi-server --port 8080 server-config.json
```

```sh
curl -X POST localhost:8080/my-route
# {"name":"value","other-name":{"some-other-name":"another-value"}}
```

## Assertions

All requests sent to the fake API are stored locally and can be queried later for assertions. This is useful for verifying that the expected requests were made to the API during automated tests.

The assert CLI checks whether a specific request was received by the server. It returns a zero exit code on success and a non-zero exit code on failure.

```sh
fakeapi assert --port 8080 POST /my-route
# Exit code 0: a POST request to /my-route was received
# Exit code 1: no matching request was found
```

## Install

### Server

```sh
nix profile add github:vncsmyrnk/fakeapi#fakeapi-server
```

```sh
docker run --rm -it -p 8080:8080 vncsmyrnk/fakeapi
```

### CLI

```sh
nix profile add github:vncsmyrnk/fakeapi
```

### APIs

Wrappers for `fakeapi` in other languages/tools.

#### Lua

```sh
luarocks install https://raw.githubusercontent.com/vncsmyrnk/fakeapi/main/api/lua/fakeapi.rockspec
```

## Roadmap and new features

Check the [milestones section](https://github.com/vncsmyrnk/fakeapi/milestones) to see what is currently being planned or in development.
