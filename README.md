# PortMaster

Fast, cross-platform port and process management for developers.

[![CI](https://github.com/RichardFlp/portmaster/actions/workflows/ci.yml/badge.svg)](https://github.com/RichardFlp/portmaster/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/RichardFlp/portmaster)](https://github.com/RichardFlp/portmaster/releases)
[![License](https://img.shields.io/github/license/RichardFlp/portmaster)](LICENSE)

PortMaster tells you what is using a port and lets you act on it. It detects
processes listening on TCP and UDP ports, shows which process owns each port,
and supports terminating processes, opening ports in the browser, finding free
ports, searching, filtering, and watching for port changes. It runs locally,
offline, and does not send any data anywhere.

```
$ portmaster

 PORT   PROTOCOL  PID    PROCESS       ADDRESS
3000   TCP       18432  node          127.0.0.1
5173   TCP       19281  node          0.0.0.0
5432   TCP       4212   postgres      127.0.0.1
6379   TCP       5511   redis-server  127.0.0.1
```

## Features

- List listening TCP and UDP ports with process names and PIDs.
- Inspect a port to see status, protocol, address, executable, command line, parent process, and start time.
- Terminate processes with explicit confirmation, protected against killing PortMaster itself.
- Find available ports starting from a given port.
- Check whether a port is in use.
- Open a listening port in the default browser.
- Search across ports, PIDs, process names, executables, and command lines.
- Filter listings by process, port, and protocol.
- Report occupied or free ports within a range.
- Watch for ports appearing and disappearing, with a configurable refresh interval.
- Interactive terminal UI with keyboard navigation.
- Machine-readable JSON output and quiet mode for scripting.
- Optional configuration file with sensible defaults.

## Supported operating systems

- Windows (64-bit: amd64 and arm64)
- Linux (amd64 and arm64)
- macOS (amd64 and arm64)

Detection backends: `/proc/net` on Linux, `lsof` on macOS, and `netstat` on
Windows. Process details come from `/proc`, `ps`, and the Windows process APIs.
No administrator or root privileges are required for normal read-only
operation.

## Installation

### npm

Requires Node.js 18+ (used only to download the binary). The correct binary
for your platform is downloaded from the matching GitHub release and verified
against its SHA256 checksum.

```sh
npm install -g @richardflp/portmaster
```

The command is `portmaster`. Run `npm rebuild -g @richardflp/portmaster` to
re-download the binary.

### Go

```sh
go install github.com/RichardFlp/portmaster@latest
```

The binary is installed to `$(go env GOPATH)/bin/portmaster`. Make sure that
directory is on your `PATH`.

### Windows

- Download `portmaster-windows-amd64.exe` or `portmaster-windows-arm64.exe`
  from the [releases page](https://github.com/RichardFlp/portmaster/releases).
- Scoop: the manifest in `packaging/scoop/portmaster.json` can be added to a
  Scoop bucket. Update the version and SHA256 hashes from the release before
  publishing.

### macOS

- Download `portmaster-darwin-amd64` or `portmaster-darwin-arm64` from the
  [releases page](https://github.com/RichardFlp/portmaster/releases).
- Homebrew: the formula in `packaging/homebrew/portmaster.rb` can be used with
  a tap. Update the source tarball SHA256 from the release before publishing.

### Linux

- Install script (downloads the latest release and verifies the SHA256 checksum):

```sh
curl -fsSL https://raw.githubusercontent.com/RichardFlp/portmaster/main/scripts/install.sh | sh
```

- `.deb` and `.rpm` packages can be built from source with
  `scripts/build-deb.sh` and `scripts/build-rpm.sh` (requires
  [fpm](https://github.com/jordansissel/fpm)).
- Raw binaries are attached to each release.

## Quick start

```sh
portmaster            # list listening ports
portmaster 3000       # inspect port 3000
portmaster free 3000  # find the first free port at or above 3000
portmaster kill 3000  # terminate the process using port 3000 (asks first)
portmaster ui         # interactive terminal UI
```

## Command reference

### `portmaster` and `portmaster list`

Lists listening ports.

```sh
portmaster
portmaster list
portmaster list --process node --protocol tcp --port 3000
portmaster list --json
portmaster list --quiet
```

Options:

| Option              | Description                          |
| ------------------- | ------------------------------------ |
| `--process <name>`  | Filter by process name (substring)   |
| `--port <n>`        | Filter by exact port                 |
| `--protocol <p>`    | Filter by protocol (`tcp` or `udp`)  |
| `--json`            | Output JSON                          |
| `--quiet`           | Output port numbers only             |

### `portmaster <port>`

Inspects a port and shows the process using it.

```sh
portmaster 3000
portmaster 3000 --json
portmaster 3000 --open
```

```
Port 3000

Status:       LISTENING
Protocol:     TCP
Address:      127.0.0.1:3000

PID:          18432
Process:      node
Executable:   /usr/bin/node
Command:
  node server.js

Parent PID:   17321
Parent:       npm

Started:      2026-08-09 21:42:18
```

Only information that the operating system actually provides is displayed. If
the process cannot be inspected, for example because of permission
restrictions, PortMaster reports that and continues. When a port has both TCP
and UDP listeners, each is shown as a separate block.

### `portmaster status <port>`

```sh
portmaster status 3000
```

```
3000 -> LISTENING
```

A free port prints `AVAILABLE`. UDP listeners report `BOUND`. Pass `--json`
for machine-readable output.

### `portmaster free`

Finds available ports by probing both TCP and UDP.

```sh
portmaster free            # 5 free ports starting at 3000
portmaster free 8000       # the first free port at or above 8000
portmaster free --from 8000 --count 10
portmaster free --from 8000 --count 3 --json
```

Options: `--from <n>` sets the start port (default 3000), `--count <n>` sets
the number of results (default 5, configurable via `free_count`), `--json`,
and `--quiet`.

### `portmaster range <start>-<end>`

```sh
portmaster range 3000-4000          # occupied ports in the range
portmaster range 3000-4000 --free   # available ports in the range
portmaster range 3000-4000 --json
```

The range must satisfy `1 <= start <= end <= 65535`.

### `portmaster search <query>`

Matches ports, PIDs, process names, executables, and command lines. Matching
is case-insensitive.

```sh
portmaster search node
portmaster search 18432
portmaster search 3000
```

Options: `--json`, `--quiet`.

### `portmaster kill`

Terminates the process using a port, or a process by PID. Always shows what
will be terminated and asks for confirmation unless `--force` is given.

```sh
portmaster kill 3000
portmaster kill 3000 --force
portmaster kill --pid 18432
```

```
Port 3000 is used by:

PID:     18432
Process: node
Command: node server.js

Kill this process? [y/N]
```

Safety rules:

- Confirmation is required by default; piping input or running non-interactively
  without `--force` results in an aborted kill.
- PortMaster never terminates itself.
- The target is resolved from the listener's process information; a port with
  no owning process is never killed.

### `portmaster open <port>`

Opens a listening port in the default browser. If the process is bound to a
specific address, that address is used; wildcard listeners open
`http://localhost:<port>`. Port 443 uses `https`.

```sh
portmaster open 3000
portmaster 3000 --open
```

### `portmaster watch`

Continuously monitors listening ports and prints changes.

```sh
portmaster watch
portmaster watch --interval 5
portmaster watch --json
```

```
PORT   PID    PROCESS

3000   18432  node
5173   19281  node

Watching for changes... (Ctrl+C to quit)

+ 4200 TCP node PID 22014
- 8080 TCP python PID 20112
```

With `--json`, each event is a single line of JSON:

```json
{"event":"add","listener":{"port":4200,"protocol":"tcp","address":"127.0.0.1","pid":22014,"process":"node"}}
```

The initial state is emitted as `add` events in JSON mode. Ctrl+C stops the
watch with exit code 0. The default interval is 2 seconds.

### `portmaster ui`

Interactive terminal interface with keyboard navigation.

```
+--------------------------------------------------------------+
| PortMaster                                      14 ports     |
+--------+----------+--------+------------+-------------------+
| PORT   | PROTOCOL | PID    | PROCESS    | ADDRESS           |
+--------+----------+--------+------------+-------------------+
| 3000   | TCP      | 18432  | node       | 127.0.0.1         |
| 5173   | TCP      | 19281  | node       | 0.0.0.0           |
| 5432   | TCP      | 4212   | postgres   | 127.0.0.1         |
+--------+----------+--------+------------+-------------------+

Up/Down Navigate  Enter Inspect  K Kill  O Open  R Refresh  Q Quit
```

Keys:

| Key        | Action                       |
| ---------- | ---------------------------- |
| Up/Down    | Navigate the list            |
| Enter, I   | Inspect the selected port    |
| K          | Kill the owning process      |
| O          | Open the selected port       |
| R          | Refresh                      |
| Q, Ctrl+C  | Quit                         |
| Esc        | Back from inspection         |

Killing asks for confirmation (`y` to proceed, `n` to cancel). The UI handles
terminal resizing, works in small terminals, and restores the terminal state
on exit. Use `portmaster ui --interval 5` to change the refresh interval.

### `portmaster version` and `portmaster help`

Prints the version or the command summary.

## JSON output

The `--json` flag produces valid JSON on stdout with consistent field names
and no terminal formatting. Errors are written to stderr and never mixed into
the JSON.

`portmaster list --json`:

```json
[
  {
    "port": 3000,
    "protocol": "tcp",
    "address": "127.0.0.1",
    "pid": 18432,
    "process": "node"
  }
]
```

`portmaster 3000 --json`:

```json
{
  "port": 3000,
  "listeners": [
    {
      "status": "LISTENING",
      "protocol": "tcp",
      "address": "127.0.0.1",
      "pid": 18432,
      "process": "node",
      "executable": "/usr/bin/node",
      "command": "node server.js",
      "parent_pid": 17321,
      "parent": "npm",
      "started": "2026-08-09 21:42:18"
    }
  ]
}
```

Fields that are not available are omitted. `portmaster status --json` emits an
array, for example `[{"port":3000,"protocol":"tcp","status":"LISTENING"}]`.
`portmaster free --json` and `portmaster range <a-b> --free --json` emit
arrays of port numbers. `portmaster range <a-b> --json` emits an array of
listener objects like `list --json`.

## Quiet mode

`--quiet` emits only port numbers, one per line, for scripting.

```sh
portmaster list --quiet
```

```
3000
5173
8080
```

Available on `list`, `free`, `range`, and `search`. When both `--json` and
`--quiet` are given, `--json` wins.

## Exit codes

| Code | Meaning                                  |
| ---- | ---------------------------------------- |
| 0    | Success                                  |
| 1    | Runtime error                            |
| 2    | Invalid arguments or usage               |
| 3    | Requested port or process not found      |
| 4    | Permission denied                        |
| 5    | Action cancelled (for example declined kill) |

`portmaster watch` and `portmaster ui` exit with code 0 when stopped with
Ctrl+C.

## Configuration

Configuration is optional. PortMaster works with no configuration file. When
present, the file is `config.json` in the platform config directory:

- Windows: `%APPDATA%\portmaster\config.json`
- Linux: `~/.config/portmaster/config.json`
- macOS: `~/Library/Application Support/portmaster/config.json`

Example:

```json
{
  "refresh_interval": 2,
  "watch_interval": 2,
  "output_format": "table",
  "ignored_processes": ["com.docker.backend"],
  "ignored_ports": [5353],
  "browser": "",
  "free_count": 5
}
```

| Field               | Type     | Description                                          |
| ------------------- | -------- | ---------------------------------------------------- |
| `refresh_interval`  | int      | Seconds between `ui` refreshes (default 2)           |
| `watch_interval`    | int      | Seconds between `watch` scans (default 2)            |
| `output_format`     | string   | `table` or `json` (default `table`)                  |
| `ignored_processes` | []string | Process names hidden from list, watch, and ui        |
| `ignored_ports`     | []int    | Ports hidden from list, watch, and ui                |
| `browser`           | string   | Preferred browser command used by `open` and the UI  |
| `free_count`        | int      | Default number of ports returned by `free` (default 5) |

Ignored processes and ports are applied to `list`, `watch`, and `ui`.
`search` is explicit and always shows matching results.

## Building from source

Requires Go 1.25 or newer.

```sh
git clone https://github.com/RichardFlp/portmaster.git
cd portmaster
make build          # builds bin/portmaster
```

Or without make:

```sh
go build -o bin/portmaster ./cmd/portmaster
```

## Development

Project layout:

```text
cmd/portmaster       entry point
internal/cli         command dispatch and flags
internal/ports       port detection (Linux, Windows, macOS backends)
internal/processes   process lookup and termination
internal/browser     browser opening
internal/config      configuration
internal/output      table and JSON rendering
internal/inspect     shared port inspection views
internal/tui         interactive terminal UI
internal/version     version string
```

Platform-specific code lives in files tagged with the operating system and is
isolated behind platform-neutral functions. The rest of the codebase does not
contain operating-system assumptions.

## Testing

```sh
go test ./...
go vet ./...
gofmt -l .
```

Tests use dynamically allocated ports, so they do not depend on specific
ports being free. Platform-specific tests run only on their operating system.
Run with `go test -race ./...` to include the race detector.

## Release process

Releases are built from Git tags by the `Release` workflow
(`.github/workflows/release.yml`).

1. Update `internal/version/version.go` and `CHANGELOG.md`.
2. Commit and push to `main`.
3. Tag the release and push the tag:

```sh
git tag v0.2.0
git push origin v0.2.0
```

The workflow builds `portmaster-windows-amd64`, `portmaster-windows-arm64`,
`portmaster-linux-amd64`, `portmaster-linux-arm64`, `portmaster-darwin-amd64`,
and `portmaster-darwin-arm64`, generates `SHA256SUMS.txt`, and attaches
everything to a GitHub Release.

To build all release binaries locally:

```sh
make release
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
