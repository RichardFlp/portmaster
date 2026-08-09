# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-08-10

### Added

- npm distribution under `@richard.flp/portmaster` with a `portmaster` command
  (the previous `portmaster-cli` name is installed as an alias).
- Auto-publish of the npm package from the Release workflow.

### Changed

- Release binaries and npm package now report version 1.1.0.

## [1.0.0] - 2026-08-10

### Fixed

- IPv6 address parsing for `/proc/net` listeners on Linux.
- macOS detection when `lsof` reports no matches instead of failing the scan.
- A race in the Windows kill test.
- CI failures across platforms, and added a race-detector job.

## [0.1.0] - 2026-08-09

### Added

- `portmaster` with no arguments lists listening TCP and UDP ports.
- `portmaster <port>` inspects a port and the process using it.
- `portmaster list` with `--process`, `--port`, and `--protocol` filters.
- `portmaster status <port>` reports LISTENING, BOUND, or AVAILABLE.
- `portmaster free [start]` finds available ports with `--from` and `--count`.
- `portmaster range <start>-<end>` lists occupied ports, with `--free` for available ones.
- `portmaster search <query>` matches ports, PIDs, process names, executables, and command lines.
- `portmaster kill <port>` and `portmaster kill --pid <n>` with confirmation and `--force`.
- `portmaster open <port>` opens a listening port in the default browser.
- `portmaster watch` monitors port changes with a configurable interval.
- `portmaster ui` provides an interactive terminal interface.
- JSON output via `--json` on list, inspect, status, free, range, search, and watch.
- Quiet mode via `--quiet` on list, free, range, and search.
- Optional JSON configuration under the per-user config directory.
- Linux detection via `/proc/net`, macOS via `lsof`, Windows via `netstat`.
- Process details including executable path, command line, parent, and start time where available.
- Protection against terminating PortMaster itself and against unconfirmed kills.
- Exit codes documented for scripting.
- Tests covering port parsing, detection, filters, free ports, ranges, JSON, kill flows, and CLI behavior.
- CI workflows for Windows, Linux, and macOS.
- Release workflow producing binaries for windows, linux, and darwin on amd64 and arm64 with SHA256 checksums.
