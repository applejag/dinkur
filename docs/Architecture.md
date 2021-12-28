# Architecture

When used via CLI:

```
  DB (Sqlite3)
      ^
      |
Dinkur CLI (Go)
      ^
      |
    (Bash)
      |
   End-user
```

When used via clients:

```
  DB (Sqlite3)
      ^
      |
Dinkur daemon (Go)
      ^
      |
    (gRPC)
      |
 Dinkur client
      ^
      |
   End-user
```

Dinkur clients talk to the Dinkur daemon through gRPC. This allows custom
clients for either web-integration, CLI usage, or desktop applications.

- Dinkur CLI
- Web-extension, e.g a Chromium extension or Greesemonkey user-script
- Desktop application, e.g WinForms app for Windows or GTK+ for GNU/Linux.

## Dinkur CLI

```console
$ date
Tue Dec 28 14:37:44 CET 2021

$ dinkur in My task name --start "10min ago"
Started tracking for task "My task name" from 14:27.

$ dinkur out --end "14:30"
Stopped tracking task "My task name". 14:27 - 14:30 (3min)
No task is currently tracked.

$ dinkur edit --end "14:35"
Updated task "My task name":
  14:27 - 14:30 (3min)   =>   14:27 - 14:35 (8min)

$ dinkur list --today
TASK          START  END    DURATION
My task name  14:27  14:35  8min
------------  -----  -----  --------
Total         14:27  14:35  8min

$ dinkur daemon
Dinkur daemon running on port 41231.
Press CTRL+C to exit...
```

## Dinkur daemon

The Dinkur daemon exposes a gRPC API over TCP/IP, which Dinkur clients rely on.

It also performs "away detection".

### gRPC API outline

- Listen to current task changes
- Update current task
- Search task history, such as for autocompletion

### Security

- IP-blocked: Only allow access from `127.0.0.1` (IPv4) and `::1` (IPv6)
- Authentication: Token-based authentication in HTTP header for gRPC connection.

### gRPC TCP/IP port selection

Randomly selected and outputted by the Dinkur backend to STDOUT. Example:

```console
$ dinkur daemon --output json
{ "port": 41231, "authToken": "ai8ve89q3jakzef9ake0k3ma9fa3" }
```

## Frontends

Can be published with an embedded Dinkur CLI and starting it's own daemon when
needed, or rely on a pre-installed one.

### Sharing daemon

If a daemon is already started, then frontends should reuse the existing daemon
instead of start up new ones.

> TODO: Missing design. To begin with, lock the database file and disallow
> multiple clients.

## Database

Sqlite3

### Optional Sqlite3 extensions

- [FTS5](https://www.sqlite.org/fts5.html) for smarter search results and
  autocompletions.

  - If available: `SELECT * FROM tasks_idx WHERE name MATCH 'foobar'`
  - If unavailable: `SELECT * FROM tasks WHERE name LIKE 'foobar%'`

It's up to the Dinkur CLI to automatically use the extensions if available and
not crash if unavailable.
