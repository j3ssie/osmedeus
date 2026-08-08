import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";
import type { DashboardStats } from "@/lib/types/asset";
import { mockStats } from "@/lib/mock/data/stats";
import type { SystemStats } from "@/lib/types/stats";
import { mockSystemStats } from "@/lib/mock/data/system-stats";

/**
 * Fetch dashboard statistics
 *
 * To replace with real API:
 * const response = await fetch(`${API_CONFIG.baseUrl}/stats`, { headers: getHeaders() });
 * return response.json();
 */
export async function fetchDashboardStats(): Promise<DashboardStats> {
  if (isDemoMode()) {
    return mockStats;
  }
  const statsRes = await http.get(`${API_PREFIX}/stats`);
  const s = statsRes.data || {};
  const workflows = s?.workflows?.total ?? 0;
  const workspaces = s?.workspaces?.total ?? 0;
  const assetsTotal = s?.assets?.total ?? 0;
  const vulnerabilities = s?.vulnerabilities?.total ?? 0;
  let subdomains = s?.subdomains?.total ?? 0;
  if (!subdomains) {
    try {
      const wsRes = await http.get(`${API_PREFIX}/workspaces`, { params: { offset: 0, limit: 1000 } });
      const wsList = Array.isArray(wsRes.data?.data) ? wsRes.data.data : [];
      subdomains = wsList.reduce((sum: number, w: any) => sum + (w.total_subdomains ?? 0), 0);
    } catch {
      subdomains = 0;
    }
  }
  return {
    workflows,
    workspaces,
    subdomains,
    httpAssets: assetsTotal,
    vulnerabilities,
  };
}

export async function fetchSystemStats(): Promise<SystemStats> {
  if (isDemoMode()) {
    return mockSystemStats;
  }
  const res = await http.get(`${API_PREFIX}/stats`);
  const data = res.data || {};
  return {
    workflows: {
      total: Number(data?.workflows?.total ?? 0),
      flows: Number(data?.workflows?.flows ?? 0),
      modules: Number(data?.workflows?.modules ?? 0),
    },
    scans: {
      total: Number(data?.scans?.total ?? 0),
      completed: Number(data?.scans?.completed ?? 0),
      running: Number(data?.scans?.running ?? 0),
      failed: Number(data?.scans?.failed ?? 0),
    },
    workspaces: {
      total: Number(data?.workspaces?.total ?? 0),
    },
    assets: {
      total: Number(data?.assets?.total ?? 0),
    },
    vulnerabilities: {
      total: Number(data?.vulnerabilities?.total ?? 0),
      critical: Number(data?.vulnerabilities?.critical ?? 0),
      high: Number(data?.vulnerabilities?.high ?? 0),
      medium: Number(data?.vulnerabilities?.medium ?? 0),
      low: Number(data?.vulnerabilities?.low ?? 0),
    },
    schedules: {
      total: Number(data?.schedules?.total ?? 0),
      enabled: Number(data?.schedules?.enabled ?? 0),
    },
  };
}
