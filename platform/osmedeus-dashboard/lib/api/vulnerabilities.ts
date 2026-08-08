import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import type { PaginatedResponse } from "@/lib/types/api";
import type {
  Vulnerability,
  VulnerabilityConfidence,
  VulnerabilitySummary,
} from "@/lib/types/vulnerability";

const confidenceValues: VulnerabilityConfidence[] = [
  "Certain",
  "Firm",
  "Tentative",
  "Manual Review Required",
];

function normalizeConfidenceKey(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[_-]+/g, " ")
    .replace(/\s+/g, " ");
}

function fromConfidenceKey(key: string): VulnerabilityConfidence | undefined {
  switch (key) {
    case "certain":
      return "Certain";
    case "firm":
      return "Firm";
    case "tentative":
      return "Tentative";
    case "manual review required":
      return "Manual Review Required";
    default:
      return undefined;
  }
}

function toConfidence(value: unknown): VulnerabilityConfidence | undefined {
  if (typeof value !== "string") return undefined;
  if (confidenceValues.includes(value as VulnerabilityConfidence)) {
    return value as VulnerabilityConfidence;
  }
  return fromConfidenceKey(normalizeConfidenceKey(value));
}

function toConfidenceQueryParam(value: unknown): string | undefined {
  if (typeof value !== "string") return undefined;
  const normalized = normalizeConfidenceKey(value);
  if (!normalized) return undefined;
  return normalized;
}

export interface VulnerabilityFilters {
  workspace?: string;
  severity?: string | string[];
  confidence?: string | string[];
  assetValue?: string;
}

export async function fetchVulnerabilities(params: {
  page?: number;
  pageSize?: number;
  filters?: VulnerabilityFilters;
}): Promise<PaginatedResponse<Vulnerability>> {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 20;
  const offset = (page - 1) * pageSize;
  const filters = params.filters ?? {};

  const normalizeToArray = (value: string | string[] | undefined): string[] => {
    if (typeof value === "string") {
      const v = value.trim();
      return v ? [v] : [];
    }
    if (Array.isArray(value)) {
      return value.map((v) => String(v).trim()).filter(Boolean);
    }
    return [];
  };

  const query: Record<string, any> = { offset, limit: pageSize };
  if (filters.workspace) query.workspace = filters.workspace;
  {
    const severities = normalizeToArray(filters.severity);
    if (severities.length) query.severity = severities.join(",");
  }
  {
    const confidences = normalizeToArray(filters.confidence)
      .map((c) => toConfidenceQueryParam(c))
      .filter((x): x is string => typeof x === "string" && x.length > 0);
    if (confidences.length) query.confidence = confidences.join(",");
  }
  if (filters.assetValue) query.asset_value = filters.assetValue;

  const res = await http.get(`${API_PREFIX}/vulnerabilities`, { params: query });
  const list = (res.data?.data || []) as Array<any>;

  const mapped: Vulnerability[] = list.map((v) => ({
    id: String(v.id),
    workspace: v.workspace ?? "",
    vulnInfo: v.vuln_info ?? "",
    vulnTitle: v.vuln_title ?? "",
    vulnDesc: v.vuln_desc ?? "",
    vulnPoc: v.vuln_poc ?? "",
    severity: v.severity ?? "info",
    confidence: toConfidence(v.confidence),
    assetType: v.asset_type ?? "",
    assetValue: v.asset_value ?? "",
    tags: v.tags ?? [],
    detailHttpRequest: v.detail_http_request,
    detailHttpResponse: v.detail_http_response,
    rawVulnJson: v.raw_vuln_json,
    createdAt: v.created_at ? new Date(v.created_at) : new Date(),
    updatedAt: v.updated_at ? new Date(v.updated_at) : new Date(),
  }));

  const total = res.data?.pagination?.total ?? mapped.length;
  const respLimit = res.data?.pagination?.limit ?? pageSize;
  const respOffset = res.data?.pagination?.offset ?? offset;
  const currentPage = Math.floor(respOffset / respLimit) + 1;

  return {
    data: mapped,
    pagination: {
      page: currentPage,
      pageSize: respLimit,
      totalItems: total,
      totalPages: Math.ceil(total / respLimit),
    },
  };
}

export async function fetchVulnerabilitySummary(
  workspace?: string
): Promise<VulnerabilitySummary> {
  const query: Record<string, any> = {};
  if (workspace) query.workspace = workspace;

  const res = await http.get(`${API_PREFIX}/vulnerabilities/summary`, { params: query });
  const data = res.data?.data || {};

  return {
    bySeverity: {
      critical: data.by_severity?.critical ?? 0,
      high: data.by_severity?.high ?? 0,
      medium: data.by_severity?.medium ?? 0,
      low: data.by_severity?.low ?? 0,
      info: data.by_severity?.info ?? 0,
    },
    total: data.total ?? 0,
    workspace: data.workspace,
  };
}

export async function fetchVulnerabilityById(id: string): Promise<Vulnerability | null> {
  try {
    const res = await http.get(`${API_PREFIX}/vulnerabilities/${id}`);
    const v = res.data?.data;
    if (!v) return null;

    return {
      id: String(v.id),
      workspace: v.workspace ?? "",
      vulnInfo: v.vuln_info ?? "",
      vulnTitle: v.vuln_title ?? "",
      vulnDesc: v.vuln_desc ?? "",
      vulnPoc: v.vuln_poc ?? "",
      severity: v.severity ?? "info",
      confidence: toConfidence(v.confidence),
      assetType: v.asset_type ?? "",
      assetValue: v.asset_value ?? "",
      tags: v.tags ?? [],
      detailHttpRequest: v.detail_http_request,
      detailHttpResponse: v.detail_http_response,
      rawVulnJson: v.raw_vuln_json,
      createdAt: v.created_at ? new Date(v.created_at) : new Date(),
      updatedAt: v.updated_at ? new Date(v.updated_at) : new Date(),
    };
  } catch {
    return null;
  }
}

export async function deleteVulnerability(id: string): Promise<boolean> {
  try {
    await http.delete(`${API_PREFIX}/vulnerabilities/${id}`);
    return true;
  } catch {
    return false;
  }
}
