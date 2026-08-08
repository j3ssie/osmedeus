/**
 * Active org selection.
 *
 * This is the UI equivalent of `osmedeus org use`: pick an org once and every
 * org-scoped list request carries it as `?org=`. An unset value means no filter,
 * so the dashboard spans all orgs — the same default the CLI and API use.
 *
 * Stored in localStorage rather than React state so the http interceptor can
 * read it without threading the value through every API call, matching how the
 * auth token and API endpoint are already handled in http.ts.
 */

const ACTIVE_ORG_KEY = "osmedeus_active_org";

/** Fired after the active org changes so listeners can refetch. */
export const ORG_CHANGED_EVENT = "osmedeus-active-org-changed";

export interface ActiveOrg {
  uuid: string;
  name: string;
}

export function getActiveOrg(): ActiveOrg | null {
  if (typeof window === "undefined") return null;
  const raw = window.localStorage.getItem(ACTIVE_ORG_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw);
    if (parsed && typeof parsed.uuid === "string" && parsed.uuid) {
      return { uuid: parsed.uuid, name: String(parsed.name ?? parsed.uuid) };
    }
  } catch {
    // Corrupt value - fall through and clear it so the UI recovers on its own.
    window.localStorage.removeItem(ACTIVE_ORG_KEY);
  }
  return null;
}

/** The UUID to send as `?org=`, or "" when no org is selected. */
export function getActiveOrgUUID(): string {
  return getActiveOrg()?.uuid ?? "";
}

export function setActiveOrg(org: ActiveOrg | null): void {
  if (typeof window === "undefined") return;
  if (org) {
    window.localStorage.setItem(ACTIVE_ORG_KEY, JSON.stringify(org));
  } else {
    window.localStorage.removeItem(ACTIVE_ORG_KEY);
  }
  window.dispatchEvent(new CustomEvent(ORG_CHANGED_EVENT, { detail: org }));
}

/**
 * Endpoints that honour `?org=` server-side. Injecting elsewhere would be
 * ignored by the API, but keeping the list exact means a stray param never shows
 * up in a request that has nothing to do with orgs.
 */
const ORG_SCOPED_PATHS = new Set(["/assets", "/vulnerabilities", "/runs", "/workspaces"]);

/** Whether a request path (relative to the API prefix) should carry `?org=`. */
export function isOrgScopedPath(pathAfterPrefix: string): boolean {
  const clean = pathAfterPrefix.split("?")[0].replace(/\/+$/, "");
  return ORG_SCOPED_PATHS.has(clean);
}
