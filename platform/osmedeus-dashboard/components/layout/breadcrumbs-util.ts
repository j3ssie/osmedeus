import {
  BarChartBigIcon,
  FolderOpenIcon,
  FlowerIcon,
  ScanSearchIcon,
  SettingsIcon,
  DatabaseIcon,
  ClipboardCheckIcon,
  ScrollText as ScrollTextIcon,
  SquareFunction as SquareFunctionIcon,
  Package as PackageIcon,
  ShieldAlertIcon,
  BrainIcon,
} from "lucide-react";
import type { ComponentType } from "react";

export type Crumb = {
  href?: string;
  label: string;
  icon?: ComponentType<{ className?: string }>;
  isCurrent?: boolean;
};

function titleCase(input: string) {
  return input
    .replace(/[-_]+/g, " ")
    .split(" ")
    .filter(Boolean)
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
}

function segmentIcon(segment: string): ComponentType<{ className?: string }> | undefined {
  if (segment === "") return BarChartBigIcon;
  if (segment === "assets") return FolderOpenIcon;
  if (segment === "inventory") return DatabaseIcon;
  if (segment === "workflows") return FlowerIcon;
  if (segment === "scans") return ScanSearchIcon;
  if (segment === "settings") return SettingsIcon;
  if (segment === "schedules") return ClipboardCheckIcon;
  if (segment === "events") return ScrollTextIcon;
  if (segment === "vulnerabilities") return ShieldAlertIcon;
  if (segment === "utilities") return SquareFunctionIcon;
  if (segment === "registry") return PackageIcon;
  if (segment === "llm") return BrainIcon;
  return undefined;
}

export function getBreadcrumbs(pathname: string): Crumb[] {
  const path = pathname || "/";
  if (path === "/") {
    return [{ label: "Dashboard", icon: BarChartBigIcon, isCurrent: true }];
  }
  const parts = path.replace(/^\/+/, "").split("/");
  if (path.startsWith("/inventory")) {
    const icon = segmentIcon("inventory");
    const last = parts[parts.length - 1] || "Inventory";
    return [{ label: titleCase(last), icon, isCurrent: true }];
  }
  const crumbs: Crumb[] = [];
  let acc = "";
  parts.forEach((part, idx) => {
    acc += `/${part}`;
    const isLast = idx === parts.length - 1;
    const icon = idx === 0 ? segmentIcon(part) : undefined;
    const label = isLast ? titleCase(part) : titleCase(part);
    crumbs.push({
      href: isLast ? undefined : acc,
      label,
      icon: isLast ? segmentIcon(parts[0]) : icon,
      isCurrent: isLast,
    });
  });
  return crumbs;
}
