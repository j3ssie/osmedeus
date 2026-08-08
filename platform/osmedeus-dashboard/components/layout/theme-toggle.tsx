"use client";

import * as React from "react";
import { useTheme } from "next-themes";
import { MoonIcon, SunIcon } from "lucide-react";
import { Button } from "@/components/ui/button";

type Variant = "default" | "outline" | "ghost";
type Size = "sm" | "icon";

/** Matches the transition duration declared for `[data-theme-transition]`. */
const THEME_TRANSITION_MS = 220;

export function ThemeToggle({
  variant = "ghost",
  size = "icon",
  className,
  ariaLabel = "Toggle theme",
  label,
}: {
  variant?: Variant;
  size?: Size;
  className?: string;
  ariaLabel?: string;
  label?: string;
}) {
  // `resolvedTheme`, not `theme`: until the user picks explicitly, `theme` is
  // "system", so a `theme === "dark"` test reads false on a dark page and the
  // first click would set the theme it is already showing — a no-op.
  const { setTheme, resolvedTheme } = useTheme();
  const [mounted, setMounted] = React.useState(false);
  const transitionTimer = React.useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  React.useEffect(() => {
    setMounted(true);
    return () => clearTimeout(transitionTimer.current);
  }, []);

  // Cross-fade the palette swap. The attribute is only set for the length of
  // the transition — a permanent global colour transition would fight every
  // other animation on the page.
  const toggle = React.useCallback(() => {
    const root = document.documentElement;
    root.setAttribute("data-theme-transition", "");
    clearTimeout(transitionTimer.current);
    transitionTimer.current = setTimeout(
      () => root.removeAttribute("data-theme-transition"),
      THEME_TRANSITION_MS
    );
    setTheme(resolvedTheme === "dark" ? "light" : "dark");
  }, [setTheme, resolvedTheme]);

  if (!mounted) {
    return (
      <Button variant={variant} size={size} className={className} aria-label={ariaLabel}>
        <SunIcon className="size-4" />
        {label ? <span className="text-xs ml-2">{label}</span> : null}
      </Button>
    );
  }

  return (
    <Button
      variant={variant}
      size={size}
      className={className}
      aria-label={ariaLabel}
      onClick={toggle}
    >
      {resolvedTheme === "dark" ? (
        <MoonIcon className="size-4" />
      ) : (
        <SunIcon className="size-4" />
      )}
      {label ? <span className="text-xs ml-2">{label}</span> : null}
    </Button>
  );
}
