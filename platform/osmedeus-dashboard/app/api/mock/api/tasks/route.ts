import { NextResponse } from "next/server";
import { mockScans } from "@/lib/mock/data/scans";

export const dynamic = "force-static";

export async function GET() {
  const running = mockScans
    .filter((s) => s.status === "running" || s.status === "pending")
    .map((s) => ({
      id: s.id,
      workflow_name: s.workflowName,
      workflow_kind: s.workflowKind,
      target: s.target,
      status: s.status === "pending" ? "pending" : "running",
      worker_id: "worker-001",
      created_at: s.startedAt?.toISOString() ?? s.createdAt.toISOString(),
      started_at: s.startedAt?.toISOString() ?? s.createdAt.toISOString(),
    }));

  const completed = mockScans
    .filter((s) => s.status === "completed" || s.status === "failed" || s.status === "cancelled")
    .map((s) => ({
      task_id: s.id,
      status: s.status === "failed" ? "failed" : s.status,
      output: s.status === "completed" ? "Scan completed successfully" : "Task ended",
      completed_at: s.completedAt?.toISOString() ?? new Date().toISOString(),
      workflow: s.workflowName,
      exports: {},
    }));

  return NextResponse.json({ running, completed });
}
