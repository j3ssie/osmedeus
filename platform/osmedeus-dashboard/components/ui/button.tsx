import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

/**
 * Every tone carries a border, so the outline and ghost variants sit on the
 * same optical baseline as the filled ones and a row of mixed buttons doesn't
 * jitter. Hover is a hairline glow derived from `--primary` rather than a
 * background shift — filled buttons have nowhere to shift to, and a shared
 * hover cue keeps the row reading as one control group.
 */
const buttonVariants = cva(
  "inline-flex shrink-0 cursor-pointer items-center justify-center gap-2 whitespace-nowrap rounded-control border text-sm font-medium leading-[1.5] outline-none transition-[color,background-color,border-color,box-shadow,translate] duration-150 hover:shadow-glow active:translate-y-px disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4 focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40",
  {
    variants: {
      variant: {
        default: "border-primary bg-primary text-primary-foreground",
        destructive:
          "border-destructive bg-destructive text-destructive-foreground focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40",
        outline:
          "border-input bg-transparent text-muted-foreground hover:text-foreground",
        secondary:
          "border-secondary bg-secondary text-secondary-foreground hover:border-border",
        ghost:
          "border-transparent bg-transparent text-muted-foreground hover:text-foreground",
        link: "border-transparent text-primary underline-offset-4 hover:underline hover:shadow-none active:translate-y-0",
      },
      size: {
        default: "h-9 px-4 py-2 has-[>svg]:px-3",
        sm: "h-8 gap-1.5 px-3 text-xs has-[>svg]:px-2.5",
        lg: "h-10 px-6 has-[>svg]:px-4",
        icon: "size-9",
        "icon-xs": "size-7",
        "icon-sm": "size-8",
        "icon-lg": "size-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

function Button({
  className,
  variant = "default",
  size = "default",
  asChild = false,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean
  }) {
  const Comp = asChild ? Slot : "button"

  return (
    <Comp
      data-slot="button"
      data-variant={variant}
      data-size={size}
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    />
  )
}

export { Button, buttonVariants }
