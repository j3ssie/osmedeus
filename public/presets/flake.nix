{
  description = "Osmedeus security tools - all binaries for security automation workflows";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };

        # Subdomain enumeration and discovery tools
        subdomainTools = with pkgs; [
          amass            # In-depth attack surface mapping and asset discovery
          subfinder        # Fast passive subdomain enumeration tool
          findomain        # Cross-platform subdomain enumerator
          shuffledns       # Wrapper around massdns for active bruteforcing
          massdns          # High-performance DNS stub resolver
        ];

        # URL and endpoint discovery tools
        urlTools = with pkgs; [
          urlfinder        # Fast URL finder for web archives
          katana           # Next-generation crawling and spidering framework
        ];

        # HTTP and web tools
        httpTools = with pkgs; [
          httpx            # Fast and multi-purpose HTTP toolkit
          ffuf             # Fast web fuzzer written in Go
        ];

        # Port scanning tools
        portScanTools = with pkgs; [
          naabu            # Fast port scanner with probing capabilities
          rustscan         # Modern port scanner
          nmap             # Network exploration and security auditing
        ];

        # DNS tools
        dnsTools = with pkgs; [
          dnsx             # Fast DNS toolkit
          tlsx             # Fast TLS grabber
          alterx           # Fast and customizable subdomain wordlist generator
        ];

        # Vulnerability scanning tools
        vulnTools = with pkgs; [
          nuclei           # Fast, customizable vulnerability scanner
          trufflehog       # Secret detection and scanning tool
          semgrep          # Static analysis tool for code scanning
          metabigor        # OSINT tool for passive reconnaissance
        ];

        # GitHub reconnaissance tools
        githubTools = with pkgs; [
          github-endpoints    # Find GitHub endpoints for a domain
          github-subdomains   # Find subdomains from GitHub
        ];

        # Utility tools
        utilityTools = with pkgs; [
          git              # Distributed version control system
          curl             # Command line tool for transferring data
          wget             # Network downloader
          unzip            # Extraction utility for zip archives
          coreutils        # Basic file, shell and text manipulation utilities
          rsync            # Fast incremental file transfer utility
          # ripgrep          # Fast regex search tool (rg)
          jq               # Lightweight command-line JSON processor
          minio-client     # MinIO client for S3-compatible object storage
        ];

        # Development tools
        devTools = with pkgs; [
          bun              # Fast JavaScript runtime, bundler, and package manager
          uv               # Extremely fast Python package and project manager
        ];

        # All tools combined
        allTools = subdomainTools ++ urlTools ++ httpTools ++ portScanTools
                   ++ dnsTools ++ vulnTools ++ githubTools ++ utilityTools ++ devTools;

      in {
        # Development shell with all tools
        devShells.default = pkgs.mkShell {
          name = "osmedeus-tools";
          buildInputs = allTools;

          shellHook = ''
            echo "🔒 Osmedeus Security Tools Environment"
            echo "======================================="
            echo ""
            echo "Subdomain Discovery:"
            echo "  amass, subfinder, findomain, shuffledns, massdns"
            echo ""
            echo "URL/Endpoint Discovery:"
            echo "  urlfinder, katana"
            echo ""
            echo "HTTP/Web Tools:"
            echo "  httpx, ffuf"
            echo ""
            echo "Port Scanning:"
            echo "  naabu, rustscan, nmap"
            echo ""
            echo "DNS Tools:"
            echo "  dnsx, tlsx, alterx"
            echo ""
            echo "Vulnerability Scanning:"
            echo "  nuclei, trufflehog, semgrep, metabigor"
            echo ""
            echo "GitHub Recon:"
            echo "  github-endpoints, github-subdomains"
            echo ""
            echo "Utilities:"
            echo "  git, curl, wget, unzip, coreutils, rsync, jq, mc (minio-client), bun, uv"
            echo ""
          '';
        };

        # Packages for individual installation
        packages = {
          default = pkgs.symlinkJoin {
            name = "osmedeus-tools";
            paths = allTools;
          };

          # Individual tool groups
          subdomain = pkgs.symlinkJoin {
            name = "osmedeus-subdomain-tools";
            paths = subdomainTools;
          };

          http = pkgs.symlinkJoin {
            name = "osmedeus-http-tools";
            paths = httpTools;
          };

          portscan = pkgs.symlinkJoin {
            name = "osmedeus-portscan-tools";
            paths = portScanTools;
          };

          dns = pkgs.symlinkJoin {
            name = "osmedeus-dns-tools";
            paths = dnsTools;
          };

          vuln = pkgs.symlinkJoin {
            name = "osmedeus-vuln-tools";
            paths = vulnTools;
          };

          utils = pkgs.symlinkJoin {
            name = "osmedeus-utility-tools";
            paths = utilityTools;
          };
        };
      }
    );
}
