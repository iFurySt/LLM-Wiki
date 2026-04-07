#!/bin/sh
set -eu

RELEASE_REPO="${DOCMESH_RELEASE_REPO:-iFurySt/DocMesh}"
VERSION="${DOCMESH_VERSION:-latest}"
DOWNLOAD_BASE_URL="${DOCMESH_DOWNLOAD_BASE_URL:-}"
INSTALL_DIR="${DOCMESH_INSTALL_DIR:-$HOME/.local/bin}"
TMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "$TMP_DIR"
}

trap cleanup EXIT INT TERM

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

need_cmd uname
need_cmd mktemp
need_cmd tar

if command -v curl >/dev/null 2>&1; then
  FETCH="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
  FETCH="wget -qO-"
else
  echo "missing downloader: need curl or wget" >&2
  exit 1
fi

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  darwin) OS="darwin" ;;
  linux) OS="linux" ;;
  *)
    echo "unsupported operating system: $OS" >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

ARCHIVE="docmesh_${VERSION}_${OS}_${ARCH}.tar.gz"
ARCHIVE="docmesh_${OS}_${ARCH}.tar.gz"

if [ -n "$DOWNLOAD_BASE_URL" ]; then
  ARCHIVE_URL="${DOWNLOAD_BASE_URL%/}/${ARCHIVE}"
elif [ "$VERSION" = "latest" ]; then
  ARCHIVE_URL="https://github.com/${RELEASE_REPO}/releases/latest/download/${ARCHIVE}"
else
  ARCHIVE_URL="https://github.com/${RELEASE_REPO}/releases/download/${VERSION}/${ARCHIVE}"
fi

echo "downloading ${ARCHIVE_URL}"
$FETCH "$ARCHIVE_URL" > "$TMP_DIR/$ARCHIVE"

mkdir -p "$INSTALL_DIR"
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

BIN_SOURCE="$TMP_DIR/docmesh"
BIN_TARGET="$INSTALL_DIR/docmesh"

if [ ! -f "$BIN_SOURCE" ]; then
  echo "archive did not contain docmesh binary" >&2
  exit 1
fi

cp "$BIN_SOURCE" "$BIN_TARGET"
chmod +x "$BIN_TARGET"

echo "installed docmesh to $BIN_TARGET"
echo "run: $BIN_TARGET version"
