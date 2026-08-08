import { NextResponse } from "next/server";
import { mockWorkspaces } from "@/lib/mock/data/workspaces";

export const dynamic = "force-static";

export async function GET() {
  const offset = 0;
  const limit = 20;

  // Mock data already has the correct structure
  const items = mockWorkspaces.map((w) => ({
    id: w.id,
    name: w.name,
    data_source: w.data_source,
    local_path: w.local_path,
    state_execution_log: w.state_execution_log,
    state_completed_file: w.state_completed_file,
    state_workflow_file: w.state_workflow_file,
    state_workflow_folder: w.state_workflow_folder,
    total_assets: w.total_assets,
    total_subdomains: w.total_subdomains,
    total_urls: w.total_urls,
    total_vulns: w.total_vulns,
    vuln_critical: w.vuln_critical,
    vuln_high: w.vuln_high,
    vuln_medium: w.vuln_medium,
    vuln_low: w.vuln_low,
    vuln_potential: w.vuln_potential,
    risk_score: w.risk_score,
    tags: w.tags,
    last_run: w.last_run,
    run_workflow: w.run_workflow,
    created_at: w.created_at,
    updated_at: w.updated_at,
  }));

  const filtered = items;

  const sliced = filtered.slice(offset, offset + limit);

  const resp = {
    data: sliced,
    pagination: {
      total: filtered.length,
      offset,
      limit,
    },
    mode: "database",
  };

  return NextResponse.json(resp);
}
