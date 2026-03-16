#!/bin/sh
# op-setup installer
# Usage: curl -sSL https://raw.githubusercontent.com/MiguelAguiarDEV/op-setup/main/install.sh | sh
set -e

REPO="MiguelAguiarDEV/op-setup"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY="op-setup"

# --- Detect OS ---
detect_os() {
  os="$(uname -s)"
  case "$os" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       echo "unsupported"; return 1 ;;
  esac
}

# --- Detect architecture ---
detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)             echo "unsupported"; return 1 ;;
  esac
}

# --- Get latest version from GitHub API ---
get_latest_version() {
  if command -v curl >/dev/null 2>&1; then
    curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/'
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/'
  else
    echo "Error: curl or wget required" >&2
    exit 1
  fi
}

# --- Download and install ---
main() {
  os="$(detect_os)"
  arch="$(detect_arch)"

  if [ "$os" = "unsupported" ] || [ "$arch" = "unsupported" ]; then
    echo "Error: unsupported platform $(uname -s)/$(uname -m)" >&2
    exit 1
  fi

  version="${VERSION:-$(get_latest_version)}"
  if [ -z "$version" ]; then
    echo "Error: could not determine latest version. Set VERSION=x.y.z manually." >&2
    exit 1
  fi

  filename="${BINARY}_${version}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/v${version}/${filename}"

  echo "Installing op-setup v${version} (${os}/${arch})..."

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  echo "  Downloading ${url}..."
  if command -v curl >/dev/null 2>&1; then
    curl -sSL "$url" -o "${tmpdir}/${filename}"
  else
    wget -q "$url" -O "${tmpdir}/${filename}"
  fi

  echo "  Extracting..."
  tar -xzf "${tmpdir}/${filename}" -C "$tmpdir"

  echo "  Installing to ${INSTALL_DIR}/${BINARY}..."
  if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR" 2>/dev/null || sudo mkdir -p "$INSTALL_DIR"
  fi
  if [ -w "$INSTALL_DIR" ]; then
    mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    chmod +x "${INSTALL_DIR}/${BINARY}"
  else
    sudo mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY}"
  fi

  echo ""
  echo "op-setup v${version} installed to ${INSTALL_DIR}/${BINARY}"
  echo ""
  echo "Run it:"
  echo "  op-setup                                          # Interactive TUI"
  echo "  op-setup --dry-run                                # Preview without changes"
  echo "  op-setup --no-interactive --profile full --dry-run # Headless dry-run"
}

main
