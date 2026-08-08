import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import type { PaginatedResponse } from "@/lib/types/api";
import type { EventLogItem } from "@/lib/types/event-log";
import { isDemoMode } from "./demo-mode";

const demoEventLogs: EventLogItem[] = [
  {
    id: "evt-001",
    topic: "webhook.received",
    eventId: "evt-001",
    name: "push",
    source: "github",
    dataType: "json",
    data: '{"repository":"example.com","ref":"refs/heads/main"}',
    workspace: "example.com",
    runId: "run-demo-001",
    workflowName: "subdomain-enum",
    processed: true,
    createdAt: new Date(Date.now() - 30 * 60 * 1000),
    processedAt: new Date(Date.now() - 29 * 60 * 1000),
  },
  {
    id: "evt-002",
    topic: "tasks.started",
    eventId: "evt-002",
    name: "scan_started",
    source: "osmedeus",
    dataType: "text",
    data: "scan started",
    workspace: "example.com",
    runId: "run-demo-002",
    workflowName: "full-recon",
    processed: true,
    createdAt: new Date(Date.now() - 12 * 60 * 1000),
    processedAt: new Date(Date.now() - 11 * 60 * 1000),
  },
  {
    id: "evt-003",
    topic: "tasks.failed",
    eventId: "evt-003",
    name: "scan_failed",
    source: "osmedeus",
    dataType: "text",
    data: "task failed",
    workspace: "acme.io",
    runId: "run-demo-003",
    workflowName: "vulnerability-scan",
    processed: false,
    createdAt: new Date(Date.now() - 5 * 60 * 1000),
    error: "Connection timeout after 5 retries",
  },
];

export interface EventLogFilters {
  topic?: string;
  name?: string;
  source?: string;
  workspace?: string;
  runId?: string;
  workflowName?: string;
  processed?: boolean;
}

export async function fetchEventLogs(params: {
  page?: number;
  pageSize?: number;
  filters?: EventLogFilters;
}): Promise<PaginatedResponse<EventLogItem>> {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 20;
  const offset = (page - 1) * pageSize;
  const filters = params.filters ?? {};

  if (isDemoMode()) {
    const topicQ = (filters.topic ?? "").trim().toLowerCase();
    const nameQ = (filters.name ?? "").trim().toLowerCase();
    const sourceQ = (filters.source ?? "").trim().toLowerCase();
    const workspaceQ = (filters.workspace ?? "").trim().toLowerCase();
    const runIdQ = (filters.runId ?? "").trim().toLowerCase();
    const workflowQ = (filters.workflowName ?? "").trim().toLowerCase();

    const filtered = demoEventLogs.filter((e) => {
      if (topicQ && !e.topic.toLowerCase().includes(topicQ)) return false;
      if (nameQ && !e.name.toLowerCase().includes(nameQ)) return false;
      if (sourceQ && !e.source.toLowerCase().includes(sourceQ)) return false;
      if (workspaceQ && String(e.workspace ?? "").toLowerCase() !== workspaceQ) return false;
      if (runIdQ && String(e.runId ?? "").toLowerCase() !== runIdQ) return false;
      if (workflowQ && String(e.workflowName ?? "").toLowerCase() !== workflowQ) return false;
      if (typeof filters.processed === "boolean" && e.processed !== filters.processed) return false;
      return true;
    });

    const sliced = filtered.slice(offset, offset + pageSize);
    return {
      data: sliced,
      pagination: {
        page,
        pageSize,
        totalItems: filtered.length,
        totalPages: Math.ceil(filtered.length / pageSize),
      },
    };
  }

  const query: Record<string, any> = { offset, limit: pageSize };
  if (filters.topic) query.topic = filters.topic;
  if (filters.name) query.name = filters.name;
  if (filters.source) query.source = filters.source;
  if (filters.workspace) query.workspace = filters.workspace;
  if (filters.runId) query.run_id = filters.runId;
  if (filters.workflowName) query.workflow_name = filters.workflowName;
  if (typeof filters.processed === "boolean") query.processed = String(filters.processed);

  const res = await http.get(`${API_PREFIX}/event-logs`, { params: query });
  const list = (res.data?.data || []) as Array<any>;
  const mapped: EventLogItem[] = list.map((e) => ({
    id: String(e.id ?? e.event_id ?? `${e.topic}-${e.created_at}`),
    topic: e.topic ?? "",
    eventId: e.event_id ?? "",
    name: e.name ?? "",
    source: e.source ?? "",
    dataType: e.data_type,
    data: e.data,
    workspace: e.workspace,
    runId: e.run_id,
    workflowName: e.workflow_name,
    processed: !!e.processed,
    createdAt: e.created_at ? new Date(e.created_at) : new Date(),
    processedAt: e.processed_at ? new Date(e.processed_at) : undefined,
    error: e.error,
  }));
  const total = res.data?.pagination?.total ?? mapped.length;
  const respLimit = res.data?.pagination?.limit ?? pageSize;
  const respOffset = res.data?.pagination?.offset ?? offset;
  const currentPage = Math.floor(respOffset / respLimit) + 1;
  const pageSz = respLimit;
  return {
    data: mapped,
    pagination: {
      page: currentPage,
      pageSize: pageSz,
      totalItems: total,
      totalPages: Math.ceil(total / pageSz),
    },
  };
}

export async function clearEventLogTables(): Promise<void> {
  if (isDemoMode()) {
    return;
  }

  await http.post(`${API_PREFIX}/database/tables/runs/clear`, { force: true });
  await http.post(`${API_PREFIX}/database/tables/step_results/clear`, { force: true });
  await http.post(`${API_PREFIX}/database/tables/event_logs/clear`, { force: true });
}
