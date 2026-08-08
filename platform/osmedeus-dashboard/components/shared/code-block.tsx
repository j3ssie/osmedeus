"use client";

import * as React from "react";
import { CheckIcon, CopyIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface CodeBlockProps {
  value: string;
  /** Cap on the rendered height before the block scrolls on its own. */
  maxHeight?: number;
  /**
   * Wrap long lines instead of scrolling sideways. On for payloads and URLs,
   * off for HTTP traffic where the line structure is the point.
   */
  wrap?: boolean;
  className?: string;
}

/**
 * Fixed-width evidence panel: raw text, its own scroll box, and a copy button
 * that only appears on hover so a stack of them stays quiet. Payloads here are
 * attacker-controlled, so the value is never interpolated as markup.
 */
export function CodeBlock({
  value,
  maxHeight = 300,
  wrap = false,
  className,
}: CodeBlockProps) {
  const [copied, setCopied] = React.useState(false);
  const timer = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  React.useEffect(
    () => () => {
      if (timer.current) clearTimeout(timer.current);
    },
    []
  );

  const copy = React.useCallback(async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      if (timer.current) clearTimeout(timer.current);
      timer.current = setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard is unavailable outside a secure context; nothing to recover.
    }
  }, [value]);

  return (
    <div
      className={cn(
        "group relative overflow-hidden rounded-control border border-border-subtle bg-sunken",
        className
      )}
    >
      <Button
        type="button"
        variant="ghost"
        size="icon-xs"
        aria-label={copied ? "Copied" : "Copy to clipboard"}
        onClick={copy}
        className="absolute right-1.5 top-1.5 z-10 bg-raised/80 opacity-0 backdrop-blur-sm transition-opacity group-hover:opacity-100 focus-visible:opacity-100"
      >
        {copied ? (
          <CheckIcon className="size-3.5 text-success" />
        ) : (
          <CopyIcon className="size-3.5" />
        )}
      </Button>
      <pre
        className={cn(
          "overflow-auto p-3 pr-10 font-mono text-xs leading-relaxed text-body",
          wrap ? "whitespace-pre-wrap break-all" : "whitespace-pre"
        )}
        style={{ maxHeight }}
      >
        {value}
      </pre>
    </div>
  );
}
