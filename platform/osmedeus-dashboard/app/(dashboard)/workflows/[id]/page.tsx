import WorkflowEditorClient from "@/components/workflow-editor/workflow-editor-client";
import fs from "fs";
import path from "path";
import * as yaml from "js-yaml";

// server component wrapper

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
export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <WorkflowEditorClient workflowId={id} />;
}
