import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import type { EventLog } from "@/lib/types/event";

export async function fetchEvents(): Promise<EventLog[]> {
  const res = await http.get(`${API_PREFIX}/tasks`);
  const running = Array.isArray(res.data?.running) ? res.data.running : [];
  const completed = Array.isArray(res.data?.completed) ? res.data.completed : [];
  const runningMapped: EventLog[] = running.map((t: any) => ({
    id: String(t.id),
    workflowName: t.workflow_name,
    workflowKind: t.workflow_kind,
    target: t.target,
    status: "running",
    createdAt: t.created_at ? new Date(t.created_at) : undefined,
    startedAt: t.started_at ? new Date(t.started_at) : undefined,
  }));
  const completedMapped: EventLog[] = completed.map((t: any) => ({
    id: String(t.task_id),
    workflowName: t.workflow || t.workflow_name,
    workflowKind: t.workflow_kind,
    target: t.target || "",
    status: t.status === "failed" ? "failed" : "completed",
    completedAt: t.completed_at ? new Date(t.completed_at) : undefined,
    output: t.output,
    error: t.status === "failed" ? "Task failed" : undefined,
  }));
  return [...runningMapped, ...completedMapped].sort(
    (a, b) => (b.startedAt?.getTime() || b.completedAt?.getTime() || 0) - (a.startedAt?.getTime() || a.completedAt?.getTime() || 0)
  );
}

export async function fetchEvent(id: string): Promise<EventLog | null> {
  try {
    const res = await http.get(`${API_PREFIX}/tasks/${encodeURIComponent(id)}`);
    const t = res.data;
    const isCompleted = !!t.completed_at || t.status === "completed" || t.status === "failed";
    return {
      id: String(t.id || t.task_id || id),
      workflowName: t.workflow_name || t.workflow || "",
      workflowKind: t.workflow_kind,
      target: t.target || "",
      status: isCompleted ? (t.status === "failed" ? "failed" : "completed") : "running",
      createdAt: t.created_at ? new Date(t.created_at) : undefined,
      startedAt: t.started_at ? new Date(t.started_at) : undefined,
      completedAt: t.completed_at ? new Date(t.completed_at) : undefined,
      output: t.output,
      error: t.status === "failed" ? "Task failed" : undefined,
    };
  } catch {
    return null;
  }
}

