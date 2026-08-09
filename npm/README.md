# portmaster-cli

npm distribution of [PortMaster](https://github.com/RichardFlp/portmaster), a
fast, cross-platform CLI for managing ports and processes.

## Install

Requires Node.js 18+ (any Node version is fine; it is only used to download
the binary). The correct binary for your platform is downloaded from the
matching [GitHub release](https://github.com/RichardFlp/portmaster/releases)
during install and verified against its SHA256 checksum.

```sh
npm install -g portmaster-cli
```

## Usage

```sh
portmaster-cli             # list listening ports
portmaster-cli 3000        # inspect port 3000
portmaster-cli free 3000   # find a free port at or above 3000
portmaster-cli kill 3000   # terminate the process using port 3000
portmaster-cli ui          # interactive terminal UI
```

Run `portmaster-cli help` for the full command reference.

## Notes

- Supported platforms: Windows, macOS, Linux on amd64 and arm64.
- The package itself contains no compiled code; the binary is downloaded at
  install time. To re-download, run `npm rebuild -g portmaster-cli`.
- In restricted networks, set `PORTMASTER_BINARY_MIRROR` to a mirror of the
  GitHub releases before installing.
- Same exit codes and behavior as the native binary. See the
  [main README](https://github.com/RichardFlp/portmaster) for details.
