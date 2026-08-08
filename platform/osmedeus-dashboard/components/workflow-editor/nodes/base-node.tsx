"use client";

import * as React from "react";
import { Handle, Position } from "@xyflow/react";
import { cn } from "@/lib/utils";
import type { WorkflowNodeData } from "@/lib/types/workflow";
import { useCanvasSettings } from "../canvas-settings";

function truncateText(input: string, maxLen: number): string {
  const text = input.trim();
  if (text.length <= maxLen) return text;
  return `${text.slice(0, Math.max(0, maxLen - 1)).trimEnd()}…`;
}

function normalizeInlineText(input: string): string {
  return input.replace(/\s+/g, " ").trim();
}

function isSensitiveKey(key: string): boolean {
  const k = key.toLowerCase();
  return (
    k.includes("password") ||
    k.includes("passwd") ||
    k.includes("secret") ||
    k.includes("token") ||
    k.includes("apikey") ||
    k.includes("api_key") ||
    k.includes("authorization")
  );
}

function stringifyInlineValue(value: unknown): string {
  if (value == null) return "null";
  if (typeof value === "string") return normalizeInlineText(value);
  if (typeof value === "number" || typeof value === "boolean" || typeof value === "bigint") {
    return String(value);
  }
  try {
    return normalizeInlineText(JSON.stringify(value));
  } catch {
    return "";
  }
}

function extractMessageContentText(content: unknown): string {
  if (typeof content === "string") return content;
  if (Array.isArray(content)) {
    const parts: string[] = [];
    for (const item of content) {
      if (typeof item === "string") {
        parts.push(item);
        continue;
      }
      if (!item || typeof item !== "object") continue;
      const anyItem = item as Record<string, unknown>;
      if (anyItem.type === "text" && typeof anyItem.text === "string") {
        parts.push(anyItem.text);
        continue;
      }
      if (typeof anyItem.content === "string") {
        parts.push(anyItem.content);
        continue;
      }
      if (typeof anyItem.input === "string") {
        parts.push(anyItem.input);
        continue;
      }
    }

    if (parts.length > 0) return parts.join(" ");
    try {
      return JSON.stringify(content);
    } catch {
      return "";
    }
  }

  if (content && typeof content === "object") {
    const anyObj = content as Record<string, unknown>;
    if (typeof anyObj.text === "string") return anyObj.text;
    if (typeof anyObj.content === "string") return anyObj.content;
    try {
      return JSON.stringify(content);
    } catch {
      return "";
    }
  }

  return content == null ? "" : String(content);
}

function buildLlmMessagesSummary(step: WorkflowNodeData["step"]): string[] {
  if (!step) return [];

  if (step.is_embedding) {
    const input = Array.isArray(step.embedding_input) ? step.embedding_input : [];
    if (input.length === 0) return [];
    const first = normalizeInlineText(String(input[0] ?? ""));
    const suffix = input.length > 1 ? ` (+${input.length - 1})` : "";
    return [truncateText(`input: ${first}${suffix}`, 140)];
  }

  const messages = Array.isArray(step.messages) ? step.messages : [];
  if (messages.length === 0) return [];

  const maxLines = 3;
  const lines: string[] = [];

  for (const msg of messages.slice(0, maxLines)) {
    if (!msg || typeof msg !== "object") continue;
    const anyMsg = msg as Record<string, unknown>;
    const role = typeof anyMsg.role === "string" ? anyMsg.role : "message";
    const name = typeof anyMsg.name === "string" ? anyMsg.name : "";
    const content = normalizeInlineText(extractMessageContentText(anyMsg.content));
    const roleLabel = name ? `${role}[${name}]` : role;
    const value = content ? truncateText(content, 140) : "";
    lines.push(value ? `${roleLabel}: ${value}` : `${roleLabel}`);
  }

  const remaining = messages.length - maxLines;
  if (remaining > 0) lines.push(`+${remaining} more`);
  return lines;
}

function buildFunctionMessagesSummary(step: WorkflowNodeData["step"]): string[] {
  if (!step) return [];
  if (step.type !== "function") return [];

  const items: string[] = [];

  if (typeof step.function === "string" && step.function.trim()) {
    items.push(`fn: ${normalizeInlineText(step.function)}`);
  }

  const functions = Array.isArray(step.functions) ? step.functions : [];
  for (const fn of functions) {
    if (typeof fn !== "string") continue;
    const text = fn.trim();
    if (!text) continue;
    items.push(`fn: ${normalizeInlineText(text)}`);
  }

  const parallelFunctions = Array.isArray(step.parallel_functions)
    ? step.parallel_functions
    : [];
  for (const fn of parallelFunctions) {
    if (typeof fn !== "string") continue;
    const text = fn.trim();
    if (!text) continue;
    items.push(`pfn: ${normalizeInlineText(text)}`);
  }

  if (items.length === 0) return [];

  const maxLines = 3;
  const lines = items.slice(0, maxLines).map((v) => truncateText(v, 140));
  const remaining = items.length - maxLines;
  if (remaining > 0) lines.push(`+${remaining} more`);
  return lines;
}

function buildHttpSummary(step: WorkflowNodeData["step"]): string[] {
  if (!step) return [];
  if (step.type !== "http") return [];

  const lines: string[] = [];

  const url = typeof step.url === "string" ? step.url.trim() : "";
  const method = typeof step.method === "string" ? step.method.trim() : "";
  if (url) {
    const methodLabel = method ? method.toUpperCase() : "HTTP";
    lines.push(truncateText(`${methodLabel} ${normalizeInlineText(url)}`, 140));
  }

  if (step.headers && typeof step.headers === "object" && !Array.isArray(step.headers)) {
    const keys = Object.keys(step.headers as Record<string, string>).filter(Boolean);
    if (keys.length > 0) {
      const shown = keys.slice(0, 3).join(", ");
      const suffix = keys.length > 3 ? ` (+${keys.length - 3})` : "";
      lines.push(truncateText(`headers: ${shown}${suffix}`, 140));
    }
  }

  const body = typeof step.request_body === "string" ? step.request_body.trim() : "";
  if (body) {
    lines.push(truncateText(`body: ${normalizeInlineText(body)}`, 140));
  }

  return lines.slice(0, 3);
}

function buildModuleSummary(module: WorkflowNodeData["module"]): string[] {
  if (!module) return [];
  const lines: string[] = [];

  const extendsValue = typeof module.extends === "string" ? module.extends.trim() : "";
  if (extendsValue) lines.push(truncateText(`extends: ${normalizeInlineText(extendsValue)}`, 140));

  const path = typeof module.path === "string" ? module.path.trim() : "";
  if (path) lines.push(truncateText(`path: ${normalizeInlineText(path)}`, 140));

  const dependsOn = Array.isArray(module.depends_on) ? module.depends_on : [];
  if (dependsOn.length > 0) {
    const shown = dependsOn.slice(0, 3).join(", ");
    const suffix = dependsOn.length > 3 ? ` (+${dependsOn.length - 3})` : "";
    lines.push(truncateText(`depends: ${normalizeInlineText(shown)}${suffix}`, 140));
  }

  const params = module.params && typeof module.params === "object" && !Array.isArray(module.params)
    ? (module.params as Record<string, unknown>)
    : null;
  if (params) {
    const entries = Object.entries(params).filter(([k]) => Boolean(k));
    if (entries.length > 0) {
      const shown = entries.slice(0, 3).map(([k, v]) => {
        if (isSensitiveKey(k)) return `${k}=***`;
        const valueText = stringifyInlineValue(v);
        return valueText ? `${k}=${valueText}` : k;
      });
      const suffix = entries.length > 3 ? ` (+${entries.length - 3})` : "";
      lines.push(truncateText(`params: ${shown.join(", ")}${suffix}`, 140));
    }
  }

  return lines.slice(0, 3);
}

function buildCommandSummary(step: WorkflowNodeData["step"]): string {
  if (!step) return "";

  const hasStructuredArgs =
    step.speed_args !== undefined ||
    step.config_args !== undefined ||
    step.input_args !== undefined ||
    step.output_args !== undefined;

  if (typeof step.command === "string" && step.command.trim()) {
    if (hasStructuredArgs) {
      const parts = [
        step.command,
        step.speed_args,
        step.config_args,
        step.input_args,
        step.output_args,
      ]
        .map((p) => (typeof p === "string" ? p.trim() : ""))
        .filter(Boolean);
      return parts.join(" ");
    }
    return step.command.trim();
  }

  if (Array.isArray(step.commands) && step.commands.length > 0) {
    const first = String(step.commands[0] || "").trim();
    if (!first) return "";
    if (step.commands.length === 1) return first;
    return `${first} (+${step.commands.length - 1})`;
  }

  if (Array.isArray(step.parallel_commands) && step.parallel_commands.length > 0) {
    const first = String(step.parallel_commands[0] || "").trim();
    if (!first) return "";
    if (step.parallel_commands.length === 1) return first;
    return `${first} (+${step.parallel_commands.length - 1})`;
  }

  return "";
}

function hasStructuredArgs(step: WorkflowNodeData["step"]): boolean {
  if (!step) return false;
  return (
    step.speed_args !== undefined ||
    step.config_args !== undefined ||
    step.input_args !== undefined ||
    step.output_args !== undefined
  );
}

interface BaseNodeProps {
  data: WorkflowNodeData;
  selected?: boolean;
  icon: React.ReactNode;
  color: string;
  showHandles?: boolean;
}

export function BaseNode({
  data,
  selected,
  icon,
  color,
  showHandles = true,
}: BaseNodeProps) {
  const { wrapLongText, showDetails } = useCanvasSettings();
  const commandSummary = React.useMemo(() => buildCommandSummary(data.step), [data.step]);
  const structured = React.useMemo(() => hasStructuredArgs(data.step), [data.step]);
  const functionMessageSummary = React.useMemo(
    () => buildFunctionMessagesSummary(data.step),
    [data.step]
  );
  const httpSummary = React.useMemo(() => buildHttpSummary(data.step), [data.step]);
  const moduleSummary = React.useMemo(() => buildModuleSummary(data.module), [data.module]);
  const llmMessageSummary = React.useMemo(() => {
    if (data.step?.type !== "llm") return [] as string[];
    return buildLlmMessagesSummary(data.step);
  }, [data.step]);

  return (
    <>
      {showHandles && (
        <Handle
          type="target"
          position={Position.Top}
          className="!bg-border !border-2 !border-background !size-3"
        />
      )}

      <div
        className={cn(
          "flex items-center gap-3 rounded-lg border bg-card px-4 py-3 shadow-sm transition-all min-w-[200px]",
          selected && "ring-2 ring-ring ring-offset-2 ring-offset-background"
        )}
      >
        <div
          className={cn(
            "flex size-8 items-center justify-center rounded-md",
            color
          )}
        >
          {icon}
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium truncate">{data.label}</p>
          {data.step?.type ? (
            <div className="flex items-center gap-2">
              <p className="text-xs text-muted-foreground capitalize">
                {data.step.type}
              </p>
              {showDetails && structured && (
                <span className="rounded border px-1.5 py-0.5 text-[10px] text-muted-foreground">
                  args
                </span>
              )}
            </div>
          ) : data.module ? (
            <p className="text-xs text-muted-foreground">module</p>
          ) : null}
          {showDetails && commandSummary && (
            <p
              className={cn(
                "mt-1 text-[11px] text-muted-foreground font-mono",
                wrapLongText ? "whitespace-pre-wrap break-words" : "truncate"
              )}
            >
              {commandSummary}
            </p>
          )}

          {showDetails && data.step?.type === "function" && functionMessageSummary.length > 0 && (
            <div
              className={cn(
                "mt-1 space-y-0.5 text-[11px] text-muted-foreground font-mono",
                wrapLongText ? "whitespace-pre-wrap break-words" : ""
              )}
            >
              {functionMessageSummary.map((line, idx) => (
                <div key={`${data.label}-fn-msg-${idx}`} className={wrapLongText ? "" : "truncate"}>
                  {line}
                </div>
              ))}
            </div>
          )}

          {showDetails && data.step?.type === "http" && httpSummary.length > 0 && (
            <div
              className={cn(
                "mt-1 space-y-0.5 text-[11px] text-muted-foreground font-mono",
                wrapLongText ? "whitespace-pre-wrap break-words" : ""
              )}
            >
              {httpSummary.map((line, idx) => (
                <div key={`${data.label}-http-${idx}`} className={wrapLongText ? "" : "truncate"}>
                  {line}
                </div>
              ))}
            </div>
          )}

          {showDetails && data.module && moduleSummary.length > 0 && (
            <div
              className={cn(
                "mt-1 space-y-0.5 text-[11px] text-muted-foreground font-mono",
                wrapLongText ? "whitespace-pre-wrap break-words" : ""
              )}
            >
              {moduleSummary.map((line, idx) => (
                <div key={`${data.label}-module-${idx}`} className={wrapLongText ? "" : "truncate"}>
                  {line}
                </div>
              ))}
            </div>
          )}

          {showDetails && data.step?.type === "llm" && llmMessageSummary.length > 0 && (
            <div
              className={cn(
                "mt-1 space-y-0.5 text-[11px] text-muted-foreground font-mono",
                wrapLongText ? "whitespace-pre-wrap break-words" : ""
              )}
            >
              {llmMessageSummary.map((line, idx) => (
                <div key={`${data.label}-llm-msg-${idx}`} className={wrapLongText ? "" : "truncate"}>
                  {line}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {showHandles && (
        <Handle
          type="source"
          position={Position.Bottom}
          className="!bg-border !border-2 !border-background !size-3"
        />
      )}
    </>
  );
}
