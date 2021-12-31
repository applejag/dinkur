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

- Go linting is performed via [Revive](https://revive.sh) and can be installed
  via `make deps`

- Markdown linting is performed via [remarklint](https://github.com/remarkjs/remark-lint)
  and can be installed via `npm install`

- Licensing linting is performed via [REUSE](https://reuse.software/), and can
  be installed via `pip3 install --user reuse`

```sh
# Lint .go & .md:
npm run lint

# Apply fixes, where possible, to .md files:
npm run lint-fix

# Only lint/fix some:
npm run lint-md
npm run lint-go
npm run lint-md-fix

# Lint licensing:
reuse lint
```
