#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
VERSION="${DOCMESH_VERSION:-$(cd "$ROOT_DIR" && go run ./cmd/cli version | awk '{print $2}')}"
RELEASE_DIR="$ROOT_DIR/dist/install/releases"
SKILL_DIR="$ROOT_DIR/dist/install/skills"
TMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "$TMP_DIR"
}

trap cleanup EXIT

mkdir -p "$RELEASE_DIR" "$SKILL_DIR"
rm -f "$RELEASE_DIR"/* "$SKILL_DIR"/*

targets=(
  "darwin amd64"
  "darwin arm64"
  "linux amd64"
  "linux arm64"
  "windows amd64"
)

for target in "${targets[@]}"; do
  read -r os arch <<<"$target"
  stage="$TMP_DIR/${os}_${arch}"
  mkdir -p "$stage"

  binary_name="docmesh"
  if [[ "$os" == "windows" ]]; then
    binary_name="docmesh.exe"
  fi

  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -o "$stage/$binary_name" "$ROOT_DIR/cmd/cli"

  archive_base="docmesh_${os}_${arch}"
  if [[ "$os" == "windows" ]]; then
    (
      cd "$stage"
      zip -qr "$RELEASE_DIR/${archive_base}.zip" "$binary_name"
    )
  else
    (
      cd "$stage"
      tar -czf "$RELEASE_DIR/${archive_base}.tar.gz" "$binary_name"
    )
  fi
done

skill_stage="$TMP_DIR/skill"
mkdir -p "$skill_stage"
cp -R "$ROOT_DIR/skills/docmesh" "$skill_stage/docmesh"
(
  cd "$skill_stage"
  zip -qr "$SKILL_DIR/DocMesh.zip" docmesh
)
cp "$SKILL_DIR/DocMesh.zip" "$SKILL_DIR/DocMesh.skill"

(
  cd "$ROOT_DIR/dist/install"
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 releases/* skills/* > checksums.txt
  else
    sha256sum releases/* skills/* > checksums.txt
  fi
  printf '%s\n' "$VERSION" > version.txt
)
