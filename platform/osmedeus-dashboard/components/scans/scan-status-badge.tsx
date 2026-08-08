import * as React from "react";
import { Badge } from "@/components/ui/badge";
import type { ScanStatus } from "@/lib/types/scan";
import {
  CheckCircleIcon,
  XCircleIcon,
  LoaderIcon,
  ClockIcon,
  BanIcon,
} from "lucide-react";

// `running` takes the accent-tinted chip rather than the solid one: a table of
// scans can have a dozen of these at once, and a dozen solid primary fills
// spend the whole accent budget on status.
const statusConfig: Record<
  ScanStatus,
  {
    label: string;
    variant: "accent" | "secondary" | "destructive" | "outline" | "success" | "warning";
    icon: React.ElementType;
  }
> = {
  completed: { label: "Completed", variant: "success", icon: CheckCircleIcon },
  running: { label: "Running", variant: "accent", icon: LoaderIcon },
  pending: { label: "Pending", variant: "secondary", icon: ClockIcon },
  failed: { label: "Failed", variant: "destructive", icon: XCircleIcon },
  cancelled: { label: "Cancelled", variant: "outline", icon: BanIcon },
};

interface ScanStatusBadgeProps {
  status: ScanStatus;
  showIcon?: boolean;
}

export function ScanStatusBadge({ status, showIcon = true }: ScanStatusBadgeProps) {
  const config = statusConfig[status];
  const Icon = config.icon;

  return (
    <Badge variant={config.variant} className="gap-1">
      {showIcon && (
        <Icon
          className={`size-3 ${status === "running" ? "animate-spin" : ""}`}
        />
      )}
      {config.label}
    </Badge>
  );
}
