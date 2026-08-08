"use client";

import * as React from "react";
import { clearEventLogTables, fetchEventLogs } from "@/lib/api/event-logs";
import type { EventLogItem } from "@/lib/types/event-log";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { RefreshCcwIcon, EyeIcon, Trash2Icon } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Select, SelectTrigger, SelectContent, SelectItem, SelectValue } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import type { PaginatedResponse } from "@/lib/types/api";
import {
  CellShell,
  DataGrid,
  GridPagination,
  type GridColDef,
} from "@/components/ui/data-grid";

export default function EventsPage() {
  const [events, setEvents] = React.useState<EventLogItem[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [query, setQuery] = React.useState("");
  const [selected, setSelected] = React.useState<EventLogItem | null>(null);
  const [open, setOpen] = React.useState(false);
  const [detail, setDetail] = React.useState<EventLogItem | null>(null);
  const [clearOpen, setClearOpen] = React.useState(false);
  const [clearing, setClearing] = React.useState(false);
  const [page, setPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(20);
  const [pagination, setPagination] = React.useState<PaginatedResponse<EventLogItem>["pagination"] | null>(null);
  const [filters, setFilters] = React.useState<{
    topic?: string;
    workspace?: string;
    processed?: boolean | undefined;
  }>({});

  const filteredEvents = React.useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return events;
    return events.filter((e) => {
      const haystack = [
        e.topic,
        e.name,
        e.source,
        e.workspace,
        e.runId,
        e.workflowName,
        e.eventId,
        e.dataType,
        e.data,
        e.error,
      ]
        .filter(Boolean)
        .join("\n")
        .toLowerCase();
      return haystack.includes(q);
    });
  }, [events, query]);

  const workspaceOptions = React.useMemo(() => {
    const values = new Set<string>();
    events.forEach((event) => {
      if (event.workspace) values.add(event.workspace);
    });
    return Array.from(values).sort((a, b) => a.localeCompare(b));
  }, [events]);

  const load = React.useCallback(async () => {
    try {
      setLoading(true);
      const res = await fetchEventLogs({
        page,
        pageSize,
        filters: {
          topic: filters.topic?.trim() || undefined,
          workspace: filters.workspace?.trim() || undefined,
          processed: typeof filters.processed === "boolean" ? filters.processed : undefined,
        },
      });
      setEvents(res.data);
      setPagination(res.pagination);
    } catch (e) {
      toast.error("Failed to load events", { description: e instanceof Error ? e.message : "" });
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, filters]);

  React.useEffect(() => {
    load();
  }, [load]);

  const openDetail = async (e: EventLogItem) => {
    setSelected(e);
    setDetail(e);
    setOpen(true);
  };

  const handleClearTables = async () => {
    try {
      setClearing(true);
      await clearEventLogTables();
      toast.success("Event tables cleared");
      setClearOpen(false);
      if (page === 1) {
        await load();
      } else {
        setPage(1);
      }
    } catch (e) {
      toast.error("Failed to clear tables", { description: e instanceof Error ? e.message : "" });
    } finally {
      setClearing(false);
    }
  };

  const topicToVariant = (t: string): "success" | "info" | "destructive" | "secondary" => {
    const x = t.toLowerCase();
    if (x.includes("completed") || x.includes("success")) return "success";
    if (x.includes("started") || x.includes("running")) return "info";
    if (x.includes("failed") || x.includes("error")) return "destructive";
    return "secondary";
  };
  const sourceToVariant = (s: string): "info" | "secondary" | "outline" => {
    const x = (s || "").toLowerCase();
    if (x === "executor" || x === "api") return "info";
    if (x === "scheduler") return "secondary";
    return "outline";
  };
  const processedToVariant = (p: boolean): "success" | "warning" => (p ? "success" : "warning");

  const columns = React.useMemo<GridColDef<EventLogItem>[]>(
    () => [
      {
        field: "topic",
        headerName: "Topic",
        minWidth: 150,
        cellRenderer: (p: { value: string }) => (
          <Badge variant={topicToVariant(p.value)}>{p.value}</Badge>
        ),
      },
      { field: "name", headerName: "Name", minWidth: 180, flex: 2 },
      {
        field: "source",
        headerName: "Source",
        minWidth: 120,
        cellRenderer: (p: { value: string }) => (
          <Badge variant={sourceToVariant(p.value)}>{p.value}</Badge>
        ),
      },
      {
        field: "workspace",
        headerName: "Workspace",
        minWidth: 140,
        valueFormatter: (p) => p.value || "-",
      },
      {
        field: "workflowName",
        colId: "workflow",
        headerName: "Workflow",
        minWidth: 140,
        valueFormatter: (p) => p.value || "-",
      },
      {
        field: "processed",
        headerName: "Processed",
        minWidth: 130,
        flex: 0,
        width: 140,
        cellRenderer: (p: { value: boolean }) => (
          <Badge variant={processedToVariant(p.value)}>
            {p.value ? "Processed" : "Unprocessed"}
          </Badge>
        ),
      },
      {
        colId: "detail",
        headerName: "Detail",
        minWidth: 90,
        flex: 0,
        width: 90,
        sortable: false,
        cellRenderer: (p: { data: EventLogItem }) => (
          <CellShell>
            <Button
              variant="outline"
              size="icon-sm"
              onClick={() => openDetail(p.data)}
              aria-label="View detail"
            >
              <EyeIcon className="size-4" />
            </Button>
          </CellShell>
        ),
      },
    ],
    // The `*ToVariant` helpers and `openDetail` are pure/stable, so the column
    // set only needs building once.
    []
  );

  return (
    <div className="space-y-4">
      

      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <CardTitle>Event Logs</CardTitle>
              <CardDescription>Filter by topic, workspace, or processed status</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" onClick={() => setPage(1)}>
                <RefreshCcwIcon className="mr-2 size-4" />
                Apply
              </Button>
              <Button
                variant="outline"
                onClick={() => {
                  setFilters({});
                  setQuery("");
                  setPage(1);
                  setPageSize(20);
                }}
              >
                Reset
              </Button>
              <Button
                variant="outline"
                onClick={() => setClearOpen(true)}
                className="gap-2 border-destructive/40 text-destructive hover:bg-destructive-soft hover:text-destructive hover:shadow-none"
              >
                <Trash2Icon className="size-4" />
                Empty Table
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3 py-2">
            <Input
              placeholder="Search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="w-[360px] lg:w-[520px]"
            />
            <Input
              placeholder="Topic"
              value={filters.topic ?? ""}
              onChange={(e) => setFilters((f) => ({ ...f, topic: e.target.value }))}
              className="max-w-[200px]"
            />
            <Select
              value={filters.workspace || "all"}
              onValueChange={(val) => {
                setFilters((f) => ({
                  ...f,
                  workspace: val === "all" ? undefined : val,
                }));
                setPage(1);
              }}
            >
              <SelectTrigger className="max-w-[200px]">
                <SelectValue placeholder="Workspace" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Workspaces</SelectItem>
                {workspaceOptions.map((workspace) => (
                  <SelectItem key={workspace} value={workspace}>
                    {workspace}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={
                typeof filters.processed === "boolean"
                  ? String(filters.processed)
                  : "all"
              }
              onValueChange={(val) => {
                setFilters((f) => ({
                  ...f,
                  processed: val === "all" ? undefined : val === "true",
                }));
                setPage(1);
              }}
            >
              <SelectTrigger className="max-w-[180px]">
                <SelectValue placeholder="Processed" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All</SelectItem>
                <SelectItem value="true">Processed</SelectItem>
                <SelectItem value="false">Unprocessed</SelectItem>
              </SelectContent>
            </Select>
            <Select
              value={String(pageSize)}
              onValueChange={(val) => {
                const n = parseInt(val, 10);
                setPageSize(Number.isNaN(n) ? 20 : n);
                setPage(1);
              }}
            >
              <SelectTrigger className="max-w-[140px]">
                <SelectValue placeholder="Page Size" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="20">20</SelectItem>
                <SelectItem value="50">50</SelectItem>
                <SelectItem value="100">100</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <DataGrid<EventLogItem>
            rows={filteredEvents}
            columns={columns}
            getRowId={(e) => String(e.id)}
            loading={loading}
            emptyState={
              <div className="py-10 text-center text-sm text-muted-foreground">
                {events.length === 0 ? "No events" : "No matching events"}
              </div>
            }
          />

          {pagination && (
            <GridPagination pagination={pagination} onPageChange={setPage} />
          )}
        </CardContent>
      </Card>

      <Dialog open={clearOpen} onOpenChange={setClearOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Empty event tables</DialogTitle>
            <DialogDescription>
              This clears runs, step results, and event logs. This cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setClearOpen(false)} disabled={clearing}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleClearTables} disabled={clearing} className="gap-2">
              <Trash2Icon className="size-4" />
              {clearing ? "Emptying..." : "Empty Table"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Event Detail</DialogTitle>
            <DialogDescription>Event log payload</DialogDescription>
          </DialogHeader>
          {selected ? (
            <div className="space-y-2 text-sm">
              <div>Topic: {selected.topic}</div>
              <div>Name: {selected.name}</div>
              <div>Source: {selected.source}</div>
              <div>Workspace: {selected.workspace || "-"}</div>
              <div>Run ID: {selected.runId || "-"}</div>
              <div>Workflow: {selected.workflowName || "-"}</div>
              <div>Processed: {selected.processed ? "true" : "false"}</div>
              <div>Created: {detail?.createdAt ? detail.createdAt.toLocaleString() : "-"}</div>
              <div>Processed At: {detail?.processedAt ? detail.processedAt.toLocaleString() : "-"}</div>
              {detail?.data ? (
                <div className="mt-2 rounded bg-muted p-2">
                  <div className="font-medium">Data</div>
                  <pre className="whitespace-pre-wrap break-words text-xs">
                    {(() => {
                      try {
                        const obj = JSON.parse(detail.data || "");
                        return JSON.stringify(obj, null, 2);
                      } catch {
                        return detail.data;
                      }
                    })()}
                  </pre>
                </div>
              ) : null}
              {detail?.error ? (
                <div className="mt-2 rounded bg-muted p-2">
                  <div className="font-medium">Error</div>
                  <pre className="whitespace-pre-wrap break-words text-xs text-destructive">{detail.error}</pre>
                </div>
              ) : null}
            </div>
          ) : null}
        </DialogContent>
      </Dialog>
    </div>
  );
}
