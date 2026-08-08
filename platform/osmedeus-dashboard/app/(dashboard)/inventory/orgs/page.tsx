"use client";

import * as React from "react";
import { useOrg } from "@/providers/org-provider";
import { createOrg, deleteOrg, updateOrg } from "@/lib/api/orgs";
import { fetchWorkspacesList } from "@/lib/api/assets";
import type { Org } from "@/lib/types/org";
import { AssignWorkspacesDialog } from "@/components/orgs/assign-workspaces-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  BuildingIcon,
  PlusIcon,
  RefreshCcwIcon,
  MoreHorizontalIcon,
  FolderPlusIcon,
  PencilIcon,
  Trash2Icon,
  GlobeIcon,
  CheckIcon,
} from "lucide-react";
import { toast } from "sonner";

export default function OrgsPage() {
  const { orgs, activeOrg, isLoading, error, selectOrg, refresh } = useOrg();

  const [workspaceNames, setWorkspaceNames] = React.useState<string[]>([]);
  const [createOpen, setCreateOpen] = React.useState(false);
  const [renameTarget, setRenameTarget] = React.useState<Org | null>(null);
  const [deleteTarget, setDeleteTarget] = React.useState<Org | null>(null);
  const [assignTarget, setAssignTarget] = React.useState<Org | null>(null);

  const [newName, setNewName] = React.useState("");
  const [newDescription, setNewDescription] = React.useState("");
  const [renameValue, setRenameValue] = React.useState("");
  const [purge, setPurge] = React.useState(false);
  const [busy, setBusy] = React.useState(false);

  // Workspace names power the assignment dialog. Fetched unscoped so an org can
  // be given workspaces that currently belong to another one.
  const loadWorkspaces = React.useCallback(async () => {
    try {
      const result = await fetchWorkspacesList({ limit: 500, org: "" });
      setWorkspaceNames(result.items.map((w) => w.name).filter(Boolean));
    } catch {
      setWorkspaceNames([]);
    }
  }, []);

  React.useEffect(() => {
    void loadWorkspaces();
  }, [loadWorkspaces]);

  const handleCreate = async () => {
    const name = newName.trim();
    if (!name) return;
    setBusy(true);
    try {
      await createOrg({ name, description: newDescription.trim() || undefined });
      toast.success(`Created org ${name}`);
      setCreateOpen(false);
      setNewName("");
      setNewDescription("");
      await refresh();
    } catch (err) {
      toast.error(cleanError(err, "Failed to create org"));
    } finally {
      setBusy(false);
    }
  };

  const handleRename = async () => {
    if (!renameTarget) return;
    const name = renameValue.trim();
    if (!name || name === renameTarget.name) {
      setRenameTarget(null);
      return;
    }
    setBusy(true);
    try {
      await updateOrg(renameTarget.uuid, { name });
      toast.success(`Renamed to ${name}`);
      setRenameTarget(null);
      await refresh();
    } catch (err) {
      toast.error(cleanError(err, "Failed to rename org"));
    } finally {
      setBusy(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setBusy(true);
    try {
      await deleteOrg(deleteTarget.uuid, purge);
      toast.success(
        purge
          ? `Deleted ${deleteTarget.name} and its data`
          : `Deleted ${deleteTarget.name} — its data moved to the default org`
      );
      if (activeOrg?.uuid === deleteTarget.uuid) {
        selectOrg(null);
        return; // selectOrg reloads
      }
      setDeleteTarget(null);
      setPurge(false);
      await refresh();
    } catch (err) {
      toast.error(cleanError(err, "Failed to delete org"));
    } finally {
      setBusy(false);
    }
  };

  const handleAssigned = async () => {
    setAssignTarget(null);
    await refresh();
    await loadWorkspaces();
  };

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader className="flex flex-row items-start justify-between gap-4">
          <div>
            <CardTitle className="flex items-center gap-2">
              <BuildingIcon className="size-5 text-primary" />
              Orgs
            </CardTitle>
            <CardDescription>
              Group workspaces under one org to query assets, findings and scans across all of
              them. Data with no org belongs to <span className="font-medium">default</span>, and
              with no org selected the dashboard spans every org.
            </CardDescription>
          </div>
          <div className="flex shrink-0 items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => void refresh()} disabled={isLoading}>
              <RefreshCcwIcon className="size-4" />
              Refresh
            </Button>
            <Button size="sm" onClick={() => setCreateOpen(true)}>
              <PlusIcon className="size-4" />
              New org
            </Button>
          </div>
        </CardHeader>

        <CardContent>
          {error ? (
            <p className="py-8 text-center text-sm text-muted-foreground">
              Could not load orgs. This server may predate the org API.
            </p>
          ) : isLoading ? (
            <div className="flex flex-col gap-2">
              {[0, 1, 2].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : orgs.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">No orgs yet.</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead className="text-right">Workspaces</TableHead>
                  <TableHead className="text-right">Assets</TableHead>
                  <TableHead className="text-right">Vulns</TableHead>
                  <TableHead className="text-right">Runs</TableHead>
                  <TableHead className="w-10" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {orgs.map((org) => {
                  const isActive = activeOrg?.uuid === org.uuid;
                  return (
                    <TableRow key={org.uuid} className={isActive ? "bg-primary-soft/40" : undefined}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <button
                            type="button"
                            onClick={() => selectOrg(isActive ? null : { uuid: org.uuid, name: org.name })}
                            className="flex items-center gap-2 text-left hover:underline"
                            title={isActive ? "Clear org scope" : `Scope dashboard to ${org.name}`}
                          >
                            {isActive ? (
                              <CheckIcon className="size-4 text-primary" />
                            ) : (
                              <BuildingIcon className="size-4 text-muted-foreground" />
                            )}
                            <span className="font-medium">{org.name}</span>
                          </button>
                          {org.is_default && (
                            <Badge variant="secondary" className="text-[10px]">
                              default
                            </Badge>
                          )}
                          {org.tags.map((t) => (
                            <Badge key={t} variant="outline" className="text-[10px]">
                              {t}
                            </Badge>
                          ))}
                        </div>
                        {org.description && (
                          <p className="mt-0.5 text-xs text-muted-foreground">{org.description}</p>
                        )}
                      </TableCell>
                      <TableCell className="text-right tabular-nums">
                        {org.stats?.total_workspaces ?? 0}
                      </TableCell>
                      <TableCell className="text-right tabular-nums">
                        {org.stats?.total_assets ?? 0}
                      </TableCell>
                      <TableCell className="text-right tabular-nums">
                        {org.stats?.total_vulns ?? 0}
                      </TableCell>
                      <TableCell className="text-right tabular-nums">
                        {org.stats?.total_runs ?? 0}
                      </TableCell>
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="size-8">
                              <MoreHorizontalIcon className="size-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onSelect={() => setAssignTarget(org)}>
                              <FolderPlusIcon className="size-4" />
                              Assign workspaces
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              disabled={org.is_default}
                              onSelect={() => {
                                setRenameTarget(org);
                                setRenameValue(org.name);
                              }}
                            >
                              <PencilIcon className="size-4" />
                              Rename
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              disabled={org.is_default}
                              className="text-destructive focus:text-destructive"
                              onSelect={() => {
                                setDeleteTarget(org);
                                setPurge(false);
                              }}
                            >
                              <Trash2Icon className="size-4" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}

          {!isLoading && !error && (
            <p className="mt-3 flex items-center gap-1.5 text-xs text-muted-foreground">
              <GlobeIcon className="size-3.5" />
              {activeOrg
                ? `Scoped to ${activeOrg.name}. Click its name again to show all orgs.`
                : "Showing all orgs. Click an org name to scope the dashboard to it."}
            </p>
          )}
        </CardContent>
      </Card>

      {/* Create */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New org</DialogTitle>
            <DialogDescription>
              Create an org, then assign workspaces to it. Existing scan data is grouped without
              re-scanning.
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-3">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="org-name">Name</Label>
              <Input
                id="org-name"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="acme"
                autoFocus
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="org-description">Description</Label>
              <Input
                id="org-description"
                value={newDescription}
                onChange={(e) => setNewDescription(e.target.value)}
                placeholder="ACME Corp"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)} disabled={busy}>
              Cancel
            </Button>
            <Button onClick={() => void handleCreate()} disabled={busy || !newName.trim()}>
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Rename */}
      <Dialog open={!!renameTarget} onOpenChange={(o) => !o && setRenameTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Rename org</DialogTitle>
            <DialogDescription>
              Renaming does not move any data — workspaces stay assigned.
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="org-rename">Name</Label>
            <Input
              id="org-rename"
              value={renameValue}
              onChange={(e) => setRenameValue(e.target.value)}
              autoFocus
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRenameTarget(null)} disabled={busy}>
              Cancel
            </Button>
            <Button onClick={() => void handleRename()} disabled={busy || !renameValue.trim()}>
              Rename
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete */}
      <Dialog
        open={!!deleteTarget}
        onOpenChange={(o) => {
          if (!o) {
            setDeleteTarget(null);
            setPurge(false);
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete {deleteTarget?.name}</DialogTitle>
            <DialogDescription>
              By default the org&apos;s workspaces, assets, vulnerabilities and runs move to the
              default org — nothing is lost, only the grouping.
            </DialogDescription>
          </DialogHeader>

          <label className="flex items-start gap-2 rounded-control border border-destructive/40 bg-destructive/5 p-3">
            <Checkbox
              checked={purge}
              onCheckedChange={(v) => setPurge(v === true)}
              className="mt-0.5"
            />
            <span className="text-sm">
              <span className="font-medium text-destructive">Also delete its data</span>
              <span className="block text-xs text-muted-foreground">
                Permanently removes {deleteTarget?.stats?.total_workspaces ?? 0} workspaces,{" "}
                {deleteTarget?.stats?.total_assets ?? 0} assets,{" "}
                {deleteTarget?.stats?.total_vulns ?? 0} vulnerabilities and{" "}
                {deleteTarget?.stats?.total_runs ?? 0} runs. This cannot be undone.
              </span>
            </span>
          </label>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)} disabled={busy}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={() => void handleDelete()} disabled={busy}>
              {purge ? "Delete org and data" : "Delete org"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {assignTarget && (
        <AssignWorkspacesDialog
          org={assignTarget}
          allWorkspaces={workspaceNames}
          onClose={() => setAssignTarget(null)}
          onAssigned={() => void handleAssigned()}
        />
      )}
    </div>
  );
}

/** Strip the `status:` prefix the http layer prepends to error messages. */
function cleanError(err: unknown, fallback: string): string {
  if (!(err instanceof Error)) return fallback;
  const match = err.message.match(/^\d+:(.*)$/);
  return (match ? match[1] : err.message) || fallback;
}
