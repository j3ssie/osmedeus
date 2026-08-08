import * as React from "react";
import { cn } from "@/lib/utils";
import { Card } from "@/components/ui/card";

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ElementType;
  description?: string;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  className?: string;
}

export function StatCard({
  title,
  value,
  icon: Icon,
  description,
  trend,
  className,
}: StatCardProps) {
  return (
    <Card className={cn("glow-card gap-0 p-5", className)}>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          {/* Eyebrow label, figure, foot — the number is the only thing on the
              tile allowed to be large, so a row of tiles scans as a row of
              numbers rather than a row of headings. */}
          <p className="text-2xs uppercase tracking-label text-faint">{title}</p>
          <div className="mt-1.5 flex items-baseline gap-2">
            <p className="font-figure text-3xl leading-[1.1] text-foreground">{value}</p>
            {trend && (
              <span
                className={cn(
                  "text-xs",
                  trend.isPositive ? "text-success" : "text-destructive"
                )}
              >
                {trend.isPositive ? "+" : "-"}{Math.abs(trend.value)}%
              </span>
            )}
          </div>
          {description && (
            <p className="mt-1 text-xs text-muted-foreground">{description}</p>
          )}
        </div>
        {/* The icon is wayfinding, not data — kept neutral so the accent budget
            stays with the one primary action on the page. */}
        <div className="flex size-9 shrink-0 items-center justify-center rounded-control bg-muted text-muted-foreground">
          <Icon className="size-4" />
        </div>
      </div>
    </Card>
  );
}
