export interface Workspace {
  id: number;
  name: string;
  data_source?: string;
  local_path: string;
  total_assets: number;
  total_subdomains: number;
  total_urls: number;
  total_vulns: number;
  vuln_critical: number;
  vuln_high: number;
  vuln_medium: number;
  vuln_low: number;
  vuln_potential: number;
  risk_score: number;
  tags: string[];
  last_run: string;
  run_workflow: string;
  state_execution_log?: string;
  state_completed_file?: string;
  state_workflow_file?: string;
  state_workflow_folder?: string;
  created_at: string;
  updated_at: string;
}

export interface HttpAsset {
  id: string;
  workspace: string;
  assetValue: string;
  url: string;
  input: string;
  scheme: string;
  method: string;
  path: string;
  statusCode: number;
  contentType: string;
  contentLength: number;
  title?: string;
  words: number;
  lines: number;
  hostIp?: string;
  aRecords: string[];
  tls?: string;
  assetType: string;
  technologies: string[];
  responseTime?: string;
  remarks?: string | string[];
  source: string;
  createdAt: Date;
  updatedAt: Date;
  lastSeenAt?: Date;
}

export interface Subdomain {
  id: string;
  workspaceId: string;
  domain: string;
  ip?: string;
  isAlive: boolean;
  createdAt: Date;
}

export interface Vulnerability {
  id: string;
  workspaceId: string;
  title: string;
  severity: "critical" | "high" | "medium" | "low" | "info";
  target: string;
  description?: string;
  createdAt: Date;
}

export interface DashboardStats {
  workflows: number;
  workspaces: number;
  subdomains: number;
  httpAssets: number;
  vulnerabilities: number;
}

export interface AssetStats {
  technologies: string[];
  sources: string[];
  remarks: string[];
  assetTypes: string[];
}

// Filter types for assets table
export interface HttpAssetFilters {
  search?: string;
  statusCodes?: number[];
  minContentLength?: number;
  maxContentLength?: number;
  location?: string;
  technologies?: string[];
  assetTypes?: string[];
  sources?: string[];
  remarks?: string[];
  contentTypes?: string[];
  tlsVersion?: string;
}

// Sorting types for assets table
export type AssetSortField =
  | "url"
  | "statusCode"
  | "contentLength"
  | "title"
  | "hostIp"
  | "remarks"
  | "technologies"
  | "responseTime";

export type SortDirection = "asc" | "desc";

export interface AssetSortState {
  field: AssetSortField | null;
  direction: SortDirection;
}

// Technology categories for badge coloring
export type TechCategory =
  | "webServer"
  | "frontend"
  | "backend"
  | "database"
  | "cms"
  | "cdn"
  | "security"
  | "language"
  | "other";

// Sorting types for workspaces table
export type WorkspaceSortField =
  | "name"
  | "total_assets"
  | "total_subdomains"
  | "total_urls"
  | "total_vulns"
  | "risk_score"
  | "last_run"
  | "state_files"
  | "actions";

export interface WorkspaceSortState {
  field: WorkspaceSortField | null;
  direction: SortDirection;
}
