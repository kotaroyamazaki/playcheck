#!/usr/bin/env bash
# Install playcheck binary from GitHub Releases.
# Supports Linux and macOS on amd64 and arm64.
set -euo pipefail

REPO="kotaroyamazaki/playcheck"
INSTALL_DIR="${HOME}/.local/bin"

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       echo "unsupported" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    arm64|aarch64)  echo "arm64" ;;
    *)              echo "unsupported" ;;
  esac
}

main() {
  local os arch version url archive_name tmp_dir

  os="$(detect_os)"
  arch="$(detect_arch)"

  if [ "$os" = "unsupported" ]; then
    echo "Error: Unsupported OS '$(uname -s)'. Only Linux and macOS are supported." >&2
    exit 1
  fi
  if [ "$arch" = "unsupported" ]; then
    echo "Error: Unsupported architecture '$(uname -m)'. Only amd64 and arm64 are supported." >&2
    exit 1
  fi

  echo "Detecting latest version..."
  version="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v?([^"]+)".*/\1/')"

  if [ -z "$version" ]; then
    echo "Error: Could not determine latest version." >&2
    exit 1
  fi
  echo "Latest version: ${version}"

  archive_name="playcheck_${version}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/v${version}/${archive_name}"

  echo "Downloading ${archive_name}..."
  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT

  if ! curl -fsSL -o "${tmp_dir}/${archive_name}" "$url"; then
    echo "Error: Failed to download from ${url}" >&2
    echo "Check that the release exists at https://github.com/${REPO}/releases" >&2
    exit 1
  fi

  echo "Extracting..."
  tar -xzf "${tmp_dir}/${archive_name}" -C "$tmp_dir"

  mkdir -p "$INSTALL_DIR"
  mv "${tmp_dir}/playcheck" "${INSTALL_DIR}/playcheck"
  chmod +x "${INSTALL_DIR}/playcheck"

  echo "Installed playcheck to ${INSTALL_DIR}/playcheck"

  if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
    echo ""
    echo "NOTE: ${INSTALL_DIR} is not in your PATH."
    echo "Add it with:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo "Or add the line above to your ~/.bashrc or ~/.zshrc"
  fi

  echo ""
  "${INSTALL_DIR}/playcheck" --version 2>/dev/null || echo "playcheck installed successfully"
}

main "$@"
