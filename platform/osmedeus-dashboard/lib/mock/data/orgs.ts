import type { Org } from "@/lib/types/org";
import { DEFAULT_ORG_UUID } from "@/lib/types/org";

/**
 * Demo-mode orgs. Workspace names line up with mockWorkspaces so the assignment
 * dialog and the per-org counts stay coherent in demo mode.
 */
export const mockOrgs: Org[] = [
  {
    uuid: DEFAULT_ORG_UUID,
    name: "default",
    description: "Default org for data not assigned to a specific org",
    tags: [],
    is_default: true,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    stats: {
      org_uuid: DEFAULT_ORG_UUID,
      org_name: "default",
      total_workspaces: 3,
      total_assets: 412,
      total_vulns: 31,
      total_runs: 18,
      workspaces: ["secure.bank.com", "shop.retail.com", "startup.dev"],
    },
  },
  {
    uuid: "f072a502-0a35-4e8f-aef0-79e048e082f7",
    name: "acme",
    description: "ACME Corp",
    tags: ["corp"],
    is_default: false,
    created_at: "2024-02-10T09:30:00Z",
    updated_at: "2024-06-02T11:12:00Z",
    stats: {
      org_uuid: "f072a502-0a35-4e8f-aef0-79e048e082f7",
      org_name: "acme",
      total_workspaces: 2,
      total_assets: 289,
      total_vulns: 24,
      total_runs: 12,
      workspaces: ["acme.io", "example.com"],
    },
  },
  {
    uuid: "ca6d8062-0ef3-472c-9d3c-e3dba920206d",
    name: "globex",
    description: "Globex Corporation",
    tags: ["bug-bounty"],
    is_default: false,
    created_at: "2024-03-22T14:05:00Z",
    updated_at: "2024-05-18T08:44:00Z",
    stats: {
      org_uuid: "ca6d8062-0ef3-472c-9d3c-e3dba920206d",
      org_name: "globex",
      total_workspaces: 2,
      total_assets: 176,
      total_vulns: 9,
      total_runs: 7,
      workspaces: ["megacorp.com", "testsite.org"],
    },
  },
];
