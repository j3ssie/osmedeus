import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";

const demoWorkers = [
  {
    id: "worker-001",
    name: "demo-worker-1",
    status: "online",
    last_seen: new Date().toISOString(),
    runner: "local",
  },
  {
    id: "worker-002",
    name: "demo-worker-2",
    status: "idle",
    last_seen: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
    runner: "docker",
  },
];

export async function getWorkers() {
  if (isDemoMode()) {
    return { data: demoWorkers };
  }
  const res = await http.get(`${API_PREFIX}/workers`);
  return res.data;
}

export async function getWorker(id: string) {
  if (isDemoMode()) {
    const w = demoWorkers.find((x) => x.id === id);
    return w ? { data: w } : { data: null };
  }
  const res = await http.get(`${API_PREFIX}/workers/${encodeURIComponent(id)}`);
  return res.data;
}

export async function getTasks() {
  const res = await http.get(`${API_PREFIX}/tasks`);
  return res.data;
}

export async function getTask(id: string) {
  const res = await http.get(`${API_PREFIX}/tasks/${encodeURIComponent(id)}`);
  return res.data;
}

export async function submitTask(payload: {
  workflow_name: string;
  workflow_kind: "flow" | "module";
  target: string;
  params?: Record<string, any>;
}) {
  if (isDemoMode()) {
    return {
      message: "Demo mode: task submitted",
      task_id: `task-${Date.now()}`,
    };
  }
  const res = await http.post(`${API_PREFIX}/tasks`, payload);
  return res.data;
}
