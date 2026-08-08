"use client";

import * as React from "react";
import Link from "next/link";
import { useOrg } from "@/providers/org-provider";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";
import { BuildingIcon, CheckIcon, GlobeIcon, SettingsIcon } from "lucide-react";

/**
 * Global org scope selector.
 *
 * Picking an org here filters assets, vulnerabilities, scans and workspaces
 * across every workspace in that org — the UI equivalent of `osmedeus org use`.
 * "All orgs" clears the filter.
 */
export function OrgSwitcher({ className }: { className?: string }) {
  const { orgs, activeOrg, isLoading, error, selectOrg } = useOrg();

  // A server that has no orgs beyond the built-in default gains nothing from a
  // scope selector, and neither does one too old to expose the endpoint.
  const hasSelectableOrgs = orgs.some((o) => !o.is_default);
  if (error || (!isLoading && !hasSelectableOrgs && !activeOrg)) {
    return null;
  }

  const label = activeOrg ? activeOrg.name : "All orgs";

  return (
    <DropdownMenu>
      <Tooltip>
        <TooltipTrigger asChild>
          <DropdownMenuTrigger asChild>
            <Button
              variant="outline"
              size="sm"
              aria-label="Select org scope"
              className={`h-8 gap-2 ${className ?? ""}`}
            >
              {activeOrg ? (
                <BuildingIcon className="size-3.5 text-primary" />
              ) : (
                <GlobeIcon className="size-3.5 text-muted-foreground" />
              )}
              <span className="max-w-[10rem] truncate text-xs">{label}</span>
            </Button>
          </DropdownMenuTrigger>
        </TooltipTrigger>
        <TooltipContent>
          {activeOrg
            ? `Scoped to ${activeOrg.name} — assets, findings and scans span its workspaces`
            : "Showing data from every org"}
        </TooltipContent>
      </Tooltip>

      <DropdownMenuContent align="end" className="w-64">
        <DropdownMenuLabel>Org scope</DropdownMenuLabel>
        <DropdownMenuSeparator />

        <DropdownMenuItem onSelect={() => selectOrg(null)} className="gap-2">
          <GlobeIcon className="size-4 text-muted-foreground" />
          <span className="flex-1">All orgs</span>
          {!activeOrg && <CheckIcon className="size-4 text-primary" />}
        </DropdownMenuItem>

        {orgs.length > 0 && <DropdownMenuSeparator />}

        {orgs.map((org) => (
          <DropdownMenuItem
            key={org.uuid}
            onSelect={() => selectOrg({ uuid: org.uuid, name: org.name })}
            className="gap-2"
          >
            <BuildingIcon className="size-4 text-muted-foreground" />
            <span className="flex-1 truncate">{org.name}</span>
            {org.stats && (
              <Badge variant="secondary" className="px-1.5 text-[10px]">
                {org.stats.total_workspaces}
              </Badge>
            )}
            {activeOrg?.uuid === org.uuid && <CheckIcon className="size-4 text-primary" />}
          </DropdownMenuItem>
        ))}

        <DropdownMenuSeparator />
        <DropdownMenuItem asChild>
          <Link href="/inventory/orgs" className="gap-2">
            <SettingsIcon className="size-4 text-muted-foreground" />
            <span>Manage orgs</span>
          </Link>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
