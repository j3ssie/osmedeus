"use client";

import * as React from "react";
import Link from "next/link";
import { fetchSchedules, createSchedule, enableSchedule, disableSchedule, triggerSchedule, deleteSchedule, updateSchedule } from "@/lib/api/schedules";
import type { Schedule } from "@/lib/types/schedule";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Checkbox } from "@/components/ui/checkbox";
import { SortableTableHead } from "@/components/ui/sortable-table-head";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { toast } from "sonner";
import { LoaderIcon, PlayIcon, PauseIcon, Trash2Icon, PencilIcon, RefreshCcwIcon, ChevronLeftIcon, ChevronRightIcon, Columns3Icon, XIcon, PlusIcon } from "lucide-react";
import type { SortDirection } from "@/lib/types/asset";

export default function SchedulesPage() {
  const [schedules, setSchedules] = React.useState<Schedule[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [creating, setCreating] = React.useState(false);
  const [formOpen, setFormOpen] = React.useState(false);
  const [name, setName] = React.useState("");
  const [workflowName, setWorkflowName] = React.useState("");
  const [target, setTarget] = React.useState("");
  const [cron, setCron] = React.useState("");
  const [enabled, setEnabled] = React.useState(true);
  const [editingId, setEditingId] = React.useState<string | null>(null);
  const [updating, setUpdating] = React.useState(false);
  const [offset, setOffset] = React.useState(0);
  const [limit, setLimit] = React.useState(20);

  const [query, setQuery] = React.useState("");
  const [enabledFilter, setEnabledFilter] = React.useState<"all" | "enabled" | "disabled">("all");
  const [triggerFilter, setTriggerFilter] = React.useState<"all" | "cron" | "event" | "watch" | "manual">("all");

  type ScheduleSortField =
    | "name"
    | "workflow"
    | "trigger"
    | "schedule_or_topic"
    | "input_config"
    | "enabled"
    | "last_run"
    | "next_run"
    | "run_count"
    | "updated"
    | "actions";

  const [sortState, setSortState] = React.useState<{
    field: ScheduleSortField;
    direction: SortDirection;
  }>({ field: "name", direction: "asc" });

  const [visibleColumns, setVisibleColumns] = React.useState({
    name: true,
    workflow: true,
    trigger: true,
    schedule_or_topic: true,
    input_config: false,
    enabled: true,
    last_run: true,
    next_run: false,
    run_count: false,
    updated: false,
    actions: true,
  });

  const inlineCodeClass = "inline-flex items-center rounded-md border bg-muted px-1.5 py-0.5 font-mono text-xs";

  const load = React.useCallback(async () => {
    try {
      setLoading(true);
      const res = await fetchSchedules({ offset: 0, limit: 5000 });
      setSchedules(res.data);
    } catch (e) {
      toast.error("Failed to load schedules", { description: e instanceof Error ? e.message : "" });
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    load();
  }, [load]);

  React.useEffect(() => {
    setOffset(0);
  }, [enabledFilter, triggerFilter, query, sortState.direction, sortState.field]);

  const toggleSort = (field: ScheduleSortField) => {
    setSortState((prev) => {
      if (prev.field === field) {
        return { field, direction: prev.direction === "asc" ? "desc" : "asc" };
      }
      return { field, direction: "asc" };
    });
  };

  const visibleColumnCount = React.useMemo(() => {
    return Object.values(visibleColumns).filter(Boolean).length;
  }, [visibleColumns]);

  const filteredSchedules = React.useMemo(() => {
    const q = query.trim().toLowerCase();

    return schedules.filter((s) => {
      if (enabledFilter === "enabled" && !s.isEnabled) return false;
      if (enabledFilter === "disabled" && s.isEnabled) return false;

      if (triggerFilter !== "all" && s.triggerType !== triggerFilter) return false;

      if (!q) return true;

      const scheduleOrTopic = s.triggerType === "event" ? (s.eventTopic ?? "") : (s.schedule ?? "");
      const haystack = [
        s.name,
        s.workflowName,
        s.triggerType ?? "",
        s.triggerName ?? "",
        scheduleOrTopic,
      ]
        .filter(Boolean)
        .join(" ")
        .toLowerCase();

      return haystack.includes(q);
    });
  }, [enabledFilter, query, schedules, triggerFilter]);

  const sortedSchedules = React.useMemo(() => {
    const getValue = (field: ScheduleSortField, s: Schedule): { missing: boolean; value: string | number } => {
      switch (field) {
        case "name":
          return { missing: !s.name, value: s.name ?? "" };
        case "workflow":
          return { missing: !s.workflowName, value: s.workflowName ?? "" };
        case "trigger":
          return {
            missing: !(s.triggerType || s.triggerName),
            value: `${s.triggerType ?? ""} ${s.triggerName ?? ""}`.trim(),
          };
        case "schedule_or_topic": {
          const v = s.triggerType === "event" ? (s.eventTopic ?? "") : (s.schedule ?? "");
          return { missing: !v, value: v };
        }
        case "input_config": {
          const v = s.inputConfig ? JSON.stringify(s.inputConfig) : "";
          return { missing: !v, value: v };
        }
        case "enabled":
          return { missing: false, value: Number(s.isEnabled) };
        case "last_run":
          return { missing: !s.lastRun, value: s.lastRun ? s.lastRun.getTime() : 0 };
        case "next_run":
          return { missing: !s.nextRun, value: s.nextRun ? s.nextRun.getTime() : 0 };
        case "run_count":
          return { missing: typeof s.runCount !== "number", value: typeof s.runCount === "number" ? s.runCount : 0 };
        case "updated":
          return { missing: !s.updatedAt, value: s.updatedAt ? s.updatedAt.getTime() : 0 };
        case "actions":
          return { missing: !s.id, value: s.id ?? "" };
      }
    };

    const items = [...filteredSchedules];
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
  }, [filteredSchedules, sortState.direction, sortState.field]);

  const effectiveTotal = sortedSchedules.length;

  const pageSchedules = React.useMemo(() => {
    return sortedSchedules.slice(offset, offset + limit);
  }, [sortedSchedules, limit, offset]);

  const columnOptions = React.useMemo(
    () =>
      [
        ["name", "Name"],
        ["workflow", "Workflow"],
        ["trigger", "Trigger"],
        ["schedule_or_topic", "Schedule / Topic"],
        ["input_config", "Input Config"],
        ["enabled", "Enabled"],
        ["last_run", "Last Run"],
        ["next_run", "Next Run"],
        ["run_count", "Run Count"],
        ["updated", "Updated"],
        ["actions", "Actions"],
      ] as const,
    []
  );

  const setColumnChecked = (
    key: keyof typeof visibleColumns,
    checked: boolean
  ) => {
    setVisibleColumns((prev) => ({ ...prev, [key]: checked }));
  };

  const resetForm = () => {
    setName("");
    setWorkflowName("");
    setTarget("");
    setCron("");
    setEnabled(true);
    setEditingId(null);
  };

  const submit = async () => {
    if (!name || !cron || (!editingId && (!workflowName || !target))) {
      toast.error("Please fill all fields");
      return;
    }
    setCreating(true);
    try {
      if (editingId) {
        setUpdating(true);
        const ok = await updateSchedule(editingId, { name, schedule: cron, enabled });
        if (!ok) throw new Error("Update failed");
        toast.success("Schedule updated");
      } else {
        await createSchedule({
          name,
          workflowName,
          workflowKind: "flow",
          target,
          schedule: cron,
          enabled,
        });
        toast.success("Schedule created");
      }
      setFormOpen(false);
      resetForm();
      await load();
    } catch (e) {
      toast.error("Failed to submit", { description: e instanceof Error ? e.message : "" });
    } finally {
      setCreating(false);
      setUpdating(false);
    }
  };

  const toggleEnable = async (s: Schedule) => {
    try {
      const ok = s.isEnabled ? await disableSchedule(s.id) : await enableSchedule(s.id);
      if (!ok) throw new Error("Action failed");
      toast.success(s.isEnabled ? "Disabled" : "Enabled");
      await load();
    } catch (e) {
      toast.error("Failed to toggle", { description: e instanceof Error ? e.message : "" });
    }
  };

  const trigger = async (s: Schedule) => {
    try {
      const ok = await triggerSchedule(s.id);
      if (!ok) throw new Error("Trigger failed");
      toast.success("Triggered");
    } catch (e) {
      toast.error("Failed to trigger", { description: e instanceof Error ? e.message : "" });
    }
  };

  const remove = async (s: Schedule) => {
    try {
      const ok = await deleteSchedule(s.id);
      if (!ok) throw new Error("Delete failed");
      toast.success("Schedule deleted");
      await load();
    } catch (e) {
      toast.error("Failed to delete", { description: e instanceof Error ? e.message : "" });
    }
  };

  return (
    <div className="space-y-4">
      

      {formOpen && (
        <Card className="max-w-2xl">
          <CardHeader>
            <CardTitle>{editingId ? "Edit Schedule" : "Create Schedule"}</CardTitle>
            <CardDescription>Define workflow, target and cron expression</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="workflow">Workflow Name</Label>
              <Input id="workflow" value={workflowName} onChange={(e) => setWorkflowName(e.target.value)} placeholder="subdomain-enum" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="target">Target</Label>
              <Input id="target" value={target} onChange={(e) => setTarget(e.target.value)} placeholder="example.com" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="cron">Cron</Label>
              <Input id="cron" value={cron} onChange={(e) => setCron(e.target.value)} placeholder="0 2 * * *" />
            </div>
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="enabled">Enabled</Label>
                <p className="text-xs text-muted-foreground">Toggle schedule activation</p>
              </div>
              <Switch id="enabled" checked={enabled} onCheckedChange={setEnabled} />
            </div>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => { setFormOpen(false); resetForm(); }}>Cancel</Button>
              <Button onClick={submit} disabled={creating || updating}>
                {creating || updating ? <LoaderIcon className="mr-2 size-4 animate-spin" /> : null}
                {editingId ? "Update" : "Create"}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader className="pb-4">
          <div className="flex items-start justify-between gap-4">
            <div>
              <CardTitle>Scheduled Jobs</CardTitle>
              <CardDescription>List of configured schedules</CardDescription>
            </div>
            <div className="flex flex-wrap items-center justify-end gap-2">
              <div className="relative">
                <Input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search schedules..."
                  className="h-9 w-[240px]"
                />
              </div>

              <Select value={enabledFilter} onValueChange={(v) => setEnabledFilter(v as any)}>
                <SelectTrigger className="h-9 w-[140px]">
                  <SelectValue placeholder="Enabled" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Enabled: All</SelectItem>
                  <SelectItem value="enabled">Enabled: Yes</SelectItem>
                  <SelectItem value="disabled">Enabled: No</SelectItem>
                </SelectContent>
              </Select>

              <Select value={triggerFilter} onValueChange={(v) => setTriggerFilter(v as any)}>
                <SelectTrigger className="h-9 w-[160px]">
                  <SelectValue placeholder="Trigger type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Trigger: All</SelectItem>
                  <SelectItem value="cron">Trigger: Cron</SelectItem>
                  <SelectItem value="event">Trigger: Event</SelectItem>
                  <SelectItem value="watch">Trigger: Watch</SelectItem>
                  <SelectItem value="manual">Trigger: Manual</SelectItem>
                </SelectContent>
              </Select>

              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="outline" className="h-9 shrink-0 rounded-md px-3">
                    <Columns3Icon className="mr-2 size-4 opacity-70" />
                    Columns
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="end" className="w-[260px] p-3">
                  <div className="text-sm font-medium">Show columns</div>
                  <div className="mt-3 space-y-2">
                    {columnOptions.map(([key, label]) => {
                      const locked = key === "name" || key === "actions";
                      return (
                        <label key={key} className="flex items-center justify-between gap-3">
                          <span className="text-sm">{label}</span>
                          <Checkbox
                            checked={visibleColumns[key]}
                            disabled={locked || (visibleColumns[key] && visibleColumnCount <= 1)}
                            onCheckedChange={(v) => setColumnChecked(key, v === true)}
                          />
                        </label>
                      );
                    })}
                  </div>
                </PopoverContent>
              </Popover>

              <Button className="shrink-0" variant="outline" onClick={load}>
                <RefreshCcwIcon className="mr-2 size-4" />
                Refresh
              </Button>

              <Button asChild className="shrink-0">
                <Link href="/scans/new?schedule=1">
                  <PlusIcon className="mr-2 size-4" />
                  New Scan
                </Link>
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="py-10 text-center text-sm text-muted-foreground">Loading...</div>
          ) : pageSchedules.length === 0 ? (
            <div className="py-10 text-center text-sm text-muted-foreground">No schedules found</div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    {visibleColumns.name && (
                      <SortableTableHead
                        field="name"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                      >
                        Name
                      </SortableTableHead>
                    )}
                    {visibleColumns.workflow && (
                      <SortableTableHead
                        field="workflow"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                      >
                        Workflow
                      </SortableTableHead>
                    )}
                    {visibleColumns.trigger && (
                      <SortableTableHead
                        field="trigger"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                      >
                        Trigger
                      </SortableTableHead>
                    )}
                    {visibleColumns.schedule_or_topic && (
                      <SortableTableHead
                        field="schedule_or_topic"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                      >
                        Schedule / Topic
                      </SortableTableHead>
                    )}
                    {visibleColumns.input_config && (
                      <SortableTableHead
                        field="input_config"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                        className="hidden xl:table-cell"
                      >
                        Input Config
                      </SortableTableHead>
                    )}
                    {visibleColumns.enabled && (
                      <SortableTableHead
                        field="enabled"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                      >
                        Enabled
                      </SortableTableHead>
                    )}
                    {visibleColumns.last_run && (
                      <SortableTableHead
                        field="last_run"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                        className="hidden md:table-cell"
                      >
                        Last Run
                      </SortableTableHead>
                    )}
                    {visibleColumns.next_run && (
                      <SortableTableHead
                        field="next_run"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                        className="hidden md:table-cell"
                      >
                        Next Run
                      </SortableTableHead>
                    )}
                    {visibleColumns.run_count && (
                      <SortableTableHead
                        field="run_count"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                        className="hidden lg:table-cell"
                      >
                        Run Count
                      </SortableTableHead>
                    )}
                    {visibleColumns.updated && (
                      <SortableTableHead
                        field="updated"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                        className="hidden xl:table-cell"
                      >
                        Updated
                      </SortableTableHead>
                    )}
                    {visibleColumns.actions && (
                      <SortableTableHead
                        field="actions"
                        currentSort={sortState}
                        onSort={(f) => toggleSort(f as ScheduleSortField)}
                        className="w-[140px] text-right"
                      >
                        Actions
                      </SortableTableHead>
                    )}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {pageSchedules.map((s) => (
                    <TableRow key={s.id}>
                      {visibleColumns.name && (
                        <TableCell className="font-medium">{s.name}</TableCell>
                      )}
                      {visibleColumns.workflow && (
                        <TableCell>
                          {s.workflowName ? (
                            <Badge variant="info" className="font-mono text-xs">
                              {s.workflowName}
                            </Badge>
                          ) : (
                            <span className="text-sm text-muted-foreground">-</span>
                          )}
                        </TableCell>
                      )}
                      {visibleColumns.trigger && (
                        <TableCell>
                          {s.triggerType || s.triggerName ? (
                            <div className="flex flex-wrap items-center gap-1">
                              {s.triggerType && (
                                <Badge variant="outline" className="capitalize">
                                  {s.triggerType}
                                </Badge>
                              )}
                              {s.triggerName && (
                                <Badge variant="secondary" className="font-mono text-xs">
                                  {s.triggerName}
                                </Badge>
                              )}
                            </div>
                          ) : (
                            <span className="text-sm text-muted-foreground">-</span>
                          )}
                        </TableCell>
                      )}
                      {visibleColumns.schedule_or_topic && (
                        <TableCell>
                          {(() => {
                            const v = s.triggerType === "event" ? (s.eventTopic ?? "-") : (s.schedule ?? "-");
                            return v === "-" ? (
                              <span className="text-sm text-muted-foreground">-</span>
                            ) : (
                              <span className={inlineCodeClass}>{v}</span>
                            );
                          })()}
                        </TableCell>
                      )}
                      {visibleColumns.input_config && (
                        <TableCell className="hidden xl:table-cell max-w-[360px] truncate">
                          {s.inputConfig ? JSON.stringify(s.inputConfig) : "-"}
                        </TableCell>
                      )}
                      {visibleColumns.enabled && (
                        <TableCell>
                          <Badge variant={s.isEnabled ? "success" : "secondary"}>
                            {s.isEnabled ? "Enabled" : "Disabled"}
                          </Badge>
                        </TableCell>
                      )}
                      {visibleColumns.last_run && (
                        <TableCell className="hidden md:table-cell">
                          {s.lastRun ? (
                            <span className={inlineCodeClass}>{s.lastRun.toLocaleString()}</span>
                          ) : (
                            <span className="text-sm text-muted-foreground">-</span>
                          )}
                        </TableCell>
                      )}
                      {visibleColumns.next_run && (
                        <TableCell className="hidden md:table-cell">
                          {s.nextRun ? (
                            <span className={inlineCodeClass}>{s.nextRun.toLocaleString()}</span>
                          ) : (
                            <span className="text-sm text-muted-foreground">-</span>
                          )}
                        </TableCell>
                      )}
                      {visibleColumns.run_count && (
                        <TableCell className="hidden lg:table-cell">
                          {typeof s.runCount === "number" ? (
                            <Badge variant="outline" className="font-mono text-xs">
                              {s.runCount}
                            </Badge>
                          ) : (
                            <span className="text-sm text-muted-foreground">-</span>
                          )}
                        </TableCell>
                      )}
                      {visibleColumns.updated && (
                        <TableCell className="hidden xl:table-cell">
                          {s.updatedAt ? s.updatedAt.toLocaleString() : "-"}
                        </TableCell>
                      )}
                      {visibleColumns.actions && (
                        <TableCell className="text-right">
                          <div className="flex justify-end gap-1">
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="outline"
                                  size="icon-sm"
                                  className="rounded-md"
                                  aria-label="Trigger"
                                  onClick={() => trigger(s)}
                                >
                                  <PlayIcon className="size-4" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent side="top">Trigger</TooltipContent>
                            </Tooltip>

                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="outline"
                                  size="icon-sm"
                                  className="rounded-md"
                                  aria-label={s.isEnabled ? "Disable" : "Enable"}
                                  onClick={() => toggleEnable(s)}
                                >
                                  {s.isEnabled ? <PauseIcon className="size-4" /> : <PlayIcon className="size-4" />}
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent side="top">{s.isEnabled ? "Disable" : "Enable"}</TooltipContent>
                            </Tooltip>

                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="outline"
                                  size="icon-sm"
                                  className="rounded-md"
                                  aria-label="Edit"
                                  onClick={() => {
                                    setFormOpen(true);
                                    setEditingId(s.id);
                                    setName(s.name);
                                    setWorkflowName(s.workflowName);
                                    setTarget(String(s.inputConfig?.target ?? ""));
                                    setCron(s.schedule ?? "");
                                    setEnabled(s.isEnabled);
                                  }}
                                >
                                  <PencilIcon className="size-4" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent side="top">Edit</TooltipContent>
                            </Tooltip>

                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="outline"
                                  size="icon-sm"
                                  className="rounded-md"
                                  aria-label="Delete"
                                  onClick={() => remove(s)}
                                >
                                  <Trash2Icon className="size-4" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent side="top">Delete</TooltipContent>
                            </Tooltip>
                          </div>
                        </TableCell>
                      )}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              <div className="flex items-center justify-between pt-2">
                <div className="text-sm text-muted-foreground">
                  Showing {effectiveTotal === 0 ? 0 : Math.min(effectiveTotal, offset + 1)}-{Math.min(offset + limit, effectiveTotal)} of {effectiveTotal}
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
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setOffset(offset + limit)}
                    disabled={offset + limit >= effectiveTotal}
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
                    <SelectTrigger className="w-[110px]">
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
