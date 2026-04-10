#!/bin/sh
set -eu

RELEASE_REPO="${LLM_WIKI_RELEASE_REPO:-iFurySt/LLM-Wiki}"
VERSION="${LLM_WIKI_VERSION:-latest}"
DEFAULT_DOWNLOAD_BASE_URL="{{LLM_WIKI_DOWNLOAD_BASE_URL}}"
DOWNLOAD_BASE_URL="${LLM_WIKI_DOWNLOAD_BASE_URL:-$DEFAULT_DOWNLOAD_BASE_URL}"
INSTALL_DIR="${LLM_WIKI_INSTALL_DIR:-$HOME/.local/bin}"
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
  DOWNLOADER="curl"
elif command -v wget >/dev/null 2>&1; then
  DOWNLOADER="wget"
else
  echo "missing downloader: need curl or wget" >&2
  exit 1
fi

fetch_to_file() {
  url="$1"
  output="$2"
  if [ "$DOWNLOADER" = "curl" ]; then
    curl -fsSL "$url" -o "$output"
  else
    wget -qO "$output" "$url"
  fi
}

download_archive() {
  primary_url="$1"
  fallback_url="$2"
  output="$3"

  if fetch_to_file "$primary_url" "$output"; then
    return 0
  fi

  if [ -n "$fallback_url" ] && [ "$fallback_url" != "$primary_url" ]; then
    echo "primary download failed, falling back to ${fallback_url}" >&2
    fetch_to_file "$fallback_url" "$output"
    return 0
  fi

  return 1
}

legacy_archive_name() {
  archive="$1"
  case "$archive" in
    llm-wiki_*)
      suffix="${archive#llm-wiki_}"
      printf '%s\n' "docmesh_${suffix}"
      ;;
    *)
      printf '%s\n' "$archive"
      ;;
  esac
}

fetch_text() {
  url="$1"
  if [ "$DOWNLOADER" = "curl" ]; then
    curl -fsSL "$url"
  else
    wget -qO- "$url"
  fi
}

extract_version() {
  binary="$1"
  if [ ! -x "$binary" ]; then
    return 1
  fi
  "$binary" version 2>/dev/null | awk 'NF >= 2 { print $2; exit }'
}

resolve_target_version() {
  if [ "$VERSION" != "latest" ]; then
    printf '%s\n' "$VERSION"
    return 0
  fi
  if [ -n "$DOWNLOAD_BASE_URL" ]; then
    printf '%s\n' "latest"
    return 0
  fi

  release_json="$(fetch_text "https://api.github.com/repos/${RELEASE_REPO}/releases/latest")"
  target_version="$(printf '%s\n' "$release_json" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  if [ -z "$target_version" ]; then
    echo "failed to resolve latest release version" >&2
    exit 1
  fi
  printf '%s\n' "$target_version"
}

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

ARCHIVE="llm-wiki_${OS}_${ARCH}.tar.gz"
TARGET_VERSION="$(resolve_target_version)"

BIN_TARGET="$INSTALL_DIR/llm-wiki"
ALIAS_TARGET="$INSTALL_DIR/lw"

CURRENT_VERSION=""
if [ -x "$BIN_TARGET" ]; then
  CURRENT_VERSION="$(extract_version "$BIN_TARGET" || true)"
elif command -v llm-wiki >/dev/null 2>&1; then
  CURRENT_VERSION="$(extract_version "$(command -v llm-wiki)" || true)"
fi

if [ -n "$CURRENT_VERSION" ] && [ "$TARGET_VERSION" != "latest" ] && [ "$CURRENT_VERSION" = "$TARGET_VERSION" ]; then
  echo "llm-wiki ${CURRENT_VERSION} is already installed"
  echo "no update needed"
  exit 0
fi

if [ -n "$DOWNLOAD_BASE_URL" ]; then
  ARCHIVE_URL="${DOWNLOAD_BASE_URL%/}/${ARCHIVE}"
else
  ARCHIVE_URL=""
fi

if [ "$TARGET_VERSION" = "latest" ]; then
  GITHUB_ARCHIVE_URL="https://github.com/${RELEASE_REPO}/releases/latest/download/${ARCHIVE}"
else
  GITHUB_ARCHIVE_URL="https://github.com/${RELEASE_REPO}/releases/download/${TARGET_VERSION}/${ARCHIVE}"
fi

LEGACY_ARCHIVE="$(legacy_archive_name "$ARCHIVE")"
if [ "$LEGACY_ARCHIVE" = "$ARCHIVE" ]; then
  LEGACY_GITHUB_ARCHIVE_URL=""
elif [ "$TARGET_VERSION" = "latest" ]; then
  LEGACY_GITHUB_ARCHIVE_URL="https://github.com/${RELEASE_REPO}/releases/latest/download/${LEGACY_ARCHIVE}"
else
  LEGACY_GITHUB_ARCHIVE_URL="https://github.com/${RELEASE_REPO}/releases/download/${TARGET_VERSION}/${LEGACY_ARCHIVE}"
fi

if [ -n "$ARCHIVE_URL" ]; then
  echo "downloading ${ARCHIVE_URL}"
else
  echo "downloading ${GITHUB_ARCHIVE_URL}"
fi
if ! download_archive "${ARCHIVE_URL:-$GITHUB_ARCHIVE_URL}" "$GITHUB_ARCHIVE_URL" "$TMP_DIR/$ARCHIVE"; then
  if [ -n "$LEGACY_GITHUB_ARCHIVE_URL" ]; then
    echo "modern archive unavailable, falling back to ${LEGACY_GITHUB_ARCHIVE_URL}" >&2
    fetch_to_file "$LEGACY_GITHUB_ARCHIVE_URL" "$TMP_DIR/$ARCHIVE"
  else
    exit 1
  fi
fi

mkdir -p "$INSTALL_DIR"
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

BIN_SOURCE="$TMP_DIR/llm-wiki"

if [ ! -f "$BIN_SOURCE" ]; then
  echo "archive did not contain llm-wiki binary" >&2
  exit 1
fi

cp "$BIN_SOURCE" "$BIN_TARGET"
chmod +x "$BIN_TARGET"

if command -v ln >/dev/null 2>&1; then
  rm -f "$ALIAS_TARGET"
  ln -s "$BIN_TARGET" "$ALIAS_TARGET" 2>/dev/null || cp "$BIN_TARGET" "$ALIAS_TARGET"
else
  cp "$BIN_TARGET" "$ALIAS_TARGET"
fi
chmod +x "$ALIAS_TARGET"

INSTALLED_VERSION="$(extract_version "$BIN_TARGET" || true)"
if [ -n "$CURRENT_VERSION" ]; then
  echo "updated llm-wiki from ${CURRENT_VERSION} to ${INSTALLED_VERSION:-unknown}"
else
  echo "installed llm-wiki ${INSTALLED_VERSION:-unknown} to $BIN_TARGET"
fi
echo "installed lw alias to $ALIAS_TARGET"
echo "run: $BIN_TARGET version"
echo "or:  $ALIAS_TARGET version"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo "warning: $INSTALL_DIR is not in PATH" >&2
    echo "run directly with: $BIN_TARGET" >&2
    ;;
esac
