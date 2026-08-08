import * as React from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatNumber, timeAgo } from "@/lib/utils";
import type { Workspace } from "@/lib/types/asset";
import {
  GlobeIcon,
  LinkIcon,
  ShieldAlertIcon,
  ArrowRightIcon,
  TagIcon,
  ClockIcon,
  LayersIcon,
  ServerIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface WorkspaceCardProps {
  workspace: Workspace;
  compact?: boolean;
}

function RiskScoreBadge({ score }: { score: number }) {
  const getRiskColor = (s: number) => {
    if (s >= 8) return "bg-destructive-soft text-destructive border-transparent";
    if (s >= 6) return "bg-orange-soft text-orange border-transparent";
    if (s >= 4) return "bg-warning-soft text-warning border-transparent";
    return "bg-success-soft text-success border-transparent";
  };

  return (
    <Badge variant="outline" className={cn("font-mono text-xs", getRiskColor(score))}>
      {score.toFixed(1)}
    </Badge>
  );
}

function VulnBadge({
  count,
  severity,
}: {
  count: number;
  severity: "critical" | "high" | "medium" | "low";
}) {
  if (count === 0) return null;

  const colors = {
    critical: "bg-destructive-soft text-destructive",
    high: "bg-orange-soft text-orange",
    medium: "bg-warning-soft text-warning",
    low: "bg-info-soft text-info",
  };

  const labels = {
    critical: "C",
    high: "H",
    medium: "M",
    low: "L",
  };

  return (
    <span
      className={cn(
        "inline-flex items-center justify-center min-w-[20px] h-5 px-1.5 rounded text-xs font-medium",
        colors[severity]
      )}
      title={`${count} ${severity}`}
    >
      {count}
    </span>
  );
}

export function WorkspaceCard({ workspace, compact = false }: WorkspaceCardProps) {
  const hasVulns =
    workspace.vuln_critical > 0 ||
    workspace.vuln_high > 0 ||
    workspace.vuln_medium > 0 ||
    workspace.vuln_low > 0;

  if (compact) {
    return (
      <Card className="hover:shadow-md transition-shadow">
        <CardContent className="p-4">
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-3 min-w-0 flex-1">
              <div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-muted">
                <ServerIcon className="size-5 text-muted-foreground" />
              </div>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <h3 className="font-medium truncate">{workspace.name}</h3>
                  <RiskScoreBadge score={workspace.risk_score} />
                </div>
                <div className="flex items-center gap-3 text-sm text-muted-foreground mt-0.5">
                  <span className="flex items-center gap-1">
                    <GlobeIcon className="size-3" />
                    {formatNumber(workspace.total_subdomains)}
                  </span>
                  <span className="flex items-center gap-1">
                    <LinkIcon className="size-3" />
                    {formatNumber(workspace.total_urls)}
                  </span>
                  {hasVulns && (
                    <span className="flex items-center gap-1">
                      <ShieldAlertIcon className="size-3" />
                      {workspace.total_vulns}
                    </span>
                  )}
                </div>
              </div>
            </div>

            <div className="flex items-center gap-2">
              {hasVulns && (
                <div className="flex items-center gap-1">
                  <VulnBadge count={workspace.vuln_critical} severity="critical" />
                  <VulnBadge count={workspace.vuln_high} severity="high" />
                  <VulnBadge count={workspace.vuln_medium} severity="medium" />
                  <VulnBadge count={workspace.vuln_low} severity="low" />
                </div>
              )}
              <Button variant="ghost" size="sm" asChild>
                <Link href={`/inventory/workspaces/${workspace.name}`}>
                  <ArrowRightIcon className="size-4" />
                </Link>
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0 flex-1">
            <CardTitle className="text-lg truncate">{workspace.name}</CardTitle>
            {workspace.tags.length > 0 && (
              <div className="flex items-center gap-1 mt-1 flex-wrap">
                {workspace.tags.slice(0, 3).map((tag) => (
                  <Badge key={tag} variant="secondary" className="text-xs">
                    <TagIcon className="size-2.5 mr-1" />
                    {tag}
                  </Badge>
                ))}
                {workspace.tags.length > 3 && (
                  <Badge variant="secondary" className="text-xs">
                    +{workspace.tags.length - 3}
                  </Badge>
                )}
              </div>
            )}
          </div>
          <RiskScoreBadge score={workspace.risk_score} />
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Vulnerability Summary */}
        {hasVulns && (
          <div className="flex items-center justify-between p-2 rounded-lg bg-muted/50">
            <div className="flex items-center gap-1.5 text-sm">
              <ShieldAlertIcon className="size-4 text-muted-foreground" />
              <span className="text-muted-foreground">{workspace.total_vulns} vulnerabilities</span>
            </div>
            <div className="flex items-center gap-1">
              <VulnBadge count={workspace.vuln_critical} severity="critical" />
              <VulnBadge count={workspace.vuln_high} severity="high" />
              <VulnBadge count={workspace.vuln_medium} severity="medium" />
              <VulnBadge count={workspace.vuln_low} severity="low" />
            </div>
          </div>
        )}

        {/* Stats Grid */}
        <div className="grid grid-cols-3 gap-3">
          <div className="flex flex-col items-center p-2 rounded-lg bg-muted/30">
            <div className="flex items-center gap-1 text-muted-foreground">
              <GlobeIcon className="size-3.5" />
            </div>
            <p className="text-lg font-semibold">{formatNumber(workspace.total_subdomains)}</p>
            <p className="text-xs text-muted-foreground">Subdomains</p>
          </div>
          <div className="flex flex-col items-center p-2 rounded-lg bg-muted/30">
            <div className="flex items-center gap-1 text-muted-foreground">
              <LinkIcon className="size-3.5" />
            </div>
            <p className="text-lg font-semibold">{formatNumber(workspace.total_urls)}</p>
            <p className="text-xs text-muted-foreground">URLs</p>
          </div>
          <div className="flex flex-col items-center p-2 rounded-lg bg-muted/30">
            <div className="flex items-center gap-1 text-muted-foreground">
              <LayersIcon className="size-3.5" />
            </div>
            <p className="text-lg font-semibold">{formatNumber(workspace.total_assets)}</p>
            <p className="text-xs text-muted-foreground">Assets</p>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between pt-2 border-t">
          <div className="text-xs text-muted-foreground">
            {workspace.last_run ? (
              <span className="flex items-center gap-1">
                <ClockIcon className="size-3" />
                {timeAgo(new Date(workspace.last_run))}
                {workspace.run_workflow && (
                  <span className="text-foreground/70">• {workspace.run_workflow}</span>
                )}
              </span>
            ) : (
              "Never scanned"
            )}
          </div>
          <Button variant="ghost" size="sm" asChild>
            <Link href={`/inventory/workspaces/${workspace.name}`}>
              View Details
              <ArrowRightIcon className="ml-1 size-4" />
            </Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
