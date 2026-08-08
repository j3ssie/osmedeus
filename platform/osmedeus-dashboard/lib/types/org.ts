/**
 * Orgs group multiple workspaces under one tenant so assets, vulnerabilities and
 * runs can be queried across all of them at once.
 *
 * Orgs are additive: every row that names no org belongs to the built-in default
 * org, and a request with no `org` filter spans every org.
 */

/** UUID of the built-in default org. Cannot be renamed or deleted. */
export const DEFAULT_ORG_UUID = "00000000-0000-0000-0000-000000000001";

export interface OrgStats {
  org_uuid: string;
  org_name: string;
  total_workspaces: number;
  total_assets: number;
  total_vulns: number;
  total_runs: number;
  workspaces: string[];
}

export interface Org {
  uuid: string;
  name: string;
  description?: string;
  tags: string[];
  is_default: boolean;
  created_at: string;
  updated_at: string;
  /** Present on list responses; absent on bare get. */
  stats?: OrgStats;
}

export interface CreateOrgInput {
  name: string;
  description?: string;
  uuid?: string;
  tags?: string[];
}

export interface UpdateOrgInput {
  name?: string;
  description?: string;
  tags?: string[];
  /** Assigns these workspaces to the org, cascading to their assets/vulns/runs. */
  workspaces?: string[];
}

/** Rows updated per table by a workspace assignment. */
export interface OrgAssignResult {
  workspaces?: number;
  assets?: number;
  vulnerabilities?: number;
  runs?: number;
}

export function isDefaultOrg(org: Pick<Org, "uuid">): boolean {
  return org.uuid === DEFAULT_ORG_UUID;
}
