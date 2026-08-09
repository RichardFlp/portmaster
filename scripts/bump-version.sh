#!/usr/bin/env bash
# Bump PortMaster version files to <version> (x.y.z or vx.y.z).
# Used by the Release workflow; also runnable locally.
set -eu

version="${1:-}"
if [ -z "$version" ]; then
  echo "usage: $0 <version>   (e.g. $0 1.2.0 or $0 v1.2.0)" >&2
  exit 1
fi
version="${version#v}"
if ! printf '%s' "$version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "invalid version: $version (expected x.y.z)" >&2
  exit 1
fi

root="$(cd "$(dirname "$0")/.." && pwd)"
if command -v python3 >/dev/null 2>&1; then
  PY=python3
else
  PY=python
fi
export VERSION="$version"
export DATE="$(date -u +%Y-%m-%d)"
export ROOT="$root"

"$PY" - <<'PY'
import json, os, re

version = os.environ["VERSION"]
date = os.environ["DATE"]
root = os.environ["ROOT"]

# internal/version/version.go
p = os.path.join(root, "internal", "version", "version.go")
text = open(p).read()
text = re.sub(r'var Version = "[^"]*"', f'var Version = "{version}"', text, count=1)
open(p, "w").write(text)

# Makefile
p = os.path.join(root, "Makefile")
text = open(p).read()
text = re.sub(r"^VERSION \?= .*$", f"VERSION ?= {version}", text, count=1, flags=re.M)
open(p, "w").write(text)

# npm/package.json
p = os.path.join(root, "npm", "package.json")
data = json.load(open(p))
data["version"] = version
with open(p, "w") as f:
    json.dump(data, f, indent=2)
    f.write("\n")

# CHANGELOG.md: rename [Unreleased] if present, otherwise insert an entry
# above the most recent release.
p = os.path.join(root, "CHANGELOG.md")
text = open(p).read()
entry = (
    f"## [{version}] - {date}\n\n"
    f"- See the [GitHub release]"
    f"(https://github.com/RichardFlp/portmaster/releases/tag/v{version}) for details.\n\n"
)
unrel = re.search(r"^## \[Unreleased\]\s*$", text, re.M)
if unrel:
    text = text[: unrel.start()] + f"## [{version}] - {date}" + text[unrel.end() :]
else:
    first = re.search(r"^## \[", text, re.M)
    if first:
        text = text[: first.start()] + entry + text[first.start() :]
    else:
        text = text.rstrip() + "\n\n" + entry
open(p, "w").write(text)
PY

if ! grep -q "var Version = \"$version\"" "$root/internal/version/version.go"; then
  echo "failed to bump internal/version/version.go" >&2
  exit 1
fi
if ! grep -q "^VERSION ?= $version$" "$root/Makefile"; then
  echo "failed to bump Makefile" >&2
  exit 1
fi

echo "bumped version files to $version"
