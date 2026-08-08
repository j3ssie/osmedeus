"use client";

import * as React from "react";
import {
  ReactFlow,
  Background,
  Controls,
  ControlButton,
  MiniMap,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
  type OnNodesChange,
  type OnEdgesChange,
  BackgroundVariant,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import { Maximize2Icon, Minimize2Icon } from "lucide-react";

import { nodeTypes } from "./nodes";
import { layoutWorkflow } from "./utils/layout-engine";
import type { WorkflowNodeData } from "@/lib/types/workflow";
import { CanvasSettingsProvider } from "./canvas-settings";

interface WorkflowCanvasProps {
  initialNodes: Node<WorkflowNodeData>[];
  initialEdges: Edge[];
  onNodeSelect?: (nodeId: string | null) => void;
  orientation?: "TB" | "LR";
  wrapLongText?: boolean;
  showDetails?: boolean;
  hideMiniMap?: boolean;
  selectedNodeId?: string | null;
  onCanvasReady?: (api: { focusNode: (nodeId: string) => void }) => void;
}

export function WorkflowCanvas({
  initialNodes,
  initialEdges,
  onNodeSelect,
  orientation = "TB",
  wrapLongText = false,
  showDetails = true,
  hideMiniMap = false,
  selectedNodeId = null,
  onCanvasReady,
}: WorkflowCanvasProps) {
  const wrapperRef = React.useRef<HTMLDivElement | null>(null);
  const [isFullscreen, setIsFullscreen] = React.useState(false);
  const instanceRef = React.useRef<any>(null);

  // Apply layout to initial nodes
  const layoutedNodes = React.useMemo(
    () => layoutWorkflow(initialNodes, initialEdges, orientation),
    [initialNodes, initialEdges, orientation]
  );

  const [nodes, setNodes, onNodesChange] = useNodesState(layoutedNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // Update nodes when initial data changes
  React.useEffect(() => {
    const newLayoutedNodes = layoutWorkflow(initialNodes, initialEdges, orientation);
    setNodes(newLayoutedNodes.map((n) => ({ ...n, selected: selectedNodeId ? n.id === selectedNodeId : false })));
    setEdges(initialEdges);
  }, [initialNodes, initialEdges, orientation, setNodes, setEdges, selectedNodeId]);

  React.useEffect(() => {
    if (selectedNodeId === null) {
      setNodes((prev) => prev.map((n) => (n.selected ? { ...n, selected: false } : n)));
      return;
    }
    setNodes((prev) => prev.map((n) => ({ ...n, selected: n.id === selectedNodeId })));
  }, [selectedNodeId, setNodes]);

  const focusNode = React.useCallback((nodeId: string) => {
    const instance = instanceRef.current;
    if (!instance) return;
    const node = typeof instance.getNode === "function" ? instance.getNode(nodeId) : null;
    if (!node) return;

    const width = (node.measured?.width ?? node.width ?? 0) as number;
    const height = (node.measured?.height ?? node.height ?? 0) as number;
    const x = (node.positionAbsolute?.x ?? node.position?.x ?? 0) + width / 2;
    const y = (node.positionAbsolute?.y ?? node.position?.y ?? 0) + height / 2;
    if (typeof instance.setCenter === "function") {
      const currentZoom = typeof instance.getZoom === "function" ? instance.getZoom() : 1;
      const targetZoom = Math.max(currentZoom ?? 1, 1.09);
      instance.setCenter(x, y, { zoom: targetZoom, duration: 500 });
    }
  }, []);

  const handleNodesChange: OnNodesChange = React.useCallback(
    (changes) => {
      onNodesChange(changes);

      // Check for selection changes
      const selectedChange = changes.find(
        (change) => change.type === "select" && change.selected
      );
      if (selectedChange && selectedChange.type === "select") {
        onNodeSelect?.(selectedChange.id);
        return;
      }

      // Also check if all nodes are deselected
      const hasSelection = changes.some(
        (change) => change.type === "select" && change.selected
      );
      if (!hasSelection && changes.some((c) => c.type === "select")) {
        // Check if any node is still selected
        const stillSelected = nodes.some(
          (n) =>
            n.selected &&
            !changes.find(
              (c) => c.type === "select" && c.id === n.id && !c.selected
            )
        );
        if (!stillSelected) {
          onNodeSelect?.(null);
        }
      }
    },
    [onNodesChange, onNodeSelect, nodes]
  );

  const handleEdgesChange: OnEdgesChange = React.useCallback(
    (changes) => {
      onEdgesChange(changes);
    },
    [onEdgesChange]
  );

  React.useEffect(() => {
    const handler = () => {
      setIsFullscreen(Boolean(document.fullscreenElement));
    };
    document.addEventListener("fullscreenchange", handler);
    document.addEventListener("webkitfullscreenchange", handler as any);
    handler();
    return () => {
      document.removeEventListener("fullscreenchange", handler);
      document.removeEventListener("webkitfullscreenchange", handler as any);
    };
  }, []);

  const toggleFullscreen = React.useCallback(async () => {
    try {
      if (document.fullscreenElement) {
        await document.exitFullscreen();
        return;
      }
      const el = wrapperRef.current as any;
      if (!el) return;
      if (typeof el.requestFullscreen === "function") {
        await el.requestFullscreen();
        return;
      }
      if (typeof el.webkitRequestFullscreen === "function") {
        el.webkitRequestFullscreen();
      }
    } catch {
    }
  }, []);

  return (
    <div ref={wrapperRef} className="h-full w-full">
      <CanvasSettingsProvider wrapLongText={wrapLongText} showDetails={showDetails}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={handleNodesChange}
          onEdgesChange={handleEdgesChange}
          nodeTypes={nodeTypes}
          onInit={(instance) => {
            instanceRef.current = instance;
            onCanvasReady?.({ focusNode });
          }}
          fitView
          fitViewOptions={{
            padding: 0.2,
          }}
          minZoom={0.1}
          maxZoom={2}
          defaultEdgeOptions={{
            type: "smoothstep",
          }}
          proOptions={{ hideAttribution: true }}
        >
          <Background
            variant={BackgroundVariant.Dots}
            gap={16}
            size={1}
            className="bg-muted/30"
          />
          <Controls
            className="rounded-lg border bg-card shadow-sm [&>button]:border-border [&>button]:bg-card [&>button:hover]:bg-muted"
            showInteractive={false}
          >
            <ControlButton onClick={toggleFullscreen} title={isFullscreen ? "Exit fullscreen" : "Fullscreen"}>
              {isFullscreen ? <Minimize2Icon className="size-4" /> : <Maximize2Icon className="size-4" />}
            </ControlButton>
          </Controls>
          {!hideMiniMap && (
            <MiniMap
              className="rounded-lg border bg-card shadow-sm"
              nodeStrokeWidth={3}
              pannable
              zoomable
            />
          )}
        </ReactFlow>
      </CanvasSettingsProvider>
    </div>
  );
}
