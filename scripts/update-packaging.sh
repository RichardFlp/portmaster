#!/usr/bin/env bash
# Update the Homebrew and Scoop packaging manifests to <version> using the
# checksums from the just-created GitHub release. Used by the Release workflow.
set -eu

version="${1:-}"
if [ -z "$version" ]; then
  echo "usage: $0 <version>   (e.g. $0 1.2.0 or $0 v1.2.0)" >&2
  exit 1
fi
version="${version#v}"

root="$(cd "$(dirname "$0")/.." && pwd)"
base="https://github.com/RichardFlp/portmaster/releases/download/v$version"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

curl -fsSL -o "$tmp/SHA256SUMS.txt" "$base/SHA256SUMS.txt"
win64="$(awk '$2 == "portmaster-windows-amd64.exe" {print $1}' "$tmp/SHA256SUMS.txt")"
winarm="$(awk '$2 == "portmaster-windows-arm64.exe" {print $1}' "$tmp/SHA256SUMS.txt")"
if [ -z "$win64" ] || [ -z "$winarm" ]; then
  echo "could not find windows hashes in SHA256SUMS.txt" >&2
  exit 1
fi

curl -fsSL -o "$tmp/src.tar.gz" "https://github.com/RichardFlp/portmaster/archive/refs/tags/v$version.tar.gz"
srcsha="$(sha256sum "$tmp/src.tar.gz" | awk '{print $1}')"

if command -v python3 >/dev/null 2>&1; then
  PY=python3
else
  PY=python
fi
export VERSION="$version" WIN64="$win64" WINARM="$winarm" SRCSHA="$srcsha" ROOT="$root"

"$PY" - <<'PY'
import json, os, re

version = os.environ["VERSION"]
root = os.environ["ROOT"]
win64 = os.environ["WIN64"]
winarm = os.environ["WINARM"]
srcsha = os.environ["SRCSHA"]

# Scoop manifest
p = os.path.join(root, "packaging", "scoop", "portmaster.json")
data = json.load(open(p))
data["version"] = version
data["architecture"]["64bit"]["url"] = (
    f"https://github.com/RichardFlp/portmaster/releases/download/v{version}/portmaster-windows-amd64.exe"
)
data["architecture"]["64bit"]["hash"] = win64
data["architecture"]["arm64"]["url"] = (
    f"https://github.com/RichardFlp/portmaster/releases/download/v{version}/portmaster-windows-arm64.exe"
)
data["architecture"]["arm64"]["hash"] = winarm
with open(p, "w") as f:
    json.dump(data, f, indent=2)
    f.write("\n")

# Homebrew formula
p = os.path.join(root, "packaging", "homebrew", "portmaster.rb")
text = open(p).read()
text = re.sub(
    r"https://github\.com/RichardFlp/portmaster/archive/refs/tags/v[0-9.]*\.tar\.gz",
    f"https://github.com/RichardFlp/portmaster/archive/refs/tags/v{version}.tar.gz",
    text,
)
text = re.sub(
    r'^  sha256 "[a-f0-9]*"',
    f'  sha256 "{srcsha}"',
    text,
    count=1,
    flags=re.M,
)
open(p, "w").write(text)
PY

echo "updated packaging manifests to $version"
