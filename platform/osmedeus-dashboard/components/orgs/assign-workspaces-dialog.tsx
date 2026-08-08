"use client";

import * as React from "react";
import { assignWorkspaces } from "@/lib/api/orgs";
import type { Org } from "@/lib/types/org";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { SearchIcon } from "lucide-react";
import { toast } from "sonner";

interface Props {
  org: Org;
  allWorkspaces: string[];
  onClose: () => void;
  onAssigned: () => void;
}

/**
 * Assign workspaces to an org.
 *
 * Assignment cascades server-side to every asset, vulnerability and run in the
 * selected workspaces, which is what lets pre-existing scan data be grouped
 * without re-scanning.
 *
 * Note this only ever adds: unchecking a workspace here does not move it out of
 * the org, because the API assigns the given list rather than replacing the
 * membership set. To move a workspace elsewhere, assign it to the other org.
 */
export function AssignWorkspacesDialog({ org, allWorkspaces, onClose, onAssigned }: Props) {
  const alreadyIn = React.useMemo(
    () => new Set(org.stats?.workspaces ?? []),
    [org.stats?.workspaces]
  );

  const [selected, setSelected] = React.useState<Set<string>>(new Set());
  const [search, setSearch] = React.useState("");
  const [busy, setBusy] = React.useState(false);

  const visible = React.useMemo(() => {
    const q = search.trim().toLowerCase();
    const names = q ? allWorkspaces.filter((n) => n.toLowerCase().includes(q)) : allWorkspaces;
    // Workspaces not yet in this org come first — they are the actionable ones.
    return [...names].sort((a, b) => {
      const aIn = alreadyIn.has(a) ? 1 : 0;
      const bIn = alreadyIn.has(b) ? 1 : 0;
      return aIn - bIn || a.localeCompare(b);
    });
  }, [allWorkspaces, search, alreadyIn]);

  const toggle = (name: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(name)) next.delete(name);
      else next.add(name);
      return next;
    });
  };

  const handleAssign = async () => {
    if (selected.size === 0) return;
    setBusy(true);
    try {
      const counts = await assignWorkspaces(org.uuid, Array.from(selected));
      const parts = [
        `${counts.workspaces ?? 0} workspaces`,
        `${counts.assets ?? 0} assets`,
        `${counts.vulnerabilities ?? 0} findings`,
        `${counts.runs ?? 0} runs`,
      ];
      toast.success(`Assigned to ${org.name}`, { description: parts.join(" · ") });
      onAssigned();
    } catch (err) {
      const msg = err instanceof Error ? err.message.replace(/^\d+:/, "") : "Assignment failed";
      toast.error(msg);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Assign workspaces to {org.name}</DialogTitle>
          <DialogDescription>
            Their assets, vulnerabilities and runs move into the org too, so existing scan data is
            grouped without re-scanning.
          </DialogDescription>
        </DialogHeader>

        <div className="relative">
          <SearchIcon className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Filter workspaces..."
            className="pl-8"
            autoFocus
          />
        </div>

        <ScrollArea className="h-72 rounded-control border border-border">
          {visible.length === 0 ? (
            <p className="p-4 text-center text-sm text-muted-foreground">
              {allWorkspaces.length === 0 ? "No workspaces found." : "No workspaces match."}
            </p>
          ) : (
            <ul className="divide-y divide-border">
              {visible.map((name) => {
                const inOrg = alreadyIn.has(name);
                return (
                  <li key={name}>
                    <label className="flex cursor-pointer items-center gap-3 px-3 py-2 hover:bg-muted">
                      <Checkbox
                        checked={inOrg || selected.has(name)}
                        disabled={inOrg}
                        onCheckedChange={() => toggle(name)}
                      />
                      <span className="flex-1 truncate text-sm">{name}</span>
                      {inOrg && (
                        <Badge variant="secondary" className="text-[10px]">
                          in org
                        </Badge>
                      )}
                    </label>
                  </li>
                );
              })}
            </ul>
          )}
        </ScrollArea>

        <DialogFooter className="items-center sm:justify-between">
          <span className="text-xs text-muted-foreground">
            {selected.size === 0 ? "Nothing selected" : `${selected.size} selected`}
          </span>
          <div className="flex gap-2">
            <Button variant="outline" onClick={onClose} disabled={busy}>
              Cancel
            </Button>
            <Button onClick={() => void handleAssign()} disabled={busy || selected.size === 0}>
              Assign
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
