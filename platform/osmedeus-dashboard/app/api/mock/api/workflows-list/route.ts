import { NextResponse } from "next/server";
import fs from "fs";
import path from "path";
import * as yaml from "js-yaml";

export const dynamic = "force-static";

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

function listYamlFiles(rootDir: string, relDir = ""): string[] {
  const fullDir = path.join(rootDir, relDir);
  const entries = fs.readdirSync(fullDir, { withFileTypes: true });
  const out: string[] = [];

  for (const e of entries) {
    if (e.isDirectory()) {
      out.push(...listYamlFiles(rootDir, path.join(relDir, e.name)));
      continue;
    }
    if (e.isFile() && e.name.endsWith(".yaml")) {
      out.push(path.join(relDir, e.name));
    }
  }

  return out;
}

export async function GET() {
  const offset = 0;
  const limit = 50;

  const dir = path.join(process.cwd(), "mock-workflows");
  let relFiles: string[] = [];
  try {
    relFiles = listYamlFiles(dir);
  } catch {
    relFiles = [];
  }

  const items = relFiles
    .map((rel) => {
      const full = path.join(dir, rel);
      let doc: any = {};
      try {
        const content = fs.readFileSync(full, "utf-8");
        doc = yaml.load(content) || {};
      } catch {
        doc = {};
      }

      const fallbackId = path.basename(rel, ".yaml");
      const name = typeof doc?.name === "string" && doc.name.trim() ? doc.name.trim() : fallbackId;
      const wfKind = (doc?.kind as string) === "flow" ? "flow" : "module";
      const description = (doc?.description as string) || `Mock workflow from ${rel}`;
      const steps = Array.isArray(doc?.steps) ? doc.steps : [];
      const modules = Array.isArray(doc?.modules) ? doc.modules : [];
      const rawTags = normalizeTags(doc?.tags);
      const tagSet = new Set(rawTags);
      tagSet.add("mock-data");
      const tags = Array.from(tagSet);
      const params = Array.isArray(doc?.params) ? doc.params : [];
      const required_params = params.filter((p: any) => p?.required).map((p: any) => p?.name ?? "");

      return {
        name,
        kind: wfKind,
        description,
        tags,
        file_path: `/mock-workflows/${rel.replace(/\\/g, "/")}`,
        params,
        required_params,
        step_count: steps.length,
        module_count: modules.length,
        checksum: "",
        indexed_at: new Date().toISOString(),
      };
    });

  const uniqueByName = new Map<string, any>();
  items.forEach((wf) => {
    const key = String(wf.name || "").trim();
    if (!key) return;
    if (!uniqueByName.has(key)) uniqueByName.set(key, wf);
  });
  const uniqueItems = Array.from(uniqueByName.values());

  const sliced = uniqueItems.slice(Math.max(0, offset), Math.max(0, offset) + Math.max(0, limit));

  return NextResponse.json({
    data: sliced,
    pagination: {
      total: uniqueItems.length,
      offset,
      limit,
    },
  });
}

