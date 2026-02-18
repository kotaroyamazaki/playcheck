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
  local os arch tag version url archive_name tmp_dir

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
  tag="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"

  if [ -z "$tag" ]; then
    echo "Error: Could not determine latest version." >&2
    exit 1
  fi

  # Strip leading 'v' for the archive filename (GoReleaser uses version without v prefix)
  version="${tag#v}"
  echo "Latest version: ${version} (tag: ${tag})"

  archive_name="playcheck_${version}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${tag}/${archive_name}"

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

  # Find the playcheck binary (may be at top level or in a subdirectory)
  local binary
  binary="$(find "$tmp_dir" -name playcheck -type f -perm -u+x 2>/dev/null | head -1)"
  if [ -z "$binary" ]; then
    # Fallback: look for any file named playcheck (may not have execute bit in archive)
    binary="$(find "$tmp_dir" -name playcheck -type f 2>/dev/null | head -1)"
  fi
  if [ -z "$binary" ]; then
    echo "Error: playcheck binary not found in archive." >&2
    exit 1
  fi

  mkdir -p "$INSTALL_DIR"
  mv "$binary" "${INSTALL_DIR}/playcheck"
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
