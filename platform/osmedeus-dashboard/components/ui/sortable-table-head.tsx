"use client";

import * as React from "react";
import { TableHead } from "@/components/ui/table";
import { ArrowUpIcon, ArrowDownIcon, ArrowUpDownIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import type { SortDirection } from "@/lib/types/asset";

interface SortableTableHeadProps {
  children: React.ReactNode;
  field: string;
  currentSort: { field: string | null; direction: SortDirection };
  onSort: (field: string) => void;
  className?: string;
}

export function SortableTableHead({
  children,
  field,
  currentSort,
  onSort,
  className,
}: SortableTableHeadProps) {
  const isActive = currentSort.field === field;
  const contentAlignClass = React.useMemo(() => {
    if (!className) return "justify-start";
    if (className.includes("text-center")) return "justify-center";
    if (className.includes("text-right")) return "justify-end";
    return "justify-start";
  }, [className]);

  return (
    <TableHead
      className={cn(
        "cursor-pointer select-none transition-colors duration-150 hover:text-foreground",
        isActive && "text-foreground",
        className
      )}
      aria-sort={
        isActive
          ? currentSort.direction === "asc"
            ? "ascending"
            : "descending"
          : "none"
      }
      onClick={() => onSort(field)}
    >
      <div className={cn("flex items-center gap-1", contentAlignClass)}>
        {children}
        {isActive ? (
          currentSort.direction === "asc" ? (
            <ArrowUpIcon className="size-3" />
          ) : (
            <ArrowDownIcon className="size-3" />
          )
        ) : (
          <ArrowUpDownIcon className="size-3 opacity-50" />
        )}
      </div>
    </TableHead>
  );
}
