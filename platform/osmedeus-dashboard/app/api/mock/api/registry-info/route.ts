import { NextResponse } from "next/server";

export const dynamic = "force-static";

type RegistryMode = "direct-fetch" | "nix-build";

type DirectBinaryMeta = {
  desc?: string;
  tags?: string[];
  version?: string;
  linux?: Record<string, string>;
  darwin?: Record<string, string>;
  windows?: Record<string, string>;
  "command-linux"?: Record<string, string>;
  "command-darwin"?: Record<string, string>;
  installed?: boolean;
  path?: string;
  optional?: boolean;
};

type NixTool = {
  name: string;
  desc?: string;
  tags?: string[];
  version?: string;
  repo_link?: string;
  installed?: boolean;
  path?: string;
};

type NixCategory = {
  name: string;
  tools: NixTool[];
};

export let mockDirectBinaries: Record<string, DirectBinaryMeta> = {
  nuclei: {
    desc: "Vulnerability scanner",
    tags: ["vuln", "scanner"],
    version: "3.0.0",
    linux: {
      amd64: "https://github.com/projectdiscovery/nuclei/releases/download/v3.0.0/nuclei_3.0.0_linux_amd64.zip",
      arm64: "https://github.com/projectdiscovery/nuclei/releases/download/v3.0.0/nuclei_3.0.0_linux_arm64.zip",
    },
    darwin: {
      amd64: "https://github.com/projectdiscovery/nuclei/releases/download/v3.0.0/nuclei_3.0.0_darwin_amd64.zip",
      arm64: "https://github.com/projectdiscovery/nuclei/releases/download/v3.0.0/nuclei_3.0.0_darwin_arm64.zip",
    },
    installed: true,
    path: "/usr/local/bin/nuclei",
    optional: false,
  },
  amass: {
    desc: "In-depth attack surface mapping",
    tags: ["recon", "subdomain"],
    version: "4.0.0",
    linux: {
      amd64: "https://github.com/owasp-amass/amass/releases/download/v4.0.0/amass_linux_amd64.zip",
    },
    darwin: {
      amd64: "https://github.com/owasp-amass/amass/releases/download/v4.0.0/amass_darwin_amd64.zip",
    },
    installed: false,
    path: "",
    optional: false,
  },
  httpx: {
    desc: "Fast HTTP probing",
    tags: ["recon", "http"],
    version: "1.6.0",
    linux: {
      amd64: "https://github.com/projectdiscovery/httpx/releases/download/v1.6.0/httpx_1.6.0_linux_amd64.zip",
    },
    darwin: {
      amd64: "https://github.com/projectdiscovery/httpx/releases/download/v1.6.0/httpx_1.6.0_darwin_amd64.zip",
    },
    installed: true,
    path: "/usr/local/bin/httpx",
    optional: true,
  },
};

const mockRegistryUrl =
  "https://raw.githubusercontent.com/osmedeus/osmedeus-base/main/registry-metadata.json";

function buildNixCategories(): NixCategory[] {
  const categories: Record<string, NixTool[]> = {
    Subdomain: [],
    Vuln: [],
    HTTP: [],
    Other: [],
  };

  for (const [name, meta] of Object.entries(mockDirectBinaries)) {
    const tags = Array.isArray(meta.tags) ? meta.tags : [];
    const installed = Boolean(meta.installed);
    const tool: NixTool = {
      name,
      desc: meta.desc,
      tags,
      version: meta.version,
      repo_link:
        name === "nuclei"
          ? "https://github.com/projectdiscovery/nuclei"
          : name === "amass"
            ? "https://github.com/owasp-amass/amass"
            : name === "httpx"
              ? "https://github.com/projectdiscovery/httpx"
              : undefined,
      installed,
      path: installed ? `/home/user/.nix-profile/bin/${name}` : undefined,
    };

    const tagSet = new Set(tags.map((t) => t.toLowerCase()));
    if (tagSet.has("subdomain")) categories.Subdomain.push(tool);
    else if (tagSet.has("vuln")) categories.Vuln.push(tool);
    else if (tagSet.has("http")) categories.HTTP.push(tool);
    else categories.Other.push(tool);
  }

  return Object.entries(categories)
    .filter(([, tools]) => tools.length > 0)
    .map(([name, tools]) => ({
      name,
      tools: tools.sort((a, b) => a.name.localeCompare(b.name)),
    }));
}

export async function GET() {
  const registry_mode = "direct-fetch" as RegistryMode;

  if (registry_mode === "nix-build") {
    return NextResponse.json({
      registry_mode: "nix-build",
      nix_installed: true,
      categories: buildNixCategories(),
    });
  }

  return NextResponse.json({
    registry_mode: "direct-fetch",
    registry_url: mockRegistryUrl,
    binaries: mockDirectBinaries,
  });
}
