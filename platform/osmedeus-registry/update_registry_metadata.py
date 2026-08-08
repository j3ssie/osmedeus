#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import argparse
import json
import os
import re
import requests
from pathlib import Path
from datetime import datetime, timezone

# Map tool names to GitHub repos
REPO_WITH_ARTIFACTS = {
    "amass": "owasp-amass/amass",
    "httprobe": "tomnomnom/httprobe",
    "subfinder": "projectdiscovery/subfinder",
    "nuclei": "projectdiscovery/nuclei",
    "httpx": "projectdiscovery/httpx",
    "katana": "projectdiscovery/katana",
    "dnsx": "projectdiscovery/dnsx",
    "naabu": "projectdiscovery/naabu",
    # "gotator": "Josue87/gotator",
    "puredns": "d3mondev/puredns",
    "ffuf": "ffuf/ffuf",
    "trivy": "aquasecurity/trivy",
    "legba": "evilsocket/legba",
    "dalfox": "hahwul/dalfox",
    "gosec": "securego/gosec",
    "gau": "lc/gau",
    "bearer": "bearer/bearer",
    "interactsh": "projectdiscovery/interactsh",
    "cariddi": "edoardottt/cariddi",
    "cloudfox": "BishopFox/cloudfox",
    "jsluice": "BishopFox/jsluice",
    "urlhunter": "utkusen/urlhunter",
    "kingfisher": "mongodb/kingfisher",
    "metabigor": "j3ssie/metabigor",
}

def get_github_token(env_var="GH_TOKEN"):
    """Get GitHub token from environment variable."""
    token = os.environ.get(env_var)
    if token:
        print("[+] Using GitHub token from {}".format(env_var))
    return token

def get_github_repo_info(repo, token=None):
    """Get latest version, description, and asset URLs from GitHub API."""
    print("[*] Getting info for {}".format(repo))
    headers = {}
    if token:
        headers["Authorization"] = "token {}".format(token)
    
    # Get latest release info
    response = requests.get(
        "https://api.github.com/repos/{}/releases/latest".format(repo),
        headers=headers
    )
    response.raise_for_status()
    data = response.json()
    version = data.get("tag_name", "") or data.get("name", "")
    print("[+] Latest version: {}".format(version))
    
    # Get repository description
    repo_response = requests.get(
        "https://api.github.com/repos/{}".format(repo),
        headers=headers
    )
    repo_response.raise_for_status()
    repo_data = repo_response.json()
    description = repo_data.get("description", "")
    if description:
        print("[+] Description: {}".format(description))
    
    # Get asset URLs from release
    assets = {}
    if "assets" in data:
        for asset in data["assets"]:
            assets[asset["name"]] = asset.get("browser_download_url")
    
    repo_link = "https://github.com/{}".format(repo)
    return version, description, assets, repo_link

def get_os_arch_from_filename(filename):
    """Extract OS and architecture from filename."""
    filename_lower = filename.lower()
    
    # Detect OS
    os_name = None
    if "darwin" in filename_lower or "macos" in filename_lower:
        os_name = "darwin"
    elif "linux" in filename_lower:
        os_name = "linux"
    elif "windows" in filename_lower or "win" in filename_lower:
        os_name = "windows"
    
    # Detect architecture
    arch = None
    if "amd64" in filename_lower or "x86_64" in filename_lower or "x64" in filename_lower:
        arch = "amd64"
    elif "arm64" in filename_lower or "aarch64" in filename_lower:
        arch = "arm64"
    elif "armv7" in filename_lower or "armv7l" in filename_lower:
        arch = "armv7"
    elif "armv6" in filename_lower or "armv6l" in filename_lower:
        arch = "armv6"
    elif "386" in filename_lower or "i386" in filename_lower or "x86" in filename_lower:
        arch = "386"
    
    return os_name, arch

def is_probably_download_artifact(filename):
    filename_lower = filename.lower()
    if any(
        marker in filename_lower
        for marker in (
            "checksum",
            "checksums",
            "sha256",
            "sha512",
            "sbom",
            ".sig",
            ".asc",
            "license",
            "readme",
        )
    ):
        return False

    if filename_lower.endswith((".zip", ".tar.gz", ".tgz", ".tar", ".gz", ".bz2", ".7z")):
        return True

    if filename_lower.endswith((".txt", ".md")):
        return False

    return "." not in Path(filename).name

def update_url(url, tool, new_version):
    """Update version in download URL."""
    # Extract version without 'v' prefix for filename
    version_num = new_version.lstrip("v")

    # Replace version in download path (e.g., /download/v1.0.0/ or /download/1.0.0/)
    url = re.sub(
        r"(/releases/download/)(v?[\d.]+(?:-[a-zA-Z0-9.-]+)?)",
        r"\g<1>{}".format(new_version),
        url
    )

    # Replace version in filename for standard patterns: tool_version_os_arch
    # Handles: subfinder_2.11.0_linux_amd64.zip, gau_2.2.4_linux_amd64.tar.gz
    url = re.sub(
        r"({}_)([\d.]+)(_)".format(re.escape(tool)),
        r"\g<1>{}\g<3>".format(version_num),
        url
    )

    # Handle patterns with 'v' in filename: tool_vX.X.X_os_arch
    # Handles: metabigor_v2.0.1_linux_amd64.tar.gz, gospider_v1.1.6_linux_x86_64.zip
    url = re.sub(
        r"({}_v)([\d.]+)(_)".format(re.escape(tool)),
        r"\g<1>{}\g<3>".format(version_num),
        url
    )

    # Handle amass pattern: amass_Linux_amd64.zip (version only in path)
    # Already handled by path replacement above

    # Handle trufflehog/gitleaks pattern: tool_version_os_arch (no 'v')
    url = re.sub(
        r"(trufflehog_)([\d.]+)(_)",
        r"\g<1>{}\g<3>".format(version_num),
        url
    )
    url = re.sub(
        r"(gitleaks_)([\d.]+)(_)",
        r"\g<1>{}\g<3>".format(version_num),
        url
    )

    # Handle httprobe pattern: httprobe-linux-amd64-0.2.tgz
    url = re.sub(
        r"(httprobe-[a-z]+-[a-z0-9]+-)([\d.]+)(\.tgz)",
        r"\g<1>{}\g<3>".format(version_num),
        url
    )


    return url

def main():
    parser = argparse.ArgumentParser(description="Update binary registry with latest GitHub versions")
    parser.add_argument(
        "--token-env",
        default="GH_TOKEN",
        help="Environment variable name for GitHub token (default: GH_TOKEN)"
    )
    parser.add_argument(
        "--tool",
        help="Update only a specific tool (default: all tools)"
    )
    args = parser.parse_args()

    # Get GitHub token
    token = get_github_token(args.token_env)

    script_dir = Path(__file__).parent
    json_path = script_dir / "registry-metadata-direct-fetch.json"

    # Load JSON
    with open(json_path, "r") as f:
        data = json.load(f)

    # Determine which tools to update
    tools_to_update = {args.tool: REPO_WITH_ARTIFACTS[args.tool]} if args.tool else REPO_WITH_ARTIFACTS

    # Update each tool
    for tool, repo in tools_to_update.items():
        if tool not in data:
            print("[!] Tool {} not found in registry, skipping".format(tool))
            continue

        try:
            new_version, description, assets, repo_link = get_github_repo_info(repo, token)
        except Exception as e:
            print("[!] Failed to get version for {}: {}".format(tool, e))
            continue

        # Update description if available
        if description:
            data[tool]["desc"] = description
            print("[*] Updated {}/desc: {}".format(tool, description))

        # Update repo_link
        if repo_link:
            data[tool]["repo_link"] = repo_link
            print("[*] Updated {}/repo_link: {}".format(tool, repo_link))

        # Update version field (strip 'v' prefix for consistency)
        if new_version:
            version_num = new_version.lstrip("v")
            data[tool]["version"] = version_num
            print("[*] Updated {}/version: {}".format(tool, version_num))

        # Update URLs using browser_download_url from assets
        for os_key, archs in data[tool].items():
            if os_key not in ("linux", "darwin", "windows"):
                continue
            if not isinstance(archs, dict):
                if isinstance(archs, str) and archs:
                    new_url = update_url(archs, tool, new_version)
                    data[tool][os_key] = new_url
                    print("[*] Updated {}/{} (fallback): {}".format(tool, os_key, new_url))
                continue

            for arch, url in archs.items():
                # Try to find matching asset
                found = False
                for filename, download_url in assets.items():
                    if download_url:
                        if not is_probably_download_artifact(filename):
                            continue
                        asset_os, asset_arch = get_os_arch_from_filename(filename)
                        if asset_os == os_key and asset_arch == arch:
                            data[tool][os_key][arch] = download_url
                            print("[*] Updated {}/{}/{}: {}".format(tool, os_key, arch, download_url))
                            found = True
                            break
                
                # Fallback to regex substitution if no direct match found
                if not found:
                    new_url = update_url(url, tool, new_version)
                    data[tool][os_key][arch] = new_url
                    print("[*] Updated {}/{}/{} (fallback): {}".format(tool, os_key, arch, new_url))

    # Add timestamp
    data["_last_update_at"] = datetime.now(timezone.utc).isoformat()
    
    # Save JSON
    with open(json_path, "w") as f:
        json.dump(data, f, indent=2)

    print("\n[+] Updated {} at {}".format(json_path, data["_last_update_at"]))

if __name__ == "__main__":
    main()
