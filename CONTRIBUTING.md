<!--
Dinkur the task time tracking utility.
<https://github.com/dinkur/dinkur>

SPDX-FileCopyrightText: 2021 Kalle Fagerberg
SPDX-License-Identifier: CC-BY-4.0
-->

# Contributing

## How to contribute

There's multiple ways to contribute:

- Spread the word! Share the project with coworkers and friends. This is the
  best way to contribute :heart:

- Report bugs or wanted features to our issue tracker:
  <https://github.com/dinkur/dinkur/issues/new>

- Tackle an issue in <https://github.com/dinkur/dinkur/issues>. If you see one
  that's tempting, ask in the issue's thread and I'll assign you so we don't get
  multiple people working on the same thing.

## Development

### Prerequisites

- To build and run the code:

  - [Go](https://go.dev/) v1.17 (or higher)

- To modify the Protocol Buffer definition and regenerate server & client:

  - [Protocol Buffer compiler](https://grpc.io/docs/protoc-installation/)
    (`protoc`) v3. Make sure to add it to your PATH

  - [Protocol Buffer compiler Go plugins](https://grpc.io/docs/languages/go/quickstart/#prerequisites).
    Can be installed by running `make deps`

```sh
# Run the Dinkur CLI:
go run --tags fts5 .

# Regenerates gRPC code (requires protoc + Go plugins):
make grpc

# Regenerates CLI markdown documentation:
make docs
```

### Formatting

- Go formatting is performed via [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)
  and can be installed via `make deps`

- Markdown formatting can be (but doesn't have to be) done via [Prettier](https://prettier.io/)

Make sure to regularly lint the project locally. We're relying on the Markdown
linting for its formatting.

### Linting

- Go linting is performed via [Revive](https://revive.sh)

- Protobuf linting is performed via [protolint](https://github.com/yoheimuta/protolint)
  and requires Go.

- Markdown linting is performed via [remarklint](https://github.com/remarkjs/remark-lint)
  and requires NPM.

- Licensing linting is performed via [REUSE](https://reuse.software/) and
  requires Python v3.

You can install all the above linting rules by running `make deps`

```sh
# Lint .go, .md, & .proto files, and REUSE compliance:
make lint

# Apply fixes, where possible, to .md & .proto files:
make lint-fix

# Only lint some:
make lint-md
make lint-go
make lint-proto
make lint-license

# Only lint & fix some:
make lint-md-fix
make lint-proto-fix
```
