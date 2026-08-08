"use client";

import * as React from "react";
import {
  fetchVulnerabilities,
  fetchVulnerabilitySummary,
} from "@/lib/api/vulnerabilities";
import type { Vulnerability, VulnerabilitySummary, VulnerabilitySeverity } from "@/lib/types/vulnerability";
import { Card, CardContent } from "@/components/ui/card";
import { SectionCardHeader } from "@/components/shared/section-card-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  ShieldAlertIcon,
  AlertTriangleIcon,
  BarChart3Icon,
  GlobeIcon,
  BadgeCheckIcon,
  Columns3Icon,
  ListIcon,
  SearchIcon,
  RotateCcwIcon,
  ChevronsUpDownIcon,
} from "lucide-react";
import { fetchWorkspaces } from "@/lib/api/assets";
import type { Workspace } from "@/lib/types/asset";
import { VulnerabilityDetailDialog } from "@/components/vulnerabilities/vulnerability-detail-dialog";
import {
  CellShell,
  DataGrid,
  GridPagination,
  type GridColDef,
} from "@/components/ui/data-grid";
import {
  severityConfig,
  severityOrder,
  getTagColor,
  SeverityBadge,
  ConfidenceBadge,
} from "@/components/vulnerabilities/vulnerability-display";
import { cn } from "@/lib/utils";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectTrigger,
  SelectContent,
  SelectItem,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { toast } from "sonner";
import type { PaginatedResponse } from "@/lib/types/api";
import type { VulnerabilityConfidence } from "@/lib/types/vulnerability";

const confidenceOptions: VulnerabilityConfidence[] = [
  "Certain",
  "Firm",
  "Tentative",
  "Manual Review Required",
];

/**
 * A findings table is read host-first: the origin says which target is
 * affected, the path is detail. Splitting them lets the host stay at full
 * strength while the path drops to the faint tier and absorbs the truncation.
 */
function AssetCell({ value }: { value?: string }) {
  const raw = (value ?? "").trim();
  if (!raw) return <span className="text-muted-foreground">-</span>;

  let host = raw;
  let rest = "";
  if (/^https?:\/\//i.test(raw)) {
    try {
      const url = new URL(raw);
      host = url.host;
      rest = `${url.pathname === "/" ? "" : url.pathname}${url.search}${url.hash}`;
    } catch {
      // Not parseable as a URL — fall back to showing it verbatim.
    }
  }

  return (
    <span className="block min-w-0 truncate font-mono text-xs" title={raw}>
      <span className="text-body">{host}</span>
      {rest && <span className="text-faint">{rest}</span>}
    </span>
  );
}

export default function VulnerabilitiesPage() {
  const [vulnerabilities, setVulnerabilities] = React.useState<Vulnerability[]>([]);
  const [summary, setSummary] = React.useState<VulnerabilitySummary | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [summaryLoading, setSummaryLoading] = React.useState(true);
  const [workspaces, setWorkspaces] = React.useState<Workspace[]>([]);
  const [workspacesLoading, setWorkspacesLoading] = React.useState(true);
  const [selected, setSelected] = React.useState<Vulnerability | null>(null);
  const [open, setOpen] = React.useState(false);
  const [page, setPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(20);
  const [pagination, setPagination] = React.useState<
    PaginatedResponse<Vulnerability>["pagination"] | null
  >(null);
  const [filters, setFilters] = React.useState<{
    workspace?: string;
    severity?: VulnerabilitySeverity[];
    confidence?: VulnerabilityConfidence[];
  }>({});
  const [searchValue, setSearchValue] = React.useState("");

  const [visibleColumns, setVisibleColumns] = React.useState<Record<string, boolean>>({
    severity: true,
    confidence: true,
    title: true,
    asset: true,
    tags: true,
  });

  const columnOptions = React.useMemo(
    () => [
      { key: "severity", label: "Severity" },
      { key: "confidence", label: "Confidence" },
      { key: "title", label: "Title" },
      { key: "asset", label: "Asset" },
      { key: "tags", label: "Tags" },
    ],
    []
  );

  const visibleColumnCount = React.useMemo(
    () => Object.values(visibleColumns).filter(Boolean).length,
    [visibleColumns]
  );

  const handleColumnToggle = React.useCallback((key: string) => {
    setVisibleColumns((prev) => {
      const next = { ...prev, [key]: !prev[key] };
      if (Object.values(next).some(Boolean)) return next;
      return prev;
    });
  }, []);

  const visibleVulnerabilities = React.useMemo(() => {
    const q = searchValue.trim().toLowerCase();
    return vulnerabilities.filter((v) => {
      if (filters.severity?.length) {
        if (!filters.severity.includes(v.severity)) return false;
      }
      if (filters.confidence?.length) {
        if (!v.confidence || !filters.confidence.includes(v.confidence)) return false;
      }
      if (!q) return true;

      const haystack = [
        v.workspace,
        v.vulnTitle,
        v.vulnInfo,
        v.vulnDesc,
        v.vulnPoc,
        v.severity,
        v.confidence,
        v.assetType,
        v.assetValue,
        ...(v.tags ?? []),
      ]
        .filter((x): x is string => typeof x === "string")
        .join("\n")
        .toLowerCase();

      return haystack.includes(q);
    });
  }, [filters.confidence, filters.severity, vulnerabilities, searchValue]);


  const loadSummary = React.useCallback(async () => {
    try {
      setSummaryLoading(true);
      const res = await fetchVulnerabilitySummary(filters.workspace?.trim() || undefined);
      setSummary(res);
    } catch (e) {
      toast.error("Failed to load summary", {
        description: e instanceof Error ? e.message : "",
      });
    } finally {
      setSummaryLoading(false);
    }
  }, [filters.workspace]);

  React.useEffect(() => {
    let cancelled = false;
    const loadWorkspaces = async () => {
      try {
        setWorkspacesLoading(true);
        const ws = await fetchWorkspaces({ offset: 0, limit: 1000 });
        if (!cancelled) setWorkspaces(ws);
      } catch (e) {
        toast.error("Failed to load workspaces", {
          description: e instanceof Error ? e.message : "",
        });
      } finally {
        if (!cancelled) setWorkspacesLoading(false);
      }
    };
    loadWorkspaces();
    return () => {
      cancelled = true;
    };
  }, []);

  const loadVulnerabilities = React.useCallback(async () => {
    try {
      setLoading(true);
      const res = await fetchVulnerabilities({
        page,
        pageSize,
        filters: {
          workspace: filters.workspace?.trim() || undefined,
          severity: filters.severity?.length ? filters.severity : undefined,
          confidence: filters.confidence?.length ? filters.confidence : undefined,
        },
      });
      setVulnerabilities(res.data);
      setPagination(res.pagination);
    } catch (e) {
      toast.error("Failed to load vulnerabilities", {
        description: e instanceof Error ? e.message : "",
      });
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, filters]);

  const handleSeverityToggle = React.useCallback((severity: VulnerabilitySeverity) => {
    setFilters((f) => {
      const current = f.severity ?? [];
      const updated = current.includes(severity)
        ? current.filter((s) => s !== severity)
        : [...current, severity];
      const normalized = updated
        .filter((s, i) => updated.indexOf(s) === i)
        .sort((a, b) => severityConfig[a].rank - severityConfig[b].rank);
      return { ...f, severity: normalized.length ? normalized : undefined };
    });
    setPage(1);
  }, []);

  const handleConfidenceToggle = React.useCallback((confidence: VulnerabilityConfidence) => {
    setFilters((f) => {
      const current = f.confidence ?? [];
      const updated = current.includes(confidence)
        ? current.filter((c) => c !== confidence)
        : [...current, confidence];
      const normalized = updated
        .filter((c, i) => updated.indexOf(c) === i)
        .sort((a, b) => a.localeCompare(b));
      return { ...f, confidence: normalized.length ? normalized : undefined };
    });
    setPage(1);
  }, []);

  React.useEffect(() => {
    loadSummary();
  }, [loadSummary]);

  React.useEffect(() => {
    loadVulnerabilities();
  }, [loadVulnerabilities]);

  const openDetail = React.useCallback((v: Vulnerability) => {
    setSelected(v);
    setOpen(true);
  }, []);

  const columns = React.useMemo<GridColDef<Vulnerability>[]>(() => {
    const all: Array<GridColDef<Vulnerability> & { key: string }> = [
      {
        key: "severity",
        colId: "severity",
        headerName: "Severity",
        minWidth: 104,
        flex: 0,
        width: 104,
        // Severity is an ordered scale, so it sorts by rank rather than
        // alphabetically — "critical" before "high" before "medium".
        valueGetter: (p) => severityConfig[p.data!.severity]?.rank ?? 99,
        cellRenderer: (p: { data: Vulnerability }) => (
          <SeverityBadge severity={p.data.severity} />
        ),
      },
      {
        key: "confidence",
        field: "confidence",
        headerName: "Confidence",
        minWidth: 110,
        flex: 0,
        width: 128,
        cellRenderer: (p: { value?: VulnerabilityConfidence }) => (
          <ConfidenceBadge confidence={p.value} />
        ),
      },
      {
        key: "title",
        field: "vulnTitle",
        colId: "title",
        headerName: "Title",
        // Fixed rather than flexible: rule names are short and repeat down the
        // column, so spare width here is dead space. The asset is the value
        // that actually varies, so it gets the slack instead.
        minWidth: 200,
        flex: 0,
        width: 300,
        cellRenderer: (p: { data: Vulnerability }) => (
          <div className="flex min-w-0 flex-col justify-center gap-0.5 leading-tight">
            <span className="truncate font-medium text-ink">
              {p.data.vulnTitle || "-"}
            </span>
            {p.data.vulnInfo && (
              <span className="truncate font-mono text-2xs text-faint">
                {p.data.vulnInfo}
              </span>
            )}
          </div>
        ),
      },
      {
        key: "asset",
        field: "assetValue",
        colId: "asset",
        headerName: "Asset",
        minWidth: 260,
        flex: 3,
        cellRenderer: (p: { value?: string }) => <AssetCell value={p.value} />,
      },
      {
        key: "tags",
        colId: "tags",
        headerName: "Tags",
        minWidth: 150,
        maxWidth: 320,
        flex: 1,
        valueGetter: (p) => (p.data?.tags ?? []).join(","),
        cellRenderer: (p: { data: Vulnerability }) => {
          const tags = p.data.tags ?? [];
          if (tags.length === 0) return <span className="text-muted-foreground">-</span>;
          return (
            <CellShell className="gap-1">
              {tags.slice(0, 2).map((tag, i) => (
                <Badge
                  key={i}
                  variant="outline"
                  className={cn("shrink-0 text-2xs", getTagColor(tag))}
                >
                  {tag}
                </Badge>
              ))}
              {tags.length > 2 && (
                <Badge
                  variant="secondary"
                  className="shrink-0 text-2xs"
                  title={tags.slice(2).join(", ")}
                >
                  +{tags.length - 2}
                </Badge>
              )}
            </CellShell>
          );
        },
      },
    ];

    return all
      .filter((c) => visibleColumns[c.key])
      .map(({ key: _key, ...col }) => col);
  }, [visibleColumns]);

  return (
    <div className="space-y-4">
      {/* Severity summary. Each tile is also the filter for its own severity —
          the counts are what you scan first, so they are what you click. */}
      <div className="rounded-card border border-border bg-card p-3">
        <div className="flex flex-wrap items-center gap-2">
          <div className="flex items-center gap-2 pr-1">
            <BarChart3Icon className="size-4 text-muted-foreground" />
            <span className="text-sm font-medium">Statistics</span>
            <span className="text-sm text-muted-foreground">
              {summaryLoading ? "-" : (summary?.total ?? 0).toLocaleString()} findings
            </span>
          </div>
          <div className="mr-1 h-6 w-px bg-border" />
          {severityOrder.map((sev) => {
            const config = severityConfig[sev];
            const Icon = config.icon;
            const count = summary?.bySeverity[sev] ?? 0;
            const active = filters.severity?.includes(sev) ?? false;
            return (
              <button
                key={sev}
                type="button"
                onClick={() => handleSeverityToggle(sev)}
                aria-pressed={active}
                title={`Filter by ${config.label.toLowerCase()} severity`}
                className={cn(
                  "inline-flex cursor-pointer items-center gap-2 rounded-control border px-2.5 py-1.5 transition-[box-shadow,border-color] hover:shadow-glow",
                  config.soft,
                  active ? "border-primary" : "border-transparent"
                )}
              >
                <Icon className="size-3.5 shrink-0 opacity-80" />
                <span className="text-2xs font-medium uppercase tracking-label">
                  {config.label}
                </span>
                <span className="text-md font-semibold tabular-nums">
                  {summaryLoading ? "-" : count.toLocaleString()}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Filters and Table */}
      <Card className="overflow-hidden">
        <SectionCardHeader
          icon={ShieldAlertIcon}
          title="Vulnerabilities"
          description="Filter by workspace, severity, or confidence"
          actions={
            <Button
              variant="outline"
              className="shrink-0"
              onClick={() => {
                setFilters({});
                setSearchValue("");
                setPage(1);
                setPageSize(20);
              }}
            >
              <RotateCcwIcon className="mr-2 size-4" />
              Reset
            </Button>
          }
        />
        <CardContent>
          <div className="flex flex-wrap gap-3 py-2">
            <div className="relative flex-1 min-w-[240px]">
              <SearchIcon className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search title, info, asset, or tags..."
                value={searchValue}
                onChange={(e) => setSearchValue(e.target.value)}
                className="pl-9 h-9"
              />
            </div>
            <Select
              value={filters.workspace || "all"}
              onValueChange={(val) => {
                setFilters((f) => ({
                  ...f,
                  workspace: val === "all" ? undefined : val,
                }));
                setPage(1);
              }}
              disabled={workspacesLoading}
            >
              <SelectTrigger className="max-w-[220px]">
                <span className="flex items-center gap-2">
                  <GlobeIcon className="size-4 text-muted-foreground" />
                  <SelectValue placeholder="Workspace" />
                </span>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Workspaces</SelectItem>
                {workspaces
                  .filter((w) => !!w.name)
                  .map((w) => (
                    <SelectItem key={`${w.id}-${w.name}`} value={w.name}>
                      {w.name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
            <Popover>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  className={(filters.severity?.length ?? 0) > 0 ? "border-primary" : undefined}
                >
                  <span className="flex items-center gap-2">
                    <AlertTriangleIcon className="size-4 text-muted-foreground" />
                    <span>Severity</span>
                    {(filters.severity?.length ?? 0) > 0 && (
                      <Badge variant="secondary" className="px-1.5 py-0 text-xs">
                        {filters.severity?.length}
                      </Badge>
                    )}
                  </span>
                  <ChevronsUpDownIcon className="ml-2 size-4 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[200px] p-2" align="start">
                <div className="space-y-1">
                  {severityOrder.map((sev) => (
                    <label
                      key={sev}
                      className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-muted cursor-pointer text-sm"
                    >
                      <Checkbox
                        checked={filters.severity?.includes(sev) ?? false}
                        onCheckedChange={() => handleSeverityToggle(sev)}
                      />
                      <span>{severityConfig[sev].label}</span>
                    </label>
                  ))}
                </div>
                {(filters.severity?.length ?? 0) > 0 && (
                  <div className="pt-2 mt-2 border-t">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full h-8"
                      onClick={() => {
                        setFilters((f) => ({ ...f, severity: undefined }));
                        setPage(1);
                      }}
                    >
                      Clear selection
                    </Button>
                  </div>
                )}
              </PopoverContent>
            </Popover>
            <Popover>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  className={(filters.confidence?.length ?? 0) > 0 ? "border-primary" : undefined}
                >
                  <span className="flex items-center gap-2">
                    <BadgeCheckIcon className="size-4 text-muted-foreground" />
                    <span>Confidence</span>
                    {(filters.confidence?.length ?? 0) > 0 && (
                      <Badge variant="secondary" className="px-1.5 py-0 text-xs">
                        {filters.confidence?.length}
                      </Badge>
                    )}
                  </span>
                  <ChevronsUpDownIcon className="ml-2 size-4 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[240px] p-2" align="start">
                <div className="space-y-1">
                  {confidenceOptions.map((c) => (
                    <label
                      key={c}
                      className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-muted cursor-pointer text-sm"
                    >
                      <Checkbox
                        checked={filters.confidence?.includes(c) ?? false}
                        onCheckedChange={() => handleConfidenceToggle(c)}
                      />
                      <span>{c}</span>
                    </label>
                  ))}
                </div>
                {(filters.confidence?.length ?? 0) > 0 && (
                  <div className="pt-2 mt-2 border-t">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="w-full h-8"
                      onClick={() => {
                        setFilters((f) => ({ ...f, confidence: undefined }));
                        setPage(1);
                      }}
                    >
                      Clear selection
                    </Button>
                  </div>
                )}
              </PopoverContent>
            </Popover>
            <Popover>
              <PopoverTrigger asChild>
                <Button variant="outline">
                  <span className="flex items-center gap-2">
                    <Columns3Icon className="size-4 text-muted-foreground" />
                    <span>Columns</span>
                    {visibleColumnCount > 0 &&
                      visibleColumnCount !== columnOptions.length && (
                        <Badge variant="secondary" className="px-1.5 py-0 text-xs">
                          {visibleColumnCount}
                        </Badge>
                      )}
                  </span>
                  <ChevronsUpDownIcon className="ml-2 size-4 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[220px] p-2" align="start">
                <div className="space-y-1">
                  {columnOptions.map((option) => {
                    const checked = visibleColumns[option.key] ?? false;
                    const disableToggle = checked && visibleColumnCount <= 1;
                    return (
                      <label
                        key={option.key}
                        className="flex items-center gap-2 px-2 py-1.5 rounded-sm hover:bg-muted cursor-pointer text-sm"
                      >
                        <Checkbox
                          checked={checked}
                          disabled={disableToggle}
                          onCheckedChange={() => handleColumnToggle(option.key)}
                        />
                        <span>{option.label}</span>
                      </label>
                    );
                  })}
                </div>
              </PopoverContent>
            </Popover>
            <Select
              value={String(pageSize)}
              onValueChange={(val) => {
                const n = parseInt(val, 10);
                setPageSize(Number.isNaN(n) ? 20 : n);
                setPage(1);
              }}
            >
              <SelectTrigger className="max-w-[140px]">
                <span className="flex items-center gap-2">
                  <ListIcon className="size-4 text-muted-foreground" />
                  <SelectValue placeholder="Page Size" />
                </span>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="20">20</SelectItem>
                <SelectItem value="50">50</SelectItem>
                <SelectItem value="100">100</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <DataGrid<Vulnerability>
            rows={visibleVulnerabilities}
            columns={columns}
            getRowId={(v) => String(v.id)}
            loading={loading}
            onRowClick={openDetail}
            // Two lines of title need the extra room; 36px crushed them.
            rowHeight={44}
            // Enough for a full page of 20 at that height. Anything shorter
            // puts a second scrollbar inside a page that is already scrolling.
            maxAutoHeight={940}
            getRowClass={(p) =>
              p.data ? `og-sev-${p.data.severity}` : undefined
            }
            emptyState={
              <div className="py-10 text-center text-sm text-muted-foreground">
                No vulnerabilities found
              </div>
            }
          />

          {pagination && (
            <GridPagination
              pagination={pagination}
              onPageChange={setPage}
            />
          )}
        </CardContent>
      </Card>

      {/* Detail Drawer */}
      <VulnerabilityDetailDialog
        vulnerability={selected}
        open={open}
        onOpenChange={setOpen}
      />
    </div>
  );
}
