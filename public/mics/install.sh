#!/usr/bin/env bash
set -euo pipefail

# Osmedeus CLI Installation Script
# Downloads pre-compiled Osmedeus CLI binary from GitHub releases

# Configuration
OSM_HOME="${OSM_HOME:-$HOME/.osmedeus}"
BIN_DIR="$HOME/.local/bin"
GITHUB_REPO="j3ssie/osmedeus"
GITHUB_RELEASES="https://github.com/${GITHUB_REPO}/releases"
FALLBACK_VERSION="v5.0.0-beta"
OSM_URL_ENV_SET=0
if [[ -n "${OSM_URL+x}" ]]; then
	OSM_URL_ENV_SET=1
fi
OSM_URL="${OSM_URL:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
LIGHT_GREEN='\033[1;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Cleanup on interrupt
cleanup() {
	echo -e "\n${YELLOW}Installation interrupted...${NC}"
	rm -f "$OSM_HOME/osm-install-"* 2>/dev/null || true
	exit 1
}

trap cleanup INT TERM

log() {
	echo -e "${BLUE}[INFO]${NC} $1" >&2
}

warn() {
	echo -e "${YELLOW}[WARN]${NC} $1" >&2
}

error() {
	echo -e "${RED}[ERROR]${NC} $1" >&2
	exit 1
}

success() {
	echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
}

# Check if command exists
command_exists() {
	command -v "$1" >/dev/null 2>&1
}

# Require command to exist or exit with error
need_cmd() {
	if ! command_exists "$1"; then
		error "need '$1' (command not found)"
	fi
}

# Check all prerequisite commands upfront
check_prereqs() {
	for cmd in uname mktemp chmod mkdir rm mv tar grep awk cut head sed basename touch; do
		need_cmd "$cmd"
	done

	# Check for sha256 checksum command (shasum on macOS/BSD, sha256sum on Linux)
	if command_exists shasum; then
		SHA256_CMD="shasum -a 256"
	elif command_exists sha256sum; then
		SHA256_CMD="sha256sum"
	else
		error "need 'shasum' or 'sha256sum' (command not found)"
	fi
}

# Detect target platform for CLI binary
detect_platform() {
	local platform
	platform="$(uname -s) $(uname -m)"

	case $platform in
		'Darwin x86_64')
			target=darwin_amd64
			;;
		'Darwin arm64')
			target=darwin_arm64
			;;
		'Linux aarch64' | 'Linux arm64')
			target=linux_arm64
			;;
		'Linux riscv64')
			error 'Not supported on riscv64'
			;;
		'Linux x86_64' | *)
			target=linux_amd64
			;;
	esac

	# Check for Rosetta 2 on macOS
	if [[ "$target" == "darwin_amd64" ]]; then
		if [[ $(sysctl -n sysctl.proc_translated 2>/dev/null) = 1 ]]; then
			target=darwin_arm64
			log "Your shell is running in Rosetta 2. Using $target instead"
		fi
	fi

	echo "$target"
}

# Robust downloader that handles snap curl issues
downloader() {
	local url="$1"
	local output_file="$2"

	# Check if we have a broken snap curl
	local snap_curl=0
	if command_exists curl; then
		local curl_path
		curl_path=$(command -v curl)
		if [[ "$curl_path" == *"/snap/"* ]]; then
			snap_curl=1
		fi
	fi

	# Check if we have a working (non-snap) curl
	if command_exists curl && [[ $snap_curl -eq 0 ]]; then
		curl -fsSL "$url" -o "$output_file"
	# Try wget for both no curl and the broken snap curl
	elif command_exists wget; then
		wget -q --show-progress "$url" -O "$output_file"
	# If we can't fall back from broken snap curl to wget, report the broken snap curl
	elif [[ $snap_curl -eq 1 ]]; then
		error "curl installed with snap cannot download files due to missing permissions. Please uninstall it and reinstall curl with a different package manager (e.g., apt)."
	else
		error "Neither curl nor wget found. Please install one of them."
	fi
}

# Download file with progress
download_file() {
	local url="$1"
	local output_file="$2"
	local version="${3:-}"

	if [[ -n "$version" ]]; then
		log "Downloading $(basename "$output_file") (${LIGHT_GREEN}${version}${NC})..."
	else
		log "Downloading $(basename "$output_file")..."
	fi

	# Use secure temporary file
	local temp_file
	temp_file=$(mktemp "$(dirname "$output_file")/tmp.XXXXXX")

	# Download to temp file first, then atomic move
	downloader "$url" "$temp_file"
	mv "$temp_file" "$output_file"
}

# Verify SHA256 checksum
verify_checksum() {
	local file="$1"
	local expected_checksum="$2"

	log "Verifying checksum..."

	local actual_checksum
	actual_checksum=$($SHA256_CMD "$file" | cut -d' ' -f1)

	if [[ "$actual_checksum" != "$expected_checksum" ]]; then
		error "Checksum verification failed!\nExpected: $expected_checksum\nActual: $actual_checksum"
	fi

	success "Checksum verified"
}

# Fetch latest CLI version from GitHub API with fallback
fetch_latest_version() {
    local api_url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    local version
    local tmp_file

    log "Fetching latest version from GitHub..."
    tmp_file=$(mktemp)

    # Try to fetch from GitHub API
    if downloader "$api_url" "$tmp_file" 2>/dev/null; then
        version=$(grep '"tag_name":' "$tmp_file" | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/')
        rm -f "$tmp_file"

        if [[ -n "$version" ]]; then
            echo "$version"
            return
        fi
    fi

    rm -f "$tmp_file" 2>/dev/null

    # Fall back to hardcoded version
    warn "Failed to fetch from GitHub API, using fallback: $FALLBACK_VERSION"
    echo "$FALLBACK_VERSION"
}

fetch_latest_version_from_metadata() {
	local metadata_url="${OSM_URL%/}/metadata.json"
	local version
	local tmp_file

	log "Fetching latest version from ${metadata_url}..."
	tmp_file=$(mktemp)
	downloader "$metadata_url" "$tmp_file"
	version=$(grep -E '"version"\s*:' "$tmp_file" | head -n 1 | sed -E 's/.*"version"\s*:\s*"([^\"]+)".*/\1/')
	rm -f "$tmp_file"

	if [[ -z "$version" ]]; then
		error "Failed to fetch latest version from ${metadata_url}"
	fi

	echo "$version"
}

# Check for existing osmedeus installation
check_existing_installation() {
	local binary_path="$BIN_DIR/osmedeus"
	local existing_binary=""

	# Check in BIN_DIR first
	if [[ -x "$binary_path" ]]; then
		existing_binary="$binary_path"
	# Also check if osmedeus is in PATH (might be installed elsewhere)
	elif command_exists osmedeus; then
		existing_binary=$(command -v osmedeus)
	fi

	if [[ -n "$existing_binary" ]]; then
		local old_version=""
		local old_build=""
		# Try to get the current version
		old_version=$("$existing_binary" version 2>/dev/null | grep 'Version:' || echo "")
		old_build=$("$existing_binary" version 2>/dev/null | grep 'Build:' || echo "")

		if [[ -n "$old_version" && -n "$old_build" ]]; then
			warn "Detected existing osmedeus installation at $existing_binary (${old_version} - ${old_build})"
		else
			warn "Detected existing osmedeus installation at $existing_binary"
		fi
		log "Will replace with the new version..."
	fi
}

# Install Osmedeus CLI binary
install_osmedeus_binary() {
	local platform="$1"
	local binary_name="osmedeus"

	# Check for existing installation before proceeding
	check_existing_installation

	local version
	version="${OSM_VERSION:-}"
	if [[ -z "$version" ]]; then
		if [[ $OSM_URL_ENV_SET -eq 1 && -n "${OSM_URL}" ]]; then
			version=$(fetch_latest_version_from_metadata)
		else
			version=$(fetch_latest_version)
		fi
	fi
	if [[ "$version" != v* ]]; then
		version="v${version}"
	fi
	log "Installing version: ${LIGHT_GREEN}${version}${NC}"

	# Strip 'v' prefix for tarball filename (e.g., v5.0.0 -> 5.0.0)
	local version_no_v="${version#v}"
	local tarball_name="osmedeus_${version_no_v}_${platform}.tar.gz"
	local base_url
	if [[ $OSM_URL_ENV_SET -eq 1 && -n "${OSM_URL}" ]]; then
		base_url="${OSM_URL%/}"
	else
		base_url="https://github.com/${GITHUB_REPO}/releases/download/${version}"
		OSM_URL="$base_url"
	fi
	local tarball_url="${base_url}/${tarball_name}"
	local checksum_url="${base_url}/checksums.txt"

	local tarball_path="$OSM_HOME/osm-install-tarball.tar.gz"
	local checksum_path="$OSM_HOME/osm-install-checksums.txt"
	local extract_dir="$OSM_HOME/osm-install-extract"

	# Ensure directories exist
	mkdir -p "$OSM_HOME"
	mkdir -p "$BIN_DIR"
	mkdir -p "$extract_dir"

	# Download checksum first
	download_file "$checksum_url" "$checksum_path" "$version"

	# Extract expected checksum for our tarball
	local expected_checksum
	expected_checksum=$(grep "$tarball_name" "$checksum_path" | awk '{print $1}')

	if [[ -z "$expected_checksum" ]]; then
		error "Could not find checksum for $tarball_name in checksums file"
	fi

	# Download tarball
	download_file "$tarball_url" "$tarball_path" "$version"

	# Verify checksum
	verify_checksum "$tarball_path" "$expected_checksum"

	# Extract tarball
	log "Extracting tarball..."
	tar -xzf "$tarball_path" -C "$extract_dir"

	# Move binary to BIN_DIR
	local binary_path="$BIN_DIR/$binary_name"
	mv "$extract_dir/$binary_name" "$binary_path"

	# Make executable
	chmod +x "$binary_path"

	# Clean up
	rm -f "$tarball_path" "$checksum_path"
	rm -rf "$extract_dir"

	success "Osmedeus CLI binary installed to $binary_path"
}

# Update PATH in shell profile
update_shell_profile() {
	# Detect shell from $SHELL or default
	local default_shell="bash"
	if [[ "$(uname -s)" == "Darwin" ]]; then
		default_shell="zsh"
	fi

	local shell_name
	shell_name=$(basename "${SHELL:-$default_shell}")

	local shell_profiles=()
	local refresh_command=""

	case "$shell_name" in
		zsh)
			shell_profiles=("$HOME/.zshrc")
			refresh_command="exec \$SHELL"
			;;
		bash)
			# Add to both .bashrc (interactive) and .bash_profile (login shells)
			[[ -f "$HOME/.bashrc" ]] && shell_profiles+=("$HOME/.bashrc")
			[[ -f "$HOME/.bash_profile" ]] && shell_profiles+=("$HOME/.bash_profile")
			# If neither exists, create .bashrc
			[[ ${#shell_profiles[@]} -eq 0 ]] && shell_profiles=("$HOME/.bashrc")
			refresh_command="source ~/.bashrc"
			;;
		fish)
			shell_profiles=("$HOME/.config/fish/config.fish")
			refresh_command="source ~/.config/fish/config.fish"
			;;
		*)
			warn "Unknown shell: $shell_name"
			warn "Please add $BIN_DIR to your PATH manually:"
			echo "  export PATH=\"$BIN_DIR:\$PATH\""
			return
			;;
	esac

	local updated=0
	for shell_profile in "${shell_profiles[@]}"; do
		# Check if PATH is already updated
		if [[ -f "$shell_profile" ]] && grep -q "$BIN_DIR" "$shell_profile" 2>/dev/null; then
			log "PATH already configured in $shell_profile"
			continue
		fi

		# Create config file if it doesn't exist
		if [[ ! -f "$shell_profile" ]]; then
			mkdir -p "$(dirname "$shell_profile")"
			touch "$shell_profile"
		fi

		# Add to PATH
		{
			echo ""
			echo "# Osmedeus CLI"
			echo "export PATH=\"$BIN_DIR:\$PATH\""
		} >> "$shell_profile"

		success "Added $BIN_DIR to PATH in $shell_profile"
		updated=1
	done

	if [[ $updated -eq 1 ]]; then
		echo ""
		log "To activate the PATH, run:"
		echo -e "  ${LIGHT_GREEN}${refresh_command}${NC}"
	fi
}

# Main installation
main() {
	log "Starting Osmedeus CLI installation..."

	# Check prerequisites
	check_prereqs

	# Detect platform
	local platform
	platform=$(detect_platform)
	log "Detected platform: $platform"

	# Install binary
	install_osmedeus_binary "$platform"

	# Update shell profile
	update_shell_profile

	echo ""
	success "Osmedeus CLI installed successfully!"
	log "Run ${LIGHT_GREEN}osmedeus health${NC} (after restarting your shell) to validate your setup and generate a sample config"
	log "Visit ${LIGHT_GREEN}https://docs.osmedeus.org${NC} for documentation"
	log "Run ${LIGHT_GREEN}osmedeus install base --preset${NC} to download the ready-to-use workflow and then start scanning"
}

main "$@"
