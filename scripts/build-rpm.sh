#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-0.1.0}"
ARCH="${ARCH:-x86_64}"

if ! command -v fpm >/dev/null 2>&1; then
  echo "fpm is required (install with: gem install fpm)" >&2
  exit 1
fi

if [ "$ARCH" != "x86_64" ] && [ "$ARCH" != "aarch64" ]; then
  echo "unsupported architecture: $ARCH" >&2
  exit 1
fi

case "$ARCH" in
  x86_64) goarch="amd64" ;;
  aarch64) goarch="arm64" ;;
esac

cd "$(dirname "$0")/.."
mkdir -p dist/pkg
CGO_ENABLED=0 GOOS=linux GOARCH="$goarch" \
  go build -trimpath \
  -ldflags "-s -w -X github.com/RichardFlp/portmaster/internal/version.Version=$VERSION" \
  -o dist/pkg/portmaster ./cmd/portmaster
mkdir -p dist/pkg/usr/share/doc/portmaster
install -m 0644 README.md dist/pkg/usr/share/doc/portmaster/README.md
install -m 0644 LICENSE dist/pkg/usr/share/doc/portmaster/LICENSE
fpm -s dir -t rpm -n portmaster -v "$VERSION" -a "$ARCH" \
  --description "Fast, cross-platform port and process management for developers" \
  --license MIT \
  --url "https://github.com/RichardFlp/portmaster" \
  -C dist/pkg .
rm -rf dist/pkg
