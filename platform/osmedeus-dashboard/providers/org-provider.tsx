"use client";

import * as React from "react";
import { fetchOrgs } from "@/lib/api/orgs";
import {
  getActiveOrg,
  setActiveOrg as persistActiveOrg,
  ORG_CHANGED_EVENT,
  type ActiveOrg,
} from "@/lib/api/active-org";
import type { Org } from "@/lib/types/org";

interface OrgContextValue {
  /** Every org known to the server, with counts. */
  orgs: Org[];
  /** The org all list views are scoped to, or null for "all orgs". */
  activeOrg: ActiveOrg | null;
  isLoading: boolean;
  error: string | null;
  /** Select an org (or null to clear). Reloads so every view refetches. */
  selectOrg: (org: ActiveOrg | null) => void;
  /** Re-read the org list, e.g. after a create/delete. */
  refresh: () => Promise<void>;
}

const OrgContext = React.createContext<OrgContextValue | undefined>(undefined);

export function OrgProvider({ children }: { children: React.ReactNode }) {
  const [orgs, setOrgs] = React.useState<Org[]>([]);
  const [activeOrg, setActiveOrgState] = React.useState<ActiveOrg | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  const refresh = React.useCallback(async () => {
    setIsLoading(true);
    try {
      const list = await fetchOrgs();
      setOrgs(list);
      setError(null);

      // Drop a stale selection if the org disappeared (deleted elsewhere, or a
      // different server). Leaving it set would silently filter every list to an
      // org that no longer exists.
      const current = getActiveOrg();
      if (current && !list.some((o) => o.uuid === current.uuid)) {
        persistActiveOrg(null);
        setActiveOrgState(null);
      }
    } catch (err) {
      // A server without the org endpoint (older build) should degrade to the
      // unfiltered view rather than breaking the whole dashboard.
      setOrgs([]);
      setError(err instanceof Error ? err.message : "Failed to load orgs");
    } finally {
      setIsLoading(false);
    }
  }, []);

  React.useEffect(() => {
    setActiveOrgState(getActiveOrg());
    void refresh();
  }, [refresh]);

  // Keep in step when another component (or another tab) changes the selection.
  React.useEffect(() => {
    const onChanged = () => setActiveOrgState(getActiveOrg());
    window.addEventListener(ORG_CHANGED_EVENT, onChanged);
    window.addEventListener("storage", onChanged);
    return () => {
      window.removeEventListener(ORG_CHANGED_EVENT, onChanged);
      window.removeEventListener("storage", onChanged);
    };
  }, []);

  const selectOrg = React.useCallback((org: ActiveOrg | null) => {
    const current = getActiveOrg();
    if (current?.uuid === org?.uuid) return;

    persistActiveOrg(org);
    setActiveOrgState(org);

    // Every page fetches in its own effect on mount, so a reload is the only
    // reliable way to get all of them onto the new org. Switching tenants is a
    // deliberate, infrequent action, so the cost is acceptable.
    if (typeof window !== "undefined") {
      window.location.reload();
    }
  }, []);

  const value = React.useMemo<OrgContextValue>(
    () => ({ orgs, activeOrg, isLoading, error, selectOrg, refresh }),
    [orgs, activeOrg, isLoading, error, selectOrg, refresh]
  );

  return <OrgContext.Provider value={value}>{children}</OrgContext.Provider>;
}

export function useOrg(): OrgContextValue {
  const ctx = React.useContext(OrgContext);
  if (!ctx) {
    throw new Error("useOrg must be used within an OrgProvider");
  }
  return ctx;
}
