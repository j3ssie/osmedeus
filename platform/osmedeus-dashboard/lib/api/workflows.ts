import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode, setDemoMode } from "./demo-mode";
import type { Workflow } from "@/lib/types/workflow";
import { MOCK_WORKFLOW_YAMLS } from "@/lib/mock/workflow-yamls";
import * as yaml from "js-yaml";

function getCustomMockYamls(): Record<string, string> {
  if (typeof window === "undefined") return {};
  try {
    const raw = window.localStorage.getItem("osmedeus_custom_workflows");
    if (!raw) return {};
    const obj = JSON.parse(raw);
    if (!obj || typeof obj !== "object") return {};
    const out: Record<string, string> = {};
    Object.entries(obj as Record<string, unknown>).forEach(([k, v]) => {
      if (typeof v !== "string") return;
      if (!v.trim()) return;
      out[String(k)] = v;
    });
    return out;
  } catch {
    return {};
  }
}

function getAllMockYamls(): Record<string, string> {
  const custom = getCustomMockYamls();
  return { ...MOCK_WORKFLOW_YAMLS, ...custom };
}

type MockYamlEntry = { id: string; content: string; source: "builtin" | "custom" };

function getMockYamlEntries(): MockYamlEntry[] {
  const out: MockYamlEntry[] = [];
  Object.entries(MOCK_WORKFLOW_YAMLS).forEach(([id, content]) => {
    if (typeof content !== "string" || !content.trim()) return;
    out.push({ id, content, source: "builtin" });
  });
  const custom = getCustomMockYamls();
  Object.entries(custom).forEach(([id, content]) => {
    if (typeof content !== "string" || !content.trim()) return;
    out.push({ id, content, source: "custom" });
  });
  return out;
}

function resolveMockYamlContent(idOrName: string): string | null {
  const all = getAllMockYamls();
  const direct = all[idOrName];
  if (typeof direct === "string" && direct.trim()) return direct;

  const entries = getMockYamlEntries().slice().reverse();
  for (const { id: fallbackId, content } of entries) {
    let doc: any = {};
    try {
      doc = yaml.load(content) || {};
    } catch {
      doc = {};
    }
    const name = typeof doc?.name === "string" ? doc.name.trim() : "";
    if (name && name === idOrName) return content;
    if (fallbackId === idOrName) return content;
  }

  return null;
}

function getUniqueMockWorkflows(): Workflow[] {
  const entries = getMockYamlEntries();
  const byName = new Map<string, { wf: Workflow; source: "builtin" | "custom" }>();
  const order: string[] = [];

  entries.forEach(({ id, content, source }) => {
    const wf = toWorkflowFromYaml(id, content);
    const key = (wf.name || "").trim() || id;
    const existing = byName.get(key);
    if (!existing) {
      byName.set(key, { wf, source });
      order.push(key);
      return;
    }
    if (existing.source === "builtin" && source === "custom") {
      byName.set(key, { wf, source });
    }
  });

  return order.map((k) => byName.get(k)!.wf);
}

function normalizeTags(raw: unknown): string[] {
  if (Array.isArray(raw)) {
    return raw
      .filter((t): t is string => typeof t === "string")
      .map((t) => t.trim())
      .filter(Boolean);
  }

  if (typeof raw === "string") {
    return raw
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);
  }

  return [];
}

function addMockDataTag(tags: string[]): string[] {
  const set = new Set(tags);
  set.add("mock-data");
  return Array.from(set);
}

function getHttpErrorCode(e: unknown): number {
  const msg = e instanceof Error ? e.message : "";
  const code = parseInt((msg.split(":")[0] || "0") as string, 10);
  return Number.isFinite(code) ? code : 0;
}

function enableDemoMode(): void {
  if (typeof window !== "undefined") {
    setDemoMode(true);
  }
}

function toWorkflowFromYaml(id: string, content: string): Workflow {
  let doc: any = {};
  try {
    doc = yaml.load(content) || {};
  } catch {
    doc = {};
  }
  const steps = Array.isArray(doc?.steps) ? doc.steps : [];
  const modules = Array.isArray(doc?.modules) ? doc.modules : [];
  const kind: "module" | "flow" = doc?.kind === "flow" ? "flow" : "module";
  const name = typeof doc?.name === "string" ? (doc.name as string) : id;
  const description = typeof doc?.description === "string" ? (doc.description as string) : "";
  const tags = addMockDataTag(normalizeTags(doc?.tags));
  const params = Array.isArray(doc?.params) ? doc.params : [];
  return {
    name,
    kind,
    description,
    tags,
    file_path: "",
    params,
    required_params: params.filter((p: any) => p?.required).map((p: any) => p?.name ?? ""),
    step_count: steps.length,
    module_count: modules.length,
    checksum: "",
    indexed_at: new Date().toISOString(),
  };
}

function getMockWorkflowTags(): string[] {
  const tagSet = new Set<string>();
  Object.values(getAllMockYamls()).forEach((content) => {
    try {
      const doc: any = yaml.load(content) || {};
      normalizeTags(doc?.tags).forEach((t) => tagSet.add(t));
    } catch {}
  });
  tagSet.add("mock-data");
  return Array.from(tagSet.values()).sort();
}

/**
 * Fetch all workflows
 */
export async function fetchWorkflows(): Promise<Workflow[]> {
  if (isDemoMode()) {
    return getUniqueMockWorkflows();
  }
  const res = await http.get(`${API_PREFIX}/workflows`);
  const data = (res.data?.data || []) as Array<any>;
  return data.map((w) => ({
    name: w.name ?? "",
    kind: w.kind === "flow" ? "flow" : "module",
    description: w.description ?? "",
    tags: Array.isArray(w.tags) ? w.tags : [],
    file_path: w.file_path ?? "",
    params: Array.isArray(w.params) ? w.params : [],
    required_params: Array.isArray(w.required_params) ? w.required_params : [],
    step_count: w.step_count ?? 0,
    module_count: w.module_count ?? 0,
    checksum: w.checksum ?? "",
    indexed_at: w.indexed_at ?? "",
  }));
}

export interface FetchWorkflowsParams {
  source?: "db" | "filesystem";
  tags?: string[];
  kind?: "module" | "flow";
  search?: string;
  offset?: number;
  limit?: number;
}

export interface WorkflowListResult {
  items: Workflow[];
  pagination: {
    total: number;
    offset: number;
    limit: number;
  };
}

export async function fetchMockWorkflowsList(params: FetchWorkflowsParams = {}): Promise<WorkflowListResult> {
  const all = getUniqueMockWorkflows();

  const filtered = all.filter((wf) => {
    if (params.kind && wf.kind !== params.kind) return false;
    if (params.tags && params.tags.length > 0) {
      const tagSet = new Set((wf.tags || []).map((t) => String(t)));
      if (!params.tags.some((t) => tagSet.has(t))) return false;
    }
    if (params.search && params.search.trim()) {
      const q = params.search.trim().toLowerCase();
      const hay = `${wf.name ?? ""} ${wf.description ?? ""} ${(wf.tags || []).join(" ")}`.toLowerCase();
      if (!hay.includes(q)) return false;
    }
    return true;
  });

  const offset = typeof params.offset === "number" ? params.offset : 0;
  const limit = typeof params.limit === "number" ? params.limit : filtered.length;
  const paged = filtered.slice(Math.max(0, offset), Math.max(0, offset) + Math.max(0, limit));

  return {
    items: paged,
    pagination: {
      total: filtered.length,
      offset,
      limit,
    },
  };
}

export async function fetchWorkflowsList(params: FetchWorkflowsParams = {}): Promise<WorkflowListResult> {
  if (isDemoMode()) {
    const items = await fetchWorkflows();

    const filtered = items.filter((wf) => {
      if (params.kind && wf.kind !== params.kind) return false;
      if (params.tags && params.tags.length > 0) {
        const tagSet = new Set((wf.tags || []).map((t) => String(t)));
        if (!params.tags.some((t) => tagSet.has(t))) return false;
      }
      if (params.search && params.search.trim()) {
        const q = params.search.trim().toLowerCase();
        const hay = `${wf.name ?? ""} ${wf.description ?? ""} ${(wf.tags || []).join(" ")}`.toLowerCase();
        if (!hay.includes(q)) return false;
      }
      return true;
    });

    const offset = typeof params.offset === "number" ? params.offset : 0;
    const limit = typeof params.limit === "number" ? params.limit : filtered.length;
    const paged = filtered.slice(Math.max(0, offset), Math.max(0, offset) + Math.max(0, limit));
    return {
      items: paged,
      pagination: { total: filtered.length, offset, limit },
    };
  }
  const query: Record<string, string | number | boolean> = {};
  if (params.source) query.source = params.source;
  if (params.tags && params.tags.length > 0) query.tags = params.tags.join(",");
  if (params.kind) query.kind = params.kind;
  if (params.search) query.search = params.search;
  if (typeof params.offset === "number") query.offset = params.offset;
  if (typeof params.limit === "number") query.limit = params.limit;
  try {
    const res = await http.get(`${API_PREFIX}/workflows`, { params: query });
    const data = (res.data?.data || []) as Array<any>;
    const pagination = res.data?.pagination || { total: data.length, offset: 0, limit: data.length };
    const items: Workflow[] = data.map((w) => ({
      name: w.name ?? "",
      kind: w.kind === "flow" ? "flow" : "module",
      description: w.description ?? "",
      tags: Array.isArray(w.tags) ? w.tags.map((t: any) => String(t)) : [],
      file_path: w.file_path ?? "",
      params: Array.isArray(w.params) ? w.params : [],
      required_params: Array.isArray(w.required_params) ? w.required_params : [],
      step_count: w.step_count ?? 0,
      module_count: w.module_count ?? 0,
      checksum: w.checksum ?? "",
      indexed_at: w.indexed_at ?? "",
    }));
    return {
      items,
      pagination: {
        total: Number(pagination.total) || items.length,
        offset: Number(pagination.offset) || 0,
        limit: Number(pagination.limit) || items.length,
      },
    };
  } catch (e) {
    const code = getHttpErrorCode(e);
    if (code === 0) {
      enableDemoMode();
      return fetchMockWorkflowsList({
        kind: params.kind,
        tags: params.tags,
        search: params.search,
        offset: params.offset,
        limit: params.limit,
      });
    }
    throw e;
  }
}

/**
 * Fetch a single workflow by ID
 */
export async function fetchWorkflow(id: string): Promise<Workflow | null> {
  if (isDemoMode()) {
    const content = resolveMockYamlContent(id);
    if (!content) return null;
    return toWorkflowFromYaml(id, content);
  }
  try {
    const res = await http.get(`${API_PREFIX}/workflows/${encodeURIComponent(id)}`, {
      params: { json: true },
    });
    const w = res.data;
    if (typeof w === "string") {
      return toWorkflowFromYaml(id, w);
    }
    return {
      name: w.name ?? "",
      kind: w.kind === "flow" ? "flow" : "module",
      description: w.description ?? "",
      tags: Array.isArray(w.tags) ? w.tags : [],
      file_path: w.file_path ?? "",
      params: Array.isArray(w.params) ? w.params : [],
      required_params: Array.isArray(w.required_params) ? w.required_params : [],
      step_count: Array.isArray(w.steps) ? w.steps.length : w.step_count ?? 0,
      module_count: w.module_count ?? 0,
      checksum: w.checksum ?? "",
      indexed_at: w.indexed_at ?? "",
    };
  } catch (e: any) {
    const code = getHttpErrorCode(e);
    if (code === 404) throw new Error("WORKFLOW_NOT_FOUND");
    if (code === 401) throw new Error("UNAUTHORIZED");
    if (code === 0) {
      enableDemoMode();
      const content = resolveMockYamlContent(id);
      return content ? toWorkflowFromYaml(id, content) : null;
    }
    throw new Error("REQUEST_FAILED");
  }
}

/**
 * Fetch workflow YAML content
 */
export async function fetchWorkflowYaml(id: string): Promise<string | null> {
  if (isDemoMode()) {
    return resolveMockYamlContent(id);
  }
  try {
    const res = await http.get(`${API_PREFIX}/workflows/${encodeURIComponent(id)}`, {
      responseType: "text",
    });
    return typeof res.data === "string" ? res.data : (res.data?.yaml ?? null);
  } catch (e: any) {
    const code = getHttpErrorCode(e);
    if (code === 404) throw new Error("WORKFLOW_NOT_FOUND");
    if (code === 401) throw new Error("UNAUTHORIZED");
    if (code === 0) {
      enableDemoMode();
      return resolveMockYamlContent(id);
    }
    throw new Error("REQUEST_FAILED");
  }
}

export async function fetchWorkflowTags(): Promise<string[]> {
  if (isDemoMode()) {
    return getMockWorkflowTags();
  }
  try {
    const res = await http.get(`${API_PREFIX}/workflows/tags`);
    const tags = res.data?.tags || [];
    return Array.isArray(tags) ? tags.map((t: any) => String(t)) : [];
  } catch (e) {
    const code = getHttpErrorCode(e);
    if (code === 0) {
      enableDemoMode();
      return getMockWorkflowTags();
    }
    throw e;
  }
}

export async function refreshWorkflowIndex(force = false): Promise<{
  message: string;
  added: number;
  updated: number;
  removed: number;
  errors: unknown[];
}> {
  const res = await http.post(`${API_PREFIX}/workflows/refresh`, undefined, {
    params: force ? { force: true } : {},
  });
  return {
    message: res.data?.message || "",
    added: Number(res.data?.added || 0),
    updated: Number(res.data?.updated || 0),
    removed: Number(res.data?.removed || 0),
    errors: Array.isArray(res.data?.errors) ? res.data.errors : [],
  };
}

/**
 * Save workflow YAML content
 */
export async function saveWorkflowYaml(
  id: string,
  yamlText: string
): Promise<boolean> {
  if (!id || !yamlText.trim()) return false;

  if (isDemoMode()) {
    if (typeof window === "undefined") return false;
    try {
      const raw = window.localStorage.getItem("osmedeus_custom_workflows");
      const obj = raw ? JSON.parse(raw) : {};
      const next = obj && typeof obj === "object" ? (obj as Record<string, unknown>) : {};
      next[id] = yamlText;
      window.localStorage.setItem("osmedeus_custom_workflows", JSON.stringify(next));
      return true;
    } catch {
      return false;
    }
  }

  try {
    if (typeof window === "undefined") return false;
    let name = id;
    let kind = "module";
    try {
      const doc: any = yaml.load(yamlText) || {};
      if (typeof doc?.name === "string" && doc.name.trim()) name = doc.name.trim();
      if (doc?.kind === "flow") kind = "flow";
    } catch {
    }

    const form = new FormData();
    const fileName = `${name || id}.yaml`;
    const blob = new Blob([yamlText], { type: "text/yaml" });
    form.append("file", blob, fileName);

    await http.post(`${API_PREFIX}/workflow-upload`, form, {
      headers: { "Content-Type": "multipart/form-data" },
      params: { kind },
    });
    return true;
  } catch (e: any) {
    const code = getHttpErrorCode(e);
    if (code === 0) {
      setDemoMode(true);
      return saveWorkflowYaml(id, yamlText);
    }
    return false;
  }
}
