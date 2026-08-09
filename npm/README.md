# @richard.flp/portmaster

npm distribution of [PortMaster](https://github.com/RichardFlp/portmaster), a
fast, cross-platform CLI for managing ports and processes.

## Install

Requires Node.js 18+ (any Node version is fine; it is only used to download
the binary). The correct binary for your platform is downloaded from the
matching [GitHub release](https://github.com/RichardFlp/portmaster/releases)
during install and verified against its SHA256 checksum.

```sh
npm install -g @richard.flp/portmaster
```

## Usage

```sh
portmaster             # list listening ports
portmaster 3000        # inspect port 3000
portmaster free 3000   # find a free port at or above 3000
portmaster kill 3000   # terminate the process using port 3000
portmaster ui          # interactive terminal UI
```

Run `portmaster help` for the full command reference.

## Notes

- Supported platforms: Windows, macOS, Linux on amd64 and arm64.
- The package itself contains no compiled code; the binary is downloaded at
  install time. To re-download, run `npm rebuild -g @richard.flp/portmaster`.
- In restricted networks, set `PORTMASTER_BINARY_MIRROR` to a mirror of the
  GitHub releases before installing.
- `portmaster-cli` is installed as an alias for compatibility with the
  previous package name.
- Same exit codes and behavior as the native binary. See the
  [main README](https://github.com/RichardFlp/portmaster) for details.
