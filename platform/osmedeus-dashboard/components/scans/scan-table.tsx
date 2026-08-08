"use client";

import * as React from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScanStatusBadge } from "./scan-status-badge";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { EmptyState } from "@/components/shared/empty-state";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { CellShell, DataGrid, type GridColDef } from "@/components/ui/data-grid";
import { cn, truncate } from "@/lib/utils";
import type { Scan } from "@/lib/types/scan";
import {
  EyeIcon,
  StopCircleIcon,
  TrashIcon,
  ScanSearchIcon,
  CalendarIcon,
  PlayIcon,
  CopyIcon,
  LoaderIcon,
} from "lucide-react";
import { toast } from "sonner";
import { cancelScan, deleteScan, duplicateScanRun, startScanRun } from "@/lib/api/scans";

interface ScanTableProps {
  scans: Scan[];
  isLoading?: boolean;
  onRefresh?: () => void;
  onSelectScan?: (scan: Scan) => void;
}

type TriggerVariant =
  | "default"
  | "secondary"
  | "destructive"
  | "outline"
  | "success"
  | "warning"
  | "info"
  | "purple"
  | "pink"
  | "cyan"
  | "orange";

const TRIGGER_CONFIG: Record<
  string,
  { label: string; variant: TriggerVariant; icon: React.ReactNode }
> = {
  cli: { label: "CLI", variant: "purple", icon: <PlayIcon className="size-3" /> },
  api: { label: "API", variant: "info", icon: <PlayIcon className="size-3" /> },
  cron: { label: "Cron", variant: "warning", icon: <CalendarIcon className="size-3" /> },
  scheduled: {
    label: "Scheduled",
    variant: "warning",
    icon: <CalendarIcon className="size-3" />,
  },
  manual: { label: "Manual", variant: "secondary", icon: <PlayIcon className="size-3" /> },
};

const PRIORITY_CONFIG: Record<string, { label: string; className: string }> = {
  high: { label: "High", className: "border-destructive/50 text-destructive" },
  medium: { label: "Medium", className: "border-warning/50 text-warning" },
  low: { label: "Low", className: "border-success/50 text-success" },
};

const PRIORITY_ORDER: Record<string, number> = { high: 3, medium: 2, low: 1 };

/**
 * Sorts blanks to the bottom in *both* directions. AG Grid flips the sign of a
 * comparator's result for a descending sort, so a blank check has to undo that
 * flip itself or empty rows float to the top the moment you reverse the order.
 */
function blanksLast<V>(
  compare: (a: V, b: V) => number,
  isBlank: (v: V) => boolean
) {
  return (a: V, b: V, _nodeA: unknown, _nodeB: unknown, isDescending: boolean) => {
    const aBlank = isBlank(a);
    const bBlank = isBlank(b);
    if (aBlank && bBlank) return 0;
    if (aBlank) return isDescending ? -1 : 1;
    if (bBlank) return isDescending ? 1 : -1;
    return compare(a, b);
  };
}

const compareText = blanksLast<string>(
  (a, b) => a.localeCompare(b, undefined, { numeric: true, sensitivity: "base" }),
  (v) => !v
);

export function ScanTable({ scans, isLoading, onRefresh, onSelectScan }: ScanTableProps) {
  const [duplicateRunId, setDuplicateRunId] = React.useState<string | null>(null);
  const [confirmState, setConfirmState] = React.useState<{
    action: "duplicate" | "cancel" | "delete";
    scan: Scan;
  } | null>(null);

  const handleCancel = React.useCallback(
    async (scan: Scan) => {
      try {
        const runUuid = scan.runUuid || scan.runId || scan.id;
        const success = await cancelScan(runUuid);
        if (success) {
          toast.success("Scan cancelled", {
            description: `Scan for ${scan.target} has been cancelled.`,
          });
          onRefresh?.();
        } else {
          toast.error("Failed to cancel scan");
        }
      } catch {
        toast.error("Failed to cancel scan");
      }
    },
    [onRefresh]
  );

  const handleDelete = React.useCallback(
    async (scan: Scan) => {
      try {
        const runUuid = scan.runUuid || scan.runId || scan.id;
        const success = await deleteScan(runUuid);
        if (success) {
          toast.success("Scan deleted", {
            description: `Scan for ${scan.target} has been deleted.`,
          });
          onRefresh?.();
        } else {
          toast.error("Failed to delete scan");
        }
      } catch {
        toast.error("Failed to delete scan");
      }
    },
    [onRefresh]
  );

  const handleDuplicateAndStart = React.useCallback(
    async (scan: Scan) => {
      const runUuid = scan.runUuid || scan.runId || scan.id;
      if (!runUuid) {
        toast.error("Missing run identifier");
        return;
      }
      try {
        setDuplicateRunId(runUuid);
        const duplicated = await duplicateScanRun(runUuid);
        const newRunUuid = duplicated.runUuid || duplicated.runId || duplicated.id;
        if (!newRunUuid) {
          toast.error("Duplicate created without run identifier");
          return;
        }
        const started = await startScanRun(newRunUuid);
        if (started) {
          toast.success("Scan duplicated and started", {
            description: `Scan for ${duplicated.target || scan.target} is running.`,
          });
        } else {
          toast.success("Scan duplicated", {
            description: `Duplicate created for ${duplicated.target || scan.target}.`,
          });
        }
        onRefresh?.();
      } catch {
        toast.error("Failed to duplicate scan");
      } finally {
        setDuplicateRunId(null);
      }
    },
    [onRefresh]
  );

  const handleConfirmAction = React.useCallback(async () => {
    if (!confirmState) return;
    const { action, scan } = confirmState;
    setConfirmState(null);
    if (action === "duplicate") return handleDuplicateAndStart(scan);
    if (action === "cancel") return handleCancel(scan);
    return handleDelete(scan);
  }, [confirmState, handleCancel, handleDelete, handleDuplicateAndStart]);

  const columns = React.useMemo<GridColDef<Scan>[]>(
    () => [
      {
        field: "status",
        headerName: "Status",
        minWidth: 120,
        flex: 0,
        width: 130,
        comparator: compareText,
        cellRenderer: (p: { data: Scan }) => <ScanStatusBadge status={p.data.status} />,
      },
      {
        field: "workflowName",
        colId: "workflow",
        headerName: "Workflow",
        minWidth: 150,
        comparator: compareText,
        cellRenderer: (p: { data: Scan }) => (
          <div className="flex min-w-0 flex-col justify-center leading-tight">
            <span className="truncate font-medium">{p.data.workflowName}</span>
            <span className="truncate text-xs text-muted-foreground">
              {p.data.workflowKind}
            </span>
          </div>
        ),
      },
      {
        field: "target",
        headerName: "Target",
        minWidth: 160,
        flex: 2,
        comparator: compareText,
        cellRenderer: (p: { value: string }) => (
          <span className="truncate font-mono text-sm">{truncate(p.value, 30)}</span>
        ),
      },
      {
        colId: "priority",
        headerName: "Priority",
        minWidth: 110,
        flex: 0,
        width: 120,
        valueGetter: (p) =>
          p.data?.priority ? String(p.data.priority).toLowerCase() : "",
        comparator: blanksLast<string>(
          (a, b) => (PRIORITY_ORDER[a] ?? 0) - (PRIORITY_ORDER[b] ?? 0),
          (v) => !v
        ),
        cellRenderer: (p: { value: string }) =>
          p.value ? (
            <Badge
              variant="outline"
              className={cn("w-fit capitalize", PRIORITY_CONFIG[p.value]?.className)}
            >
              {PRIORITY_CONFIG[p.value]?.label ?? p.value}
            </Badge>
          ) : (
            <span className="text-muted-foreground">-</span>
          ),
      },
      {
        colId: "progress",
        headerName: "Steps",
        minWidth: 160,
        valueGetter: (p) => {
          const total = p.data?.totalSteps ?? 0;
          return total > 0 ? (p.data?.completedSteps ?? 0) / total : -1;
        },
        comparator: blanksLast<number>((a, b) => a - b, (v) => v < 0),
        cellRenderer: (p: { data: Scan }) => {
          const scan = p.data;
          if (scan.totalSteps > 0) {
            return (
              <CellShell className="gap-2">
                <span className="whitespace-nowrap text-sm">
                  {scan.completedSteps}/{scan.totalSteps} steps
                </span>
                <div className="h-2 w-16 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full bg-primary transition-all"
                    style={{
                      width: `${Math.round((scan.completedSteps / scan.totalSteps) * 100)}%`,
                    }}
                  />
                </div>
              </CellShell>
            );
          }
          return (
            <span className="text-muted-foreground">
              {scan.status === "running" ? "In progress..." : "-"}
            </span>
          );
        },
      },
      {
        colId: "trigger",
        headerName: "Trigger",
        minWidth: 130,
        valueGetter: (p) => p.data?.triggerType ?? "",
        comparator: compareText,
        cellRenderer: (p: { data: Scan }) => {
          const type = (p.data.triggerType || "manual").toLowerCase();
          const cfg = TRIGGER_CONFIG[type] ?? {
            label: p.data.triggerType || "manual",
            variant: "outline" as TriggerVariant,
            icon: <PlayIcon className="size-3" />,
          };
          return (
            <div className="flex min-w-0 flex-col justify-center gap-1">
              <Badge variant={cfg.variant} className="w-fit gap-1">
                {cfg.icon}
                <span>{cfg.label}</span>
              </Badge>
              {p.data.triggerName ? (
                <span className="truncate text-xs text-muted-foreground">
                  {truncate(p.data.triggerName, 22)}
                </span>
              ) : null}
            </div>
          );
        },
      },
      {
        colId: "actions",
        headerName: "Actions",
        minWidth: 132,
        flex: 0,
        width: 132,
        sortable: false,
        cellRenderer: (p: { data: Scan }) => {
          const scan = p.data;
          const runUuid = scan.runUuid || scan.runId || scan.id;
          const isDuplicating = duplicateRunId === runUuid;
          const isActive = scan.status === "running" || scan.status === "pending";

          return (
            <CellShell className="w-full justify-end gap-1">
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="outline"
                    size="icon-sm"
                    className="border-info/40 text-info hover:bg-info-soft hover:text-info hover:shadow-none"
                    onClick={() => onSelectScan?.(scan)}
                    aria-label="View scan details"
                  >
                    <EyeIcon className="size-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="top">View details</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="outline"
                    size="icon-sm"
                    className="border-purple/40 text-purple hover:bg-purple-soft hover:text-purple hover:shadow-none"
                    onClick={() => setConfirmState({ action: "duplicate", scan })}
                    aria-label="Duplicate and start scan"
                    disabled={isDuplicating}
                  >
                    {isDuplicating ? (
                      <LoaderIcon className="size-4 animate-spin" />
                    ) : (
                      <CopyIcon className="size-4" />
                    )}
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="top">Duplicate &amp; start</TooltipContent>
              </Tooltip>
              {isActive ? (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="outline"
                      size="icon-sm"
                      className="border-warning/40 text-warning hover:bg-warning-soft hover:text-warning hover:shadow-none"
                      onClick={() => setConfirmState({ action: "cancel", scan })}
                      aria-label="Stop scan"
                    >
                      <StopCircleIcon className="size-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="top">Stop scan</TooltipContent>
                </Tooltip>
              ) : (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="outline"
                      size="icon-sm"
                      className="border-destructive/40 text-destructive hover:bg-destructive-soft hover:text-destructive hover:shadow-none"
                      onClick={() => setConfirmState({ action: "delete", scan })}
                      aria-label="Delete scan"
                    >
                      <TrashIcon className="size-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="top">Delete scan</TooltipContent>
                </Tooltip>
              )}
            </CellShell>
          );
        },
      },
    ],
    [duplicateRunId, onSelectScan]
  );

  if (isLoading) {
    return <TableSkeleton rows={5} columns={7} />;
  }

  return (
    <>
      <TooltipProvider>
        <DataGrid<Scan>
          rows={scans}
          columns={columns}
          getRowId={(scan) => String(scan.id)}
          // Row heights carry two stacked lines in the workflow and trigger
          // cells, so they need a little more room than the shared default.
          rowHeight={44}
          initialState={{ sort: { sortModel: [{ colId: "status", sort: "asc" }] } }}
          emptyState={
            <EmptyState
              icon={ScanSearchIcon}
              title="No scans found"
              description="Start your first security scan to see results here."
              action={{
                label: "New Scan",
                onClick: () => (window.location.href = "/scans/new"),
              }}
            />
          }
        />
      </TooltipProvider>
      <Dialog
        open={!!confirmState}
        onOpenChange={(open) => (!open ? setConfirmState(null) : null)}
      >
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>
              {confirmState?.action === "duplicate"
                ? "Duplicate and start scan"
                : confirmState?.action === "cancel"
                  ? "Stop running scan"
                  : "Delete scan"}
            </DialogTitle>
            <DialogDescription>
              {confirmState?.action === "duplicate"
                ? "Create a new run and start it immediately?"
                : confirmState?.action === "cancel"
                  ? "Stop the selected scan? This cannot be undone."
                  : "Delete the selected scan? This cannot be undone."}
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setConfirmState(null)}>
              Cancel
            </Button>
            <Button
              variant={confirmState?.action === "delete" ? "destructive" : "default"}
              onClick={handleConfirmAction}
            >
              Confirm
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
