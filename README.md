<!--
Dinkur the task time tracking utility.
<https://github.com/dinkur/dinkur>

SPDX-FileCopyrightText: 2021 Kalle Fagerberg
SPDX-License-Identifier: CC-BY-4.0
-->

# Dinkur

[![REUSE status](https://api.reuse.software/badge/github.com/dinkur/dinkur)](https://api.reuse.software/info/github.com/dinkur/dinkur)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/956b94a743244ce2a971ce572e05be3e)](https://www.codacy.com/gh/dinkur/dinkur/dashboard?utm_source=github.com&utm_medium=referral&utm_content=dinkur/dinkur&utm_campaign=Badge_Grade)

Task and time tracking utility.

## Install

Requires [Go](https://go.dev/) v1.17 (or higher)

```console
$ go install -tags='fts5' -ldflags='-s -w' github.com/dinkur/dinkur@latest
```

> The `-tags='fts5'` flag adds [Sqlite FTS5](https://www.sqlite.org/fts5.html)
> support, which is used for better and more performant full-text search.
>
> The `-ldflag='-s -w'` removes debug symbols, reducing the binary size from
> about 24M down to 8M.

For you CLI-power users, we recommend aliasing it to `ur`.

```sh
alias ur=dinkur
```

### CLI Autocompletion

Automatic generation of completions are powered by [Cobra](https://github.com/spf13/cobra),
which supports Bash, Zsh, Fish, and PowerShell.

Completions are provided by the `completion` subcommand. To get a more detailed
guide on how to install them, run the following (where `bash` can be exchanged
with `zsh`, `fish`, and `powershell`):

```sh
dinkur completion bash --help
```

## Usage

Full documentation can be found at [docs/cmd/dinkur.md](docs/cmd/dinkur.md)

## Development

Prerequisites:

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

## License

This repository is created and maintained by Kalle Fagerberg
([@jilleJr](https://github.com/jilleJr)).

The code in this project is licensed under GNU General Public License v3.0
or later ([LICENSES/GPL-3.0-or-later.txt](LICENSES/GPL-3.0-or-later.txt)),
and documentation is licensed under Creative Commons Attribution 4.0
International ([LICENSES/CC-BY-4.0.txt](LICENSES/CC-BY-4.0.txt)).
