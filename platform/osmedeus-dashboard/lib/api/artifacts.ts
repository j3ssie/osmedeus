import { http, getHttpBaseURL } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";
import type { PaginatedResponse } from "@/lib/types/api";
import type { Artifact } from "@/lib/types/artifact";

export interface ArtifactFilters {
  workspace?: string;
  artifactType?: string;
}

const DEMO_ARTIFACTS: Artifact[] = [
  {
    id: "e746aecd-d0db-4331-a44c-3cd94b7f74b5",
    runId: "988c6fff",
    workspace: "bin.com",
    name: "security-report",
    artifactPath: "/home/osmedeus/workspaces-osmedeus/bin.com/security-report.md",
    artifactType: "report",
    contentType: "md",
    sizeBytes: 0,
    lineCount: 0,
    description: "Security report summary",
    createdAt: new Date("2026-01-10T09:18:58.677214Z"),
  },
  {
    id: "71f516ed-7e72-4559-b803-8252f774c2ee",
    runId: "aee682c2",
    workspace: "example.com",
    name: "screenshots",
    artifactPath: "/home/osmedeus/workspaces-osmedeus/example.com/screenshots/",
    artifactType: "screenshot",
    contentType: "folder",
    sizeBytes: 15728640,
    lineCount: 78,
    description: "GoWitness screenshot captures",
    createdAt: new Date("2026-01-10T08:12:29.227691Z"),
  },
  {
    id: "c6159b4e-1927-4888-b9ee-5aa12996a131",
    runId: "19ac73d1",
    workspace: "staging.test.local",
    name: "nuclei-partial.json",
    artifactPath: "/home/osmedeus/workspaces-osmedeus/staging.test.local/vuln/nuclei-partial.json",
    artifactType: "output",
    contentType: "json",
    sizeBytes: 8923,
    lineCount: 45,
    description: "Partial nuclei results before failure",
    createdAt: new Date("2026-01-10T07:12:29.227691Z"),
  },
];

function mapArtifact(a: any): Artifact {
  return {
    id: String(a?.id ?? ""),
    runId: String(a?.run_id ?? a?.runId ?? ""),
    workspace: String(a?.workspace ?? ""),
    name: String(a?.name ?? ""),
    artifactPath: String(a?.artifact_path ?? a?.artifactPath ?? ""),
    artifactType: String(a?.artifact_type ?? a?.artifactType ?? ""),
    contentType: String(a?.content_type ?? a?.contentType ?? ""),
    sizeBytes: Number(a?.size_bytes ?? a?.sizeBytes ?? 0) || 0,
    lineCount: Number(a?.line_count ?? a?.lineCount ?? 0) || 0,
    description: String(a?.description ?? ""),
    createdAt: a?.created_at ? new Date(a.created_at) : new Date(),
  };
}

export async function fetchArtifacts(params: {
  page?: number;
  pageSize?: number;
  verifyExist?: boolean;
  filters?: ArtifactFilters;
} = {}): Promise<PaginatedResponse<Artifact>> {
  const page = Math.max(1, params.page ?? 1);
  const pageSize = Math.max(1, params.pageSize ?? 20);
  const filters = params.filters ?? {};
  const verifyExist = !!params.verifyExist;

  if (isDemoMode()) {
    const ws = (filters.workspace ?? "").trim().toLowerCase();
    const t = (filters.artifactType ?? "").trim().toLowerCase();
    const filtered = DEMO_ARTIFACTS.filter((a) => {
      if (ws && a.workspace.toLowerCase() !== ws) return false;
      if (t && a.artifactType.toLowerCase() !== t) return false;
      return true;
    });
    const total = filtered.length;
    const start = (page - 1) * pageSize;
    const data = filtered.slice(start, start + pageSize);
    return {
      data,
      pagination: {
        page,
        pageSize,
        totalItems: total,
        totalPages: Math.max(1, Math.ceil(total / pageSize)),
      },
    };
  }

  const offset = (page - 1) * pageSize;
  const query: Record<string, any> = { offset, limit: pageSize };
  if (filters.workspace) query.workspace = filters.workspace;
  if (filters.artifactType) query.artifact_type = filters.artifactType;
  if (verifyExist) query.verify_exist = true;

  const res = await http.get(`${API_PREFIX}/artifacts`, { params: query });
  const list = (res.data?.data || []) as Array<any>;
  const mapped = list.map(mapArtifact);
  const total = Number(res.data?.pagination?.total ?? mapped.length);

  return {
    data: mapped,
    pagination: {
      page,
      pageSize,
      totalItems: Number.isFinite(total) ? total : mapped.length,
      totalPages: Math.max(1, Math.ceil((Number.isFinite(total) ? total : mapped.length) / pageSize)),
    },
  };
}

export const WORKSPACE_PREFIX_KEY_STORAGE = "osmedeus_workspace_prefix_key";

// The workspace_prefix_key gates the backend's unauthenticated static workspace
// serving route (GET /ws/{workspace_prefix_key}/*). It is a secret configured on
// the server, so the dashboard reads it from local config (or a build-time env var).
export function getWorkspacePrefixKey(): string {
  if (typeof window !== "undefined") {
    const saved = localStorage.getItem(WORKSPACE_PREFIX_KEY_STORAGE);
    if (saved && saved.trim()) return saved.trim();
  }
  const env = process.env.NEXT_PUBLIC_WORKSPACE_PREFIX_KEY;
  return env ? env.trim() : "";
}

// Build the raw, unauthenticated static URL for an artifact served by the backend
// at /ws/{workspace_prefix_key}/{workspace}/{relative_path}. Returns null when the
// prefix key has not been configured. The browser can open this URL directly in a
// new tab — the server renders HTML natively and lists folders (directory browsing).
export function buildWorkspaceArtifactUrl(params: {
  workspace: string;
  artifactPath: string;
}): string | null {
  const workspace = (params.workspace ?? "").trim();
  const artifactPath = (params.artifactPath ?? "").trim();
  if (!workspace || !artifactPath) return null;

  const prefixKey = getWorkspacePrefixKey();
  if (!prefixKey) return null;

  const base = getHttpBaseURL().replace(/\/+$/, "");
  const rel = toRelativeArtifactPath(artifactPath, workspace);
  const segments = [workspace, ...rel.split("/")]
    .filter(Boolean)
    .map((s) => encodeURIComponent(s))
    .join("/");

  return `${base}/ws/${encodeURIComponent(prefixKey)}/${segments}`;
}

function toRelativeArtifactPath(fullPath: string, workspace: string): string {
  const ws = (workspace ?? "").trim();
  if (!ws) return String(fullPath ?? "").trim().replace(/^\/+/, "");

  const normalized = String(fullPath ?? "").trim().replace(/\\/g, "/");
  const marker = `/${ws}/`;
  const idx = normalized.indexOf(marker);
  if (idx >= 0) {
    return normalized.slice(idx + marker.length).replace(/^\/+/, "");
  }

  const alt = `${ws}/`;
  const idx2 = normalized.indexOf(alt);
  if (idx2 >= 0) {
    return normalized.slice(idx2 + alt.length).replace(/^\/+/, "");
  }

  return normalized.replace(/^\/+/, "");
}

export async function fetchArtifactContent(params: {
  workspace: string;
  artifactPath: string;
}): Promise<string> {
  const workspace = (params.workspace ?? "").trim();
  const artifactPath = (params.artifactPath ?? "").trim();

  if (!workspace) {
    throw new Error("0:workspace is required");
  }
  if (!artifactPath) {
    throw new Error("0:artifact_path is required");
  }

  if (isDemoMode()) {
    return `# Sample Repository Findings - ${workspace}\n\n**Target**: ${workspace}\n**Date**: 2026-01-10\n\n\n---\n*Report generated by Osmedeus v5.0.0*\n`;
  }

  const rel = toRelativeArtifactPath(artifactPath, workspace);
  const url = `${API_PREFIX}/artifacts/${encodeURIComponent(workspace)}`;
  const res = await http.get(url, {
    params: { artifact_path: rel },
    responseType: "text",
    transformResponse: (x) => x,
    headers: { Accept: "text/plain, text/markdown, */*" },
  });

  return typeof res.data === "string" ? res.data : String(res.data ?? "");
}
