export type RegistryMode = "direct-fetch" | "nix-build";

export interface DirectFetchBinaryMeta {
  desc?: string;
  tags?: string[];
  version?: string;
  linux?: Record<string, string>;
  darwin?: Record<string, string>;
  windows?: Record<string, string>;
  ["command-linux"]?: Record<string, string>;
  ["command-darwin"]?: Record<string, string>;
  installed?: boolean;
  path?: string;
  optional?: boolean;
}

export interface RegistryInfoDirectFetch {
  registry_mode: "direct-fetch";
  registry_url: string;
  binaries: Record<string, DirectFetchBinaryMeta>;
}

export interface NixTool {
  name: string;
  desc?: string;
  tags?: string[];
  version?: string;
  repo_link?: string;
  installed?: boolean;
  path?: string;
}

export interface NixCategory {
  name: string;
  tools: NixTool[];
}

export interface RegistryInfoNixBuild {
  registry_mode: "nix-build";
  nix_installed?: boolean;
  categories: NixCategory[];
}

export type RegistryMetadata = RegistryInfoDirectFetch | RegistryInfoNixBuild;

export interface BinaryInstallResponse {
  message: string;
  registry_mode?: RegistryMode;
  installed: string[];
  installed_count: number;
  binaries_folder?: string;
  failed: Array<{ name: string; error?: string }>;
  failed_count: number;
}

export interface WorkflowInstallResponse {
  message: string;
  source: string;
  workflow_folder?: string;
}
