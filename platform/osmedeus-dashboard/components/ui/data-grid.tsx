"use client";

import * as React from "react";
import { AgGridReact, type AgGridReactProps } from "ag-grid-react";
import {
  AllCommunityModule,
  ModuleRegistry,
  themeQuartz,
  type ColDef,
} from "ag-grid-community";
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

// Community modules are registered once for the whole app. AG Grid throws at
// render time if a feature is used before its module is registered, and this
// file is the only entry point to the grid.
ModuleRegistry.registerModules([AllCommunityModule]);

/**
 * The grid runs on the app's own tokens rather than a stock AG Grid palette.
 * Every value here is a `var(--…)`, so the grid re-paints with the rest of the
 * UI when `.dark` flips — no second theme object, no remount, no flash. AG Grid
 * derives a handful of shades from these with `color-mix()`, which resolves
 * against the same variables at paint time.
 */
export const osmedeusGridTheme = themeQuartz.withParams({
  backgroundColor: "var(--card)",
  foregroundColor: "var(--foreground)",
  borderColor: "var(--og-border-subtle)",
  chromeBackgroundColor: "var(--card)",
  accentColor: "var(--primary)",

  // Header: the same small uppercase eyebrow the rest of the app uses, closed
  // by the full-strength rule while body rows only get the subtle one.
  headerBackgroundColor: "var(--card)",
  headerTextColor: "var(--og-text-faint)",
  headerFontSize: "10.5px",
  headerFontWeight: 400,
  headerRowBorder: { color: "var(--og-border)" },
  headerColumnBorder: false,
  headerColumnResizeHandleColor: "var(--og-border)",
  headerCellHoverBackgroundColor: "var(--muted)",

  rowBorder: { color: "var(--og-border-subtle)" },
  rowHoverColor: "var(--muted)",
  // `--wash-primary`, not the Tailwind `--color-*` alias: `@theme inline`
  // substitutes those into utilities rather than emitting them as variables,
  // so there is nothing for AG Grid's `var()` to resolve against.
  selectedRowBackgroundColor: "var(--wash-primary)",
  oddRowBackgroundColor: "transparent",
  columnBorder: false,
  wrapperBorder: false,
  wrapperBorderRadius: 0,

  cellTextColor: "var(--foreground)",
  cellFontFamily: "var(--font-sans)",
  fontFamily: "var(--font-sans)",
  dataFontSize: "12.5px",
  fontSize: "12.5px",

  inputBackgroundColor: "var(--card)",
  inputBorder: { color: "var(--og-border)" },
  inputFocusBorder: { color: "var(--primary)" },
  menuBackgroundColor: "var(--popover)",
  menuTextColor: "var(--foreground)",
  borderRadius: "8px",
  checkboxUncheckedBorderColor: "var(--og-border)",
  checkboxCheckedBackgroundColor: "var(--primary)",
  checkboxCheckedBorderColor: "var(--primary)",
  iconButtonHoverBackgroundColor: "var(--muted)",
  focusShadow: { spread: 3, color: "color-mix(in oklab, var(--ring) 40%, transparent)" },
});

export type GridDensity = "comfortable" | "compact";

/** Row / header heights per density, matching the HTML tables they replace. */
const DENSITY: Record<GridDensity, { rowHeight: number; headerHeight: number }> = {
  comfortable: { rowHeight: 36, headerHeight: 30 },
  compact: { rowHeight: 28, headerHeight: 28 },
};

export interface DataGridProps<T>
  extends Omit<AgGridReactProps<T>, "theme" | "getRowId"> {
  rows: T[];
  columns: ColDef<T>[];
  /**
   * Stable row identity, taking the row itself rather than AG Grid's params
   * wrapper — it lets the grid patch rows in place instead of redrawing the
   * body on every data refresh.
   */
  getRowId?: (row: T) => string;
  density?: GridDensity;
  /** Fixed viewport height. Omit to let the grid size itself to its rows. */
  height?: number | string;
  /** Cap on the auto-height the grid grows to before it starts scrolling. */
  maxAutoHeight?: number;
  loading?: boolean;
  /** Rendered in place of the grid when there are no rows. */
  emptyState?: React.ReactNode;
  /** Rendered in place of the grid on the first load, before any rows exist. */
  loadingState?: React.ReactNode;
  onRowClick?: (row: T) => void;
  className?: string;
}

/**
 * Virtualised data grid. Only the rows and columns inside the viewport are in
 * the DOM, so a 500-row page costs about the same as a 20-row one — which is
 * the whole reason these tables moved off plain `<table>` markup.
 */
export function DataGrid<T>({
  rows,
  columns,
  getRowId,
  density = "comfortable",
  height,
  maxAutoHeight = 640,
  loading = false,
  emptyState,
  loadingState = (
    <div className="py-10 text-center text-sm text-muted-foreground">Loading...</div>
  ),
  onRowClick,
  className,
  // Destructured rather than spread so an explicit override also feeds the
  // auto-height sum below — otherwise taller rows silently overflow the box.
  rowHeight: rowHeightProp,
  headerHeight: headerHeightProp,
  ...rest
}: DataGridProps<T>) {
  const rowHeight = rowHeightProp ?? DENSITY[density].rowHeight;
  const headerHeight = headerHeightProp ?? DENSITY[density].headerHeight;

  const defaultColDef = React.useMemo<ColDef<T>>(
    () => ({
      sortable: true,
      resizable: true,
      // Filters and the floating filter row are opt-in per column: turning them
      // on globally puts a menu button in every header, which reads as noise on
      // a table whose filtering already lives in the toolbar above it.
      filter: false,
      suppressHeaderMenuButton: true,
      suppressMovable: true,
      minWidth: 80,
      // Every column shares the grid width unless it opts out with `flex: 0`
      // plus a `width` — status chips and action buttons do, so they keep their
      // size while the text columns absorb whatever is left.
      flex: 1,
    }),
    []
  );

  const resolvedGetRowId = React.useMemo(
    () => (getRowId ? ({ data }: { data: T }) => getRowId(data) : undefined),
    [getRowId]
  );

  // The first load and the empty result both replace the grid outright rather
  // than dressing it with an overlay — an empty grid is just a header over a
  // blank box, which reads as broken. A *refresh* keeps the grid (and its
  // scroll position) and lets the page dim it instead.
  if (rows.length === 0) {
    if (loading) return <>{loadingState}</>;
    if (emptyState) return <>{emptyState}</>;
  }

  // Beyond the header and rows themselves: 1px for the header's bottom border
  // and 10px for the horizontal-scroll gutter AG Grid always keeps in the
  // layout — that is the `::-webkit-scrollbar` height set in globals.css. Leave
  // them out and the box lands ~11px short, which puts a vertical scrollbar on
  // a table that would otherwise fit exactly.
  const autoHeight = Math.min(
    maxAutoHeight,
    headerHeight + 1 + Math.max(rows.length, 3) * rowHeight + 10
  );

  return (
    <div
      className={cn("og-grid w-full", className)}
      style={{ height: height ?? autoHeight }}
    >
      <AgGridReact<T>
        theme={osmedeusGridTheme}
        rowData={rows}
        columnDefs={columns}
        defaultColDef={defaultColDef}
        getRowId={resolvedGetRowId}
        rowHeight={rowHeight}
        headerHeight={headerHeight}
        // Row animation is the one thing that gets expensive at 500 rows, and
        // these tables re-sort on every header click.
        animateRows={false}
        suppressCellFocus
        suppressColumnVirtualisation={false}
        rowBuffer={8}
        enableCellTextSelection
        ensureDomOrder
        onRowClicked={onRowClick ? (e) => e.data && onRowClick(e.data) : undefined}
        rowClass={onRowClick ? "cursor-pointer" : undefined}
        overlayNoRowsTemplate="&nbsp;"
        {...rest}
      />
    </div>
  );
}

/* -------------------------------------------------------------------------- */
/* Pagination                                                                  */
/* -------------------------------------------------------------------------- */

export interface GridPaginationState {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
}

/**
 * Server-side pager shared by every grid. The grid itself never paginates —
 * the page of rows it is handed is the page the API returned — so this only
 * reports the window and asks for a different one.
 */
export function GridPagination({
  pagination,
  onPageChange,
  noun = "results",
}: {
  pagination: GridPaginationState;
  onPageChange?: (page: number) => void;
  noun?: string;
}) {
  if (pagination.totalPages <= 1) return null;

  const { page, pageSize, totalItems, totalPages } = pagination;
  const windowSize = Math.min(5, totalPages);

  // Keep the current page inside a five-wide window without letting the window
  // run off either end of the range.
  const firstInWindow =
    totalPages <= windowSize
      ? 1
      : page <= 3
        ? 1
        : page >= totalPages - 2
          ? totalPages - windowSize + 1
          : page - 2;

  return (
    <div className="flex items-center justify-between border-t border-border px-2 pt-3">
      <p className="text-sm text-muted-foreground">
        Showing{" "}
        <span className="font-medium text-foreground">{(page - 1) * pageSize + 1}</span>{" "}
        to{" "}
        <span className="font-medium text-foreground">
          {Math.min(page * pageSize, totalItems)}
        </span>{" "}
        of{" "}
        <span className="font-medium text-foreground">{totalItems.toLocaleString()}</span>{" "}
        {noun}
      </p>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange?.(page - 1)}
          disabled={page <= 1}
        >
          <ChevronLeftIcon className="mr-1 size-4" />
          Previous
        </Button>
        <div className="flex items-center gap-1">
          {Array.from({ length: windowSize }, (_, i) => {
            const pageNum = firstInWindow + i;
            return (
              <Button
                key={pageNum}
                variant={page === pageNum ? "default" : "outline"}
                size="sm"
                className="w-9"
                onClick={() => onPageChange?.(pageNum)}
              >
                {pageNum}
              </Button>
            );
          })}
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange?.(page + 1)}
          disabled={page >= totalPages}
        >
          Next
          <ChevronRightIcon className="ml-1 size-4" />
        </Button>
      </div>
    </div>
  );
}

/** Dimming overlay for a refresh that lands on top of already-visible rows. */
export function GridRefreshOverlay() {
  return (
    <div className="absolute inset-0 z-20 flex items-center justify-center bg-background/50">
      <div className="flex items-center gap-2 rounded-control border border-border bg-background px-3 py-2 text-sm text-muted-foreground">
        <div className="size-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        Refreshing...
      </div>
    </div>
  );
}

/* -------------------------------------------------------------------------- */
/* Shared cell renderers                                                       */
/* -------------------------------------------------------------------------- */

/**
 * Monospace value with the app's muted-dash fallback for empties.
 *
 * `truncate` is not optional: cells are flex containers (see `.og-grid .ag-cell`
 * in globals.css), and AG Grid's own `text-overflow: ellipsis` cannot reach the
 * anonymous flex item a bare value becomes — it has to sit on a real element.
 * A long value in a renderer without it clips mid-character instead.
 *
 * `block` is the other half: `overflow` is ignored on a non-replaced inline
 * box, and AG Grid's `.ag-cell-value` leaves this span inline, so without it
 * `truncate` renders no ellipsis at all.
 */
export function MonoCell({
  value,
  className,
}: {
  value?: string | number | null;
  className?: string;
}) {
  if (value === null || value === undefined || value === "") {
    return <span className="text-muted-foreground">-</span>;
  }
  return (
    <span
      className={cn("block min-w-0 truncate font-mono text-xs", className)}
      title={String(value)}
    >
      {value}
    </span>
  );
}

/** Plain text with the same empty fallback. */
export function TextCell({ value }: { value?: string | number | null }) {
  if (value === null || value === undefined || value === "") {
    return <span className="text-muted-foreground">-</span>;
  }
  return (
    <span className="block min-w-0 truncate" title={String(value)}>
      {value}
    </span>
  );
}

/**
 * Wraps cell content so it centres on the row and clips rather than wraps.
 * AG Grid cells are flex containers; chips and buttons need the alignment.
 */
export function CellShell({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex h-full min-w-0 items-center gap-1.5", className)}>
      {children}
    </div>
  );
}

export type { ColDef as GridColDef };
