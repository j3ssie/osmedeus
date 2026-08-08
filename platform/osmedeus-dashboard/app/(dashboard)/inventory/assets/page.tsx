"use client";

import * as React from "react";
import { useSearchParams } from "next/navigation";
import { fetchWorkspaces, fetchHttpAssets, fetchAssetStats } from "@/lib/api/assets";
import type { Workspace } from "@/lib/types/asset";
import type { PaginatedResponse } from "@/lib/types/api";
import type {
  HttpAsset,
  HttpAssetFilters,
  AssetSortState,
  AssetSortField,
  AssetStats,
} from "@/lib/types/asset";
import { Card, CardContent } from "@/components/ui/card";
import { SectionCardHeader } from "@/components/shared/section-card-header";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Checkbox } from "@/components/ui/checkbox";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  ChevronsUpDownIcon,
  FolderOpenIcon,
  RefreshCcwIcon,
  Rows3Icon,
  Rows4Icon,
} from "lucide-react";
import { toast } from "sonner";
import { AssetFilters } from "@/components/assets/asset-filters";
import {
  HttpAssetsTable,
  type HttpAssetsColumnKey,
} from "@/components/assets/http-assets-table";
import { AssetDetailDialog } from "@/components/assets/asset-detail-dialog";
import { cn, sortAssets, STATUS_BANDS } from "@/lib/utils";

export default function InventoryAssetsPage() {
  const searchParams = useSearchParams();
  const workspaceParam = (searchParams.get("workspace") ?? "").trim();

  const [workspaces, setWorkspaces] = React.useState<Workspace[]>([]);
  // Workspace ids, or empty for every workspace. `?workspace=` accepts a
  // comma-separated list so a multi-workspace view stays linkable, and a single
  // value still works the way it did when this was a one-of-many select.
  const [selectedWorkspaces, setSelectedWorkspaces] = React.useState<string[]>(
    () =>
      workspaceParam
        ? workspaceParam
            .split(",")
            .map((value) => value.trim())
            .filter(Boolean)
        : []
  );
  const [workspaceSearch, setWorkspaceSearch] = React.useState("");
  const [assetsResponse, setAssetsResponse] =
    React.useState<PaginatedResponse<HttpAsset> | null>(null);
  const [isLoading, setIsLoading] = React.useState(false);
  const [filters, setFilters] = React.useState<HttpAssetFilters>({});
  const [pageSize, setPageSize] = React.useState(50);
  const [assetStats, setAssetStats] = React.useState<AssetStats | null>(null);
  const [page, setPage] = React.useState(1);
  const [selectedAsset, setSelectedAsset] = React.useState<HttpAsset | null>(
    null
  );
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [sortState, setSortState] = React.useState<AssetSortState>({
    field: null,
    direction: "asc",
  });
  const [density, setDensity] = React.useState<"comfortable" | "compact">(
    "comfortable"
  );
  const [visibleColumns, setVisibleColumns] = React.useState<
    Record<HttpAssetsColumnKey, boolean>
  >(() => ({
    id: false,
    workspace: false,
    assetValue: true,
    url: false,
    input: false,
    scheme: false,
    method: false,
    path: false,
    statusCode: true,
    contentType: false,
    contentLength: true,
    title: true,
    words: false,
    lines: false,
    hostIp: true,
    dnsRecords: false,
    tls: false,
    assetType: true,
    remarks: true,
    technologies: true,
    responseTime: false,
    source: false,
    createdAt: false,
    updatedAt: false,
    lastSeenAt: false,
  }));
  const lastFetchRef = React.useRef<number>(0);
  const forceNextRef = React.useRef<boolean>(false);
  const COOLDOWN_MS = 20000;

  React.useEffect(() => {
    const loadWs = async () => {
      try {
        const ws = await fetchWorkspaces();
        setWorkspaces(ws);
      } catch (e) {
        toast.error("Failed to load workspaces", {
          description: e instanceof Error ? e.message : "",
        });
      }
    };
    loadWs();
  }, []);

  // A `?workspace=` value can arrive as a name; once the list is in, swap any
  // such entry for its id so the checkboxes below match on identity.
  React.useEffect(() => {
    if (!workspaces.length || !selectedWorkspaces.length) return;
    let changed = false;
    const resolved = selectedWorkspaces.map((entry) => {
      if (workspaces.some((w) => String(w.id) === entry)) return entry;
      const byName =
        workspaces.find((w) => w.name === entry) ||
        workspaces.find((w) => w.name.toLowerCase() === entry.toLowerCase());
      if (!byName) return entry;
      changed = true;
      return String(byName.id);
    });
    if (changed) setSelectedWorkspaces(Array.from(new Set(resolved)));
  }, [workspaces, selectedWorkspaces]);

  // Names are what the API filters on; an empty list means every workspace.
  const effectiveWorkspaces = React.useMemo(
    () =>
      selectedWorkspaces.map(
        (id) => workspaces.find((w) => String(w.id) === id)?.name ?? id
      ),
    [selectedWorkspaces, workspaces]
  );
  const workspaceFilterKey = effectiveWorkspaces.join(",");

  const selectedWorkspaceNames = React.useMemo(
    () =>
      selectedWorkspaces
        .map((id) => workspaces.find((w) => String(w.id) === id)?.name)
        .filter((name): name is string => Boolean(name)),
    [selectedWorkspaces, workspaces]
  );

  const toggleWorkspace = React.useCallback((id: string) => {
    setSelectedWorkspaces((prev) =>
      prev.includes(id) ? prev.filter((value) => value !== id) : [...prev, id]
    );
    setPage(1);
    forceNextRef.current = true;
  }, []);

  // Rows from several workspaces are indistinguishable without the column that
  // names them, so it comes along on the way into a multi-workspace view —
  // whether that came from the checkboxes or straight from `?workspace=`. Only
  // on the crossing, and never forced back off: hiding it again is the reader's
  // call, and re-showing it on every render would overrule them.
  const multiWorkspaceRef = React.useRef(false);
  React.useEffect(() => {
    const multi = selectedWorkspaces.length > 1;
    if (multi && !multiWorkspaceRef.current) {
      setVisibleColumns((columns) =>
        columns.workspace ? columns : { ...columns, workspace: true }
      );
    }
    multiWorkspaceRef.current = multi;
  }, [selectedWorkspaces]);

  const filteredWorkspaces = React.useMemo(() => {
    const query = workspaceSearch.trim().toLowerCase();
    if (!query) return workspaces;
    return workspaces.filter((ws) => ws.name.toLowerCase().includes(query));
  }, [workspaceSearch, workspaces]);

  React.useEffect(() => {
    const loadStats = async () => {
      try {
        const stats = await fetchAssetStats(
          workspaceFilterKey ? workspaceFilterKey.split(",") : undefined
        );
        setAssetStats(stats);
      } catch (e) {
        toast.error("Failed to load asset statistics", {
          description: e instanceof Error ? e.message : "",
        });
      }
    };
    loadStats();
  }, [workspaceFilterKey]);

  const loadAssets = React.useCallback(
    async (force?: boolean) => {
      const now = Date.now();
      if (!force && now - lastFetchRef.current < COOLDOWN_MS) return;
      try {
        setIsLoading(true);
        const res = await fetchHttpAssets(
          workspaceFilterKey ? workspaceFilterKey.split(",") : undefined,
          {
            page,
            pageSize,
            filters,
          }
        );
        const tp = res.pagination?.totalPages ?? 0;
        if (tp > 0 && page > tp) {
          setPage(tp);
          forceNextRef.current = true;
          return;
        }
        setAssetsResponse(res);
        lastFetchRef.current = Date.now();
      } catch (e) {
        toast.error("Failed to load assets", {
          description: e instanceof Error ? e.message : "",
        });
      } finally {
        setIsLoading(false);
      }
    },
    [workspaceFilterKey, page, pageSize, filters]
  );

  React.useEffect(() => {
    const doLoad = async () => {
      await loadAssets(forceNextRef.current);
      forceNextRef.current = false;
    };
    doLoad();
  }, [loadAssets]);

  const handleFiltersChange = React.useCallback((next: HttpAssetFilters) => {
    setFilters(next);
    setPage(1);
  }, []);

  const handleSort = React.useCallback((field: AssetSortField) => {
    setSortState((prev) => ({
      field,
      direction:
        prev.field === field && prev.direction === "asc" ? "desc" : "asc",
    }));
  }, []);

  const clientFilteredAssets = React.useMemo(() => {
    const assets = assetsResponse?.data ?? [];
    const q = (filters.search ?? "").trim().toLowerCase();

    return assets.filter((a) => {
      if (q) {
        const hay = [a.url, a.title ?? "", a.assetValue, a.hostIp ?? ""]
          .join(" ")
          .toLowerCase();
        if (!hay.includes(q)) return false;
      }
      if (filters.statusCodes?.length) {
        if (!filters.statusCodes.includes(a.statusCode)) return false;
      }
      if (filters.technologies?.length) {
        const techSet = new Set(
          a.technologies.map((t) => String(t).trim().toLowerCase())
        );
        const wanted = filters.technologies.map((t) => t.trim().toLowerCase());
        if (!wanted.some((t) => techSet.has(t))) return false;
      }
      if (filters.assetTypes?.length) {
        const assetType = String(a.assetType ?? "").trim().toLowerCase();
        const wanted = filters.assetTypes.map((t) => t.trim().toLowerCase());
        if (!wanted.includes(assetType)) return false;
      }
      if (filters.sources?.length) {
        const source = String(a.source ?? "").trim().toLowerCase();
        const wanted = filters.sources.map((t) => t.trim().toLowerCase());
        if (!wanted.includes(source)) return false;
      }
      if (filters.remarks?.length) {
        const remarks = Array.isArray(a.remarks)
          ? a.remarks
          : a.remarks
            ? [a.remarks]
            : [];
        const remarkSet = new Set(
          remarks.map((t) => String(t).trim().toLowerCase()).filter(Boolean)
        );
        const wanted = filters.remarks.map((t) => t.trim().toLowerCase());
        if (!wanted.some((t) => remarkSet.has(t))) return false;
      }
      return true;
    });
  }, [
    assetsResponse?.data,
    filters.search,
    filters.statusCodes,
    filters.technologies,
    filters.assetTypes,
    filters.sources,
    filters.remarks,
  ]);

  // Apply client-side sorting
  const sortedAssets = React.useMemo(() => {
    return sortAssets(clientFilteredAssets, sortState.field, sortState.direction);
  }, [clientFilteredAssets, sortState]);

  const fallbackAssetTypes = React.useMemo(() => {
    const values = (assetsResponse?.data ?? [])
      .map((asset) => asset.assetType)
      .filter((value) => value && value.trim().length > 0);
    return Array.from(new Set(values)).sort((a, b) => a.localeCompare(b));
  }, [assetsResponse?.data]);

  const fallbackSources = React.useMemo(() => {
    const values = (assetsResponse?.data ?? [])
      .map((asset) => asset.source)
      .filter((value) => value && value.trim().length > 0);
    return Array.from(new Set(values)).sort((a, b) => a.localeCompare(b));
  }, [assetsResponse?.data]);

  const fallbackRemarks = React.useMemo(() => {
    const values = (assetsResponse?.data ?? [])
      .flatMap((asset) =>
        Array.isArray(asset.remarks)
          ? asset.remarks
          : asset.remarks
            ? [asset.remarks]
            : []
      )
      .map((value) => String(value).trim())
      .filter((value) => value.length > 0);
    return Array.from(new Set(values)).sort((a, b) => a.localeCompare(b));
  }, [assetsResponse?.data]);

  const assetTypeOptions =
    assetStats?.assetTypes?.length ? assetStats.assetTypes : fallbackAssetTypes;
  const sourceOptions =
    assetStats?.sources?.length ? assetStats.sources : fallbackSources;
  const technologyOptions =
    assetStats?.technologies?.length ? assetStats.technologies : [];
  const remarkOptions =
    assetStats?.remarks?.length ? assetStats.remarks : fallbackRemarks;

  const columnOptions = React.useMemo(
    () => [
      { key: "id", label: "ID" },
      { key: "workspace", label: "Workspace" },
      { key: "assetValue", label: "Asset Value" },
      { key: "url", label: "URL" },
      { key: "input", label: "Input" },
      { key: "scheme", label: "Scheme" },
      { key: "method", label: "Method" },
      { key: "path", label: "Path" },
      { key: "statusCode", label: "Status Code" },
      { key: "contentType", label: "Content Type" },
      { key: "contentLength", label: "Content Length" },
      { key: "title", label: "Title" },
      { key: "words", label: "Words" },
      { key: "lines", label: "Lines" },
      { key: "hostIp", label: "Host IP" },
      { key: "dnsRecords", label: "DNS Records" },
      { key: "tls", label: "TLS" },
      { key: "assetType", label: "Asset Type" },
      { key: "technologies", label: "Tech" },
      { key: "responseTime", label: "Time" },
      { key: "remarks", label: "Remarks" },
      { key: "source", label: "Source" },
      { key: "createdAt", label: "Created At" },
      { key: "updatedAt", label: "Updated At" },
      { key: "lastSeenAt", label: "Last Seen At" },
    ],
    []
  );
  const pageSizeOptions = React.useMemo(() => [50, 100, 200, 500], []);

  // Check if any filters are active
  const hasActiveFilters = React.useMemo(() => {
    return !!(
      filters.search ||
      filters.statusCodes?.length ||
      filters.technologies?.length ||
      filters.assetTypes?.length ||
      filters.sources?.length ||
      filters.remarks?.length ||
      filters.contentTypes?.length ||
      filters.tlsVersion ||
      filters.location
    );
  }, [filters]);

  return (
    <div className="space-y-4">
      {/* Main Content Card */}
      <Card className="overflow-hidden">
        <SectionCardHeader
          title="Assets Inventory"
          description={
            assetsResponse?.pagination?.totalItems !== undefined ? (
              <>
                <span className="font-medium text-foreground">
                  {assetsResponse.pagination.totalItems.toLocaleString()}
                </span>{" "}
                assets found
                {selectedWorkspaces.length > 0 && (
                  <>
                    {" "}
                    in{" "}
                    <span className="font-medium text-foreground">
                      {selectedWorkspaces.length <= 2
                        ? (selectedWorkspaceNames.length
                            ? selectedWorkspaceNames
                            : selectedWorkspaces
                          ).join(" and ")
                        : `${selectedWorkspaces.length} workspaces`}
                    </span>
                  </>
                )}
              </>
            ) : (
              "Loading assets..."
            )
          }
          actions={
            <>
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className={cn(
                      "h-9 w-56 justify-between font-normal",
                      selectedWorkspaces.length > 0 && "border-primary"
                    )}
                  >
                    <span className="flex min-w-0 items-center gap-2">
                      <FolderOpenIcon className="size-4 shrink-0" />
                      <span className="truncate">
                        {selectedWorkspaces.length === 0
                          ? "All Workspaces"
                          : selectedWorkspaces.length === 1
                            ? (selectedWorkspaceNames[0] ?? selectedWorkspaces[0])
                            : `${selectedWorkspaces.length} workspaces`}
                      </span>
                    </span>
                    <ChevronsUpDownIcon className="size-4 shrink-0 opacity-50" />
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-64 p-0" align="end">
                  <div className="border-b p-2">
                    <Input
                      placeholder="Search workspaces..."
                      value={workspaceSearch}
                      onChange={(e) => setWorkspaceSearch(e.target.value)}
                      className="h-8"
                    />
                  </div>
                  <ScrollArea className="h-[240px]">
                    <div className="space-y-1 p-2">
                      {filteredWorkspaces.map((ws) => (
                        <label
                          key={ws.id}
                          className="flex cursor-pointer items-center gap-2 rounded-sm px-2 py-1.5 text-sm hover:bg-muted"
                        >
                          <Checkbox
                            checked={selectedWorkspaces.includes(String(ws.id))}
                            onCheckedChange={() => toggleWorkspace(String(ws.id))}
                          />
                          <span className="truncate">{ws.name}</span>
                        </label>
                      ))}
                      {filteredWorkspaces.length === 0 && (
                        <p className="py-4 text-center text-sm text-muted-foreground">
                          No workspaces found
                        </p>
                      )}
                    </div>
                  </ScrollArea>
                  {selectedWorkspaces.length > 0 && (
                    <div className="border-t p-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 w-full"
                        onClick={() => {
                          setSelectedWorkspaces([]);
                          setPage(1);
                          forceNextRef.current = true;
                        }}
                      >
                        All workspaces
                      </Button>
                    </div>
                  )}
                </PopoverContent>
              </Popover>
              <Button
                variant="outline"
                size="sm"
                onClick={() => loadAssets(true)}
                disabled={isLoading}
              >
                <RefreshCcwIcon
                  className={`size-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
                />
                Refresh
              </Button>
              <Button
                variant="outline"
                size="icon-sm"
                onClick={() =>
                  setDensity((d) =>
                    d === "compact" ? "comfortable" : "compact"
                  )
                }
                aria-label={
                  density === "compact"
                    ? "Switch to comfortable rows"
                    : "Switch to compact rows"
                }
                title={
                  density === "compact" ? "Comfortable rows" : "Compact rows"
                }
              >
                {density === "compact" ? (
                  <Rows4Icon className="size-4" />
                ) : (
                  <Rows3Icon className="size-4" />
                )}
              </Button>
              {sortedAssets.length > 0 && (
                <div className="hidden lg:flex items-center gap-2">
                  {(() => {
                    const counts = sortedAssets.reduce((acc, a) => {
                      const band = STATUS_BANDS.find(
                        (b) => a.statusCode >= b.min && a.statusCode < b.max
                      );
                      if (band) acc[band.label] = (acc[band.label] ?? 0) + 1;
                      return acc;
                    }, {} as Record<string, number>);
                    return STATUS_BANDS.map((b) => (
                      <Badge key={b.label} variant={b.variant} className="gap-1">
                        {b.label}: {counts[b.label] ?? 0}
                      </Badge>
                    ));
                  })()}
                </div>
              )}
            </>
          }
        />

        {/* Filters section */}
        <div className="border-b px-4 py-2.5 bg-muted/10">
          <AssetFilters
            filters={filters}
            onFiltersChange={handleFiltersChange}
            technologyOptions={technologyOptions}
            assetTypeOptions={assetTypeOptions}
            sourceOptions={sourceOptions}
            remarkOptions={remarkOptions}
            columnOptions={columnOptions}
            visibleColumns={visibleColumns}
            onVisibleColumnsChange={setVisibleColumns}
            pageSize={pageSize}
            pageSizeOptions={pageSizeOptions}
            onPageSizeChange={(size) => {
              setPageSize(size);
              setPage(1);
              forceNextRef.current = true;
            }}
          />
        </div>

        {/* Table section */}
        <CardContent className="p-0">
          <HttpAssetsTable
            assets={sortedAssets}
            isLoading={isLoading}
            pagination={assetsResponse?.pagination}
            sortState={sortState}
            onSort={handleSort}
            onPageChange={(p) => {
              setPage(p);
            }}
            onSelect={(asset) => {
              setSelectedAsset(asset);
              setDialogOpen(true);
            }}
            hasActiveFilters={hasActiveFilters}
            visibleColumns={visibleColumns}
            density={density}
          />
        </CardContent>
      </Card>

      <AssetDetailDialog
        asset={selectedAsset}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
      />
    </div>
  );
}
