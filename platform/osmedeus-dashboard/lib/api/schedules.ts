import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import type { Schedule } from "@/lib/types/schedule";
import { isDemoMode } from "./demo-mode";

let demoSchedules: Schedule[] = [
  {
    id: "schedule-001",
    name: "daily-full-recon",
    workflowName: "full-recon",
    triggerType: "cron",
    schedule: "0 2 * * *",
    inputConfig: { target: "example.com" },
    isEnabled: true,
    runCount: 12,
    lastRun: new Date(Date.now() - 24 * 60 * 60 * 1000),
    nextRun: new Date(Date.now() + 2 * 60 * 60 * 1000),
    createdAt: new Date(Date.now() - 20 * 24 * 60 * 60 * 1000),
    updatedAt: new Date(Date.now() - 24 * 60 * 60 * 1000),
  },
  {
    id: "schedule-002",
    name: "hourly-http-probe",
    workflowName: "http-probe",
    triggerType: "cron",
    schedule: "0 * * * *",
    inputConfig: { target: "api.example.com" },
    isEnabled: true,
    runCount: 120,
    lastRun: new Date(Date.now() - 60 * 60 * 1000),
    nextRun: new Date(Date.now() + 60 * 60 * 1000),
    createdAt: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000),
    updatedAt: new Date(Date.now() - 60 * 60 * 1000),
  },
  {
    id: "schedule-003",
    name: "weekly-vuln-scan",
    workflowName: "vulnerability-scan",
    triggerType: "cron",
    schedule: "0 0 * * 0",
    inputConfig: { target: "shop.retail.com" },
    isEnabled: false,
    runCount: 3,
    lastRun: new Date(Date.now() - 8 * 24 * 60 * 60 * 1000),
    createdAt: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
    updatedAt: new Date(Date.now() - 8 * 24 * 60 * 60 * 1000),
  },
];

export interface FetchSchedulesResult {
  data: Schedule[];
  pagination: {
    total: number;
    offset: number;
    limit: number;
  };
}

function mapSchedule(s: any): Schedule {
  return {
    id: String(s.id ?? ""),
    name: s.name ?? "",
    workflowName: s.workflow_name ?? "",
    workflowPath: s.workflow_path ?? undefined,
    triggerName: s.trigger_name ?? undefined,
    triggerType: s.trigger_type ?? undefined,
    schedule: s.schedule ?? undefined,
    eventTopic: s.event_topic ?? undefined,
    watchPath: s.watch_path ?? undefined,
    inputConfig: s.input_config && typeof s.input_config === "object" ? s.input_config : undefined,
    isEnabled: !!(s.is_enabled ?? s.enabled),
    lastRun: s.last_run ? new Date(s.last_run) : undefined,
    nextRun: s.next_run ? new Date(s.next_run) : undefined,
    runCount: typeof s.run_count === "number" ? s.run_count : undefined,
    createdAt: s.created_at ? new Date(s.created_at) : undefined,
    updatedAt: s.updated_at ? new Date(s.updated_at) : undefined,
  };
}

export async function fetchSchedules(params?: { offset?: number; limit?: number }): Promise<FetchSchedulesResult> {
  if (isDemoMode()) {
    const offset = params?.offset ?? 0;
    const limit = params?.limit ?? 50;
    const data = demoSchedules.slice(offset, offset + limit);
    return {
      data,
      pagination: { total: demoSchedules.length, offset, limit },
    };
  }
  const res = await http.get(`${API_PREFIX}/schedules`, {
    params: { offset: params?.offset ?? 0, limit: params?.limit ?? 50 },
  });
  const data = Array.isArray(res.data?.data) ? res.data.data : [];
  const paginationRaw = res.data?.pagination ?? {};
  const total = Number(paginationRaw.total) || data.length;
  const offset = Number(paginationRaw.offset) || (params?.offset ?? 0);
  const limit = Number(paginationRaw.limit) || (params?.limit ?? 50);
  return {
    data: data.map(mapSchedule),
    pagination: { total, offset, limit },
  };
}

export async function createSchedule(input: {
  name: string;
  workflowName: string;
  workflowKind?: "flow" | "module";
  target: string;
  schedule: string;
  enabled?: boolean;
  params?: Record<string, any>;
  runner_type?: string;
}): Promise<Schedule> {
  if (isDemoMode()) {
    const now = new Date();
    const schedule: Schedule = {
      id: `schedule-${Math.random().toString(16).slice(2, 10)}`,
      name: input.name,
      workflowName: input.workflowName,
      triggerType: "cron",
      schedule: input.schedule,
      inputConfig: { ...(input.params ?? {}), target: input.target },
      isEnabled: input.enabled ?? true,
      runCount: 0,
      createdAt: now,
      updatedAt: now,
    };
    demoSchedules = [schedule, ...demoSchedules];
    return schedule;
  }
  const body = {
    name: input.name,
    workflow_name: input.workflowName,
    workflow_kind: input.workflowKind ?? "flow",
    target: input.target,
    schedule: input.schedule,
    enabled: input.enabled ?? true,
    params: input.params,
    runner_type: input.runner_type,
  };
  const res = await http.post(`${API_PREFIX}/schedules`, body);
  const s = res.data?.data ?? {};
  return {
    id: String(s.id ?? ""),
    name: s.name ?? input.name,
    workflowName: s.workflow_name ?? input.workflowName,
    triggerType: "cron",
    schedule: s.schedule ?? input.schedule,
    inputConfig: {
      ...(input.params ?? {}),
      target: input.target,
    },
    isEnabled: !!(s.is_enabled ?? s.enabled ?? input.enabled ?? true),
    runCount: 0,
  };
}

export async function getSchedule(id: string): Promise<Schedule | null> {
  if (isDemoMode()) {
    return demoSchedules.find((s) => s.id === id) ?? null;
  }
  try {
    const res = await http.get(`${API_PREFIX}/schedules/${encodeURIComponent(id)}`);
    const s = res.data?.data;
    if (!s) return null;
    return mapSchedule(s);
  } catch {
    return null;
  }
}

export async function updateSchedule(id: string, patch: Partial<{ name: string; schedule: string; enabled: boolean }>): Promise<boolean> {
  if (isDemoMode()) {
    const idx = demoSchedules.findIndex((s) => s.id === id);
    if (idx === -1) return false;
    const now = new Date();
    const next = [...demoSchedules];
    next[idx] = {
      ...next[idx],
      name: typeof patch.name === "string" ? patch.name : next[idx].name,
      schedule: typeof patch.schedule === "string" ? patch.schedule : next[idx].schedule,
      isEnabled: typeof patch.enabled === "boolean" ? patch.enabled : next[idx].isEnabled,
      updatedAt: now,
    };
    demoSchedules = next;
    return true;
  }
  try {
    await http.put(`${API_PREFIX}/schedules/${encodeURIComponent(id)}`, patch);
    return true;
  } catch {
    return false;
  }
}

export async function deleteSchedule(id: string): Promise<boolean> {
  if (isDemoMode()) {
    const before = demoSchedules.length;
    demoSchedules = demoSchedules.filter((s) => s.id !== id);
    return demoSchedules.length !== before;
  }
  try {
    await http.delete(`${API_PREFIX}/schedules/${encodeURIComponent(id)}`);
    return true;
  } catch {
    return false;
  }
}

export async function enableSchedule(id: string): Promise<boolean> {
  if (isDemoMode()) {
    return updateSchedule(id, { enabled: true });
  }
  try {
    await http.post(`${API_PREFIX}/schedules/${encodeURIComponent(id)}/enable`);
    return true;
  } catch {
    return false;
  }
}

export async function disableSchedule(id: string): Promise<boolean> {
  if (isDemoMode()) {
    return updateSchedule(id, { enabled: false });
  }
  try {
    await http.post(`${API_PREFIX}/schedules/${encodeURIComponent(id)}/disable`);
    return true;
  } catch {
    return false;
  }
}

export async function triggerSchedule(id: string): Promise<boolean> {
  if (isDemoMode()) {
    const idx = demoSchedules.findIndex((s) => s.id === id);
    if (idx === -1) return false;
    const now = new Date();
    const next = [...demoSchedules];
    next[idx] = {
      ...next[idx],
      lastRun: now,
      nextRun: next[idx].schedule ? new Date(now.getTime() + 60 * 60 * 1000) : next[idx].nextRun,
      runCount: (next[idx].runCount ?? 0) + 1,
      updatedAt: now,
    };
    demoSchedules = next;
    return true;
  }
  try {
    await http.post(`${API_PREFIX}/schedules/${encodeURIComponent(id)}/trigger`);
    return true;
  } catch {
    return false;
  }
}
