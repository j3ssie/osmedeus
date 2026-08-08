"use client";

import * as React from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { cn } from "@/lib/utils";

/**
 * Scanner descriptions arrive as markdown — headings, lists, links, fenced
 * payloads — so rendering them as plain text leaves `##` and `**` on screen.
 * The component map keeps them on the app's type ramp instead of the browser
 * defaults, which are far too large next to a 12.5px UI.
 *
 * `remark-gfm` is not optional here: probe results come as pipe tables, and
 * CommonMark has no table syntax at all — without it those descriptions render
 * as a wall of raw `|` characters.
 */
export function Markdown({
  children,
  className,
}: {
  children: string;
  className?: string;
}) {
  return (
    <div className={cn("space-y-2.5 text-sm text-body", className)}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          /*
           * Headings stay in the case they were authored in. The uppercase
           * eyebrow the rest of the app uses for section labels cannot be
           * reused here: scanner headings carry URL paths and payloads, and
           * upper-casing those destroys information the reader needs.
           */
          h1: (props) => (
            <h1
              {...props}
              className="mt-4 text-md font-semibold text-ink [overflow-wrap:anywhere] first:mt-0"
            />
          ),
          h2: (props) => (
            <h2
              {...props}
              className="mt-4 text-sm font-semibold text-ink [overflow-wrap:anywhere] first:mt-0"
            />
          ),
          h3: (props) => (
            <h3
              {...props}
              className="mt-3 text-xs font-semibold text-body [overflow-wrap:anywhere] first:mt-0"
            />
          ),
          p: (props) => <p {...props} className="leading-relaxed" />,
          ul: (props) => <ul {...props} className="list-disc space-y-1 pl-5" />,
          ol: (props) => <ol {...props} className="list-decimal space-y-1 pl-5" />,
          li: (props) => <li {...props} className="leading-relaxed marker:text-faint" />,
          strong: (props) => <strong {...props} className="font-semibold text-ink" />,
          a: (props) => (
            <a
              {...props}
              className="text-primary underline underline-offset-4 [overflow-wrap:anywhere]"
              target="_blank"
              rel="noreferrer"
            />
          ),
          /*
           * `overflow-wrap: anywhere` rather than `break-all`: a payload only
           * breaks when it genuinely cannot fit, instead of greedily splitting
           * every token so `!STANDARD` lands as `!STANDAR` / `D`.
           */
          code: (props) => (
            <code
              {...props}
              className="rounded bg-sunken px-1 py-0.5 font-mono text-xs [overflow-wrap:anywhere]"
            />
          ),
          pre: (props) => (
            <pre
              {...props}
              className="overflow-x-auto rounded-control border border-border-subtle bg-sunken p-3 font-mono text-xs [&_code]:bg-transparent [&_code]:p-0"
            />
          ),
          blockquote: (props) => (
            <blockquote
              {...props}
              className="border-l-2 border-border pl-3 text-muted-foreground"
            />
          ),
          hr: () => <hr className="border-border-subtle" />,
          /*
           * A probe table runs to nine columns of payloads and status codes.
           * Wrapping the cells to fit the panel turns it into mush, so the
           * table keeps its natural width (`min-w-max`, cells `nowrap`) and the
           * box scrolls sideways instead — the row stays one readable line.
           */
          table: (props) => (
            <div className="overflow-x-auto rounded-control border border-border">
              <table {...props} className="w-full min-w-max border-collapse text-xs" />
            </div>
          ),
          thead: (props) => (
            <thead {...props} className="border-b border-border bg-sunken" />
          ),
          tbody: (props) => (
            <tbody {...props} className="divide-y divide-border-subtle" />
          ),
          tr: (props) => <tr {...props} className="hover:bg-muted/60" />,
          th: (props) => (
            <th
              {...props}
              className="whitespace-nowrap px-2.5 py-1.5 text-left text-2xs font-medium uppercase tracking-label text-faint"
            />
          ),
          td: (props) => (
            <td
              {...props}
              className="whitespace-nowrap px-2.5 py-1.5 align-top [&_code]:bg-transparent [&_code]:px-0"
            />
          ),
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  );
}
