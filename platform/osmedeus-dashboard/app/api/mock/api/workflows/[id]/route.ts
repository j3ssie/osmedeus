import { NextResponse } from "next/server";
import fs from "fs";
import path from "path";
import * as yaml from "js-yaml";

export const dynamic = "force-static";

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

export async function generateStaticParams() {
  try {
    const dir = path.join(process.cwd(), "mock-workflows");
    const files = listYamlFiles(dir);
    const ids = new Set<string>();

    for (const rel of files) {
      const full = path.join(dir, rel);
      try {
        const content = fs.readFileSync(full, "utf-8");
        const doc: any = yaml.load(content) || {};
        const name = typeof doc?.name === "string" && doc.name.trim() ? doc.name.trim() : "";
        ids.add(name || path.basename(rel, ".yaml"));
      } catch {
        ids.add(path.basename(rel, ".yaml"));
      }
    }

    return Array.from(ids).map((id) => ({ id }));
  } catch {
    return [{ id: "mock-workflow" }];
  }
}

function resolveWorkflowFile(dir: string, id: string): { rel: string; content: string } | null {
  let relFiles: string[] = [];
  try {
    relFiles = listYamlFiles(dir);
  } catch {
    relFiles = [];
  }

  const wanted = id.trim();
  if (!wanted) return null;

  for (const rel of relFiles) {
    const full = path.join(dir, rel);
    try {
      const content = fs.readFileSync(full, "utf-8");
      let doc: any = {};
      try {
        doc = yaml.load(content) || {};
      } catch {
        doc = {};
      }
      const fallbackId = path.basename(rel, ".yaml");
      const name = typeof doc?.name === "string" && doc.name.trim() ? doc.name.trim() : fallbackId;
      if (name === wanted || fallbackId === wanted) {
        return { rel, content };
      }
    } catch {
      continue;
    }
  }

  return null;
}

export async function GET(_request: Request, ctx: { params: Promise<{ id: string }> }) {
  const { id } = await ctx.params;

  const dir = path.join(process.cwd(), "mock-workflows");
  const found = resolveWorkflowFile(dir, id);
  if (!found) {
    return NextResponse.json({ error: true, message: "Workflow not found" }, { status: 404 });
  }

  return new NextResponse(found.content, {
    headers: {
      "content-type": "text/yaml; charset=utf-8",
    },
  });
}
