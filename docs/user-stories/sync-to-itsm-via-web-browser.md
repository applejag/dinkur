<!--
Dinkur the task time tracking utility.
<https://github.com/dinkur/dinkur>

SPDX-FileCopyrightText: 2021 Kalle Fagerberg
SPDX-License-Identifier: CC-BY-4.0
-->

# User story: Sync to ITSM via web browser

As an end user, I can let my time reporting be automatically submitted to our
internal Jira instance, so that I don't have to manually copy data from one
system to another.

## Prerequisites

- Dinkur gRPC API supporting TLS.
- Dinkur gRPC API supporting authentication.

## Scope

- Focus on writing a [userscript](https://en.wikipedia.org/wiki/Userscript) for
  simplicity's sake.

- Focus on Jira. Other [ITSM](https://en.wikipedia.org/wiki/IT_service_management)
  systems can have custom userscripts or web browser extensions later, while
  taking inspiration from this first userscript.

## Specifications

- The userscript needs to be able to correlate tasks with Jira tickets. Adding
  some queryable metadata to tasks in the Dinkur database could be an option
  (such as an additional SQL table). This can allow smart suggestions from
  Dinkur such as "as you previously reported task with this (or similar) name to
  Jira ticket #12345, do you want to report this other task there as well?"
