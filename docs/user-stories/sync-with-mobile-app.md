<!--
Dinkur the task time tracking utility.
<https://github.com/dinkur/dinkur>

SPDX-FileCopyrightText: 2021 Kalle Fagerberg
SPDX-License-Identifier: CC-BY-4.0
-->

# User story: Sync with mobile app

As an end user, I can use use a mobile app that syncs with the Dinkur daemon, so
that I can keep tracking time when having outdoor meetings.

## Prerequisites

- Dinkur gRPC API supporting TLS.
- Dinkur gRPC API supporting authentication.
- Dinkur daemon "eventual consistent" syncing across daemon instances.

## Scope

- Android only. If cross-platform (iOS + Android) tools are too tedious, then
  focus only on Android.

## Specifications

- Sync between phone app and desktop Dinkur daemon needs to be able to only rely
  on local Wi-Fi. Maybe UDP for discovery and then TCP for connecting.
  For security reasons, proper authentication and TLS is a must here.

- Needs eventual consistent data storage. The phone app must be able to perform
  full CRUD operations on its own offline-first storage. The sync feature is an
  addition, that needs to be able to resolve the differences between the desktop
  and phone stores.

  Hosting a separate local Dinkur daemon instance on the phone via
  [Go mobile](https://pkg.go.dev/golang.org/x/mobile#section-readme) could be a
  wise solution.

- Data needs to be merged. Consider making the Dinkur tasks behave as an
  [CRDT](https://en.wikipedia.org/wiki/Conflict-free_replicated_data_type).

  Could timestamp all changes, such as via
  [event sourcing](https://martinfowler.com/eaaDev/EventSourcing.html), and only
  use the Sqlite database for faster queries.

  Prioritize keeping data. So if one the desktop Dinkur daemon instance deletes
  a task and the mobile app Dinkur daemon edits that same task, then it should
  resolve as non-deleted edited task.

  May be of use: [RON](http://replicated.cc/)
