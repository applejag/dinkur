<!--
Dinkur the task time tracking utility.
<https://github.com/dinkur/dinkur>

SPDX-FileCopyrightText: 2021 Kalle Fagerberg
SPDX-License-Identifier: CC-BY-4.0
-->

# Glossary

## Dinkur

Name of the product. Combination of "dinka", Swedish slang for pocket watch, and
"ur", Swedish term for a watch.

## Dinkur daemon

A [daemon](https://en.wikipedia.org/wiki/Daemon_\(computing\)) with an
[gRPC](https://grpc.io/) API to allow reading+writing to the task tracking from
other applications, such as [Dinkur clients](#dinkur-client).

The Dinkur daemon also performs ["away detection"](away-detection.md).

## Dinkur client

An application/process that talks to the [Dinkur daemon](#dinkur-daemon) through
[gRPC](https://grpc.io/).

## Dinkur CLI

A [Command-line interface (CLI)](https://en.wikipedia.org/wiki/Command-line_interface)
usable by humans and scripts to read+write to the Dinkur task tracking database.

The Dinkur CLI can be a [Dinkur client](#dinkur-client), if it detects that a
[Dinkur daemon](#dinkur-daemon) is already running. Otherwise, it will act on
the local database file directly.

## Dinkur frontend

A Dinkur frontend is a [Dinkur client](#dinkur-client) in the form of a
[Graphical User interface (GUI)](https://en.wikipedia.org/wiki/Graphical_user_interface)
that allows users to easily view and change their current task.

For users not using Dinkur via the command-line, this is how they interact with
their task tracking.

## Task

A Dinkur task is a work item, an event, or a task that the end-user wants to
track. A task has the following attributions:

- Name of the task, as entered by the end-user
- Start time, which was set when starting the task
- End time, which is left unset until the task is stopped

Tasks can only be listed, started, and stopped by
[Dinkur clients](#dinkur-client) or the [Dinkur CLI](#dinkur-cli).
