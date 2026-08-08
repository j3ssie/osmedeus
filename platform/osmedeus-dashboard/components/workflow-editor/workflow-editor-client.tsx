"use client";

import * as React from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { WorkflowCanvas } from "@/components/workflow-editor/workflow-canvas";
import { WorkflowSidebar } from "@/components/workflow-editor/workflow-sidebar";
import {
  parseWorkflowYaml,
  serializeWorkflowToYaml,
  updateStepInWorkflow,
  getFlowModules,
  type ParsedWorkflow,
} from "@/components/workflow-editor/utils/yaml-parser";
import { ErrorState } from "@/components/shared/error-state";
import { fetchWorkflow, fetchWorkflowYaml, saveWorkflowYaml } from "@/lib/api/workflows";
import { getHttpBaseURL } from "@/lib/api/http";
import { usePathname } from "next/navigation";
import type { Workflow, WorkflowStep, WorkflowYaml, WorkflowFlowModule, ModuleWorkflowYaml, FlowWorkflowYaml } from "@/lib/types/workflow";
import { toast } from "sonner";
import { ArrowLeftIcon, LoaderIcon, ArrowUpDownIcon, ArrowLeftRightIcon, ClipboardIcon, AlignJustifyIcon, MapIcon, SaveIcon, EyeIcon } from "lucide-react";

interface WorkflowEditorClientProps {
  workflowId: string;
}

export default function WorkflowEditorClient({ workflowId }: WorkflowEditorClientProps) {
  const pathname = usePathname();
  const splitRef = React.useRef<HTMLDivElement | null>(null);
  const canvasApiRef = React.useRef<{ focusNode: (nodeId: string) => void } | null>(null);
  const pendingFocusIdRef = React.useRef<string | null>(null);
  const effectiveId = React.useMemo(() => {
    if (workflowId && workflowId.length > 0) return workflowId;
    const seg = pathname?.split("/").filter(Boolean).pop() || "";
    return decodeURIComponent(seg);
  }, [workflowId, pathname]);
  const [workflow, setWorkflow] = React.useState<Workflow | null>(null);
  const [parsedWorkflow, setParsedWorkflow] = React.useState<ParsedWorkflow | null>(null);
  const [workflowData, setWorkflowData] = React.useState<WorkflowYaml | null>(null);
  const [yamlPreview, setYamlPreview] = React.useState("");
  const [isLoading, setIsLoading] = React.useState(true);
  const [isSaving, setIsSaving] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [selectedStepName, setSelectedStepName] = React.useState<string | null>(null);
  const [orientation, setOrientation] = React.useState<"TB" | "LR">("TB");
  const baseURL = getHttpBaseURL();
  const [sidebarWidth, setSidebarWidth] = React.useState<number>(320);
  const [wrapCanvasText, setWrapCanvasText] = React.useState<boolean>(true);
  const [showCanvasDetails, setShowCanvasDetails] = React.useState<boolean>(true);
  const [hideMiniMap, setHideMiniMap] = React.useState<boolean>(true);

  const toolbarButtonClassName =
    "border-info/40 text-info hover:bg-info-soft hover:text-info";

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    const raw = window.localStorage.getItem("osmedeus_workflow_sidebar_width");
    const n = raw ? Number(raw) : NaN;
    if (Number.isFinite(n) && n > 0) setSidebarWidth(n);
  }, []);

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    const raw = window.localStorage.getItem("osmedeus_workflow_canvas_wrap");
    if (raw === null) return;
    setWrapCanvasText(raw === "1");
  }, []);

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    const raw = window.localStorage.getItem("osmedeus_workflow_canvas_details");
    if (raw === null) return;
    setShowCanvasDetails(raw === "1");
  }, []);

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    const raw = window.localStorage.getItem("osmedeus_workflow_canvas_hide_minimap");
    if (raw === null) return;
    setHideMiniMap(raw === "1");
  }, []);

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    window.localStorage.setItem("osmedeus_workflow_sidebar_width", String(sidebarWidth));
  }, [sidebarWidth]);

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    window.localStorage.setItem("osmedeus_workflow_canvas_wrap", wrapCanvasText ? "1" : "0");
  }, [wrapCanvasText]);

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    window.localStorage.setItem("osmedeus_workflow_canvas_details", showCanvasDetails ? "1" : "0");
  }, [showCanvasDetails]);

  React.useEffect(() => {
    if (typeof window === "undefined") return;
    window.localStorage.setItem("osmedeus_workflow_canvas_hide_minimap", hideMiniMap ? "1" : "0");
  }, [hideMiniMap]);

  const loadWorkflow = React.useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      const [wf, yaml] = await Promise.all([
        fetchWorkflow(effectiveId),
        fetchWorkflowYaml(effectiveId),
      ]);

      if (!wf || !yaml) {
        setError(`Workflow not found: ${effectiveId}`);
        return;
      }

      setWorkflow(wf);
      const parsed = parseWorkflowYaml(yaml);
      setParsedWorkflow(parsed);
      setWorkflowData(parsed.raw);
      setYamlPreview(yaml);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "";
      if (msg === "WORKFLOW_NOT_FOUND") {
        setError(`Workflow not found: ${effectiveId}`);
      } else if (msg === "NETWORK_ERROR") {
        setError(`Cannot reach API at ${baseURL}`);
      } else if (msg === "UNAUTHORIZED") {
        setError("Session expired. Please log in.");
      } else {
        setError("Failed to load workflow");
      }
    } finally {
      setIsLoading(false);
    }
  }, [effectiveId, baseURL]);

  React.useEffect(() => {
    loadWorkflow();
  }, [loadWorkflow]);

  const selectedStep = React.useMemo(() => {
    if (!selectedStepName || !workflowData) return null;
    if (workflowData.kind !== "module") return null;
    const steps = (workflowData as ModuleWorkflowYaml).steps ?? [];
    return steps.find((s) => s.name === selectedStepName) ?? null;
  }, [selectedStepName, workflowData]);

  const selectedModule = React.useMemo(() => {
    if (!selectedStepName || !workflowData) return null;
    if (workflowData.kind !== "flow") return null;
    const wf = workflowData as FlowWorkflowYaml;
    const name = typeof wf.name === "string" ? wf.name : "";
    const modules = getFlowModules(wf, name);
    return modules.find((m) => m.name === selectedStepName) ?? null;
  }, [selectedStepName, workflowData]);

  const selectedTrigger = React.useMemo(() => {
    if (selectedStepName !== "_trigger") return null;
    const node = parsedWorkflow?.nodes.find((n) => n.id === "_trigger");
    const triggers = (node?.data as any)?.triggers;
    return Array.isArray(triggers) ? triggers : null;
  }, [parsedWorkflow, selectedStepName]);

  const selectedOverride = React.useMemo(() => {
    if (selectedStepName !== "_override") return null;
    const node = parsedWorkflow?.nodes.find((n) => n.id === "_override");
    const overrideModule = (node?.data as any)?.module;
    if (!overrideModule || typeof overrideModule !== "object") return null;
    return overrideModule as { name?: string; params?: Record<string, unknown> };
  }, [parsedWorkflow, selectedStepName]);

  const allSteps = React.useMemo(() => {
    if (!workflowData || workflowData.kind !== "module") return [] as WorkflowStep[];
    return (workflowData as ModuleWorkflowYaml).steps ?? [];
  }, [workflowData]);

  const allModules = React.useMemo(() => {
    if (!workflowData || workflowData.kind !== "flow") return [] as WorkflowFlowModule[];
    const wf = workflowData as FlowWorkflowYaml;
    const name = typeof wf.name === "string" ? wf.name : "";
    return getFlowModules(wf, name);
  }, [workflowData]);

  const handleNavigateToNode = React.useCallback((nodeId: string) => {
    pendingFocusIdRef.current = nodeId;
    setSelectedStepName(nodeId);
    requestAnimationFrame(() => {
      canvasApiRef.current?.focusNode(nodeId);
    });
  }, []);

  React.useEffect(() => {
    const id = pendingFocusIdRef.current;
    if (!id) return;
    if (id !== selectedStepName) return;
    pendingFocusIdRef.current = null;
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        canvasApiRef.current?.focusNode(id);
      });
    });
  }, [selectedStepName]);

  const handleNodeSelect = React.useCallback((nodeId: string | null) => {
    if (!nodeId || nodeId === "_start" || nodeId === "_end") {
      setSelectedStepName(null);
    } else {
      setSelectedStepName(nodeId);
    }
  }, []);

  const handleStepUpdate = React.useCallback(
    (stepName: string, updates: Partial<WorkflowStep>) => {
      if (!workflowData) return;
      if (workflowData.kind !== "module") return;

      const updatedWorkflow = updateStepInWorkflow(workflowData as ModuleWorkflowYaml, stepName, updates);
      setWorkflowData(updatedWorkflow);

      const newYaml = serializeWorkflowToYaml(updatedWorkflow);
      setYamlPreview(newYaml);

      const parsed = parseWorkflowYaml(newYaml);
      setParsedWorkflow(parsed);
    },
    [workflowData]
  );

  if (error) {
    return (
      <div className="flex h-[calc(100vh-10rem)] items-center justify-center">
        <div className="space-y-4 text-center">
          <ErrorState title="Workflow Error" message={error} onRetry={loadWorkflow} />
          <div className="flex items-center justify-center gap-2">
            <Button variant="outline" asChild>
              <Link href="/workflows">Back to Workflows</Link>
            </Button>
            <Button variant="outline" asChild>
              <Link href="/settings">Settings</Link>
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-[calc(100vh-7rem)] flex-col">
      <div className="flex items-center justify-between border-b px-4 py-3">
        <div className="flex items-center gap-4">
          <Button variant="outline" size="icon" asChild>
            <Link href="/workflows">
              <ArrowLeftIcon className="size-4" />
              <span className="sr-only">Back to workflows</span>
            </Link>
          </Button>
          <div>
            {isLoading ? (
              <div className="space-y-1">
                <Skeleton className="h-6 w-40" />
                <Skeleton className="h-4 w-24" />
              </div>
            ) : workflow ? (
              <>
                <div className="flex items-center gap-2">
                  <h1 className="text-lg font-semibold">{workflow.name}</h1>
                  <Badge variant="secondary" className="capitalize">
                    {workflow.kind}
                  </Badge>
                </div>
                <p className="text-sm text-muted-foreground">
                  {workflow.description}
                </p>
              </>
            ) : null}
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setOrientation((prev) => (prev === "TB" ? "LR" : "TB"))}
            className={toolbarButtonClassName}
          >
            {orientation === "TB" ? (
              <ArrowUpDownIcon className="mr-2 size-4" />
            ) : (
              <ArrowLeftRightIcon className="mr-2 size-4" />
            )}
            {orientation === "TB" ? "Vertical" : "Horizontal"}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setWrapCanvasText((v) => !v)}
            className={toolbarButtonClassName}
          >
            <AlignJustifyIcon className="mr-2 size-4" />
            {wrapCanvasText ? "Wrap lines on" : "Wrap lines off"}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowCanvasDetails((v) => !v)}
            className={toolbarButtonClassName}
          >
            <EyeIcon className="mr-2 size-4" />
            {showCanvasDetails ? "Details on" : "Details off"}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setHideMiniMap((v) => !v)}
            className={toolbarButtonClassName}
          >
            <MapIcon className="mr-2 size-4" />
            {hideMiniMap ? "Minimap off" : "Minimap on"}
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!yamlPreview}
            onClick={async () => {
              try {
                await navigator.clipboard.writeText(yamlPreview);
                toast.success("Copied to clipboard");
              } catch {
                toast.error("Failed to copy");
              }
            }}
          >
            <ClipboardIcon className="mr-2 size-4" />
            Copy YAML
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!yamlPreview || isSaving || isLoading}
            onClick={async () => {
              if (!yamlPreview) return;
              setIsSaving(true);
              try {
                const ok = await saveWorkflowYaml(effectiveId, yamlPreview);
                if (ok) {
                  toast.success("Saved");
                } else {
                  toast.error("Save failed");
                }
              } catch {
                toast.error("Save failed");
              } finally {
                setIsSaving(false);
              }
            }}
          >
            <SaveIcon className="mr-2 size-4" />
            {isSaving ? "Saving..." : "Save"}
          </Button>
        </div>
      </div>

      <div ref={splitRef} className="flex flex-1 overflow-hidden">
        <div className="flex-1">
          {isLoading ? (
            <div className="flex h-full items-center justify-center">
              <div className="flex flex-col items-center gap-4">
                <LoaderIcon className="size-8 animate-spin text-muted-foreground" />
                <p className="text-sm text-muted-foreground">Loading workflow...</p>
              </div>
            </div>
          ) : parsedWorkflow ? (
            <WorkflowCanvas
              initialNodes={parsedWorkflow.nodes}
              initialEdges={parsedWorkflow.edges}
              onNodeSelect={handleNodeSelect}
              orientation={orientation}
              wrapLongText={wrapCanvasText}
              showDetails={showCanvasDetails}
              hideMiniMap={hideMiniMap}
              selectedNodeId={selectedStepName}
              onCanvasReady={(api) => {
                canvasApiRef.current = api;
                if (pendingFocusIdRef.current) {
                  api.focusNode(pendingFocusIdRef.current);
                }
              }}
            />
          ) : null}
        </div>

        <div
          role="separator"
          aria-orientation="vertical"
          tabIndex={0}
          className="w-1 cursor-col-resize bg-border/60 hover:bg-border focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          onPointerDown={(e) => {
            const startX = e.clientX;
            const startWidth = sidebarWidth;
            (e.currentTarget as HTMLDivElement).setPointerCapture(e.pointerId);

            const handleMove = (ev: PointerEvent) => {
              const rect = splitRef.current?.getBoundingClientRect();
              const containerWidth = rect?.width ?? 0;
              const deltaX = ev.clientX - startX;
              const maxWidth = Math.max(260, containerWidth - 200);
              const next = Math.min(maxWidth, Math.max(260, startWidth - deltaX));
              setSidebarWidth(next);
            };

            const handleUp = (ev: PointerEvent) => {
              window.removeEventListener("pointermove", handleMove);
              window.removeEventListener("pointerup", handleUp);
            };

            window.addEventListener("pointermove", handleMove);
            window.addEventListener("pointerup", handleUp);
          }}
          onKeyDown={(e) => {
            const step = e.shiftKey ? 40 : 20;
            if (e.key === "ArrowLeft") setSidebarWidth((w) => Math.min(w + step, 900));
            if (e.key === "ArrowRight") setSidebarWidth((w) => Math.max(260, w - step));
          }}
        />

        <div className="shrink-0" style={{ width: sidebarWidth }}>
          <WorkflowSidebar
            selectedStep={selectedStep}
            selectedModule={selectedModule}
              selectedTrigger={selectedTrigger}
              selectedOverride={selectedOverride}
            yamlPreview={yamlPreview}
            wrapLongText={wrapCanvasText}
            onStepUpdate={handleStepUpdate}
            workflowKind={workflowData?.kind ?? null}
            allSteps={allSteps}
            allModules={allModules}
            onNavigateToNode={handleNavigateToNode}
          />
        </div>
      </div>
    </div>
  );
}
