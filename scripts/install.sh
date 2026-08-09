#!/usr/bin/env bash
set -eu
# pipefail is a bash/zsh extension; enable it only where supported so the
# script also runs under POSIX sh (e.g. dash on Debian/Ubuntu).
if [ -n "${BASH_VERSION:-}" ] || [ -n "${ZSH_VERSION:-}" ]; then
  set -o pipefail
fi

REPO="RichardFlp/portmaster"
BIN="/usr/local/bin/portmaster"
VERSION="${VERSION:-latest}"

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) echo "amd64" ;;
    aarch64 | arm64) echo "arm64" ;;
    *)
      echo "unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

release_url() {
  local file="$1"
  if [ "$VERSION" = "latest" ]; then
    echo "https://github.com/$REPO/releases/latest/download/$file"
  else
    echo "https://github.com/$REPO/releases/download/$VERSION/$file"
  fi
}

main() {
  if [ "$(uname -s)" != "Linux" ]; then
    echo "install.sh supports Linux only" >&2
    exit 1
  fi
  arch=$(detect_arch)
  binary="portmaster-linux-$arch"
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT
  echo "Downloading $binary"
  curl -fsSL -o "$tmp/$binary" "$(release_url "$binary")"
  curl -fsSL -o "$tmp/SHA256SUMS.txt" "$(release_url "SHA256SUMS.txt")"
  (cd "$tmp" && sha256sum -c --ignore-missing SHA256SUMS.txt) >/dev/null
  chmod +x "$tmp/$binary"
  if [ -w /usr/local/bin ]; then
    install -m 0755 "$tmp/$binary" "$BIN"
  else
    sudo install -m 0755 "$tmp/$binary" "$BIN"
  fi
  echo "Installed portmaster to $BIN"
}

main "$@"
