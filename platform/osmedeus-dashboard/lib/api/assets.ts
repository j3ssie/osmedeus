import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";
import type { Workspace, HttpAsset, HttpAssetFilters, AssetStats } from "@/lib/types/asset";
import type { PaginatedResponse } from "@/lib/types/api";
import { mockWorkspaces } from "@/lib/mock/data/workspaces";
import { mockHttpAssets } from "@/lib/mock/data/http-assets";

export interface FetchWorkspacesParams {
  offset?: number;
  limit?: number;
  search?: string;
  filesystem?: boolean;
  data_source?: string;
  /**
   * Org scope override. Omit to inherit the active org (see lib/api/active-org).
   * Pass "" to force an unscoped listing — the org management screen needs every
   * workspace, including ones already in another org.
   */
  org?: string;
}

export interface WorkspacesListResult {
  items: Workspace[];
  pagination: {
    total: number;
    offset: number;
    limit: number;
  };
  mode?: string;
}

/**
 * Fetch all workspaces with pagination
 */
export async function fetchWorkspaces(params: FetchWorkspacesParams = {}): Promise<Workspace[]> {
  const result = await fetchWorkspacesList(params);
  return result.items;
}

/**
 * Fetch workspaces with pagination info
 */
export async function fetchWorkspacesList(params: FetchWorkspacesParams = {}): Promise<WorkspacesListResult> {
  if (isDemoMode()) {
    const normalized = mockWorkspaces.map((w: any) => mapWorkspace(w));
    const search = params.search?.trim().toLowerCase();
    const ds = params.data_source?.trim().toLowerCase();
    const filteredBase = search
      ? normalized.filter((w) => {
          if (w.name.toLowerCase().includes(search)) return true;
          if (w.local_path.toLowerCase().includes(search)) return true;
          if (w.data_source?.toLowerCase().includes(search)) return true;
          return w.tags.some((t) => t.toLowerCase().includes(search));
        })
      : normalized;
    const filtered = ds && ds !== "all" ? filteredBase.filter((w) => (w.data_source ?? "").toLowerCase() === ds) : filteredBase;
    const offset = typeof params.offset === "number" ? params.offset : 0;
    const limit = typeof params.limit === "number" ? params.limit : filtered.length;
    const items = filtered.slice(offset, offset + limit);
    return {
      items,
      pagination: { total: filtered.length, offset, limit },
      mode: params.filesystem ? "filesystem" : "database",
    };
  }
  const query: Record<string, any> = {};
  if (typeof params.offset === "number") query.offset = params.offset;
  if (typeof params.limit === "number") query.limit = params.limit;
  if (params.search) query.search = params.search;
  if (params.filesystem) query.filesystem = true;
  if (params.data_source) query.data_source = params.data_source;
  // Defined-but-empty deliberately suppresses the active-org interceptor.
  if (typeof params.org === "string") query.org = params.org;
  const res = await http.get(`${API_PREFIX}/workspaces`, { params: query });
  const payload = (res.data || {}) as any;
  const list = (Array.isArray(payload.data)
    ? payload.data
    : Array.isArray(payload.items)
      ? payload.items
      : []) as Array<any>;
  const paginationRaw = payload.pagination || payload.meta?.pagination || {};
  const totalRaw = paginationRaw.total ?? paginationRaw.totalItems;
  const offsetRaw = paginationRaw.offset;
  const limitRaw = paginationRaw.limit;
  const pageRaw = paginationRaw.page;
  const pageSizeRaw = paginationRaw.pageSize;
  const computedOffset =
    typeof offsetRaw !== "undefined"
      ? offsetRaw
      : typeof pageRaw === "number" && typeof pageSizeRaw === "number"
        ? Math.max(0, (pageRaw - 1) * pageSizeRaw)
        : typeof params.offset === "number"
          ? params.offset
          : 0;
  const computedLimit =
    typeof limitRaw !== "undefined"
      ? limitRaw
      : typeof pageSizeRaw === "number"
        ? pageSizeRaw
        : typeof params.limit === "number"
          ? params.limit
          : list.length;
  const computedTotal = Number(totalRaw);
  return {
    items: list.map(mapWorkspace),
    pagination: {
      total: Number.isFinite(computedTotal) ? computedTotal : list.length,
      offset: Number(computedOffset) || 0,
      limit: Number(computedLimit) || list.length,
    },
    mode: payload.mode ?? payload.meta?.mode ?? (params.filesystem ? "filesystem" : "database"),
  };
}

function mapWorkspace(w: any): Workspace {
  return {
    id: Number(w?.id ?? w?.workspace_id ?? 0) || 0,
    name: String(w?.name ?? w?.workspace ?? w?.target ?? ""),
    data_source: typeof w?.data_source === "string" ? w.data_source : typeof w?.dataSource === "string" ? w.dataSource : undefined,
    local_path: String(w?.local_path ?? w?.workspace_path ?? w?.path ?? ""),
    total_assets: Number(w?.total_assets ?? w?.assets_total ?? w?.assets?.total ?? 0) || 0,
    total_subdomains: Number(w?.total_subdomains ?? w?.subdomains_total ?? w?.subdomains?.total ?? 0) || 0,
    total_urls: Number(w?.total_urls ?? w?.urls_total ?? w?.http_assets_total ?? w?.http_assets?.total ?? 0) || 0,
    total_vulns: Number(w?.total_vulns ?? w?.vulns_total ?? w?.vulnerabilities?.total ?? 0) || 0,
    vuln_critical: Number(w?.vuln_critical ?? w?.vulnerabilities?.critical ?? 0) || 0,
    vuln_high: Number(w?.vuln_high ?? w?.vulnerabilities?.high ?? 0) || 0,
    vuln_medium: Number(w?.vuln_medium ?? w?.vulnerabilities?.medium ?? 0) || 0,
    vuln_low: Number(w?.vuln_low ?? w?.vulnerabilities?.low ?? 0) || 0,
    vuln_potential: Number(w?.vuln_potential ?? w?.vulnerabilities?.potential ?? w?.vulnerabilities?.info ?? 0) || 0,
    risk_score: Number(w?.risk_score ?? w?.risk?.score ?? w?.score ?? 0) || 0,
    tags: Array.isArray(w?.tags) ? w.tags : Array.isArray(w?.labels) ? w.labels : [],
    last_run: String(w?.last_run ?? w?.last_scan ?? w?.latest_run_at ?? w?.last_run_at ?? ""),
    run_workflow: String(w?.run_workflow ?? w?.last_workflow ?? w?.workflow ?? ""),
    state_execution_log:
      typeof w?.state_execution_log === "string"
        ? w.state_execution_log
        : typeof w?.state?.execution_log === "string"
          ? w.state.execution_log
          : undefined,
    state_completed_file:
      typeof w?.state_completed_file === "string"
        ? w.state_completed_file
        : typeof w?.state?.completed_file === "string"
          ? w.state.completed_file
          : undefined,
    state_workflow_file:
      typeof w?.state_workflow_file === "string"
        ? w.state_workflow_file
        : typeof w?.state?.workflow_file === "string"
          ? w.state.workflow_file
          : undefined,
    state_workflow_folder:
      typeof w?.state_workflow_folder === "string"
        ? w.state_workflow_folder
        : typeof w?.state?.workflow_folder === "string"
          ? w.state.workflow_folder
          : undefined,
    created_at: String(w?.created_at ?? w?.createdAt ?? ""),
    updated_at: String(w?.updated_at ?? w?.updatedAt ?? ""),
  };
}

/**
 * Fetch a single workspace by ID
 */
export async function fetchWorkspace(id: string): Promise<Workspace | null> {
  if (isDemoMode()) {
    const w = mockWorkspaces.find((x: any) => String(x.id) === id || x.name === id);
    if (!w) return null;
    return mapWorkspace(w);
  }
  try {
    const res = await http.get(`${API_PREFIX}/workspaces`, { params: { offset: 0, limit: 1000 } });
    const list = (res.data?.data || []) as Array<any>;
    const w = list.find((x) => String(x.id) === id || x.name === id);
    if (!w) return null;
    return mapWorkspace(w);
  } catch {
    return null;
  }
}

function normalizeStatsList(list: unknown): string[] {
  if (!Array.isArray(list)) return [];
  const values = list
    .map((item) => String(item ?? "").trim())
    .filter((value) => value.length > 0);
  return Array.from(new Set(values)).sort((a, b) => a.localeCompare(b));
}

/**
 * Which workspaces a query covers: one name, several, or none at all — and none
 * means every workspace, the same thing the old single-value `undefined` meant.
 */
export type WorkspaceSelection = string | string[] | undefined;

function normalizeWorkspaceSelection(workspace: WorkspaceSelection): string[] {
  const list = Array.isArray(workspace) ? workspace : workspace ? [workspace] : [];
  return Array.from(
    new Set(list.map((name) => String(name ?? "").trim()).filter(Boolean))
  );
}

function resolveMockWorkspaceKey(workspace: string | undefined): string | null {
  const normalizedWorkspace = (workspace ?? "").trim();
  if (!normalizedWorkspace) return null;
  const match = mockWorkspaces.find((w) => w.name === normalizedWorkspace) ||
    mockWorkspaces.find((w) => w.name.toLowerCase() === normalizedWorkspace.toLowerCase());
  if (match) return `ws-${String(match.id).padStart(3, "0")}`;
  if (normalizedWorkspace.startsWith("ws-")) return normalizedWorkspace;
  return null;
}

/**
 * Mock bucket keys for a selection. An empty selection spans every bucket; a
 * selection whose names resolve to nothing spans none, rather than silently
 * falling back to the first workspace the way the single-value version did.
 */
function resolveMockWorkspaceKeys(selection: string[]): string[] {
  if (!selection.length) return Object.keys(mockHttpAssets);
  const keys = selection
    .map((name) => resolveMockWorkspaceKey(name))
    .filter((key): key is string => Boolean(key));
  return Array.from(new Set(keys));
}

/** Mock buckets are keyed `ws-001`; rows should still name their workspace. */
function mockWorkspaceLabel(key: string): string {
  const id = Number(key.replace(/^ws-0*/, ""));
  return mockWorkspaces.find((w) => w.id === id)?.name ?? key;
}

export async function fetchAssetStats(
  workspace?: WorkspaceSelection
): Promise<AssetStats> {
  const selection = normalizeWorkspaceSelection(workspace);
  if (isDemoMode()) {
    const allAssets = resolveMockWorkspaceKeys(selection).flatMap(
      (key) => mockHttpAssets[key] ?? []
    );
    const technologies = new Set<string>();
    const sources = new Set<string>();
    const remarks = new Set<string>();
    const assetTypes = new Set<string>();
    allAssets.forEach((asset) => {
      asset.technologies.forEach((tech) => {
        const value = String(tech ?? "").trim();
        if (value) technologies.add(value);
      });
      const source = String(asset.source ?? "").trim();
      if (source) sources.add(source);
      const assetType = String(asset.assetType ?? "").trim();
      if (assetType) assetTypes.add(assetType);
      const assetRemarks = Array.isArray(asset.remarks)
        ? asset.remarks
        : asset.remarks
          ? [asset.remarks]
          : [];
      assetRemarks.forEach((remark) => {
        const value = String(remark ?? "").trim();
        if (value) remarks.add(value);
      });
    });
    return {
      technologies: Array.from(technologies).sort((a, b) => a.localeCompare(b)),
      sources: Array.from(sources).sort((a, b) => a.localeCompare(b)),
      remarks: Array.from(remarks).sort((a, b) => a.localeCompare(b)),
      assetTypes: Array.from(assetTypes).sort((a, b) => a.localeCompare(b)),
    };
  }
  const params = selection.length ? { workspace: selection.join(",") } : undefined;
  const res = await http.get(`${API_PREFIX}/asset-stats`, { params });
  const data = res.data?.data ?? {};
  return {
    technologies: normalizeStatsList(data.technologies),
    sources: normalizeStatsList(data.sources),
    remarks: normalizeStatsList(data.remarks),
    assetTypes: normalizeStatsList(data.asset_types ?? data.assetTypes),
  };
}

/**
 * Fetch HTTP assets across one, several, or all workspaces, with filtering and
 * pagination. Several workspaces go to the server as one comma-joined
 * `workspace` param — the same shape the other multi-value filters below use —
 * so paging stays server-side instead of stitching per-workspace responses.
 */
export async function fetchHttpAssets(
  workspace: WorkspaceSelection,
  params: {
    page?: number;
    pageSize?: number;
    filters?: HttpAssetFilters;
  }
): Promise<PaginatedResponse<HttpAsset>> {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 20;
  const offset = (page - 1) * pageSize;
  const filters = params.filters ?? {};
  const selection = normalizeWorkspaceSelection(workspace);
  if (isDemoMode()) {
    const all = resolveMockWorkspaceKeys(selection).flatMap((key) =>
      (mockHttpAssets[key] ?? []).map((asset) =>
        // The buckets store their own key as the workspace; rows need the name,
        // which is the only way to tell workspaces apart once several are on.
        asset.workspace === key
          ? { ...asset, workspace: mockWorkspaceLabel(key) }
          : asset
      )
    );
    const q = (filters.search ?? "").trim().toLowerCase();
    const statusSet = new Set(filters.statusCodes ?? []);
    const wantedTech = (filters.technologies ?? [])
      .map((t) => t.trim().toLowerCase())
      .filter(Boolean);
    const wantedAssetTypes = (filters.assetTypes ?? [])
      .map((t) => t.trim().toLowerCase())
      .filter(Boolean);
    const wantedSources = (filters.sources ?? [])
      .map((t) => t.trim().toLowerCase())
      .filter(Boolean);
    const wantedRemarks = (filters.remarks ?? [])
      .map((t) => t.trim().toLowerCase())
      .filter(Boolean);
    const wantedContentTypes = (filters.contentTypes ?? [])
      .map((t) => t.trim().toLowerCase())
      .filter(Boolean);
    const wantedTls = (filters.tlsVersion ?? "").trim().toLowerCase();
    const wantedLocation = (filters.location ?? "").trim().toLowerCase();

    const filtered = all.filter((a) => {
      if (q) {
        const hay = [a.url, a.assetValue, a.title ?? "", a.hostIp ?? ""]
          .join(" ")
          .toLowerCase();
        if (!hay.includes(q)) return false;
      }
      if (statusSet.size > 0 && !statusSet.has(a.statusCode)) return false;
      if (wantedTech.length > 0) {
        const techSet = new Set(a.technologies.map((t) => String(t).trim().toLowerCase()));
        if (!wantedTech.some((t) => techSet.has(t))) return false;
      }
      if (wantedAssetTypes.length > 0) {
        const assetType = String(a.assetType ?? "").trim().toLowerCase();
        if (!wantedAssetTypes.includes(assetType)) return false;
      }
      if (wantedSources.length > 0) {
        const source = String(a.source ?? "").trim().toLowerCase();
        if (!wantedSources.includes(source)) return false;
      }
      if (wantedRemarks.length > 0) {
        const assetRemarks = Array.isArray(a.remarks)
          ? a.remarks
          : a.remarks
            ? [a.remarks]
            : [];
        const remarkSet = new Set(
          assetRemarks.map((r) => String(r).trim().toLowerCase()).filter(Boolean)
        );
        if (!wantedRemarks.some((r) => remarkSet.has(r))) return false;
      }
      if (wantedContentTypes.length > 0) {
        const ct = (a.contentType ?? "").toLowerCase();
        if (!wantedContentTypes.some((t) => ct.includes(t))) return false;
      }
      if (wantedTls && String(a.tls ?? "").toLowerCase() !== wantedTls) return false;
      if (wantedLocation) {
        const locHay = [a.url, a.assetValue, a.hostIp ?? ""]
          .join(" ")
          .toLowerCase();
        if (!locHay.includes(wantedLocation)) return false;
      }
      if (typeof filters.minContentLength === "number" && a.contentLength < filters.minContentLength) return false;
      if (typeof filters.maxContentLength === "number" && a.contentLength > filters.maxContentLength) return false;
      return true;
    });

    const sliced = filtered.slice(offset, offset + pageSize);
    const totalItems = filtered.length;
    return {
      data: sliced,
      pagination: {
        page,
        pageSize,
        totalItems,
        totalPages: Math.ceil(totalItems / pageSize),
      },
    };
  }
  const query: Record<string, any> = { offset, limit: pageSize };
  if (selection.length) query.workspace = selection.join(",");
  if (filters.search) query.search = filters.search;
  if (filters.statusCodes?.length) query.status_code = filters.statusCodes.join(",");
  if (typeof filters.minContentLength === "number") query.min_content_length = filters.minContentLength;
  if (typeof filters.maxContentLength === "number") query.max_content_length = filters.maxContentLength;
  if (filters.location) query.location = filters.location;
  // New filter parameters
  if (filters.technologies?.length) query.tech = filters.technologies.join(",");
  if (filters.assetTypes?.length) query.asset_type = filters.assetTypes.join(",");
  if (filters.sources?.length) query.source = filters.sources.join(",");
  if (filters.remarks?.length) query.remarks = filters.remarks.join(",");
  if (filters.contentTypes?.length) query.content_type = filters.contentTypes.join(",");
  if (filters.tlsVersion) query.tls = filters.tlsVersion;
  const res = await http.get(`${API_PREFIX}/assets`, { params: query });
  const list = (res.data?.data || []) as Array<any>;
  const mapped: HttpAsset[] = list.map((a) => ({
    id: String(a.id ?? a.url),
    workspace: a.workspace ?? "",
    assetValue: a.asset_value ?? "",
    url: a.url ?? "",
    input: a.input ?? "",
    scheme: a.scheme ?? "",
    method: a.method ?? "GET",
    path: a.path ?? "/",
    statusCode: a.status_code ?? 0,
    contentType: a.content_type ?? "",
    contentLength: a.content_length ?? 0,
    title: a.title,
    words: a.words ?? 0,
    lines: a.lines ?? 0,
    hostIp: a.host_ip,
    aRecords: a.dns_records ?? a.a ?? [],
    tls: a.tls,
    assetType: a.asset_type ?? "web",
    technologies: a.tech ?? [],
    responseTime: a.time,
    remarks: a.remarks,
    source: a.source ?? "",
    createdAt: a.created_at ? new Date(a.created_at) : new Date(),
    updatedAt: a.updated_at ? new Date(a.updated_at) : new Date(),
    lastSeenAt: a.last_seen_at ? new Date(a.last_seen_at) : undefined,
  }));
  const total = res.data?.pagination?.total ?? mapped.length;
  return {
    data: mapped,
    pagination: {
      page,
      pageSize,
      totalItems: total,
      totalPages: Math.ceil(total / pageSize),
    },
  };
}
