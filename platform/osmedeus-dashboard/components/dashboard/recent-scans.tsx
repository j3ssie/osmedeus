"use client";

import * as React from "react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { EmptyState } from "@/components/shared/empty-state";
import { ErrorState } from "@/components/shared/error-state";
import { fetchRecentScans } from "@/lib/api/scans";
import { timeAgo } from "@/lib/utils";
import type { Scan, ScanStatus } from "@/lib/types/scan";
import {
  ArrowRightIcon,
  ScanSearchIcon,
  CheckCircleIcon,
  XCircleIcon,
  LoaderIcon,
  ClockIcon,
  BanIcon,
} from "lucide-react";

const statusConfig: Record<
  ScanStatus,
  { label: string; variant: "default" | "secondary" | "destructive" | "outline" | "success" | "warning"; icon: React.ElementType }
> = {
  completed: { label: "Completed", variant: "success", icon: CheckCircleIcon },
  running: { label: "Running", variant: "default", icon: LoaderIcon },
  pending: { label: "Pending", variant: "secondary", icon: ClockIcon },
  failed: { label: "Failed", variant: "destructive", icon: XCircleIcon },
  cancelled: { label: "Cancelled", variant: "outline", icon: BanIcon },
};

export function RecentScans() {
  const [scans, setScans] = React.useState<Scan[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  const loadScans = React.useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await fetchRecentScans(5);
      setScans(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load scans");
    } finally {
      setIsLoading(false);
    }
  }, []);

  React.useEffect(() => {
    loadScans();
  }, [loadScans]);

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="text-lg">Recent Scans</CardTitle>
        <Button variant="ghost" size="sm" asChild>
          <Link href="/scans">
            View all
            <ArrowRightIcon className="ml-1 size-4" />
          </Link>
        </Button>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center gap-4">
                <Skeleton className="size-9 rounded-control" />
                <div className="flex-1 space-y-2">
                  <Skeleton className="h-4 w-1/3" />
                  <Skeleton className="h-3 w-1/2" />
                </div>
                <Skeleton className="h-6 w-20" />
              </div>
            ))}
          </div>
        ) : error ? (
          <ErrorState message={error} onRetry={loadScans} />
        ) : scans.length === 0 ? (
          <EmptyState
            icon={ScanSearchIcon}
            title="No scans yet"
            description="Start your first security scan to see results here."
            action={{
              label: "New Scan",
              onClick: () => (window.location.href = "/scans/new"),
            }}
          />
        ) : (
          <div className="space-y-4">
            {scans.map((scan) => {
              const status = statusConfig[scan.status];
              const StatusIcon = status.icon;
              return (
                <Link
                  key={scan.id}
                  href={`/scans`}
                  className="-mx-2 flex items-center gap-4 rounded-control p-2 transition-colors duration-150 hover:bg-sunken"
                >
                  <div className="flex size-9 items-center justify-center rounded-control bg-muted">
                    <StatusIcon
                      className={`size-5 ${
                        scan.status === "running" ? "animate-spin" : ""
                      } ${
                        scan.status === "completed"
                          ? "text-success"
                          : scan.status === "failed"
                          ? "text-destructive"
                          : "text-muted-foreground"
                      }`}
                    />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{scan.target}</p>
                    <p className="text-sm text-muted-foreground truncate">
                      {scan.workflowName} &middot; {scan.startedAt ? timeAgo(scan.startedAt) : "Pending"}
                    </p>
                  </div>
                  <Badge variant={status.variant}>{status.label}</Badge>
                </Link>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
