#!/bin/bash

# Configuration
REPO_OWNER="${REPO_OWNER:-Eutika}"      # Repository owner
REPO_NAME="${REPO_NAME:-eu-missions-cli}"        # Repository name
OUTPUT_DIR="${OUTPUT_DIR:-./}"     # Directory to save the binary

# Validate required environment variables
if [  [ -z "$REPO_OWNER" ] || [ -z "$REPO_NAME" ]; then
    echo "Error: Missing required environment variables"
    echo "Please set: REPO_OWNER, and REPO_NAME"
    exit 1
fi

# Detect OS and architecture for GoReleaser naming convention
detect_platform() {
    local os arch

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="Linux";;
        Darwin*)    os="Darwin";;
        MINGW*)     os="Windows";;
        *)          echo "Unsupported OS: $(uname -s)"; exit 1;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64*)    arch="x86_64";;
        aarch64*)   arch="arm64";;
        arm64*)     arch="arm64";;
        i386*)      arch="i386";;
        i686*)      arch="i386";;
        *)          echo "Unsupported architecture: $(uname -m)"; exit 1;;
    esac

    echo "${os}_${arch}"
}

# Get the latest release information
get_latest_release() {
    curl -sL \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
}

# Download and extract the appropriate archive
download_archive() {
    local platform="$1"
    local release_data="$2"
    local asset_url
    local asset_name

    # Parse release data to find the matching asset
    # Handle both .tar.gz (Linux/Darwin) and .zip (Windows) formats
    if [[ $platform == *"Windows"* ]]; then
        asset_url=$(echo "$release_data" | jq -r ".assets[] | select(.name | contains(\"${platform}\") and endswith(\".zip\")) | .url" | head -n1)
    else
        asset_url=$(echo "$release_data" | jq -r ".assets[] | select(.name | contains(\"${platform}\") and endswith(\".tar.gz\")) | .url" | head -n1)
    fi

    if [ -z "$asset_url" ]; then
        echo "Error: No matching archive found for platform: ${platform}"
        exit 1
    fi

    # Get the asset name
    asset_name=$(echo "$release_data" | jq -r ".assets[] | select(.url == \"${asset_url}\") | .name")
    local temp_dir=$(mktemp -d)
    local archive_path="${temp_dir}/${asset_name}"

    echo "Downloading ${asset_name}..."
    
    # Download the archive
    curl -sL \
        -H "Accept: application/octet-stream" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "$asset_url" \
        -o "$archive_path"

    echo "Extracting archive..."
    
    # Extract based on format
    if [[ $asset_name == *.tar.gz ]]; then
        tar xzf "$archive_path" -C "$temp_dir"
    elif [[ $asset_name == *.zip ]]; then
        unzip -q "$archive_path" -d "$temp_dir"
    else
        echo "Error: Unsupported archive format"
        rm -rf "$temp_dir"
        exit 1
    fi

    # Find and move the binary
    local binary_name="${REPO_NAME}"
    if [[ $platform == *"Windows"* ]]; then
        binary_name="${binary_name}.exe"
    fi

    # Find the binary in the extracted files
    local binary_path=$(find "$temp_dir" -type f -name "$binary_name")
    if [ -z "$binary_path" ]; then
        echo "Error: Binary not found in archive"
        rm -rf "$temp_dir"
        exit 1
    fi

    # Move binary to output directory
    mkdir -p "$OUTPUT_DIR"
    mv "$binary_path" "${OUTPUT_DIR}/${binary_name}"

    # Make binary executable on Unix-like systems
    if [ "$(uname -s)" != "MINGW"* ]; then
        chmod +x "${OUTPUT_DIR}/${binary_name}"
    fi

    # Cleanup
    rm -rf "$temp_dir"

    echo "Successfully extracted binary to: ${OUTPUT_DIR}/${binary_name}"
}

main() {
    # Create output directory if it doesn't exist
    OUTPUT_DIR="${HOME}/.local/bin"
    mkdir -p "$OUTPUT_DIR"

    # Detect platform
    platform=$(detect_platform)
    echo "Detected platform: ${platform}"

    # Get latest release
    echo "Fetching latest release information..."
    release_data=$(get_latest_release)

    # Check if release fetch was successful
    if [ "$(echo "$release_data" | jq -r '.message')" = "Not Found" ]; then
        echo "Error: Repository not found or no access. Check your credentials and repository name."
        exit 1
    fi

    # Download and extract archive
    download_archive "$platform" "$release_data"

    # Ensure the binary is named "missions"
    mv "${OUTPUT_DIR}/${REPO_NAME}" "${OUTPUT_DIR}/missions"

    # Add to PATH if not already present
    if [[ ":$PATH:" != *":${HOME}/.local/bin:"* ]]; then
        echo "Adding ${HOME}/.local/bin to PATH"
        echo "export PATH=\"\$HOME/.local/bin:\$PATH\"" >> "$HOME/.bashrc"
        echo "export PATH=\"\$HOME/.local/bin:\$PATH\"" >> "$HOME/.zshrc" 2>/dev/null
        echo "Please restart your terminal or run 'source ~/.bashrc' to update PATH"
    fi

    echo "Installed 'missions' CLI to ${OUTPUT_DIR}/missions"
}

main
