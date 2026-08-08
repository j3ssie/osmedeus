import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";
import type {
  RegistryMetadata,
  RegistryMode,
  BinaryInstallResponse,
  WorkflowInstallResponse,
} from "@/lib/types/registry";

export async function getRegistryMetadata(params?: {
  registry_mode?: RegistryMode;
}): Promise<RegistryMetadata> {
  const registry_mode = params?.registry_mode ?? "direct-fetch";
  if (isDemoMode()) {
    if (registry_mode === "nix-build") {
      return {
        registry_mode: "nix-build",
        nix_installed: true,
        categories: [
          {
            name: "Subdomain",
            tools: [
              {
                name: "amass",
                desc: "In-depth attack surface mapping and asset discovery",
                tags: ["recon", "subdomain"],
                version: "4.2.0",
                repo_link: "https://github.com/owasp-amass/amass",
                installed: true,
                path: "/home/user/.nix-profile/bin/amass",
              },
              {
                name: "subfinder",
                desc: "Fast passive subdomain enumeration tool",
                tags: ["recon", "subdomain"],
                version: "2.6.0",
                installed: false,
              },
            ],
          },
          {
            name: "Vuln",
            tools: [
              {
                name: "nuclei",
                desc: "Fast, customizable vulnerability scanner",
                tags: ["vuln", "scanner"],
                version: "3.0.0",
                installed: true,
                path: "/home/user/.nix-profile/bin/nuclei",
                repo_link: "https://github.com/projectdiscovery/nuclei",
              },
            ],
          },
        ],
      };
    }

    return {
      registry_mode: "direct-fetch",
      registry_url:
        "https://raw.githubusercontent.com/osmedeus/osmedeus-base/main/registry-metadata.json",
      binaries: {
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
      },
    };
  }

  const res = await http.get(`${API_PREFIX}/registry-info`, {
    params: { registry_mode },
  });
  return (res.data || {}) as RegistryMetadata;
}

export async function installBinaries(params: {
  names?: string[];
  install_all?: boolean;
  install_optional?: boolean;
  registry_url?: string;
  registry_mode?: RegistryMode;
}): Promise<BinaryInstallResponse> {
  if (isDemoMode()) {
    const names = Array.isArray(params.names) ? params.names : [];
    return {
      message: "Binary installation completed",
      registry_mode: params.registry_mode ?? "direct-fetch",
      installed: names,
      installed_count: names.length,
      binaries_folder: "/home/user/osmedeus-base/binaries",
      failed: [],
      failed_count: 0,
    };
  }
  const body: Record<string, any> = {
    type: "binary",
  };
  if (Array.isArray(params.names) && params.names.length > 0) {
    body.names = params.names;
  }
  if (params.install_all) {
    body.install_all = true;
  }
  if (params.install_optional) {
    body.install_optional = true;
  }
  if (params.registry_url) {
    body.registry_url = params.registry_url;
  }
  if (params.registry_mode) {
    body.registry_mode = params.registry_mode;
  }
  const res = await http.post(`${API_PREFIX}/registry-install`, body);
  return res.data as BinaryInstallResponse;
}

export async function installWorkflow(source: string): Promise<WorkflowInstallResponse> {
  if (isDemoMode()) {
    return {
      message: "Workflow installed successfully",
      source,
      workflow_folder: "/home/user/osmedeus-base/workflow",
    };
  }
  const body = { type: "workflow", source };
  const res = await http.post(`${API_PREFIX}/registry-install`, body);
  return res.data as WorkflowInstallResponse;
}
