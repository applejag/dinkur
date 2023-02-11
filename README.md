<!--
Dinkur the task time tracking utility.
<https://github.com/dinkur/dinkur>

SPDX-FileCopyrightText: 2021 Kalle Fagerberg
SPDX-License-Identifier: CC-BY-4.0
-->

<div align="center">
  <img width="200" src="docs/dinkur-large-512.svg" alt="Dinkur icon" />
  <h1 style="color: #c43000">Dinkur</h1>
  <p>Task and time tracking CLI and framework.</p>

[![REUSE status](https://api.reuse.software/badge/github.com/dinkur/dinkur)](https://api.reuse.software/info/github.com/dinkur/dinkur)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/956b94a743244ce2a971ce572e05be3e)](https://www.codacy.com/gh/dinkur/dinkur/dashboard?utm_source=github.com\&utm_medium=referral\&utm_content=dinkur/dinkur\&utm_campaign=Badge_Grade)

</div>

## Install

Requires [Go](https://go.dev/) v1.20 (or higher)

```console
$ go install -tags='fts5' -ldflags='-s -w' github.com/dinkur/dinkur@latest
```

> The `-tags='fts5'` flag adds [Sqlite FTS5](https://www.sqlite.org/fts5.html)
> support, which is used for better and more performant full-text search.
>
> The `-ldflag='-s -w'` removes debug symbols, reducing the binary size from
> about 34M down to 13M.

For you CLI-power users, we recommend aliasing it to `u`.

```sh
alias u=dinkur
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

### Run daemon on boot (systemd)

Copy the [`dinkur.service`](./dinkur.service) file from the repo into your
local user systemd unit store:

```sh
sudo cp dinkur.service /usr/local/lib/systemd/user/dinkur.service
```

Then enable it only for your user:

```sh
## Does not need sudo
systemctl --user enable dinkur --now

## Read logs, does not need sudo either
journalctl --user --unit dinkur --follow

## Try it out
dinkur --client=grpc status
```

## Usage

```console
$ u in Speedrun Minecraft
                ID  NAME                  START  END      DURATION
Started entry:  #1  `Speedrun Minecraft`  18:39  active…  -

$ u in Boil minute-rice
                ID  NAME                  START  END      DURATION
Stopped entry:  #1  `Speedrun Minecraft`  18:39  18:39    0:00:06
Started entry:  #2  `Boil minute-rice`    18:39  active…  -

$ u out
                ID  NAME                START  END    DURATION
Stopped entry:  #2  `Boil minute-rice`  18:39  18:40  0:01:01

$ u list
  ID  NAME                  DAY     START  END    DURATION
  #1  `Speedrun Minecraft`  Jan-20  18:39  18:39  0:00:06
  #2  `Boil minute-rice`    -       18:39  18:40  0:01:01

  -   TOTAL: 2 entries      -       18:39  18:40  0:01:07
```

Full documentation can be found at [docs/cmd/dinkur.md](docs/cmd/dinkur.md).

## Contributing

Read how to contribute over at [CONTRIBUTING.md](CONTRIBUTING.md), including
how to set up your development environment, if you so feel inclined.

## Inspiration sources

<!--lint disable maximum-line-length-->

| Project         | License?       | CLI? | GUI? | Sync?   | AFK detect?   | OS?                   |
| --------------- | -------------- | ---- | ---- | ------- | ------------- | --------------------- |
| [Grindstone][g] | Proprietary    | ❌    | ✅    | ✅ *($)* | ✅ *(Windows)* | Windows, Android, iOS |
| [𝑓𝑓][ff]      | FOSS *(GPLv3)* | ✅    | ✅    | ✅       | ❌             | Linux, Mac, Android   |
| [timetrap][t]   | OSS *(MIT)*    | ✅    | ❌    | ❌       | ❌             | Windows, Linux, Mac   |

<!--lint enable maximum-line-length-->

No code is taken from the above projects. However, they all have some distinct
features each that I'm greatly inspired by and have implemented into Dinkur.

## License

This repository is created and maintained by Kalle Fagerberg
([@jilleJr](https://github.com/jilleJr)).

The code in this project is licensed under GNU General Public License v3.0
or later ([LICENSES/GPL-3.0-or-later.txt](LICENSES/GPL-3.0-or-later.txt)),
and documentation is licensed under Creative Commons Attribution 4.0
International ([LICENSES/CC-BY-4.0.txt](LICENSES/CC-BY-4.0.txt)).

[g]: https://epiforge.com/grindstone

[ff]: https://github.com/ff-notes/ff

[t]: https://github.com/samg/timetrap
