import dagre from "dagre";
import type { Node, Edge } from "@xyflow/react";

const NODE_WIDTH = 220;
const NODE_HEIGHT = 80;

export function layoutWorkflow(
  nodes: Node[],
  edges: Edge[],
  orientation: "TB" | "LR" = "TB"
): Node[] {
  const g = new dagre.graphlib.Graph();

  const dashedEdges = edges.filter((e) => typeof (e.style as any)?.strokeDasharray === "string");
  const labeledEdges = edges.filter((e) => typeof e.label === "string" && e.label.trim().length > 0);
  const branchingEdges = dashedEdges.length > 0 ? dashedEdges : labeledEdges;

  const decisionOutDegree = new Map<string, number>();
  for (const e of branchingEdges) {
    decisionOutDegree.set(e.source, (decisionOutDegree.get(e.source) ?? 0) + 1);
  }
  const maxDecisionOut = Math.max(0, ...Array.from(decisionOutDegree.values()));

  const hasDecisions = dashedEdges.length > 0 || labeledEdges.length > 0;
  const ranksep = hasDecisions ? 170 : 80;
  const nodesep = hasDecisions ? 120 + Math.max(0, maxDecisionOut - 1) * 40 : 50;
  const edgesep = hasDecisions ? 40 : 10;

  g.setGraph({
    rankdir: orientation,
    ranksep,
    nodesep,
    edgesep,
    marginx: hasDecisions ? 60 : 20,
    marginy: hasDecisions ? 60 : 20,
  });

  g.setDefaultEdgeLabel(() => ({}));

  // Add nodes to dagre
  nodes.forEach((node) => {
    g.setNode(node.id, {
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
    });
  });

  // Add edges to dagre
  edges.forEach((edge) => {
    if (edge.source === "_trigger" && edge.target === "_start") {
      g.setEdge(edge.source, edge.target, { minlen: 2 });
      return;
    }
    g.setEdge(edge.source, edge.target);
  });

  // Run layout
  dagre.layout(g);

  // Apply positions back to nodes
  return nodes.map((node) => {
    const nodeWithPosition = g.node(node.id);
    if (nodeWithPosition) {
      return {
        ...node,
        position: {
          x: nodeWithPosition.x - NODE_WIDTH / 2,
          y: nodeWithPosition.y - NODE_HEIGHT / 2,
        },
      };
    }
    return node;
  });
}
