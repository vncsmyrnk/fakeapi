[![CI workflow](https://github.com/vncsmyrnk/fakeapi/actions/workflows/ci.yml/badge.svg)](https://github.com/vncsmyrnk/fakeapi/actions/workflows/ci.yml)
[![Release workflow](https://github.com/vncsmyrnk/fakeapi/actions/workflows/release.yml/badge.svg)](https://github.com/vncsmyrnk/fakeapi/actions/workflows/release.yml)
[![CLI version](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fvncsmyrnk%2Ffakeapi%2Frefs%2Fheads%2Fmain%2Fcmd%2Fcli%2FVERSION&search=.*&label=CLI%20version)](https://github.com/vncsmyrnk/fakeapi/blob/main/flake.nix)
[![Docker image version](https://img.shields.io/docker/v/vncsmyrnk/fakeapi)](https://hub.docker.com/r/vncsmyrnk/fakeapi/tags)
[![LuaRocks version](https://img.shields.io/luarocks/v/vncsmyrnk/fakeapi)](https://luarocks.org/modules/vncsmyrnk/fakeapi)
<br>
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

You can also filter for body attributes, headers and query strings using [GJSON Path Syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md).

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
luarocks install fakeapi --local --lua-version=5.1
```

## Roadmap and new features

Check the [milestones section](https://github.com/vncsmyrnk/fakeapi/milestones) to see what is currently being planned or in development.
