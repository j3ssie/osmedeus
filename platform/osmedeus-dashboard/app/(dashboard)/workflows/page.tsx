"use client";

import * as React from "react";
import { fetchMockWorkflowsList, fetchWorkflowsList, fetchWorkflowTags, refreshWorkflowIndex } from "@/lib/api/workflows";
import type { Workflow } from "@/lib/types/workflow";
import { isDemoMode } from "@/lib/api/demo-mode";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { ScrollArea } from "@/components/ui/scroll-area";
import { SortableTableHead } from "@/components/ui/sortable-table-head";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogTrigger,
} from "@/components/ui/dialog";
import Link from "next/link";
import { CheckIcon, ChevronsUpDownIcon, ClipboardIcon, DatabaseIcon, EyeIcon, FileCodeIcon, HardDriveIcon, RefreshCwIcon, SearchIcon, TagIcon, LayersIcon, BoxIcon, ChevronLeftIcon, ChevronRightIcon } from "lucide-react";
import { toast } from "sonner";
import * as yaml from "js-yaml";
import type { SortDirection } from "@/lib/types/asset";
import { cn } from "@/lib/utils";

export default function WorkflowsListingPage() {
  const [items, setItems] = React.useState<Workflow[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [query, setQuery] = React.useState("");
  const [kindFilter, setKindFilter] = React.useState<"all" | "module" | "flow">("all");
  const [source, setSource] = React.useState<"db" | "filesystem">("db");
  const [allTags, setAllTags] = React.useState<string[]>([]);
  const [selectedTag, setSelectedTag] = React.useState<string>("all");
  const [mockOnly, setMockOnly] = React.useState(() => isDemoMode());
  const prevSelectedTagRef = React.useRef<string>("all");
  const [tagPopoverOpen, setTagPopoverOpen] = React.useState(false);
  const [tagSearch, setTagSearch] = React.useState("");
  const [offset, setOffset] = React.useState(0);
  const [limit, setLimit] = React.useState(20);
  const [total, setTotal] = React.useState(0);

  const filteredTags = React.useMemo(() => {
    const q = tagSearch.trim().toLowerCase();
    if (!q) return allTags;
    return allTags.filter((t) => t.toLowerCase().includes(q));
  }, [allTags, tagSearch]);

  type WorkflowSortField =
    | "name"
    | "kind"
    | "description"
    | "steps"
    | "modules"
    | "params"
    | "tags"
    | "action";

  const [sortState, setSortState] = React.useState<{
    field: WorkflowSortField;
    direction: SortDirection;
  }>({ field: "name", direction: "asc" });

  // Upload dialog state
  const [uploadOpen, setUploadOpen] = React.useState(false);
  const [uploadYaml, setUploadYaml] = React.useState("");
  const [uploadId, setUploadId] = React.useState("");
  const [refreshing, setRefreshing] = React.useState(false);

  const load = React.useCallback(async () => {
    try {
      setLoading(true);
      const res = mockOnly
        ? await fetchMockWorkflowsList({
            kind: kindFilter === "all" ? undefined : kindFilter,
            search: query.trim() || undefined,
            tags: ["mock-data"],
            offset,
            limit,
          })
        : await fetchWorkflowsList({
            source,
            kind: kindFilter === "all" ? undefined : kindFilter,
            search: query.trim() || undefined,
            tags: selectedTag !== "all" ? [selectedTag] : undefined,
            offset,
            limit,
          });
      setItems(res.items);
      setTotal(res.pagination.total);
      if (!mockOnly && isDemoMode()) {
        setMockOnly(true);
        prevSelectedTagRef.current = selectedTag;
        setSelectedTag("all");
      }
    } catch (e) {
      toast.error("Failed to load workflows", { description: e instanceof Error ? e.message : "" });
    } finally {
      setLoading(false);
    }
  }, [source, kindFilter, query, selectedTag, mockOnly, offset, limit]);

  React.useEffect(() => {
    load();
  }, [load]);

  React.useEffect(() => {
    const loadTags = async () => {
      try {
        const tags = await fetchWorkflowTags();
        setAllTags(tags);
      } catch {
        setAllTags([]);
      }
    };
    loadTags();
  }, []);

  const handleRefreshIndex = async () => {
    try {
      setRefreshing(true);
      const res = await refreshWorkflowIndex(false);
      toast.success(res.message || "Workflows indexed");
      setOffset(0);
      load();
    } catch (e) {
      toast.error("Failed to refresh index", { description: e instanceof Error ? e.message : "" });
    } finally {
      setRefreshing(false);
    }
  };

  const handleUpload = () => {
    try {
      const doc: any = yaml.load(uploadYaml) || {};
      const id = (uploadId || doc?.name || "").toString().trim();
      if (!id) {
        toast.error("Invalid YAML or missing name/ID");
        return;
      }
      if (typeof window !== "undefined") {
        const raw = window.localStorage.getItem("osmedeus_custom_workflows");
        const obj = raw ? JSON.parse(raw) : {};
        obj[id] = uploadYaml;
        window.localStorage.setItem("osmedeus_custom_workflows", JSON.stringify(obj));
      }
      toast.success("Workflow uploaded", { description: `Added ${id}` });
      setUploadOpen(false);
      setUploadYaml("");
      setUploadId("");
      load();
    } catch (e) {
      toast.error("Failed to parse YAML", { description: e instanceof Error ? e.message : "" });
    }
  };

  const currentPage = Math.floor(offset / limit) + 1;
  const totalPages = Math.ceil(total / limit);

  const toggleSort = React.useCallback((field: WorkflowSortField) => {
    setSortState((prev) => {
      if (prev.field === field) {
        return { field, direction: prev.direction === "asc" ? "desc" : "asc" };
      }
      return { field, direction: "asc" };
    });
  }, []);

  const sortedItems = React.useMemo(() => {
    const getValue = (
      field: WorkflowSortField,
      wf: Workflow
    ): { missing: boolean; value: string | number } => {
      switch (field) {
        case "name":
          return { missing: !wf.name, value: wf.name ?? "" };
        case "kind":
          return { missing: !wf.kind, value: wf.kind ?? "" };
        case "description":
          return {
            missing: !wf.description,
            value: (wf.description ?? "").toString(),
          };
        case "steps":
          return { missing: wf.step_count == null, value: wf.step_count ?? 0 };
        case "modules":
          return {
            missing: wf.module_count == null,
            value: wf.module_count ?? 0,
          };
        case "params":
          return {
            missing: !wf.params || wf.params.length === 0,
            value: wf.params?.length ?? 0,
          };
        case "tags": {
          const tags = wf.tags ?? [];
          const v = tags.join(",");
          return { missing: tags.length === 0, value: v };
        }
        case "action":
          return { missing: !wf.name, value: wf.name ?? "" };
      }
    };

    const out = [...items];
    out.sort((a, b) => {
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
    return out;
  }, [items, sortState.direction, sortState.field]);

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader className="pb-4">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <LayersIcon className="size-5" />
                Workflows
              </CardTitle>
              <CardDescription>Browse and manage workflow definitions</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={handleRefreshIndex}
                disabled={refreshing}
              >
                <RefreshCwIcon className={`mr-2 size-4 ${refreshing ? "animate-spin" : ""}`} />
                Refresh Index
              </Button>
              <div className="flex items-center gap-2 rounded-md border px-2 py-1">
                <Label htmlFor="mock-only" className="text-xs text-muted-foreground">
                  Show Mock Workflow
                </Label>
                <Switch
                  id="mock-only"
                  checked={mockOnly}
                  onCheckedChange={(checked) => {
                    setMockOnly(checked);
                    setOffset(0);
                    if (checked) {
                      prevSelectedTagRef.current = selectedTag;
                      setSelectedTag("all");
                    } else {
                      setSelectedTag(prevSelectedTagRef.current || "all");
                    }
                  }}
                />
              </div>
              <Dialog open={uploadOpen} onOpenChange={setUploadOpen}>
                <DialogTrigger asChild>
                  <Button size="sm">
                    <FileCodeIcon className="mr-2 size-4" />
                    Upload
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-xl">
                  <DialogHeader>
                    <DialogTitle>Upload Workflow YAML</DialogTitle>
                    <DialogDescription>
                      Paste YAML content. If ID is empty, name from YAML is used.
                    </DialogDescription>
                  </DialogHeader>
                  <div className="space-y-3">
                    <Input
                      placeholder="Workflow ID (optional)"
                      value={uploadId}
                      onChange={(e) => setUploadId(e.target.value)}
                    />
                    <div className="flex items-center justify-end">
                      <Button
                        type="button"
                        variant="outline"
                        size="icon"
                        disabled={!uploadYaml}
                        onClick={async () => {
                          try {
                            await navigator.clipboard.writeText(uploadYaml);
                            toast.success("Copied to clipboard");
                          } catch {
                            toast.error("Failed to copy");
                          }
                        }}
                      >
                        <ClipboardIcon className="size-4" />
                        <span className="sr-only">Copy YAML</span>
                      </Button>
                    </div>
                    <textarea
                      value={uploadYaml}
                      onChange={(e) => setUploadYaml(e.target.value)}
                      className="min-h-48 w-full rounded-md border bg-background p-3 font-mono text-sm"
                      placeholder="Paste YAML here..."
                    />
                  </div>
                  <DialogFooter>
                    <Button
                      variant="outline"
                      onClick={() => {
                        setUploadOpen(false);
                        setUploadYaml("");
                        setUploadId("");
                      }}
                    >
                      Cancel
                    </Button>
                    <Button onClick={handleUpload}>Save</Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </div>
          </div>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* Filters */}
          <div className="flex flex-wrap items-center gap-3">
            <div className="relative flex-1 min-w-[200px] max-w-sm">
              <SearchIcon className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search workflows..."
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  setOffset(0);
                }}
                className="pl-9"
              />
            </div>

            <Select
              value={source}
              onValueChange={(val) => {
                setSource(val as "db" | "filesystem");
                setOffset(0);
              }}
            >
              <SelectTrigger className="w-[140px]">
                <span className="flex items-center gap-2">
                  {source === "filesystem" ? (
                    <HardDriveIcon className="size-4 text-muted-foreground" />
                  ) : (
                    <DatabaseIcon className="size-4 text-muted-foreground" />
                  )}
                  <SelectValue placeholder="Source" />
                </span>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="db">Database</SelectItem>
                <SelectItem value="filesystem">Filesystem</SelectItem>
              </SelectContent>
            </Select>

            <Select
              value={kindFilter}
              onValueChange={(val) => {
                setKindFilter(val as "all" | "module" | "flow");
                setOffset(0);
              }}
            >
              <SelectTrigger className="w-[130px]">
                <span className="flex items-center gap-2">
                  {kindFilter === "module" ? (
                    <BoxIcon className="size-4 text-muted-foreground" />
                  ) : (
                    <LayersIcon className="size-4 text-muted-foreground" />
                  )}
                  <SelectValue placeholder="Kind" />
                </span>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Kinds</SelectItem>
                <SelectItem value="module">Module</SelectItem>
                <SelectItem value="flow">Flow</SelectItem>
              </SelectContent>
            </Select>

            <Popover
              open={tagPopoverOpen}
              onOpenChange={(open) => {
                setTagPopoverOpen(open);
                if (!open) setTagSearch("");
              }}
            >
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  disabled={mockOnly}
                  className="w-[200px] justify-between"
                >
                  <span className="flex items-center gap-2 min-w-0">
                    <TagIcon className="size-4 text-muted-foreground" />
                    <span className="truncate">
                      {selectedTag === "all" ? "All Tags" : selectedTag}
                    </span>
                  </span>
                  <ChevronsUpDownIcon className="size-4 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[240px] p-0" align="start">
                <div className="p-2 border-b">
                  <Input
                    placeholder="Search tags..."
                    value={tagSearch}
                    onChange={(e) => setTagSearch(e.target.value)}
                    className="h-8"
                  />
                </div>
                <ScrollArea className="h-[240px]">
                  <div className="p-2 space-y-1">
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className={cn(
                        "w-full justify-start h-8",
                        selectedTag === "all" && "bg-muted"
                      )}
                      onClick={() => {
                        setSelectedTag("all");
                        setOffset(0);
                        setTagPopoverOpen(false);
                      }}
                    >
                      <span className="inline-flex items-center gap-2">
                        <span className="size-4 inline-flex items-center justify-center">
                          {selectedTag === "all" && <CheckIcon className="size-4" />}
                        </span>
                        All Tags
                      </span>
                    </Button>

                    {filteredTags.map((t) => (
                      <Button
                        key={t}
                        type="button"
                        variant="ghost"
                        size="sm"
                        className={cn(
                          "w-full justify-start h-8",
                          selectedTag === t && "bg-muted"
                        )}
                        onClick={() => {
                          setSelectedTag(t);
                          setOffset(0);
                          setTagPopoverOpen(false);
                        }}
                      >
                        <span className="inline-flex items-center gap-2 min-w-0">
                          <span className="size-4 inline-flex items-center justify-center">
                            {selectedTag === t && <CheckIcon className="size-4" />}
                          </span>
                          <span className="truncate">{t}</span>
                        </span>
                      </Button>
                    ))}
                    {filteredTags.length === 0 && (
                      <p className="text-sm text-muted-foreground text-center py-4">
                        No tags found
                      </p>
                    )}
                  </div>
                </ScrollArea>
              </PopoverContent>
            </Popover>
          </div>

          {/* Table */}
          {loading ? (
            <div className="py-16 text-center text-sm text-muted-foreground">Loading...</div>
          ) : items.length === 0 ? (
            <div className="py-16 text-center text-sm text-muted-foreground">No workflows found</div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <SortableTableHead
                      field="name"
                      currentSort={sortState}
                      onSort={(f) => toggleSort(f as WorkflowSortField)}
                    >
                      Name
                    </SortableTableHead>
                    <SortableTableHead
                      field="kind"
                      currentSort={sortState}
                      onSort={(f) => toggleSort(f as WorkflowSortField)}
                      className="w-[100px]"
                    >
                      Kind
                    </SortableTableHead>
                    <SortableTableHead
                      field="description"
                      currentSort={sortState}
                      onSort={(f) => toggleSort(f as WorkflowSortField)}
                      className="hidden md:table-cell"
                    >
                      Description
                    </SortableTableHead>
                    <SortableTableHead
                      field="steps"
                      currentSort={sortState}
                      onSort={(f) => toggleSort(f as WorkflowSortField)}
                      className="w-[80px] text-center"
                    >
                      Steps
                    </SortableTableHead>
                    <SortableTableHead
                      field="modules"
                      currentSort={sortState}
                      onSort={(f) => toggleSort(f as WorkflowSortField)}
                      className="w-[80px] text-center hidden sm:table-cell"
                    >
                      Modules
                    </SortableTableHead>
                    <SortableTableHead
                      field="params"
                      currentSort={sortState}
                      onSort={(f) => toggleSort(f as WorkflowSortField)}
                      className="w-[80px] text-center hidden sm:table-cell"
                    >
                      Params
                    </SortableTableHead>
                    <SortableTableHead
                      field="tags"
                      currentSort={sortState}
                      onSort={(f) => toggleSort(f as WorkflowSortField)}
                      className="hidden lg:table-cell"
                    >
                      Tags
                    </SortableTableHead>
                    <SortableTableHead
                      field="action"
                      currentSort={sortState}
                      onSort={(f) => toggleSort(f as WorkflowSortField)}
                      className="w-[80px]"
                    >
                      Action
                    </SortableTableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sortedItems.map((wf) => (
                    <TableRow key={wf.name}>
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          <BoxIcon className="size-4 text-muted-foreground" />
                          <span>{wf.name}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant="outline"
                          className={
                            wf.kind === "flow"
                              ? "border-transparent bg-purple-soft text-purple"
                              : "border-transparent bg-info-soft text-info"
                          }
                        >
                          {wf.kind}
                        </Badge>
                      </TableCell>
                      <TableCell className="hidden md:table-cell text-muted-foreground max-w-[300px] truncate">
                        {wf.description || "-"}
                      </TableCell>
                      <TableCell className="text-center">{wf.step_count}</TableCell>
                      <TableCell className="text-center hidden sm:table-cell">{wf.module_count}</TableCell>
                      <TableCell className="text-center hidden sm:table-cell">
                        {wf.params?.length || 0}
                        {wf.required_params?.length > 0 && (
                          <span className="text-muted-foreground text-xs ml-1">
                            ({wf.required_params.length} req)
                          </span>
                        )}
                      </TableCell>
                      <TableCell className="hidden lg:table-cell">
                        <div className="flex gap-1 flex-wrap">
                          {(wf.tags || []).slice(0, 3).map((t) => (
                            <Badge key={t} variant="secondary" className="text-xs">
                              {t}
                            </Badge>
                          ))}
                          {(wf.tags || []).length > 3 && (
                            <Badge variant="secondary" className="text-xs">
                              +{wf.tags.length - 3}
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Button variant="outline" size="sm" asChild>
                          <Link href={`/workflows-editor?workflow=${encodeURIComponent(wf.name)}`}>
                            <EyeIcon className="size-4" />
                            <span className="sr-only">Open</span>
                          </Link>
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {/* Pagination */}
              <div className="flex items-center justify-between pt-2">
                <div className="text-sm text-muted-foreground">
                  Showing {offset + 1}-{Math.min(offset + limit, total)} of {total}
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setOffset(Math.max(0, offset - limit))}
                    disabled={offset <= 0}
                  >
                    <ChevronLeftIcon className="size-4" />
                    Prev
                  </Button>
                  <span className="text-sm text-muted-foreground">
                    Page {currentPage} of {totalPages}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setOffset(offset + limit)}
                    disabled={offset + limit >= total}
                  >
                    Next
                    <ChevronRightIcon className="size-4" />
                  </Button>
                  <Select
                    value={String(limit)}
                    onValueChange={(val) => {
                      setLimit(Number(val));
                      setOffset(0);
                    }}
                  >
                    <SelectTrigger className="w-[90px]">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="20">20/page</SelectItem>
                      <SelectItem value="50">50/page</SelectItem>
                      <SelectItem value="100">100/page</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
