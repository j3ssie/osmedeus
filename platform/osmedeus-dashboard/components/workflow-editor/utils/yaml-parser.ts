import * as yaml from "js-yaml";
import type { Node, Edge } from "@xyflow/react";
import type {
  WorkflowYaml,
  WorkflowStep,
  WorkflowNodeData,
  ModuleWorkflowYaml,
  FlowWorkflowYaml,
  WorkflowFlowModule,
  WorkflowDecision,
} from "@/lib/types/workflow";

// Use a compatible node type that works with React Flow's typing constraints
type WorkflowNode = Node & { data: WorkflowNodeData };

export interface ParsedWorkflow {
  nodes: WorkflowNode[];
  edges: Edge[];
  metadata: {
    name: string;
    description: string;
    kind: string;
  };
  raw: WorkflowYaml;
}

type DecisionEdge = {
  label: string;
  next: string;
  kind: "rule" | "case" | "default";
};

type NormalizedTrigger = {
  name?: string;
  on?: string;
  enabled?: boolean;
  schedule?: string;
  path?: string;
  input?: {
    type?: string;
    field?: string;
    function?: string;
  };
  event?: {
    topic?: string;
    filters?: string[];
    filterFunctions?: string[];
  };
};

function isSwitchDecision(decision: unknown): decision is Extract<WorkflowDecision, { switch: string }> {
  if (!decision || typeof decision !== "object") return false;
  const d = decision as any;
  return typeof d.switch === "string" && d.cases && typeof d.cases === "object" && !Array.isArray(d.cases);
}

function normalizeDecisionEdges(decision: unknown): DecisionEdge[] {
  if (Array.isArray(decision)) {
    return decision
      .map((r) => ({
        condition: typeof (r as any)?.condition === "string" ? (r as any).condition : "",
        next: typeof (r as any)?.next === "string" ? (r as any).next : "",
      }))
      .filter((r) => r.next.trim().length > 0)
      .map((r) => ({ label: r.condition, next: r.next, kind: "rule" as const }));
  }

  if (isSwitchDecision(decision)) {
    const d = decision as any;
    const cases: Record<string, any> = d.cases ?? {};
    const edges: DecisionEdge[] = Object.entries(cases)
      .filter(([key]) => typeof key === "string" && key.trim().length > 0)
      .map(([key, value]) => {
        const goto = typeof value?.goto === "string" ? value.goto : (typeof value?.next === "string" ? value.next : "");
        return { label: key, next: goto, kind: "case" as const };
      })
      .filter((e) => e.next.trim().length > 0);

    const defGoto = typeof d.default?.goto === "string" ? d.default.goto : (typeof d.default?.next === "string" ? d.default.next : "");
    if (typeof defGoto === "string" && defGoto.trim().length > 0) {
      edges.push({ label: "default", next: defGoto, kind: "default" });
    }

    return edges;
  }

  return [];
}

function truncateLabel(label: string, maxLen = 30): string {
  const trimmed = label.trim();
  if (trimmed.length <= maxLen) return trimmed;
  return trimmed.substring(0, maxLen) + "...";
}

function normalizeTriggerList(value: unknown): NormalizedTrigger[] {
  if (!value) return [];
  const rawList = Array.isArray(value) ? value : (typeof value === "object" ? [value] : []);
  return rawList.map((entry) => {
    const anyEntry = entry as Record<string, unknown>;
    const event = anyEntry.event && typeof anyEntry.event === "object" ? (anyEntry.event as Record<string, unknown>) : null;
    const input = anyEntry.input && typeof anyEntry.input === "object" ? (anyEntry.input as Record<string, unknown>) : null;
    const rawFilters = event?.filters;
    const rawFilterFunctions = (event as any)?.filter_functions ?? (event as any)?.filterFunctions;
    const filters = Array.isArray(rawFilters) ? rawFilters.filter((f) => typeof f === "string") as string[] : [];
    const filterFunctions = Array.isArray(rawFilterFunctions)
      ? rawFilterFunctions.filter((f) => typeof f === "string") as string[]
      : [];

    return {
      name: typeof anyEntry.name === "string" ? anyEntry.name : undefined,
      on: typeof anyEntry.on === "string" ? anyEntry.on : undefined,
      enabled: typeof anyEntry.enabled === "boolean" ? anyEntry.enabled : undefined,
      schedule: typeof anyEntry.schedule === "string" ? anyEntry.schedule : undefined,
      path: typeof anyEntry.path === "string" ? anyEntry.path : undefined,
      input: input
        ? {
            type: typeof input.type === "string" ? input.type : undefined,
            field: typeof input.field === "string" ? input.field : undefined,
            function: typeof (input as any).function === "string" ? (input as any).function : undefined,
          }
        : undefined,
      event: event
        ? {
            topic: typeof event.topic === "string" ? event.topic : undefined,
            filters: filters.length > 0 ? filters : undefined,
            filterFunctions: filterFunctions.length > 0 ? filterFunctions : undefined,
          }
        : undefined,
    };
  });
}

export function getFlowModules(workflow: FlowWorkflowYaml, workflowName: string): WorkflowFlowModule[] {
  const modules = Array.isArray(workflow.modules) ? workflow.modules : [];
  if (modules.length > 0) return modules;

  const extendsValue = typeof (workflow as any).extends === "string" ? (workflow as any).extends.trim() : "";
  if (!extendsValue) return [];

  const override = (workflow as any).override;
  const overrideParams =
    override && typeof override === "object" && !Array.isArray(override)
      ? (override as Record<string, unknown>).params
      : undefined;
  const params =
    overrideParams && typeof overrideParams === "object" && !Array.isArray(overrideParams)
      ? (overrideParams as Record<string, unknown>)
      : undefined;
  const name = workflowName.trim() || extendsValue;

  return [
    {
      name,
      extends: extendsValue,
      params,
    },
  ];
}

export function parseWorkflowYaml(yamlText: string): ParsedWorkflow {
  const workflowAny = (yaml.load(yamlText) as any) ?? {};
  const kind: "module" | "flow" = workflowAny?.kind === "flow" ? "flow" : "module";
  const workflow = workflowAny as WorkflowYaml;
  const nodes: WorkflowNode[] = [];
  const edges: Edge[] = [];
  const triggers = normalizeTriggerList(workflowAny?.triggers ?? workflowAny?.trigger);
  const overrideParams =
    workflowAny?.override && typeof workflowAny.override === "object" && !Array.isArray(workflowAny.override)
      ? (workflowAny.override as Record<string, unknown>).params
      : undefined;
  const overrideNodeParams =
    overrideParams && typeof overrideParams === "object" && !Array.isArray(overrideParams)
      ? (overrideParams as Record<string, unknown>)
      : null;

  if (triggers.length > 0) {
    nodes.push({
      id: "_trigger",
      type: "trigger",
      position: { x: 0, y: 0 },
      data: { label: "Triggers", step: null, module: null, triggers },
    });
  }

  // Add start node
  nodes.push({
    id: "_start",
    type: "start",
    position: { x: 0, y: 0 },
    data: { label: "Start", step: null, module: null },
  });

  if (overrideNodeParams) {
    nodes.push({
      id: "_override",
      type: "override",
      position: { x: 0, y: 0 },
      data: {
        label: "Override",
        step: null,
        module: {
          name: "override",
          params: overrideNodeParams,
        },
      },
    });
    edges.push({
      id: `_override->_start`,
      source: "_override",
      target: "_start",
      type: "smoothstep",
    });
  }

  if (triggers.length > 0) {
    edges.push({
      id: overrideNodeParams ? `_trigger->_override` : `_trigger->_start`,
      source: "_trigger",
      target: overrideNodeParams ? "_override" : "_start",
      type: "smoothstep",
    });
  }

  if (kind === "module") {
    const wf = workflow as ModuleWorkflowYaml;
    const steps = Array.isArray((wf as any)?.steps) ? (wf as any).steps : [];
    const stepIds = new Set<string>(steps.map((s: any) => (typeof s?.name === "string" ? s.name : "")).filter(Boolean));
    const missingNodeIds = new Set<string>();
    const outDegree = new Map<string, number>();
    const decisionTargets = new Set<string>();
    const hasDependsOn = steps.some(
      (s: any) => Array.isArray(s?.depends_on) && s.depends_on.filter((d: any) => typeof d === "string").length > 0
    );

    steps.forEach((step: WorkflowStep) => {
      const edges = normalizeDecisionEdges((step as any).decision);
      edges.forEach((edge) => {
        if (edge.next !== "_end") decisionTargets.add(edge.next);
      });
    });

    steps.forEach((step: WorkflowStep, index: number) => {
      const nodeId = step.name;

      nodes.push({
        id: nodeId,
        type: step.type,
        position: { x: 0, y: 0 },
        data: {
          label: step.name,
          step,
          module: null,
        },
      });

      const deps: string[] = Array.isArray((step as any).depends_on)
        ? (step as any).depends_on.filter((d: any) => typeof d === "string")
        : [];
      const prevNodeId = index === 0 ? "_start" : steps[index - 1].name;
      const prevStep = index > 0 ? steps[index - 1] : null;
      const prevDecisionEdges = prevStep ? normalizeDecisionEdges((prevStep as any).decision) : [];
      const hasDecisionToThis = prevDecisionEdges.some((d) => d.next === nodeId);

      if (hasDependsOn) {
        if (deps.length > 0) {
          deps.forEach((d) => {
            if (!stepIds.has(d)) missingNodeIds.add(d);
            edges.push({
              id: `${d}->${nodeId}`,
              source: d,
              target: nodeId,
              type: "smoothstep",
              animated: step.type === "parallel" || step.type === "parallel-steps",
            });
            if (stepIds.has(d)) {
              outDegree.set(d, (outDegree.get(d) || 0) + 1);
            }
          });
        } else if (!decisionTargets.has(nodeId)) {
          edges.push({
            id: `_start->${nodeId}`,
            source: "_start",
            target: nodeId,
            type: "smoothstep",
            animated: step.type === "parallel" || step.type === "parallel-steps",
          });
        }
      } else if (prevDecisionEdges.length === 0 && !hasDecisionToThis && !decisionTargets.has(nodeId)) {
        edges.push({
          id: `${prevNodeId}->${nodeId}`,
          source: prevNodeId,
          target: nodeId,
          type: "smoothstep",
          animated: step.type === "parallel" || step.type === "parallel-steps",
        });
        if (prevNodeId !== "_start" && stepIds.has(prevNodeId)) {
          outDegree.set(prevNodeId, (outDegree.get(prevNodeId) || 0) + 1);
        }
      }

      const decisionEdges = normalizeDecisionEdges((step as any).decision);
      if (decisionEdges.length > 0) {
        decisionEdges.forEach((rule) => {
          const next = rule.next;
          if (next !== "_end" && !stepIds.has(next)) missingNodeIds.add(next);
          edges.push({
            id: `${nodeId}->${next}:${rule.kind}:${rule.label}`,
            source: nodeId,
            target: next,
            type: "smoothstep",
            label: truncateLabel(rule.label),
            style: { strokeDasharray: "5 5" },
          });
          outDegree.set(nodeId, (outDegree.get(nodeId) || 0) + 1);
        });
      }
    });

    for (const missingId of missingNodeIds) {
      if (stepIds.has(missingId)) continue;
      nodes.push({
        id: missingId,
        position: { x: 0, y: 0 },
        data: {
          label: `Missing: ${missingId}`,
          step: null,
          module: null,
        },
      });
    }

    if (steps.length === 0) {
      edges.push({
        id: `_start->_end`,
        source: "_start",
        target: "_end",
        type: "smoothstep",
      });
    } else {
      steps.forEach((step: WorkflowStep) => {
        const od = outDegree.get(step.name) || 0;
        if (od === 0) {
          edges.push({
            id: `${step.name}->_end`,
            source: step.name,
            target: "_end",
            type: "smoothstep",
          });
        }
      });
    }
  } else {
    const wf = workflow as FlowWorkflowYaml;
    const name = typeof workflowAny?.name === "string" ? workflowAny.name : "";
    const modules = getFlowModules(wf, name);

    modules.forEach((m) => {
      nodes.push({
        id: m.name,
        type: "module",
        position: { x: 0, y: 0 },
        data: {
          label: m.name,
          step: null,
          module: m,
        },
      });
    });

    const nodeIds = new Set(modules.map((m) => m.name));
    const missingNodeIds = new Set<string>();
    const outDegree = new Map<string, number>();

    modules.forEach((m) => {
      const deps = Array.isArray(m.depends_on) ? m.depends_on.filter((d) => typeof d === "string") : [];
      if (deps.length > 0) {
        deps.forEach((d) => {
          if (!nodeIds.has(d)) return;
          edges.push({
            id: `${d}->${m.name}`,
            source: d,
            target: m.name,
            type: "smoothstep",
          });
          outDegree.set(d, (outDegree.get(d) || 0) + 1);
        });
      } else {
        edges.push({
          id: `_start->${m.name}`,
          source: "_start",
          target: m.name,
          type: "smoothstep",
        });
      }

      const decisionEdges = normalizeDecisionEdges((m as any).decision);
      if (decisionEdges.length > 0) {
        decisionEdges.forEach((rule) => {
          const next = rule.next;
          if (next !== "_end" && !nodeIds.has(next)) missingNodeIds.add(next);
          edges.push({
            id: `${m.name}->${next}:${rule.kind}:${rule.label}`,
            source: m.name,
            target: next,
            type: "smoothstep",
            label: truncateLabel(rule.label),
            style: { strokeDasharray: "5 5" },
          });
          outDegree.set(m.name, (outDegree.get(m.name) || 0) + 1);
        });
      }
    });

    for (const missingId of missingNodeIds) {
      if (nodeIds.has(missingId)) continue;
      nodes.push({
        id: missingId,
        position: { x: 0, y: 0 },
        data: {
          label: `Missing: ${missingId}`,
          step: null,
          module: null,
        },
      });
    }

    if (modules.length === 0) {
      edges.push({
        id: `_start->_end`,
        source: "_start",
        target: "_end",
        type: "smoothstep",
      });
    } else {
      modules.forEach((m) => {
        const od = outDegree.get(m.name) || 0;
        if (od === 0) {
          edges.push({
            id: `${m.name}->_end`,
            source: m.name,
            target: "_end",
            type: "smoothstep",
          });
        }
      });
    }
  }

  // Add end node
  nodes.push({
    id: "_end",
    type: "end",
    position: { x: 0, y: 0 },
    data: { label: "End", step: null, module: null },
  });

  return {
    nodes,
    edges,
    metadata: {
      name: (workflowAny?.name as string) || "",
      description: (workflowAny?.description as string) || "",
      kind,
    },
    raw: workflow,
  };
}

export function serializeWorkflowToYaml(workflow: WorkflowYaml): string {
  return yaml.dump(workflow, {
    indent: 2,
    lineWidth: -1,
    noRefs: true,
    quoteStyle: "double",
  });
}

export function updateStepInWorkflow(
  workflow: ModuleWorkflowYaml,
  stepName: string,
  updates: Partial<WorkflowStep>
): ModuleWorkflowYaml {
  return {
    ...workflow,
    steps: workflow.steps.map((step) =>
      step.name === stepName ? { ...step, ...updates } : step
    ),
  };
}
