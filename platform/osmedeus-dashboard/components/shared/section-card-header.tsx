import * as React from "react";
import {
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface SectionCardHeaderProps {
  title: React.ReactNode;
  description?: React.ReactNode;
  icon?: React.ElementType;
  actions?: React.ReactNode;
  className?: string;
}

/**
 * Canonical header for a content section Card: a faintly tinted band closed by
 * a subtle rule, with a compact title, optional muted icon + description, and
 * right-aligned actions. Use this for every page-level section so headers stay
 * consistent. The tint is the sunken plane at half strength — enough to read as
 * a band against the raised card without becoming a second surface.
 *
 * It assumes it is the Card's first child: the `-mt-3` cancels the Card's top
 * padding so the band sits flush against the top edge.
 */
export function SectionCardHeader({
  title,
  description,
  icon: Icon,
  actions,
  className,
}: SectionCardHeaderProps) {
  return (
    <CardHeader className={cn("-mt-3 border-b border-border-subtle bg-muted/50 py-2", className)}>
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-0.5">
          <CardTitle className="flex items-center gap-2 text-md font-medium">
            {Icon ? <Icon className="size-4 text-muted-foreground" /> : null}
            {title}
          </CardTitle>
          {description ? (
            <CardDescription>{description}</CardDescription>
          ) : null}
        </div>
        {actions ? (
          <div className="flex flex-wrap items-center gap-3 sm:justify-end">
            {actions}
          </div>
        ) : null}
      </div>
    </CardHeader>
  );
}
