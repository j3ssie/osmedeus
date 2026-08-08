"use client";

import * as React from "react";
import Link from "next/link";
import { useTheme } from "next-themes";
import { Light as SyntaxHighlighter } from "react-syntax-highlighter";
import bashLang from "react-syntax-highlighter/dist/esm/languages/hljs/bash";
import github from "react-syntax-highlighter/dist/esm/styles/hljs/github";
import atomOneDark from "react-syntax-highlighter/dist/esm/styles/hljs/atom-one-dark";
import {
  CalendarIcon,
  ClipboardIcon,
  ClipboardListIcon,
  PlayIcon,
  PlusIcon,
  RefreshCcwIcon,
  ScanSearchIcon,
  XIcon,
  TerminalIcon,
  FunctionSquareIcon,
  RepeatIcon,
} from "lucide-react";
import { toast } from "sonner";
import { Card, CardContent } from "@/components/ui/card";
import { SectionCardHeader } from "@/components/shared/section-card-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ScanTable } from "@/components/scans/scan-table";
import { ErrorState } from "@/components/shared/error-state";
import { fetchScans } from "@/lib/api/scans";
import { ScanStatusBadge } from "@/components/scans/scan-status-badge";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { fetchStepResults } from "@/lib/api/steps";
import type { Scan, StepResult } from "@/lib/types/scan";
import type { PaginatedResponse } from "@/lib/types/api";
import { cn, formatDateTime } from "@/lib/utils";

SyntaxHighlighter.registerLanguage("bash", bashLang);
const CodeHighlighter = SyntaxHighlighter as unknown as React.ComponentType<any>;

export default function ScansPage() {
  const { resolvedTheme } = useTheme();
  const [scans, setScans] = React.useState<Scan[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [page, setPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(20);
  const [pagination, setPagination] = React.useState<PaginatedResponse<Scan>["pagination"] | null>(null);
  const [filters, setFilters] = React.useState<{
    status?: string;
    workspace?: string;
  }>({});
  const [tableSearch, setTableSearch] = React.useState<string>("");
  const [selectedScan, setSelectedScan] = React.useState<Scan | null>(null);
  const [stepResults, setStepResults] = React.useState<StepResult[]>([]);
  const [stepPagination, setStepPagination] = React.useState<{
    limit: number;
    offset: number;
    total: number;
  } | null>(null);
  const [isStepsLoading, setIsStepsLoading] = React.useState(false);
  const [stepsError, setStepsError] = React.useState<string | null>(null);
  const [areStepsExpanded, setAreStepsExpanded] = React.useState(true);

  const loadScans = React.useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await fetchScans({
        page,
        pageSize,
        filters: {
          status: filters.status || undefined,
          workspace: filters.workspace || undefined,
        },
      });
      setScans(response.data);
      setPagination(response.pagination);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load scans");
      toast.error("Failed to load scans", {
        description: err instanceof Error ? err.message : "",
      });
    } finally {
      setIsLoading(false);
    }
  }, [page, pageSize, filters]);

  React.useEffect(() => {
    loadScans();
  }, [loadScans]);

  const filteredScans = React.useMemo(() => {
    const q = tableSearch.trim().toLowerCase();
    return scans.filter((s) => {
      const parts = [
        s.id,
        s.runId,
        s.runUuid,
        s.workflowName,
        s.workflowKind,
        s.target,
        s.status,
        s.priority,
        s.triggerType,
        s.triggerName,
        s.workspacePath,
        s.errorMessage,
        s.startedAt?.toISOString(),
        s.completedAt?.toISOString(),
        s.createdAt?.toISOString(),
        s.updatedAt?.toISOString(),
      ].filter(Boolean);

      let paramsText = "";
      if (s.params && typeof s.params === "object") {
        try {
          paramsText = JSON.stringify(s.params);
        } catch {
          paramsText = "";
        }
      }

      const haystack = `${parts.join(" ")} ${paramsText}`.toLowerCase();
      return q ? haystack.includes(q) : true;
    });
  }, [scans, tableSearch]);

  const formatDetailValue = React.useCallback((key: string, value: unknown) => {
    if (value === null || value === undefined || value === "") return "-";
    if (key.endsWith("At")) {
      if (value instanceof Date) return value.toISOString();
      if (typeof value === "string") return value;
    }
    if (typeof value === "number" || typeof value === "boolean") return String(value);
    if (typeof value === "string") return value;
    if (typeof value === "object") {
      try {
        return JSON.stringify(value, null, 2);
      } catch {
        return "[object]";
      }
    }
    return String(value);
  }, []);

  const detailItems = React.useMemo(() => {
    if (!selectedScan)
      return [] as Array<{
        key: keyof Scan | string;
        label: string;
        value: string;
        mono?: boolean;
      }>;
    const stepsSummary =
      selectedScan.completedSteps || selectedScan.totalSteps
        ? `${selectedScan.completedSteps ?? 0} / ${selectedScan.totalSteps ?? 0}`
        : "-";
    const items: Array<{ key: keyof Scan | string; label: string; mono?: boolean; value?: string }> = [
      { key: "runUuid", label: "Run UUID", mono: true },
      { key: "workflowName", label: "Workflow Name" },
      { key: "workflowKind", label: "Workflow Kind" },
      { key: "target", label: "Target", mono: true },
      { key: "status", label: "Status" },
      { key: "priority", label: "Priority" },
      { key: "triggerType", label: "Trigger Type" },
      { key: "triggerName", label: "Trigger Name" },
      { key: "runGroupId", label: "Run Group ID", mono: true },
      { key: "workspace", label: "Workspace", mono: true },
      { key: "workspacePath", label: "Workspace Path", mono: true },
      { key: "stepsSummary", label: "Completed / Total", value: stepsSummary },
      { key: "errorMessage", label: "Error" },
    ];

    return items
      .map((it) => {
        const value =
          it.value ?? formatDetailValue(String(it.key), (selectedScan as any)[it.key]);
        return { key: it.key, label: it.label, value, mono: it.mono };
      })
      .filter((it) => it.value !== "-");
  }, [formatDetailValue, selectedScan]);

  const paramsEntries = selectedScan?.params
    ? Object.entries(selectedScan.params).map(([k, v]) => [k, String(v)] as const)
    : [];

  const workspaceOptions = React.useMemo(() => {
    const values = new Set<string>();
    scans.forEach((scan) => {
      if (scan.workspace) values.add(scan.workspace);
    });
    return Array.from(values).sort((a, b) => a.localeCompare(b));
  }, [scans]);

  const statusFilterStyles = React.useMemo(() => {
    const status = (filters.status || "all").toLowerCase();
    const map: Record<string, { trigger: string; dot: string }> = {
      all: { trigger: "border-muted-foreground/30", dot: "bg-muted-foreground/50" },
      pending: { trigger: "border-warning/50 text-warning", dot: "bg-warning" },
      running: { trigger: "border-info/50 text-info", dot: "bg-info" },
      completed: { trigger: "border-success/50 text-success", dot: "bg-success" },
      failed: { trigger: "border-destructive/50 text-destructive", dot: "bg-destructive" },
      cancelled: { trigger: "border-muted-foreground/40 text-muted-foreground", dot: "bg-muted-foreground" },
    };
    return map[status] ?? map.all;
  }, [filters.status]);

  const triggerConfig = React.useMemo(() => {
    const conf: Record<
      string,
      {
        label: string;
        variant: "default" | "secondary" | "destructive" | "outline" | "success" | "warning" | "info" | "purple" | "pink" | "cyan" | "orange";
        icon: React.ReactNode;
        className?: string;
      }
    > = {
      cli: { label: "CLI", variant: "purple", icon: <PlayIcon className="size-3" /> },
      api: { label: "API", variant: "info", icon: <PlayIcon className="size-3" /> },
      cron: { label: "Cron", variant: "warning", icon: <CalendarIcon className="size-3" /> },
      scheduled: { label: "Scheduled", variant: "warning", icon: <CalendarIcon className="size-3" /> },
      manual: { label: "Manual", variant: "secondary", icon: <PlayIcon className="size-3" /> },
    };
    return conf;
  }, []);

  const priorityConfig = React.useMemo(() => {
    const conf: Record<
      string,
      {
        label: string;
        className: string;
      }
    > = {
      high: { label: "High", className: "border-destructive/50 text-destructive" },
      medium: { label: "Medium", className: "border-warning/50 text-warning" },
      low: { label: "Low", className: "border-success/50 text-success" },
    };
    return conf;
  }, []);

  const stepStatusConfig = React.useMemo(() => {
    const conf: Record<
      string,
      {
        label: string;
        variant: "default" | "secondary" | "destructive" | "outline" | "success" | "warning";
      }
    > = {
      pending: { label: "Pending", variant: "secondary" },
      running: { label: "Running", variant: "default" },
      completed: { label: "Completed", variant: "success" },
      success: { label: "Success", variant: "success" },
      failed: { label: "Failed", variant: "destructive" },
      error: { label: "Error", variant: "destructive" },
      skipped: { label: "Skipped", variant: "outline" },
    };
    return conf;
  }, []);

  const stepTypeConfig = React.useMemo(() => {
    const conf: Record<
      string,
      {
        label: string;
        variant: "default" | "secondary" | "destructive" | "outline" | "success" | "warning" | "info" | "purple" | "pink" | "cyan" | "orange";
        icon: React.ReactNode;
      }
    > = {
      bash: { label: "Bash", variant: "purple", icon: <TerminalIcon className="size-3" /> },
      function: { label: "Function", variant: "info", icon: <FunctionSquareIcon className="size-3" /> },
      foreach: { label: "Foreach", variant: "orange", icon: <RepeatIcon className="size-3" /> },
    };
    return conf;
  }, []);

  const formatStepDuration = (durationMs?: number): string => {
    if (durationMs === undefined || durationMs === null) return "-";
    if (durationMs < 1000) return `${durationMs}ms`;
    const sec = Math.floor(durationMs / 1000);
    if (sec < 60) return `${sec}s`;
    const min = Math.floor(sec / 60);
    if (min < 60) return `${min}m ${sec % 60}s`;
    const hour = Math.floor(min / 60);
    return `${hour}h ${min % 60}m`;
  };

  React.useEffect(() => {
    if (!selectedScan) {
      setStepResults([]);
      setStepPagination(null);
      setStepsError(null);
      return;
    }
    const runUuid =
      typeof selectedScan.runUuid === "string" ? selectedScan.runUuid.trim() : "";
    const runId =
      selectedScan.runId !== undefined && selectedScan.runId !== null
        ? String(selectedScan.runId).trim()
        : "";
    if (!runUuid && !runId) {
      setStepResults([]);
      setStepPagination(null);
      setStepsError(null);
      return;
    }
    let cancelled = false;
    setIsStepsLoading(true);
    setStepsError(null);
    fetchStepResults({
      runUuid: runUuid || undefined,
      runId: !runUuid ? runId : undefined,
      limit: 200,
      offset: 0,
    })
      .then((resp) => {
        if (cancelled) return;
        setStepResults(resp.data);
        setStepPagination(resp.pagination);
      })
      .catch((err) => {
        if (cancelled) return;
        setStepsError(err instanceof Error ? err.message : "Failed to load step results");
        setStepResults([]);
        setStepPagination(null);
      })
      .finally(() => {
        if (!cancelled) setIsStepsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [selectedScan]);

  React.useEffect(() => {
    setAreStepsExpanded(true);
  }, [selectedScan]);

  return (
    <div className="space-y-4">
      {/* Filters */}
      <Card className="overflow-hidden">
        <SectionCardHeader
          icon={ScanSearchIcon}
          title="Scans"
          description="Filter by status, workflow, or target"
          actions={
            <Button asChild>
              <Link href="/scans/new">
                <PlusIcon className="mr-2 size-4" />
                New Scan
              </Link>
            </Button>
          }
        />
        <CardContent>
          <div className="flex flex-wrap gap-2">
            <Input
              placeholder="Search"
              value={tableSearch}
              onChange={(e) => setTableSearch(e.target.value)}
              className="h-9 w-full md:w-1/4 max-w-none"
            />
            <Select
              value={filters.status || "all"}
              onValueChange={(val) => {
                setFilters((f) => ({
                  ...f,
                  status: val === "all" ? undefined : val,
                }));
                setPage(1);
              }}
            >
              <SelectTrigger className={cn("max-w-[180px]", statusFilterStyles.trigger)}>
                <span className="flex items-center gap-2">
                  <span className={cn("size-2 rounded-full", statusFilterStyles.dot)} />
                  <SelectValue placeholder="Status" />
                </span>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">
                  <span className="flex items-center gap-2">
                    <span className="size-2 rounded-full bg-muted-foreground/50" />
                    All Statuses
                  </span>
                </SelectItem>
                <SelectItem value="pending">
                  <span className="flex items-center gap-2 text-warning">
                    <span className="size-2 rounded-full bg-warning" />
                    Pending
                  </span>
                </SelectItem>
                <SelectItem value="running">
                  <span className="flex items-center gap-2 text-info">
                    <span className="size-2 rounded-full bg-info" />
                    Running
                  </span>
                </SelectItem>
                <SelectItem value="completed">
                  <span className="flex items-center gap-2 text-success">
                    <span className="size-2 rounded-full bg-success" />
                    Completed
                  </span>
                </SelectItem>
                <SelectItem value="failed">
                  <span className="flex items-center gap-2 text-destructive">
                    <span className="size-2 rounded-full bg-destructive" />
                    Failed
                  </span>
                </SelectItem>
              </SelectContent>
            </Select>
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
            <Button variant="outline" onClick={loadScans} disabled={isLoading}>
              <RefreshCcwIcon className={`mr-2 size-4 ${isLoading ? "animate-spin" : ""}`} />
              Refresh
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                setSelectedScan(null);
                setFilters({});
                setTableSearch("");
                setPage(1);
              }}
              disabled={!selectedScan && !filters.status && !filters.workspace && !tableSearch}
            >
              <XIcon className="mr-2 size-4" />
              Clear
            </Button>
          </div>

          {error ? (
            <ErrorState message={error} onRetry={loadScans} />
          ) : (
            <>
              <ScanTable
                scans={filteredScans}
                isLoading={isLoading}
                onRefresh={loadScans}
                onSelectScan={setSelectedScan}
              />

              {pagination && pagination.totalPages > 1 && (
                <div className="flex items-center justify-between px-2 py-3">
                  <p className="text-sm text-muted-foreground">
                    Showing {(pagination.page - 1) * pagination.pageSize + 1} to{" "}
                    {Math.min(
                      pagination.page * pagination.pageSize,
                      pagination.totalItems
                    )}{" "}
                    of {pagination.totalItems.toLocaleString()} results
                  </p>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage((p) => Math.max(1, p - 1))}
                      disabled={pagination.page <= 1}
                    >
                      Previous
                    </Button>
                    <div className="flex items-center gap-1">
                      {Array.from(
                        { length: Math.min(5, pagination.totalPages) },
                        (_, i) => {
                          let pageNum: number;
                          if (pagination.totalPages <= 5) {
                            pageNum = i + 1;
                          } else if (pagination.page <= 3) {
                            pageNum = i + 1;
                          } else if (pagination.page >= pagination.totalPages - 2) {
                            pageNum = pagination.totalPages - 4 + i;
                          } else {
                            pageNum = pagination.page - 2 + i;
                          }
                          return (
                            <Button
                              key={pageNum}
                              variant={
                                pagination.page === pageNum ? "default" : "outline"
                              }
                              size="sm"
                              className="w-9"
                              onClick={() => setPage(pageNum)}
                            >
                              {pageNum}
                            </Button>
                          );
                        }
                      )}
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        setPage((p) => Math.min(p + 1, pagination.totalPages))
                      }
                      disabled={pagination.page >= pagination.totalPages}
                    >
                      Next
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      <Card className="overflow-hidden">
        <SectionCardHeader
          icon={ClipboardListIcon}
          title="Scan Details"
          description={
            selectedScan
              ? "Selected scan details and step output"
              : "Select a scan to view details"
          }
          actions={
            <Button
              className="rounded-md"
              variant="outline"
              size="icon"
              disabled={!selectedScan?.id}
              onClick={async () => {
                if (!selectedScan?.id) return;
                try {
                  await navigator.clipboard.writeText(
                    JSON.stringify(selectedScan, null, 2)
                  );
                  toast.success("Run JSON copied");
                } catch {
                  toast.error("Failed to copy");
                }
              }}
            >
              <ClipboardIcon className="size-4" />
              <span className="sr-only">Copy JSON</span>
            </Button>
          }
        />
        <CardContent>
          <ScrollArea className="h-[70vh] lg:h-[75vh] rounded-md border bg-muted/20">
            <div className="p-5 space-y-5 text-sm">
              {selectedScan ? (
                <>
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                    {detailItems.map((item) => (
                      <div
                        key={item.label}
                        className={cn(
                          "space-y-1 rounded-md border border-transparent px-2 py-1.5 transition-colors",
                          item.key === "runUuid"
                            ? "cursor-pointer hover:border-primary/30 hover:bg-primary-soft"
                            : "hover:border-muted-foreground/10"
                        )}
                        onClick={async () => {
                          if (item.key !== "runUuid") return;
                          if (!selectedScan?.runUuid) return;
                          try {
                            await navigator.clipboard.writeText(selectedScan.runUuid);
                            toast.success("Run UUID copied");
                          } catch {
                            toast.error("Failed to copy");
                          }
                        }}
                        role={item.key === "runUuid" ? "button" : undefined}
                        tabIndex={item.key === "runUuid" ? 0 : undefined}
                      >
                        <div className="text-xs text-muted-foreground">{item.label}</div>
                        {item.key === "status" && selectedScan ? (
                          <div className="flex items-center gap-2">
                            <ScanStatusBadge status={selectedScan.status} />
                          </div>
                        ) : item.key === "triggerType" && selectedScan ? (
                          <Badge
                            variant={
                              triggerConfig[(selectedScan.triggerType || "manual").toLowerCase()]
                                ?.variant ?? "outline"
                            }
                            className="gap-1 w-fit"
                          >
                            {
                              triggerConfig[(selectedScan.triggerType || "manual").toLowerCase()]
                                ?.icon ?? <PlayIcon className="size-3" />
                            }
                            <span>{item.value}</span>
                          </Badge>
                        ) : item.key === "priority" && selectedScan ? (
                          selectedScan.priority ? (
                            <Badge
                              variant="outline"
                              className={cn(
                                "w-fit capitalize",
                                priorityConfig[String(selectedScan.priority).toLowerCase()]?.className
                              )}
                            >
                              {priorityConfig[String(selectedScan.priority).toLowerCase()]?.label ??
                                String(selectedScan.priority)}
                            </Badge>
                          ) : (
                            <span className="text-sm text-muted-foreground">-</span>
                          )
                        ) : item.key === "errorMessage" ? (
                          <div className="rounded-md border border-destructive/20 bg-destructive-soft px-2 py-1 font-mono text-destructive break-all">
                            {item.value}
                          </div>
                        ) : (
                          <div className={item.mono ? "font-mono break-all" : "break-all"}>
                            {item.value}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>

                  {paramsEntries.length > 0 && (
                    <div className="space-y-2">
                      <div className="text-xs text-muted-foreground">Params</div>
                      <textarea
                        readOnly
                        value={JSON.stringify(
                          Object.fromEntries(paramsEntries.map(([k, v]) => [k, v]))
                        )}
                        rows={4}
                        className="w-full rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words"
                      />
                    </div>
                  )}

                  <div className="space-y-2">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <div className="text-xs text-muted-foreground">Step Results</div>
                      <div className="flex flex-wrap items-center gap-2">
                        {stepPagination?.total ? (
                          <div className="text-xs text-muted-foreground">
                            {stepResults.length} of {stepPagination.total}
                          </div>
                        ) : null}
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-7 px-2 text-xs"
                          onClick={() => setAreStepsExpanded((prev) => !prev)}
                          disabled={isStepsLoading}
                        >
                          {areStepsExpanded ? "Hide Steps" : "Show Steps"}
                        </Button>
                      </div>
                    </div>
                    {stepResults.length > 0 ? (
                      <div className="flex flex-wrap gap-2">
                        {stepResults.map((step, idx) => {
                          const anchorId = step.id
                            ? `step-result-${step.id}`
                            : `step-result-${idx}`;
                          const label = step.stepName || `Step ${idx + 1}`;
                          const stepTypeKey = String(step.stepType || "unknown").toLowerCase();
                          const stepType =
                            stepTypeConfig[stepTypeKey] ?? {
                              label: step.stepType ? String(step.stepType) : "Unknown",
                              variant: "outline" as const,
                              icon: <PlayIcon className="size-3" />,
                            };
                          return (
                            <Button
                              key={anchorId}
                              variant="outline"
                              size="sm"
                              className="h-7 px-2 text-xs"
                              onClick={() => {
                                setAreStepsExpanded(true);
                                setTimeout(() => {
                                  const el = document.getElementById(anchorId);
                                  el?.scrollIntoView({ behavior: "smooth", block: "start" });
                                }, 0);
                              }}
                            >
                              <span className="flex items-center gap-1.5">
                                {stepType.icon}
                                <span className="max-w-[140px] truncate">{label}</span>
                              </span>
                            </Button>
                          );
                        })}
                      </div>
                    ) : null}
                    {!areStepsExpanded ? null : (
                    <>
                    {isStepsLoading ? (
                      <TableSkeleton rows={3} columns={3} />
                    ) : stepsError ? (
                      <div className="rounded-md border border-destructive/30 bg-destructive-soft px-3 py-2 text-sm text-destructive">
                        {stepsError}
                      </div>
                    ) : stepResults.length === 0 ? (
                      <div className="text-sm text-muted-foreground">
                        No step results available for this run.
                      </div>
                    ) : (
                      <div className="space-y-3">
                        {stepResults.map((step, idx) => {
                          const statusKey = String(step.status || "unknown").toLowerCase();
                          const status =
                            stepStatusConfig[statusKey] ?? {
                              label: step.status ? String(step.status) : "Unknown",
                              variant: "outline" as const,
                            };
                          const stepTypeKey = String(step.stepType || "unknown").toLowerCase();
                          const stepType =
                            stepTypeConfig[stepTypeKey] ?? {
                              label: step.stepType ? String(step.stepType) : "Unknown",
                              variant: "outline" as const,
                              icon: <PlayIcon className="size-3" />,
                            };
                          const markdownOutput = step.output
                            ? `\`\`\`\n${step.output}\n\`\`\``
                            : "";
                          const anchorId = step.id
                            ? `step-result-${step.id}`
                            : `step-result-${idx}`;
                          return (
                            <div
                              key={step.id || `${step.stepName}-${idx}`}
                              id={anchorId}
                              className="rounded-md border bg-background/60 p-3"
                            >
                              <div className="flex flex-wrap items-start justify-between gap-2">
                                <div className="flex flex-wrap items-center gap-2">
                                  <div className="font-medium">
                                    {step.stepName || "Unnamed step"}
                                  </div>
                                  <span className="text-muted-foreground">-</span>
                                  <Badge variant={stepType.variant} className="w-fit gap-1 font-normal">
                                    {stepType.icon}
                                    <span>{stepType.label}</span>
                                  </Badge>
                                </div>
                                <div className="flex flex-wrap items-center gap-2">
                                  <Badge variant={status.variant}>{status.label}</Badge>
                                  <span className="text-xs text-muted-foreground">
                                    {formatStepDuration(step.durationMs)}
                                  </span>
                                </div>
                              </div>
                              <div className="mt-2 grid gap-2 text-xs text-muted-foreground sm:grid-cols-2">
                                <div>
                                  Started:{" "}
                                  {step.startedAt ? formatDateTime(step.startedAt) : "-"}
                                </div>
                                <div>
                                  Completed:{" "}
                                  {step.completedAt ? formatDateTime(step.completedAt) : "-"}
                                </div>
                              </div>
                              {step.command ? (
                                <div className="mt-3 space-y-1">
                                  <div className="text-xs text-muted-foreground">Command</div>
                                  <div className="max-h-40 overflow-auto rounded-md border bg-muted/30 p-2">
                                    <CodeHighlighter
                                      language="bash"
                                      style={resolvedTheme === "dark" ? atomOneDark : github}
                                      customStyle={{
                                        margin: 0,
                                        background: "transparent",
                                        fontSize: "0.75rem",
                                        whiteSpace: "pre-wrap",
                                        wordBreak: "break-word",
                                      }}
                                      codeTagProps={{
                                        style: {
                                          whiteSpace: "pre-wrap",
                                          wordBreak: "break-word",
                                        },
                                      }}
                                    >
                                      {step.command}
                                    </CodeHighlighter>
                                  </div>
                                </div>
                              ) : null}
                              {step.output ? (
                                <div className="mt-3 space-y-1">
                                  <div className="text-xs text-muted-foreground">Output</div>
                                  <pre className="max-h-60 overflow-auto rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                                    {step.output}
                                  </pre>
                                </div>
                              ) : null}
                              {step.output ? (
                                <div className="mt-3 space-y-1">
                                  <div className="text-xs text-muted-foreground">Output (Markdown)</div>
                                  <textarea
                                    readOnly
                                    value={markdownOutput}
                                    rows={4}
                                    className="w-full rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words"
                                  />
                                </div>
                              ) : null}
                              {step.error ? (
                                <div className="mt-3 space-y-1">
                                  <div className="text-xs text-muted-foreground">Error</div>
                                  <div className="rounded-md border border-destructive/30 bg-destructive-soft p-2 font-mono text-xs text-destructive whitespace-pre-wrap break-words">
                                    {step.error}
                                  </div>
                                </div>
                              ) : null}
                            </div>
                          );
                        })}
                      </div>
                    )}
                    </>
                    )}
                  </div>
                </>
              ) : (
                <div className="text-sm text-muted-foreground">
                  Select a scan from the table to see its details here.
                </div>
              )}
            </div>
          </ScrollArea>
        </CardContent>
      </Card>
    </div>
  );
}
