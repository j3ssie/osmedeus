import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import {
  TECH_CATEGORY_MAP,
  CATEGORY_BADGE_VARIANT,
  CATEGORY_DOT_COLOR,
} from "@/lib/constants/tech-categories";
import type { TechCategory, HttpAsset, AssetSortField, SortDirection, Workspace, WorkspaceSortField } from "@/lib/types/asset";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: Date | string): string {
  const d = typeof date === "string" ? new Date(date) : date;
  return d.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function formatDateTime(date: Date | string): string {
  const d = typeof date === "string" ? new Date(date) : date;
  return d.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export function formatDuration(seconds: number): string {
  if (seconds < 60) {
    return `${seconds}s`;
  }
  if (seconds < 3600) {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return secs > 0 ? `${mins}m ${secs}s` : `${mins}m`;
  }
  const hours = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export function formatNumber(num: number): string {
  if (num >= 1000000) {
    return `${(num / 1000000).toFixed(1)}M`;
  }
  if (num >= 1000) {
    return `${(num / 1000).toFixed(1)}K`;
  }
  return num.toString();
}

export function timeAgo(date: Date | string): string {
  const d = typeof date === "string" ? new Date(date) : date;
  const now = new Date();
  const seconds = Math.floor((now.getTime() - d.getTime()) / 1000);

  if (seconds < 60) return "just now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)}d ago`;
  return formatDate(d);
}

export function truncate(value: unknown, length: number): string {
  const text = typeof value === "string" ? value : value == null ? "" : String(value);
  if (text.length <= length) return text;
  return text.slice(0, length) + "...";
}

export function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// Tech badge utilities
export function getTechCategory(tech: string): TechCategory {
  const normalized = tech.toLowerCase();
  // Check for exact match first
  if (TECH_CATEGORY_MAP[normalized]) {
    return TECH_CATEGORY_MAP[normalized];
  }
  // Check for partial match (e.g., "nginx/1.21.0" should match "nginx")
  for (const [key, category] of Object.entries(TECH_CATEGORY_MAP)) {
    if (normalized.includes(key) || key.includes(normalized)) {
      return category as TechCategory;
    }
  }
  return "other";
}

export function getTechBadgeVariant(tech: string): string {
  const category = getTechCategory(tech);
  return CATEGORY_BADGE_VARIANT[category];
}

export function getTechDotColor(tech: string): string {
  return CATEGORY_DOT_COLOR[getTechCategory(tech)];
}

export function groupTechnologiesByCategory(
  technologies: string[]
): Record<TechCategory, string[]> {
  const groups: Record<TechCategory, string[]> = {
    webServer: [],
    frontend: [],
    backend: [],
    database: [],
    cms: [],
    cdn: [],
    security: [],
    language: [],
    other: [],
  };

  technologies.forEach((tech) => {
    const category = getTechCategory(tech);
    groups[category].push(tech);
  });

  return groups;
}

// Parse response time string to milliseconds for sorting
export function parseResponseTime(time?: string): number {
  if (!time) return Infinity;
  const match = time.match(/(\d+(?:\.\d+)?)\s*(ms|s)/i);
  if (!match) return Infinity;
  const [, value, unit] = match;
  return unit.toLowerCase() === "s"
    ? parseFloat(value) * 1000
    : parseFloat(value);
}

// Client-side sorting for assets
export function sortAssets(
  assets: HttpAsset[],
  field: AssetSortField | null,
  direction: SortDirection
): HttpAsset[] {
  if (!field) return assets;

  return [...assets].sort((a, b) => {
    let comparison = 0;

    switch (field) {
      case "url":
        comparison = a.url.localeCompare(b.url);
        break;
      case "statusCode":
        comparison = a.statusCode - b.statusCode;
        break;
      case "contentLength":
        comparison = a.contentLength - b.contentLength;
        break;
      case "title":
        comparison = (a.title ?? "").localeCompare(b.title ?? "");
        break;
      case "hostIp":
        comparison = (a.hostIp ?? "").localeCompare(b.hostIp ?? "");
        break;
      case "technologies":
        comparison = a.technologies.length - b.technologies.length;
        break;
      case "responseTime":
        comparison =
          parseResponseTime(a.responseTime) - parseResponseTime(b.responseTime);
        break;
    }

    return direction === "asc" ? comparison : -comparison;
  });
}

// Client-side sorting for workspaces
export function sortWorkspaces(
  workspaces: Workspace[],
  field: WorkspaceSortField | null,
  direction: SortDirection
): Workspace[] {
  if (!field) return workspaces;

  return [...workspaces].sort((a, b) => {
    let comparison = 0;

    switch (field) {
      case "name":
        comparison = a.name.localeCompare(b.name);
        break;
      case "total_assets":
        comparison = a.total_assets - b.total_assets;
        break;
      case "total_subdomains":
        comparison = a.total_subdomains - b.total_subdomains;
        break;
      case "total_urls":
        comparison = a.total_urls - b.total_urls;
        break;
      case "total_vulns":
        comparison = a.total_vulns - b.total_vulns;
        break;
      case "risk_score":
        comparison = a.risk_score - b.risk_score;
        break;
      case "last_run":
        const dateA = a.last_run ? new Date(a.last_run).getTime() : 0;
        const dateB = b.last_run ? new Date(b.last_run).getTime() : 0;
        comparison = dateA - dateB;
        break;
      case "state_files": {
        const countStateFiles = (ws: Workspace) =>
          Number(Boolean(ws.state_execution_log)) +
          Number(Boolean(ws.state_completed_file)) +
          Number(Boolean(ws.state_workflow_file)) +
          Number(Boolean(ws.state_workflow_folder));
        comparison = countStateFiles(a) - countStateFiles(b);
        break;
      }
      case "actions":
        comparison = a.name.localeCompare(b.name);
        break;
    }

    return direction === "asc" ? comparison : -comparison;
  });
}

// HTTP status-code → badge variant. Single source of truth shared by the assets
// table, the asset detail dialog, and the status summary badges so the bands
// never drift apart.
export type StatusBadgeVariant =
  | "success"
  | "info"
  | "warning"
  | "destructive"
  | "secondary";

export const STATUS_BANDS: Array<{
  label: string;
  variant: Exclude<StatusBadgeVariant, "secondary">;
  min: number;
  max: number;
}> = [
  { label: "2xx", variant: "success", min: 200, max: 300 },
  { label: "3xx", variant: "info", min: 300, max: 400 },
  { label: "4xx", variant: "warning", min: 400, max: 500 },
  { label: "5xx", variant: "destructive", min: 500, max: 600 },
];

export function getStatusBadgeVariant(statusCode: number): StatusBadgeVariant {
  return (
    STATUS_BANDS.find((b) => statusCode >= b.min && statusCode < b.max)?.variant ??
    "secondary"
  );
}
