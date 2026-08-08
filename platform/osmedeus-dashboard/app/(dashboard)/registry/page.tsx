"use client";

import * as React from "react";
import { Card, CardContent } from "@/components/ui/card";
import { SectionCardHeader } from "@/components/shared/section-card-header";
import { Table, TableHeader, TableHead, TableBody, TableRow, TableCell } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "sonner";
import {
  LoaderIcon,
  PackageIcon,
  CheckCircle2Icon,
  XCircleIcon,
  DownloadIcon,
  SearchIcon,
  RefreshCwIcon,
  GitBranchIcon,
  FolderArchiveIcon,
  PlusCircleIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  FilterIcon,
  TagsIcon,
  ListIcon,
} from "lucide-react";
import { getRegistryMetadata, installBinaries, installWorkflow } from "@/lib/api/registry";
import type {
  DirectFetchBinaryMeta,
  NixTool,
  RegistryMetadata,
  RegistryMode,
} from "@/lib/types/registry";
import { CardSkeleton, TableSkeleton } from "@/components/shared/loading-skeleton";
import { SortableTableHead } from "@/components/ui/sortable-table-head";
import { cn } from "@/lib/utils";
import type { SortDirection } from "@/lib/types/asset";

type FilterOption = "all" | "installed" | "not-installed";

const getTagColor = (tag: string): string => {
  const t = tag.toLowerCase();
  if (["vuln", "scanner"].includes(t)) {
    return "bg-destructive-soft text-destructive border-transparent";
  }
  if (["recon", "subdomain", "discovery"].includes(t)) {
    return "bg-info-soft text-info border-transparent";
  }
  if (["http", "web", "url"].includes(t)) {
    return "bg-purple-soft text-purple border-transparent";
  }
  if (["network", "port", "dns"].includes(t)) {
    return "bg-cyan-soft text-cyan border-transparent";
  }
  return "bg-secondary text-secondary-foreground border-transparent";
};

type RegistryRow = {
  name: string;
  category?: string;
  meta: DirectFetchBinaryMeta | NixTool;
};

export default function RegistryPage() {
  const [loading, setLoading] = React.useState(true);
  const [metadata, setMetadata] = React.useState<RegistryMetadata | null>(null);
  const [registryMode] = React.useState<RegistryMode>("direct-fetch");
  const [search, setSearch] = React.useState("");
  const [filterStatus, setFilterStatus] = React.useState<FilterOption>("all");
  const [tagFilter, setTagFilter] = React.useState<string>("all");
  const [includeOptional, setIncludeOptional] = React.useState(false);
  const [selected, setSelected] = React.useState<Record<string, boolean>>({});
  const [perPage, setPerPage] = React.useState(15);
  const [page, setPage] = React.useState(1);

  type RegistrySortField =
    | "name"
    | "version"
    | "tags"
    | "description"
    | "optional"
    | "status";

  const [sortState, setSortState] = React.useState<{
    field: RegistrySortField;
    direction: SortDirection;
  }>({ field: "name", direction: "asc" });
  const [installingBatch, setInstallingBatch] = React.useState(false);
  const [workflowSource, setWorkflowSource] = React.useState("");
  const [installingWorkflow, setInstallingWorkflow] = React.useState(false);

  const load = React.useCallback(async () => {
    try {
      setLoading(true);
      const data = await getRegistryMetadata({ registry_mode: registryMode });
      setMetadata(data);
    } catch (e) {
      toast.error("Failed to load registry metadata", {
        description: e instanceof Error ? e.message : "",
      });
    } finally {
      setLoading(false);
    }
  }, [registryMode]);

  React.useEffect(() => {
    load();
  }, [load]);

  const rows = React.useMemo<RegistryRow[]>(() => {
    if (!metadata) return [];
    if (metadata.registry_mode === "nix-build") {
      const out: RegistryRow[] = [];
      for (const cat of metadata.categories || []) {
        for (const tool of cat.tools || []) {
          if (!tool?.name) continue;
          out.push({ name: tool.name, category: cat.name, meta: tool });
        }
      }
      return out.sort((a, b) => a.name.localeCompare(b.name));
    }
    return Object.entries(metadata.binaries || {})
      .map(([name, meta]) => ({ name, meta }))
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [metadata]);

  const tagOptions = React.useMemo(() => {
    const set = new Set<string>();
    for (const row of rows) {
      const tags = (row.meta as any)?.tags;
      if (!Array.isArray(tags)) continue;
      for (const t of tags) {
        if (typeof t === "string" && t.trim()) set.add(t);
      }
    }
    return Array.from(set).sort((a, b) => a.localeCompare(b));
  }, [rows]);

  const filtered = React.useMemo(() => {
    const q = search.trim().toLowerCase();
    return rows.filter((row) => {
      const isOptional = Boolean((row.meta as any)?.optional);
      if (!includeOptional && isOptional) return false;
      const rowTags = Array.isArray((row.meta as any)?.tags)
        ? ((row.meta as any).tags as string[])
        : [];
      if (tagFilter !== "all" && !rowTags.includes(tagFilter)) return false;

      const installed = Boolean((row.meta as any)?.installed);
      if (filterStatus === "installed" && !installed) return false;
      if (filterStatus === "not-installed" && installed) return false;
      if (!q) return true;
      const desc = String((row.meta as any)?.desc || "").toLowerCase();
      const tags = rowTags.join(" ").toLowerCase();
      return (
        row.name.toLowerCase().includes(q) ||
        desc.includes(q) ||
        (row.category || "").toLowerCase().includes(q) ||
        tags.includes(q)
      );
    });
  }, [rows, search, filterStatus, tagFilter, includeOptional]);

  const sorted = React.useMemo(() => {
    const getValue = (
      field: RegistrySortField,
      row: RegistryRow
    ): { missing: boolean; value: string | number } => {
      const meta = row.meta as any;
      switch (field) {
        case "name":
          return { missing: !row.name, value: row.name ?? "" };
        case "version": {
          const v = typeof meta.version === "string" ? meta.version : "";
          return { missing: !v, value: v };
        }
        case "tags": {
          const tags = Array.isArray(meta.tags) ? (meta.tags as string[]) : [];
          const v = tags.join(",");
          return { missing: tags.length === 0, value: v };
        }
        case "description": {
          const v = typeof meta.desc === "string" ? meta.desc : "";
          return { missing: !v, value: v };
        }
        case "optional": {
          const isOptional = typeof meta.optional === "boolean" ? meta.optional : null;
          return { missing: isOptional === null, value: Number(isOptional) };
        }
        case "status":
          return { missing: false, value: Number(Boolean(meta.installed)) };
      }
    };

    const items = [...filtered];
    items.sort((a, b) => {
      const av = getValue(sortState.field, a);
      const bv = getValue(sortState.field, b);

      if (av.missing && bv.missing) return 0;
      if (av.missing) return 1;
      if (bv.missing) return -1;

      let cmp = 0;
      if (typeof av.value === "number" && typeof bv.value === "number") {
        cmp = av.value - bv.value;
      } else {
        cmp = String(av.value).localeCompare(String(bv.value), undefined, {
          numeric: true,
          sensitivity: "base",
        });
      }

      return sortState.direction === "asc" ? cmp : -cmp;
    });

    return items;
  }, [filtered, sortState.direction, sortState.field]);

  const totalPages = React.useMemo(() => {
    return Math.max(1, Math.ceil(sorted.length / perPage));
  }, [perPage, sorted.length]);

  const currentPage = Math.min(Math.max(page, 1), totalPages);
  const startIndex = (currentPage - 1) * perPage;
  const endIndexExclusive = Math.min(startIndex + perPage, filtered.length);

  const pageRows = React.useMemo(() => {
    return sorted.slice(startIndex, endIndexExclusive);
  }, [endIndexExclusive, sorted, startIndex]);

  React.useEffect(() => {
    setPage(1);
  }, [
    search,
    filterStatus,
    tagFilter,
    includeOptional,
    registryMode,
    perPage,
    sortState.direction,
    sortState.field,
  ]);

  const toggleSort = (field: RegistrySortField) => {
    setSortState((prev) => {
      if (prev.field === field) {
        return { field, direction: prev.direction === "asc" ? "desc" : "asc" };
      }
      return { field, direction: "asc" };
    });
  };

  const selectedCount = Object.values(selected).filter(Boolean).length;
  const allSelected = pageRows.length > 0 && pageRows.every((r) => selected[r.name]);

  const toggleSelect = (name: string, value: boolean) => {
    setSelected((prev) => ({ ...prev, [name]: value }));
  };

  const toggleSelectAll = (value: boolean) => {
    setSelected((prev) => {
      const next: Record<string, boolean> = { ...prev };
      pageRows.forEach((r) => {
        next[r.name] = value;
      });
      return next;
    });
  };

  const doInstallSelected = async () => {
    const names = Object.entries(selected)
      .filter(([, v]) => v)
      .map(([k]) => k);
    if (names.length === 0) {
      toast.error("No tools selected");
      return;
    }
    setInstallingBatch(true);
    try {
      const res = await installBinaries({
        names,
        registry_mode: registryMode,
      });
      toast.success(res.message || "Installation complete", {
        description: `Installed: ${res.installed_count}, Failed: ${res.failed_count}`,
      });
      setSelected({});
      await load();
    } catch (e) {
      toast.error("Installation failed", {
        description: e instanceof Error ? e.message : "",
      });
    } finally {
      setInstallingBatch(false);
    }
  };

  const doInstallWorkflow = async () => {
    const src = workflowSource.trim();
    if (!src) {
      toast.error("Please provide a workflow source");
      return;
    }
    setInstallingWorkflow(true);
    try {
      const res = await installWorkflow(src);
      toast.success(res.message || "Workflow installed", { description: res.source });
      setWorkflowSource("");
    } catch (e) {
      toast.error("Workflow install failed", {
        description: e instanceof Error ? e.message : "",
      });
    } finally {
      setInstallingWorkflow(false);
    }
  };

  const stats = React.useMemo(() => {
    const installed = rows.filter((r) => Boolean((r.meta as any)?.installed)).length;
    const optional = rows.filter((r) => Boolean((r.meta as any)?.optional)).length;
    return {
      total: rows.length,
      installed,
      notInstalled: rows.length - installed,
      optional,
    };
  }, [rows]);

  if (loading) {
    return (
      <div className="space-y-4">
        <CardSkeleton />
        <div className="rounded-xl border bg-card p-6">
          <TableSkeleton rows={8} columns={6} />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Binary Registry Card */}
      <Card className="overflow-hidden">
        <SectionCardHeader
          icon={PackageIcon}
          title="Binary Registry"
          description={
            <>
              Installing binaries is better handled through the CLI (
              <span className="font-mono">osmedeus install binary -h</span>), as it
              provides a more flexible way to manage binary installation. This page
              is mainly for visualization purposes.
            </>
          }
          actions={
            <>
              <Button variant="outline" size="sm" onClick={load} disabled={loading}>
                <RefreshCwIcon className="mr-2 size-4" />
                Refresh
              </Button>
              <Button
                type="button"
                size="sm"
                variant={includeOptional ? "secondary" : "outline"}
                onClick={() => setIncludeOptional((prev) => !prev)}
              >
                <PlusCircleIcon className="mr-2 size-4" />
                Include Optional
              </Button>
            </>
          }
        />

        <CardContent className="space-y-4">
          {/* Filters */}
          <div className="flex flex-wrap items-end gap-3">
            <div className="relative flex-1 min-w-[200px] max-w-sm">
              <SearchIcon className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search tools..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>

            <Select
              value={filterStatus}
              onValueChange={(val) => setFilterStatus(val as FilterOption)}
            >
              <SelectTrigger className="w-[160px]">
                <span className="flex items-center gap-2">
                  <FilterIcon className="size-4 text-muted-foreground" />
                  <SelectValue placeholder="Filter" />
                </span>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Tools</SelectItem>
                <SelectItem value="installed">Installed</SelectItem>
                <SelectItem value="not-installed">Not Installed</SelectItem>
              </SelectContent>
            </Select>

            <Select
              value={tagFilter}
              onValueChange={(val) => {
                setTagFilter(val);
                setSelected({});
              }}
            >
              <SelectTrigger className="w-[160px]">
                <span className="flex items-center gap-2">
                  <TagsIcon className="size-4 text-muted-foreground" />
                  <SelectValue placeholder="Tag" />
                </span>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Tags</SelectItem>
                {tagOptions.map((tag) => (
                  <SelectItem key={tag} value={tag}>
                    {tag}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select value={String(perPage)} onValueChange={(v) => setPerPage(Number(v))}>
              <SelectTrigger className="w-[160px]">
                <span className="flex items-center gap-2">
                  <ListIcon className="size-4 text-muted-foreground" />
                  <SelectValue placeholder="Per page" />
                </span>
              </SelectTrigger>
              <SelectContent>
                {[10, 15, 30, 50, 100].map((n) => (
                  <SelectItem key={n} value={String(n)}>
                    {n}/page
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <div className="flex items-center gap-2 text-sm">
              <Badge variant="outline" className="h-9 gap-2 px-3 text-sm font-medium">
                <CheckCircle2Icon className="size-4 text-success" />
                {stats.installed} installed
              </Badge>
              <Badge variant="outline" className="h-9 gap-2 px-3 text-sm font-medium">
                <XCircleIcon className="size-4 text-muted-foreground" />
                {stats.notInstalled} missing
              </Badge>
              <Badge variant="outline" className="h-9 gap-2 px-3 text-sm font-medium">
                <PlusCircleIcon className="size-4 text-muted-foreground" />
                {stats.optional} optional
              </Badge>
            </div>

          </div>

        </CardContent>
      </Card>

      {/* Batch Selection Bar */}
      {selectedCount > 0 && (
        <div className="sticky top-0 z-10 flex items-center justify-between rounded-lg border bg-card p-3 shadow-sm">
          <div className="text-sm">
            <span className="font-medium">{selectedCount}</span> tool
            {selectedCount > 1 ? "s" : ""} selected
          </div>
          <div className="flex items-center gap-2">
            <Button size="sm" variant="outline" onClick={() => setSelected({})}>
              Clear Selection
            </Button>
            <Button size="sm" onClick={doInstallSelected} disabled={installingBatch}>
              {installingBatch ? (
                <LoaderIcon className="mr-2 size-4 animate-spin" />
              ) : (
                <DownloadIcon className="mr-2 size-4" />
              )}
              Install Selected
            </Button>
          </div>
        </div>
      )}

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-12 pl-4">
                  <Checkbox
                    checked={allSelected}
                    onCheckedChange={(v) => toggleSelectAll(Boolean(v))}
                  />
                </TableHead>
                <SortableTableHead
                  field="name"
                  currentSort={sortState}
                  onSort={(f) => toggleSort(f as RegistrySortField)}
                >
                  Tool
                </SortableTableHead>
                <SortableTableHead
                  field="version"
                  currentSort={sortState}
                  onSort={(f) => toggleSort(f as RegistrySortField)}
                  className="w-[110px]"
                >
                  Version
                </SortableTableHead>
                <SortableTableHead
                  field="tags"
                  currentSort={sortState}
                  onSort={(f) => toggleSort(f as RegistrySortField)}
                  className="hidden md:table-cell"
                >
                  Tags
                </SortableTableHead>
                <SortableTableHead
                  field="description"
                  currentSort={sortState}
                  onSort={(f) => toggleSort(f as RegistrySortField)}
                  className="hidden lg:table-cell"
                >
                  Description
                </SortableTableHead>
                <SortableTableHead
                  field="optional"
                  currentSort={sortState}
                  onSort={(f) => toggleSort(f as RegistrySortField)}
                  className="hidden md:table-cell w-[110px]"
                >
                  Optional
                </SortableTableHead>
                <SortableTableHead
                  field="status"
                  currentSort={sortState}
                  onSort={(f) => toggleSort(f as RegistrySortField)}
                  className="w-[120px]"
                >
                  Status
                </SortableTableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-12 text-muted-foreground">
                    No tools found
                  </TableCell>
                </TableRow>
              ) : (
                pageRows.map((row) => {
                  const name = row.name;
                  const meta = row.meta as any;
                  const isInstalled = Boolean(meta.installed);
                  const version = typeof meta.version === "string" ? meta.version : "";
                  const tags = Array.isArray(meta.tags) ? (meta.tags as string[]) : [];

                  return (
                    <TableRow key={name} className={cn(selected[name] && "bg-muted/50")}>
                      <TableCell className="pl-4">
                        <Checkbox
                          checked={Boolean(selected[name])}
                          onCheckedChange={(v) => toggleSelect(name, Boolean(v))}
                        />
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <PackageIcon className="size-4 text-muted-foreground" />
                          <span className="font-medium">{name}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        {version ? (
                          <Badge variant="outline" className="font-mono text-xs">
                            {version}
                          </Badge>
                        ) : (
                          "-"
                        )}
                      </TableCell>
                      <TableCell className="hidden md:table-cell">
                        {tags.length > 0 ? (
                          <div className="flex flex-wrap gap-1 max-w-[280px]">
                            {tags.slice(0, 3).map((t: string) => (
                              <Badge
                                key={t}
                                variant="outline"
                                className={`text-xs ${getTagColor(t)}`}
                              >
                                {t}
                              </Badge>
                            ))}
                            {tags.length > 3 && (
                              <Badge
                                variant="outline"
                                className="text-xs bg-secondary text-secondary-foreground border-transparent"
                              >
                                +{tags.length - 3}
                              </Badge>
                            )}
                          </div>
                        ) : (
                          "-"
                        )}
                      </TableCell>
                      <TableCell className="hidden lg:table-cell text-muted-foreground">
                        {meta.desc || "-"}
                      </TableCell>
                      <TableCell className="hidden md:table-cell">
                        {typeof meta.optional === "boolean" ? (
                          <Badge variant="outline" className="text-xs">
                            {meta.optional ? "Yes" : "No"}
                          </Badge>
                        ) : (
                          "-"
                        )}
                      </TableCell>
                      <TableCell>
                        {isInstalled ? (
                          <Badge
                            variant="outline"
                            className="gap-1 border-transparent bg-success-soft text-success"
                          >
                            <CheckCircle2Icon className="size-3" />
                            Installed
                          </Badge>
                        ) : (
                          <Badge variant="secondary" className="gap-1">
                            <XCircleIcon className="size-3" />
                            Missing
                          </Badge>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>

          {filtered.length > 0 && (
            <div className="flex items-center justify-between gap-3 border-t px-4 py-3">
              <div className="text-sm text-muted-foreground">
                Showing {startIndex + 1}-{endIndexExclusive} of {filtered.length}
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  className="rounded-md"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={currentPage <= 1}
                >
                  <ChevronLeftIcon className="size-4" />
                  Prev
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="rounded-md"
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={currentPage >= totalPages}
                >
                  Next
                  <ChevronRightIcon className="size-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Install Workflow Card */}
      <Card className="overflow-hidden">
        <SectionCardHeader
          icon={GitBranchIcon}
          title="Install Workflow"
          description="Install workflow from a Git repository or ZIP archive URL"
        />
        <CardContent>
          <div className="flex flex-col gap-4 sm:flex-row">
            <div className="flex-1">
              <Label htmlFor="workflow-source" className="sr-only">
                Workflow Source
              </Label>
              <Input
                id="workflow-source"
                placeholder="https://github.com/osmedeus/osmedeus-workflow.git"
                value={workflowSource}
                onChange={(e) => setWorkflowSource(e.target.value)}
              />
            </div>
            <Button onClick={doInstallWorkflow} disabled={installingWorkflow}>
              {installingWorkflow ? (
                <LoaderIcon className="mr-2 size-4 animate-spin" />
              ) : (
                <FolderArchiveIcon className="mr-2 size-4" />
              )}
              Install
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
