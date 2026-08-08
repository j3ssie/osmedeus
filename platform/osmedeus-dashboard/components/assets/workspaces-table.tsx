"use client";

import * as React from "react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { EmptyState } from "@/components/shared/empty-state";
import {
  CellShell,
  DataGrid,
  GridPagination,
  GridRefreshOverlay,
  type GridColDef,
} from "@/components/ui/data-grid";
import { formatNumber } from "@/lib/utils";
import type { Workspace, WorkspaceSortState, WorkspaceSortField } from "@/lib/types/asset";
import { ArchiveIcon, FolderOpenIcon, EyeIcon, SearchXIcon } from "lucide-react";

interface WorkspacesTableProps {
  workspaces: Workspace[];
  isLoading?: boolean;
  pagination?: {
    page: number;
    pageSize: number;
    totalItems: number;
    totalPages: number;
  };
  sortState: WorkspaceSortState;
  onSort: (field: WorkspaceSortField) => void;
  onPageChange?: (page: number) => void;
  hasActiveFilters?: boolean;
}

function getRiskBadgeClass(score: number): string {
  if (score >= 8) return "bg-destructive-soft text-destructive border-transparent";
  if (score >= 6) return "bg-orange-soft text-orange border-transparent";
  if (score >= 4) return "bg-warning-soft text-warning border-transparent";
  return "bg-success-soft text-success border-transparent";
}

function VulnBadge({
  count,
  severity,
}: {
  count: number;
  severity: "critical" | "high" | "medium" | "low";
}) {
  if (count === 0) return null;

  const colors = {
    critical: "bg-destructive-soft text-destructive",
    high: "bg-orange-soft text-orange",
    medium: "bg-warning-soft text-warning",
    low: "bg-info-soft text-info",
  };

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span
          className={`inline-flex h-5 min-w-[20px] items-center justify-center rounded px-1.5 text-xs font-medium ${colors[severity]}`}
        >
          {count}
        </span>
      </TooltipTrigger>
      <TooltipContent side="top">
        {count} {severity}
      </TooltipContent>
    </Tooltip>
  );
}

type TagBadgeVariant =
  | "success"
  | "warning"
  | "info"
  | "purple"
  | "pink"
  | "cyan"
  | "orange"
  | "secondary";

function hashString(input: string): number {
  let h = 0;
  for (let i = 0; i < input.length; i++) {
    h = (h * 31 + input.charCodeAt(i)) >>> 0;
  }
  return h;
}

/** Stable pseudo-random tone per tag, so a tag keeps its colour across pages. */
function getTagVariant(tag: string): TagBadgeVariant {
  const variants: TagBadgeVariant[] = [
    "success",
    "warning",
    "info",
    "purple",
    "pink",
    "cyan",
    "orange",
    "secondary",
  ];
  return variants[hashString(tag) % variants.length];
}

function TagsCell({ tags }: { tags: string[] }) {
  if (!tags?.length) {
    return <span className="text-muted-foreground">-</span>;
  }

  const visible = tags.slice(0, 2);
  const remaining = tags.length - visible.length;

  return (
    <CellShell className="gap-1">
      {visible.map((t, i) => (
        <Badge key={`${t}-${i}`} variant={getTagVariant(t)} className="text-xs">
          {t}
        </Badge>
      ))}
      {remaining > 0 ? (
        <Tooltip>
          <TooltipTrigger asChild>
            <Badge variant="outline" className="cursor-default text-xs">
              +{remaining}
            </Badge>
          </TooltipTrigger>
          <TooltipContent side="top">
            <div className="max-w-[260px] break-words text-xs">{tags.join(", ")}</div>
          </TooltipContent>
        </Tooltip>
      ) : null}
    </CellShell>
  );
}

export function WorkspacesTable({
  workspaces,
  isLoading,
  pagination,
  sortState,
  onSort,
  onPageChange,
  hasActiveFilters,
}: WorkspacesTableProps) {
  const columns = React.useMemo<GridColDef<Workspace>[]>(
    () => [
      {
        field: "name",
        headerName: "Name",
        minWidth: 180,
        flex: 2,
        cellRenderer: (p: { value: string }) => (
          <Link
            href={`/inventory/workspaces/${p.value}`}
            className="truncate font-medium text-primary hover:underline"
          >
            {p.value}
          </Link>
        ),
      },
      {
        colId: "tags",
        headerName: "Tags",
        minWidth: 180,
        flex: 2,
        sortable: false,
        cellRenderer: (p: { data: Workspace }) => <TagsCell tags={p.data.tags} />,
      },
      {
        field: "total_assets",
        headerName: "Assets",
        minWidth: 90,
        flex: 0,
        width: 100,
        type: "numericColumn",
        cellClass: "font-mono tabular-nums text-muted-foreground",
        valueFormatter: (p) => formatNumber(p.value ?? 0),
      },
      {
        field: "total_subdomains",
        headerName: "Subdomains",
        minWidth: 110,
        flex: 0,
        width: 120,
        type: "numericColumn",
        cellClass: "font-mono tabular-nums text-muted-foreground",
        valueFormatter: (p) => formatNumber(p.value ?? 0),
      },
      {
        field: "total_urls",
        headerName: "URLs",
        minWidth: 90,
        flex: 0,
        width: 100,
        type: "numericColumn",
        cellClass: "font-mono tabular-nums text-muted-foreground",
        valueFormatter: (p) => formatNumber(p.value ?? 0),
      },
      {
        field: "total_vulns",
        headerName: "Vulnerabilities",
        minWidth: 170,
        cellRenderer: (p: { data: Workspace }) => {
          const ws = p.data;
          const hasVulns =
            ws.vuln_critical > 0 ||
            ws.vuln_high > 0 ||
            ws.vuln_medium > 0 ||
            ws.vuln_low > 0;
          if (!hasVulns) return <span className="text-muted-foreground">-</span>;
          return (
            <CellShell className="gap-1">
              <VulnBadge count={ws.vuln_critical} severity="critical" />
              <VulnBadge count={ws.vuln_high} severity="high" />
              <VulnBadge count={ws.vuln_medium} severity="medium" />
              <VulnBadge count={ws.vuln_low} severity="low" />
              <span className="ml-1 text-xs text-muted-foreground">({ws.total_vulns})</span>
            </CellShell>
          );
        },
      },
      {
        field: "risk_score",
        headerName: "Risk",
        minWidth: 90,
        flex: 0,
        width: 100,
        cellRenderer: (p: { value: number }) => (
          <Badge
            variant="outline"
            className={`font-mono text-xs ${getRiskBadgeClass(p.value)}`}
          >
            {p.value.toFixed(1)}
          </Badge>
        ),
      },
      {
        colId: "actions",
        headerName: "Actions",
        minWidth: 110,
        flex: 0,
        width: 120,
        sortable: false,
        cellRenderer: (p: { data: Workspace }) => (
          <CellShell className="w-full justify-center gap-2">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" size="icon" className="size-7" asChild>
                  <Link
                    href={{
                      pathname: "/inventory/assets",
                      query: { workspace: p.data.name },
                    }}
                  >
                    <EyeIcon className="size-3.5" />
                  </Link>
                </Button>
              </TooltipTrigger>
              <TooltipContent side="top">View assets</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" size="icon" className="size-7" asChild>
                  <Link
                    href={{
                      pathname: "/inventory/artifacts",
                      query: { workspace: p.data.name },
                    }}
                  >
                    <ArchiveIcon className="size-3.5" />
                  </Link>
                </Button>
              </TooltipTrigger>
              <TooltipContent side="top">View artifacts</TooltipContent>
            </Tooltip>
          </CellShell>
        ),
      },
    ],
    []
  );

  // Sorting is server-side, so the grid never reorders rows itself — it only
  // reports which header was clicked and paints the arrow for the active field.
  const columnsWithSort = React.useMemo<GridColDef<Workspace>[]>(
    () =>
      columns.map((col) => {
        const id = col.colId ?? (col.field as string);
        return {
          ...col,
          comparator: () => 0,
          sort:
            sortState.field === id ? (sortState.direction as "asc" | "desc") : null,
        };
      }),
    [columns, sortState]
  );

  const handleSortChanged = React.useCallback(
    (event: { columns?: { getColId(): string }[] | null }) => {
      const changed = event.columns?.[0];
      if (changed) onSort(changed.getColId() as WorkspaceSortField);
    },
    [onSort]
  );

  return (
    <TooltipProvider>
      <div className="relative space-y-4">
        {isLoading && workspaces.length > 0 && <GridRefreshOverlay />}

        <DataGrid<Workspace>
          rows={workspaces}
          columns={columnsWithSort}
          getRowId={(ws) => String(ws.id || ws.name)}
          loading={isLoading}
          onSortChanged={handleSortChanged}
          emptyState={
            <div className="relative flex min-h-[360px] items-center justify-center">
              <EmptyState
                icon={hasActiveFilters ? SearchXIcon : FolderOpenIcon}
                title={hasActiveFilters ? "No matching workspaces" : "No workspaces found"}
                description={
                  hasActiveFilters
                    ? "No workspaces match your current search. Try adjusting your search criteria."
                    : "Workspaces are created when you run scans. Start a new scan to create a workspace."
                }
              />
            </div>
          }
        />

        {pagination && (
          <GridPagination
            pagination={pagination}
            onPageChange={onPageChange}
            noun="workspaces"
          />
        )}
      </div>
    </TooltipProvider>
  );
}
