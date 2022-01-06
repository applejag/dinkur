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
> about 34M down to 13M.

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

Full documentation can be found at [docs/cmd/dinkur.md](docs/cmd/dinkur.md).

## Contributing

Read how to contribute over at [CONTRIBUTING.md](CONTRIBUTING.md), including
how to set up your development environment, if you so feel inclined.

## Similar projects comparison

<!--lint disable maximum-line-length-->

| Project                                       | License?          | CLI? | GUI?                | Sync?                        | OS?                   |
| --------------------------------------------- | ----------------- | ---- | ------------------- | ---------------------------- | --------------------- |
| [Grindstone](https://epiforge.com/grindstone) | Proprietary       | ❌   | ✅                  | ✅ _($)_                     | Windows, Android, iOS |
| [𝑓𝑓](https://github.com/ff-notes/ff)          | FOSS _(GPL 3.0)_  | ✅   | ✅ _(Experimental)_ | ✅ _(via external, eg. Git)_ | Linux, Mac, Android   |
| [timetrap](https://github.com/samg/timetrap)  | OSS _(MIT)_       | ✅   | ❌                  | ❌                           | Windows, Linux, Mac   |
| Dinkur _(this repo)_                          | FOSS _(GPL 3.0+)_ | ✅   | ❌                  | ❌                           | Windows, Linux, Mac   |

<!--lint enable maximum-line-length-->

## License

This repository is created and maintained by Kalle Fagerberg
([@jilleJr](https://github.com/jilleJr)).

The code in this project is licensed under GNU General Public License v3.0
or later ([LICENSES/GPL-3.0-or-later.txt](LICENSES/GPL-3.0-or-later.txt)),
and documentation is licensed under Creative Commons Attribution 4.0
International ([LICENSES/CC-BY-4.0.txt](LICENSES/CC-BY-4.0.txt)).
