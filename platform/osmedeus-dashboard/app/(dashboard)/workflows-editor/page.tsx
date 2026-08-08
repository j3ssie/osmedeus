"use client";

import * as React from "react";
import { useRouter, useSearchParams } from "next/navigation";
import WorkflowEditorClient from "@/components/workflow-editor/workflow-editor-client";
import { ErrorState } from "@/components/shared/error-state";
import { fetchWorkflows } from "@/lib/api/workflows";
import type { Workflow } from "@/lib/types/workflow";

export default function WorkflowsEditorPage() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const requested = (searchParams.get("workflow") || "").trim();

  const [workflows, setWorkflows] = React.useState<Workflow[]>([]);
  const [selectedWorkflow, setSelectedWorkflow] = React.useState<string>(requested);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  const syncUrl = React.useCallback(
    (id: string) => {
      const url = id ? `/workflows-editor?workflow=${encodeURIComponent(id)}` : "/workflows-editor";
      router.replace(url);
    },
    [router]
  );

  const load = React.useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const list = await fetchWorkflows();
      const sorted = list.slice().sort((a, b) => a.name.localeCompare(b.name));
      setWorkflows(sorted);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load workflows");
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    load();
  }, [load]);

  React.useEffect(() => {
    if (loading) return;
    if (workflows.length === 0) return;

    const requestedValid = requested && workflows.some((w) => w.name === requested);

    if (requestedValid && requested !== selectedWorkflow) {
      setSelectedWorkflow(requested);
      return;
    }

    if (!selectedWorkflow) {
      const next = requestedValid ? requested : workflows[0]!.name;
      setSelectedWorkflow(next);
      if (next && requested !== next) syncUrl(next);
    }
  }, [loading, requested, selectedWorkflow, syncUrl, workflows]);

  if (error) {
    return (
      <div className="flex h-[calc(100vh-10rem)] items-center justify-center">
        <div className="space-y-4">
          <ErrorState title="Workflow Error" message={error} onRetry={load} />
        </div>
      </div>
    );
  }

  return (
    <div>
      {selectedWorkflow ? <WorkflowEditorClient key={selectedWorkflow} workflowId={selectedWorkflow} /> : null}
    </div>
  );
}
