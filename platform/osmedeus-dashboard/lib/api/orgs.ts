import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";
import { mockOrgs } from "@/lib/mock/data/orgs";
import type {
  Org,
  OrgStats,
  CreateOrgInput,
  UpdateOrgInput,
  OrgAssignResult,
} from "@/lib/types/org";

function mapStats(raw: any, fallbackUUID: string, fallbackName: string): OrgStats {
  return {
    org_uuid: String(raw?.org_uuid ?? fallbackUUID),
    org_name: String(raw?.org_name ?? fallbackName),
    total_workspaces: Number(raw?.total_workspaces ?? 0) || 0,
    total_assets: Number(raw?.total_assets ?? 0) || 0,
    total_vulns: Number(raw?.total_vulns ?? 0) || 0,
    total_runs: Number(raw?.total_runs ?? 0) || 0,
    workspaces: Array.isArray(raw?.workspaces) ? raw.workspaces.map(String) : [],
  };
}

function mapOrg(raw: any): Org {
  const uuid = String(raw?.uuid ?? "");
  const name = String(raw?.name ?? "");
  return {
    uuid,
    name,
    description: typeof raw?.description === "string" ? raw.description : undefined,
    tags: Array.isArray(raw?.tags) ? raw.tags.map(String) : [],
    is_default: Boolean(raw?.is_default),
    created_at: String(raw?.created_at ?? ""),
    updated_at: String(raw?.updated_at ?? ""),
    stats: raw?.stats ? mapStats(raw.stats, uuid, name) : undefined,
  };
}

/**
 * List every org with its workspace/asset/vulnerability counts.
 *
 * The counts come embedded in this response, so there is no reason to also call
 * fetchOrgStats when rendering a list.
 */
export async function fetchOrgs(): Promise<Org[]> {
  if (isDemoMode()) {
    return mockOrgs.map((o) => ({ ...o }));
  }
  const res = await http.get(`${API_PREFIX}/orgs`);
  const payload = (res.data || {}) as any;
  const list = Array.isArray(payload.data) ? payload.data : [];
  return list.map(mapOrg);
}

/** Fetch one org by name or UUID. */
export async function fetchOrg(ref: string): Promise<Org> {
  if (isDemoMode()) {
    const found = mockOrgs.find((o) => o.uuid === ref || o.name === ref);
    if (!found) throw new Error(`404:org not found: ${ref}`);
    return { ...found };
  }
  const res = await http.get(`${API_PREFIX}/orgs/${encodeURIComponent(ref)}`);
  return mapOrg((res.data || {}).data);
}

/** Fetch aggregate counts for one org. */
export async function fetchOrgStats(ref: string): Promise<OrgStats> {
  if (isDemoMode()) {
    const found = mockOrgs.find((o) => o.uuid === ref || o.name === ref);
    if (!found?.stats) throw new Error(`404:org not found: ${ref}`);
    return { ...found.stats };
  }
  const res = await http.get(`${API_PREFIX}/orgs/${encodeURIComponent(ref)}/stats`);
  const payload = (res.data || {}) as any;
  return mapStats(payload.data, ref, ref);
}

export async function createOrg(input: CreateOrgInput): Promise<Org> {
  if (isDemoMode()) {
    throw new Error("400:Creating orgs is disabled in demo mode");
  }
  const res = await http.post(`${API_PREFIX}/orgs`, input);
  return mapOrg((res.data || {}).data);
}

export interface UpdateOrgResult {
  org: Org;
  assigned?: OrgAssignResult;
}

/**
 * Update an org's metadata and optionally assign workspaces to it.
 *
 * Assigning cascades the org stamp to every asset, vulnerability and run in
 * those workspaces — this is how pre-existing data gets grouped without
 * re-scanning.
 */
export async function updateOrg(ref: string, input: UpdateOrgInput): Promise<UpdateOrgResult> {
  if (isDemoMode()) {
    throw new Error("400:Editing orgs is disabled in demo mode");
  }
  const res = await http.put(`${API_PREFIX}/orgs/${encodeURIComponent(ref)}`, input);
  const payload = (res.data || {}) as any;
  return { org: mapOrg(payload.data), assigned: payload.assigned };
}

/** Assign workspaces to an org. Thin wrapper over updateOrg. */
export async function assignWorkspaces(ref: string, workspaces: string[]): Promise<OrgAssignResult> {
  const result = await updateOrg(ref, { workspaces });
  return result.assigned ?? {};
}

/**
 * Delete an org.
 *
 * By default its data is reassigned to the default org — nothing is lost, only
 * the grouping. Pass purge to delete the workspaces, assets, vulnerabilities and
 * runs along with it.
 */
export async function deleteOrg(ref: string, purge = false): Promise<void> {
  if (isDemoMode()) {
    throw new Error("400:Deleting orgs is disabled in demo mode");
  }
  await http.delete(`${API_PREFIX}/orgs/${encodeURIComponent(ref)}`, {
    params: purge ? { purge: true } : undefined,
  });
}
