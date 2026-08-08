import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";
import type { StepResult, StepResultsResponse } from "@/lib/types/scan";

function mapStepResultFromApi(r: any): StepResult {
  return {
    id: String(r?.id ?? ""),
    runId: r?.run_id ?? r?.runId ?? undefined,
    runUuid: r?.run_uuid ?? r?.runUuid ?? undefined,
    stepName: r?.step_name ?? r?.stepName ?? "",
    stepType: r?.step_type ?? r?.stepType ?? "",
    status: r?.status ?? "unknown",
    command: r?.command ?? "",
    output: r?.output ?? "",
    error: r?.error ?? "",
    durationMs: typeof r?.duration_ms === "number" ? r.duration_ms : undefined,
    startedAt: r?.started_at ? new Date(r.started_at) : undefined,
    completedAt: r?.completed_at ? new Date(r.completed_at) : undefined,
    createdAt: r?.created_at ? new Date(r.created_at) : undefined,
  };
}

export async function fetchStepResults(params: {
  runUuid?: string;
  runId?: string | number;
  limit?: number;
  offset?: number;
}): Promise<StepResultsResponse> {
  const limit = params.limit ?? 200;
  const offset = params.offset ?? 0;

  if (isDemoMode()) {
    return {
      data: [],
      pagination: { limit, offset, total: 0 },
    };
  }

  const query: Record<string, any> = { limit, offset };
  if (params.runUuid) query.run_uuid = params.runUuid;
  if (!query.run_uuid && params.runId !== undefined) query.run_id = params.runId;

  const res = await http.get(`${API_PREFIX}/step-results`, { params: query });
  const list = (res.data?.data || []) as Array<any>;
  const pagination = res.data?.pagination || {};

  return {
    data: list.map(mapStepResultFromApi),
    pagination: {
      limit: pagination?.limit ?? limit,
      offset: pagination?.offset ?? offset,
      total: pagination?.total ?? list.length,
    },
  };
}
