import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

/**
 * Every tone is a soft fill plus a mark that reads on it. The fills are tuned
 * to roughly 1.2:1 against the row a chip sits in, which is quiet enough to
 * stack a dozen of them in a table without the page turning into a christmas
 * tree. `default` is the only solid brand variant — it spends the accent
 * budget, so reserve it for the one thing on screen that deserves it.
 */
const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-pill border border-transparent px-2.5 py-[3px] text-xs leading-[1.5] whitespace-nowrap transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-ring",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground",
        secondary: "bg-secondary text-secondary-foreground",
        destructive: "bg-destructive-soft text-destructive",
        outline: "border-border text-muted-foreground",
        success: "bg-success-soft text-success",
        warning: "bg-warning-soft text-warning",
        info: "bg-info-soft text-info",
        purple: "bg-purple-soft text-purple",
        pink: "bg-pink-soft text-pink",
        cyan: "bg-cyan-soft text-cyan",
        orange: "bg-orange-soft text-orange",
        /** Accent-tinted chip — the quiet counterpart to `default`. */
        accent: "bg-primary-soft text-primary-soft-fg",
        /** Inverted chip, for the rare label that has to out-shout the rest. */
        solid: "bg-foreground text-background",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

function Badge({
  className,
  variant,
  ...props
}: React.ComponentProps<"div"> & VariantProps<typeof badgeVariants>) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
