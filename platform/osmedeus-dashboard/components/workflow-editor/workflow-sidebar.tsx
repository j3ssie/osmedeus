"use client";

import * as React from "react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import type { WorkflowStep, WorkflowFlowModule } from "@/lib/types/workflow";
import {
  TerminalIcon,
  LayersIcon,
  FunctionSquareIcon,
  RepeatIcon,
  ClockIcon,
  ClipboardIcon,
  GlobeIcon,
  BrainIcon,
  BoxIcon,
  ZapIcon,
  SlidersHorizontalIcon,
} from "lucide-react";
import { Light as SyntaxHighlighter } from "react-syntax-highlighter";
import yamlLang from "react-syntax-highlighter/dist/esm/languages/hljs/yaml";
import bashLang from "react-syntax-highlighter/dist/esm/languages/hljs/bash";
import javascriptLang from "react-syntax-highlighter/dist/esm/languages/hljs/javascript";
import github from "react-syntax-highlighter/dist/esm/styles/hljs/github";
import atomOneDark from "react-syntax-highlighter/dist/esm/styles/hljs/atom-one-dark";
import { useTheme } from "next-themes";
import { toast } from "sonner";
import * as jsyaml from "js-yaml";

interface WorkflowSidebarProps {
  selectedStep: WorkflowStep | null;
  selectedModule?: WorkflowFlowModule | null;
  selectedTrigger?: Record<string, unknown>[] | null;
  selectedOverride?: { name?: string; params?: Record<string, unknown> } | null;
  yamlPreview: string;
  wrapLongText?: boolean;
  onStepUpdate?: (stepName: string, updates: Partial<WorkflowStep>) => void;
  workflowKind?: "module" | "flow" | null;
  allSteps?: WorkflowStep[];
  allModules?: WorkflowFlowModule[];
  onNavigateToNode?: (nodeId: string) => void;
}

export function WorkflowSidebar({
  selectedStep,
  selectedModule = null,
  selectedTrigger = null,
  selectedOverride = null,
  yamlPreview,
  wrapLongText = false,
  onStepUpdate,
  workflowKind = null,
  allSteps = [],
  allModules = [],
  onNavigateToNode,
}: WorkflowSidebarProps) {
  SyntaxHighlighter.registerLanguage("yaml", yamlLang);
  SyntaxHighlighter.registerLanguage("bash", bashLang);
  SyntaxHighlighter.registerLanguage("javascript", javascriptLang);
  // `resolvedTheme` so a "system" preference still picks the dark code style.
  const { resolvedTheme: theme } = useTheme();
  const stepTypeIcons: Record<string, React.ReactNode> = {
    bash: <TerminalIcon className="size-4" />,
    parallel: <LayersIcon className="size-4" />,
    "parallel-steps": <LayersIcon className="size-4" />,
    function: <FunctionSquareIcon className="size-4" />,
    foreach: <RepeatIcon className="size-4" />,
    http: <GlobeIcon className="size-4" />,
    llm: <BrainIcon className="size-4" />,
    container: <BoxIcon className="size-4" />,
    "remote-bash": <BoxIcon className="size-4" />,
    module: <BoxIcon className="size-4" />,
    trigger: <ZapIcon className="size-4" />,
    override: <SlidersHorizontalIcon className="size-4" />,
  };
  const CodeHighlighter = SyntaxHighlighter as unknown as React.ComponentType<any>;
  const selectionType =
    selectedStep?.type ?? (selectedOverride ? "override" : selectedModule ? "module" : selectedTrigger ? "trigger" : "");
  const selectionName =
    selectedStep?.name ??
    (selectedOverride
      ? selectedOverride.name === "override"
        ? "Override"
        : selectedOverride.name || "Override"
      : selectedModule?.name ?? (selectedTrigger ? "Triggers" : ""));
  const [activeTab, setActiveTab] = React.useState("properties");

  const normalizedTriggerEntries = React.useMemo(() => {
    if (!selectedTrigger) return [];
    return selectedTrigger.map((entry) => {
      const anyEntry = entry as Record<string, unknown>;
      const event = anyEntry.event && typeof anyEntry.event === "object" ? (anyEntry.event as Record<string, unknown>) : null;
      const rawFilters = event?.filters;
      const rawFilterFunctions = (event as any)?.filter_functions ?? (event as any)?.filterFunctions;
      const filters = Array.isArray(rawFilters) ? rawFilters.filter((f) => typeof f === "string") as string[] : [];
      const filterFunctions = Array.isArray(rawFilterFunctions)
        ? rawFilterFunctions.filter((f) => typeof f === "string") as string[]
        : [];

      return {
        name: typeof anyEntry.name === "string" ? anyEntry.name : "",
        on: typeof anyEntry.on === "string" ? anyEntry.on : "",
        enabled: typeof anyEntry.enabled === "boolean" ? anyEntry.enabled : undefined,
        schedule: typeof anyEntry.schedule === "string" ? anyEntry.schedule : "",
        path: typeof anyEntry.path === "string" ? anyEntry.path : "",
        event: event
          ? {
              topic: typeof event.topic === "string" ? event.topic : "",
              filters,
              filterFunctions,
            }
          : null,
      };
    });
  }, [selectedTrigger]);

  const badgeVariantForStepType = React.useCallback(
    (
      type: string
    ):
      | "default"
      | "secondary"
      | "destructive"
      | "outline"
      | "success"
      | "warning"
      | "info"
      | "purple"
      | "pink"
      | "cyan"
      | "orange" => {
      switch (type) {
        case "bash":
          return "info";
        case "remote-bash":
          return "info";
        case "container":
          return "warning";
        case "parallel":
          return "cyan";
        case "parallel-steps":
          return "cyan";
        case "function":
          return "purple";
        case "foreach":
          return "orange";
        case "http":
          return "success";
        case "llm":
          return "pink";
        default:
          return "secondary";
      }
    },
    []
  );

  const bashResolvedCommand = React.useMemo(() => {
    if (!selectedStep) return "";
    const hasStructuredArgs =
      selectedStep.speed_args !== undefined ||
      selectedStep.config_args !== undefined ||
      selectedStep.input_args !== undefined ||
      selectedStep.output_args !== undefined;
    if (!hasStructuredArgs) return "";

    const parts = [
      selectedStep.command,
      selectedStep.speed_args,
      selectedStep.config_args,
      selectedStep.input_args,
      selectedStep.output_args,
    ]
      .map((p) => (typeof p === "string" ? p.trim() : ""))
      .filter(Boolean);

    return parts.join(" ");
  }, [selectedStep]);

  const foreachStepYaml = React.useMemo(() => {
    if (!selectedStep || selectedStep.type !== "foreach") return "";
    if (!selectedStep.step) return "";
    try {
      return jsyaml.dump(selectedStep.step, {
        indent: 2,
        lineWidth: -1,
        noRefs: true,
        quoteStyle: "double",
      });
    } catch {
      return "";
    }
  }, [selectedStep]);

  const renderStringList = React.useCallback(
    (
      label: string,
      items: string[],
      opts?: {
        language?: "bash" | "javascript" | "json";
        copyAllText?: string;
      }
    ) => {
      const language = opts?.language;
      const copyAllText = opts?.copyAllText;

      return (
        <div className="space-y-2">
          <div className="flex items-center justify-between gap-2">
            <Label>
              {label} ({items.length})
            </Label>
            {copyAllText !== undefined && (
              <Button
                className="rounded-md"
                variant="outline"
                size="icon"
                onClick={async () => {
                  try {
                    await navigator.clipboard.writeText(copyAllText);
                    toast.success("Copied to clipboard");
                  } catch {
                    toast.error("Failed to copy");
                  }
                }}
              >
                <ClipboardIcon className="size-4" />
                <span className="sr-only">Copy all</span>
              </Button>
            )}
          </div>

          <div className="space-y-2">
            {items.map((item, idx) => (
              <div key={`${idx}-${item}`} className="rounded-md border bg-muted/30 p-2">
                <div className="mb-2 flex items-center justify-between gap-2">
                  <Badge variant="secondary" className="text-[10px]">
                    {idx + 1}
                  </Badge>
                </div>
                {language ? (
                  <CodeHighlighter
                    language={language}
                    style={theme === "dark" ? atomOneDark : github}
                    customStyle={{
                      margin: 0,
                      background: "transparent",
                      fontSize: "0.75rem",
                      whiteSpace: "pre-wrap",
                      wordBreak: "break-word",
                    }}
                    codeTagProps={{
                      style: {
                        whiteSpace: "pre-wrap",
                        wordBreak: "break-word",
                      },
                    }}
                  >
                    {item}
                  </CodeHighlighter>
                ) : (
                  <div className="font-mono text-xs whitespace-pre-wrap break-words">{item}</div>
                )}
              </div>
            ))}
          </div>
        </div>
      );
    },
    [CodeHighlighter, theme]
  );

  const renderDecisionBlock = React.useCallback((decision: any) => {
    if (!decision) return null;

    if (Array.isArray(decision) && decision.length > 0) {
      return (
        <div className="space-y-2">
          <Label>Decision Rules</Label>
          <div className="space-y-2">
            {decision.map((rule, i) => (
              <div key={i} className="rounded-md border p-2 text-xs font-mono overflow-hidden">
                <div className="text-muted-foreground whitespace-pre-wrap break-words">if {rule?.condition}</div>
                <div className="text-primary whitespace-pre-wrap break-words">→ {rule?.next}</div>
              </div>
            ))}
          </div>
        </div>
      );
    }

    if (typeof decision === "object" && decision && typeof (decision as any).switch === "string" && (decision as any).cases && typeof (decision as any).cases === "object") {
      const cases = (decision as any).cases as Record<string, any>;
      const entries = Object.entries(cases).filter(([k]) => typeof k === "string" && k.trim().length > 0);
      const hasDefault = (decision as any).default && (typeof (decision as any).default.goto === "string" || typeof (decision as any).default.next === "string");
      const defaultGoto = typeof (decision as any).default?.goto === "string" ? (decision as any).default.goto : (typeof (decision as any).default?.next === "string" ? (decision as any).default.next : "");

      return (
        <div className="space-y-2">
          <Label>Decision (switch)</Label>
          <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
            switch {(decision as any).switch}
          </div>
          <div className="space-y-2">
            {entries.map(([key, value]) => {
              const goto = typeof value?.goto === "string" ? value.goto : (typeof value?.next === "string" ? value.next : "");
              return (
                <div key={key} className="rounded-md border p-2 text-xs font-mono overflow-hidden">
                  <div className="text-muted-foreground whitespace-pre-wrap break-words">case {key}</div>
                  <div className="text-primary whitespace-pre-wrap break-words">→ {goto}</div>
                </div>
              );
            })}
            {hasDefault && (
              <div className="rounded-md border p-2 text-xs font-mono overflow-hidden">
                <div className="text-muted-foreground whitespace-pre-wrap break-words">default</div>
                <div className="text-primary whitespace-pre-wrap break-words">→ {defaultGoto}</div>
              </div>
            )}
          </div>
        </div>
      );
    }

    return null;
  }, []);

  return (
    <div className="flex h-full flex-col border-l bg-card">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="flex h-full flex-col">
        <div className="border-b px-4 py-2">
          <TabsList className="w-full">
            <TabsTrigger value="properties" className="flex-1">
              Properties
            </TabsTrigger>
            <TabsTrigger value="items" className="flex-1">
              {workflowKind === "flow" ? "Modules" : "Steps"}
            </TabsTrigger>
            <TabsTrigger value="yaml" className="flex-1">
              YAML
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="properties" className="flex-1 m-0 min-h-0">
          <ScrollArea className="h-full">
            {selectedStep || selectedModule || selectedTrigger || selectedOverride ? (
              <div className="p-4 space-y-6">
                {/* Step Header */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    {stepTypeIcons[selectionType]}
                    <Badge variant="secondary" className="capitalize">
                      {selectionType}
                    </Badge>
                  </div>
                  <h3 className="text-lg font-semibold">{selectionName}</h3>
                </div>

                <Separator />

                {/* Basic Properties */}
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="stepName">Name</Label>
                    <Input
                      id="stepName"
                      value={selectionName}
                      onChange={(e) =>
                        selectedStep ? onStepUpdate?.(selectedStep.name, { name: e.target.value }) : undefined
                      }
                      disabled={!selectedStep}
                    />
                  </div>

                  {selectedTrigger && normalizedTriggerEntries.length > 0 && (
                    <div className="space-y-4">
                      <Label>Triggers</Label>
                      <div className="space-y-3">
                        {normalizedTriggerEntries.map((trigger, idx) => (
                          <div key={`${trigger.name || "trigger"}-${idx}`} className="rounded-md border p-3 space-y-3">
                            <div className="flex flex-wrap items-center gap-2">
                              <Badge variant="secondary" className="uppercase">
                                {trigger.on || "trigger"}
                              </Badge>
                              <span className="text-sm font-medium">
                                {trigger.name || `Trigger ${idx + 1}`}
                              </span>
                              {typeof trigger.enabled === "boolean" && (
                                <Badge variant={trigger.enabled ? "success" : "outline"}>
                                  {trigger.enabled ? "enabled" : "disabled"}
                                </Badge>
                              )}
                            </div>

                            {trigger.schedule && (
                              <div className="space-y-2">
                                <Label>Schedule</Label>
                                <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                                  {trigger.schedule}
                                </div>
                              </div>
                            )}

                            {trigger.path && (
                              <div className="space-y-2">
                                <Label>Path</Label>
                                <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                                  {trigger.path}
                                </div>
                              </div>
                            )}

                            {trigger.event?.topic && (
                              <div className="space-y-2">
                                <Label>Topic</Label>
                                <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                                  {trigger.event.topic}
                                </div>
                              </div>
                            )}

                            {trigger.event?.filters && trigger.event.filters.length > 0 &&
                              renderStringList("Filters", trigger.event.filters, {
                                language: "javascript",
                                copyAllText: trigger.event.filters.join("\n"),
                              })}

                            {trigger.event?.filterFunctions && trigger.event.filterFunctions.length > 0 &&
                              renderStringList("Filter Functions", trigger.event.filterFunctions, {
                                language: "javascript",
                                copyAllText: trigger.event.filterFunctions.join("\n"),
                              })}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {selectedOverride && selectedOverride.params && Object.keys(selectedOverride.params).length > 0 && (
                    <div className="space-y-4">
                      <div className="space-y-2">
                        <Label>Override Params</Label>
                        <div className="rounded-md border bg-muted/30 p-2">
                          <CodeHighlighter
                            language="json"
                            style={theme === "dark" ? atomOneDark : github}
                            customStyle={{
                              margin: 0,
                              background: "transparent",
                              fontSize: "0.75rem",
                              whiteSpace: "pre-wrap",
                              wordBreak: "break-word",
                            }}
                            codeTagProps={{
                              style: {
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              },
                            }}
                          >
                            {JSON.stringify(selectedOverride.params, null, 2)}
                          </CodeHighlighter>
                        </div>
                      </div>
                    </div>
                  )}

                  {selectedModule && (
                    <div className="space-y-4">
                      {selectedModule.extends && (
                        <div className="space-y-2">
                          <Label>Extends</Label>
                          <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                            {selectedModule.extends}
                          </div>
                        </div>
                      )}

                      {selectedModule.path && (
                        <div className="space-y-2">
                          <Label>Path</Label>
                          <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                            {selectedModule.path}
                          </div>
                        </div>
                      )}

                      {Array.isArray(selectedModule.depends_on) && selectedModule.depends_on.length > 0 &&
                        renderStringList(
                          "Depends On",
                          selectedModule.depends_on.map((d) => String(d)),
                          { copyAllText: selectedModule.depends_on.join("\n") }
                        )}

                      {selectedModule.condition && (
                        <div className="space-y-2">
                          <Label>Condition</Label>
                          <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                            {selectedModule.condition}
                          </div>
                        </div>
                      )}

                      {selectedModule.params && Object.keys(selectedModule.params).length > 0 && (
                        <div className="space-y-2">
                          <Label>Params</Label>
                          <div className="rounded-md border bg-muted/30 p-2">
                            <CodeHighlighter
                              language="json"
                              style={theme === "dark" ? atomOneDark : github}
                              customStyle={{
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                              codeTagProps={{
                                style: {
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                },
                              }}
                            >
                              {JSON.stringify(selectedModule.params, null, 2)}
                            </CodeHighlighter>
                          </div>
                        </div>
                      )}

                      {selectedModule?.decision && renderDecisionBlock(selectedModule.decision)}

                      {Array.isArray(selectedModule.on_success) && selectedModule.on_success.length > 0 && (
                        <div className="space-y-2">
                          <Label>On Success</Label>
                          <div className="rounded-md border bg-muted/30 p-2">
                            <CodeHighlighter
                              language="json"
                              style={theme === "dark" ? atomOneDark : github}
                              customStyle={{
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                              codeTagProps={{
                                style: {
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                },
                              }}
                            >
                              {JSON.stringify(selectedModule.on_success, null, 2)}
                            </CodeHighlighter>
                          </div>
                        </div>
                      )}

                      {Array.isArray(selectedModule.on_error) && selectedModule.on_error.length > 0 && (
                        <div className="space-y-2">
                          <Label>On Error</Label>
                          <div className="rounded-md border bg-muted/30 p-2">
                            <CodeHighlighter
                              language="json"
                              style={theme === "dark" ? atomOneDark : github}
                              customStyle={{
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                              codeTagProps={{
                                style: {
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                },
                              }}
                            >
                              {JSON.stringify(selectedModule.on_error, null, 2)}
                            </CodeHighlighter>
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {selectedStep && (selectedStep.type === "bash" || selectedStep.type === "remote-bash" || selectedStep.type === "container") && selectedStep.command && (
                    <div className="space-y-2">
                      <div className="flex items-center justify-between gap-2">
                        <Label htmlFor="command">Command</Label>
                        <Button
                          className="rounded-md"
                          variant="outline"
                          size="icon"
                          onClick={async () => {
                            try {
                              await navigator.clipboard.writeText(selectedStep.command || "");
                              toast.success("Copied to clipboard");
                            } catch {
                              toast.error("Failed to copy");
                            }
                          }}
                        >
                          <ClipboardIcon className="size-4" />
                          <span className="sr-only">Copy command</span>
                        </Button>
                      </div>
                      <div className="rounded-md border bg-muted/30 p-2">
                        <CodeHighlighter
                          language="bash"
                          style={theme === "dark" ? atomOneDark : github}
                          customStyle={{
                            margin: 0,
                            background: "transparent",
                            fontSize: "0.75rem",
                            whiteSpace: "pre-wrap",
                            wordBreak: "break-word",
                          }}
                          codeTagProps={{
                            style: {
                              whiteSpace: "pre-wrap",
                              wordBreak: "break-word",
                            },
                          }}
                        >
                          {selectedStep.command}
                        </CodeHighlighter>
                      </div>
                    </div>
                  )}

                  {selectedStep && (selectedStep.type === "bash" || selectedStep.type === "remote-bash" || selectedStep.type === "container") && (
                    selectedStep.speed_args !== undefined ||
                    selectedStep.config_args !== undefined ||
                    selectedStep.input_args !== undefined ||
                    selectedStep.output_args !== undefined
                  ) && (
                    <div className="space-y-2">
                      <Label>Structured Args</Label>
                      <div className="space-y-2">
                        <div className="rounded-md border p-2 text-xs font-mono overflow-hidden">
                          <div className="text-muted-foreground whitespace-pre-wrap break-words">speed_args</div>
                          <div className="whitespace-pre-wrap break-words">{selectedStep.speed_args ?? ""}</div>
                        </div>
                        <div className="rounded-md border p-2 text-xs font-mono overflow-hidden">
                          <div className="text-muted-foreground whitespace-pre-wrap break-words">config_args</div>
                          <div className="whitespace-pre-wrap break-words">{selectedStep.config_args ?? ""}</div>
                        </div>
                        <div className="rounded-md border p-2 text-xs font-mono overflow-hidden">
                          <div className="text-muted-foreground whitespace-pre-wrap break-words">input_args</div>
                          <div className="whitespace-pre-wrap break-words">{selectedStep.input_args ?? ""}</div>
                        </div>
                        <div className="rounded-md border p-2 text-xs font-mono overflow-hidden">
                          <div className="text-muted-foreground whitespace-pre-wrap break-words">output_args</div>
                          <div className="whitespace-pre-wrap break-words">{selectedStep.output_args ?? ""}</div>
                        </div>
                      </div>

                      {bashResolvedCommand && (
                        <div className="space-y-2">
                          <Label>Resolved Command</Label>
                          <div className="rounded-md border bg-muted/30 p-2">
                            <CodeHighlighter
                              language="bash"
                              style={theme === "dark" ? atomOneDark : github}
                              customStyle={{
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                              codeTagProps={{
                                style: {
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                },
                              }}
                            >
                              {bashResolvedCommand}
                            </CodeHighlighter>
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {selectedStep && (selectedStep.type === "bash" || selectedStep.type === "remote-bash" || selectedStep.type === "container") &&
                    Array.isArray(selectedStep.commands) && selectedStep.commands.length > 0 &&
                    renderStringList("Commands", selectedStep.commands, {
                      language: "bash",
                      copyAllText: selectedStep.commands.join("\n"),
                    })}

                  {selectedStep && (selectedStep.type === "bash" || selectedStep.type === "remote-bash" || selectedStep.type === "container") &&
                    Array.isArray(selectedStep.parallel_commands) && selectedStep.parallel_commands.length > 0 &&
                    renderStringList("Parallel Commands", selectedStep.parallel_commands, {
                      language: "bash",
                      copyAllText: selectedStep.parallel_commands.join("\n"),
                    })}

                  {selectedStep && selectedStep.type === "function" && selectedStep.function && (
                    <div className="space-y-2">
                      <Label>Function</Label>
                      <div className="rounded-md border bg-muted/30 p-2">
                        <CodeHighlighter
                          language="javascript"
                          style={theme === "dark" ? atomOneDark : github}
                          customStyle={{
                            margin: 0,
                            background: "transparent",
                            fontSize: "0.75rem",
                            whiteSpace: "pre-wrap",
                            wordBreak: "break-word",
                          }}
                          wrapLongLines
                          codeTagProps={{
                            style: {
                              whiteSpace: "pre-wrap",
                              wordBreak: "break-word",
                            },
                          }}
                        >
                          {selectedStep.function}
                        </CodeHighlighter>
                      </div>
                    </div>
                  )}

                  {selectedStep && selectedStep.type === "function" &&
                    Array.isArray(selectedStep.functions) && selectedStep.functions.length > 0 &&
                    renderStringList("Functions", selectedStep.functions, {
                      language: "javascript",
                      copyAllText: selectedStep.functions.join("\n"),
                    })}

                  {selectedStep && selectedStep.type === "function" &&
                    Array.isArray(selectedStep.parallel_functions) && selectedStep.parallel_functions.length > 0 &&
                    renderStringList("Parallel Functions", selectedStep.parallel_functions, {
                      language: "javascript",
                      copyAllText: selectedStep.parallel_functions.join("\n"),
                    })}

                  {selectedStep && selectedStep.type === "http" && (
                    <div className="space-y-4">
                      {selectedStep.url && (
                        <div className="space-y-2">
                          <Label>URL</Label>
                          <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                            {selectedStep.url}
                          </div>
                        </div>
                      )}

                      {selectedStep.method && (
                        <div className="space-y-2">
                          <Label>Method</Label>
                          <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                            {selectedStep.method}
                          </div>
                        </div>
                      )}

                      {selectedStep.headers && Object.keys(selectedStep.headers).length > 0 && (
                        <div className="space-y-2">
                          <Label>Headers</Label>
                          <div className="rounded-md border bg-muted/30 p-2">
                            <CodeHighlighter
                              language="json"
                              style={theme === "dark" ? atomOneDark : github}
                              customStyle={{
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                              codeTagProps={{
                                style: {
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                },
                              }}
                            >
                              {JSON.stringify(selectedStep.headers, null, 2)}
                            </CodeHighlighter>
                          </div>
                        </div>
                      )}

                      {selectedStep.request_body && (
                        <div className="space-y-2">
                          <Label>Request Body</Label>
                          <div className="rounded-md border bg-muted/30 p-2">
                            <CodeHighlighter
                              language="json"
                              style={theme === "dark" ? atomOneDark : github}
                              customStyle={{
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                              codeTagProps={{
                                style: {
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                },
                              }}
                            >
                              {String(selectedStep.request_body)}
                            </CodeHighlighter>
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {selectedStep && selectedStep.type === "llm" && (
                    <div className="space-y-4">
                      {typeof selectedStep.is_embedding === "boolean" && (
                        <div className="space-y-2">
                          <Label>Embedding</Label>
                          <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs">
                            {String(selectedStep.is_embedding)}
                          </div>
                        </div>
                      )}

                      {Array.isArray(selectedStep.embedding_input) && selectedStep.embedding_input.length > 0 &&
                        renderStringList(
                          "Embedding Input",
                          selectedStep.embedding_input.map((v) => String(v)),
                          { copyAllText: selectedStep.embedding_input.join("\n") }
                        )}

                      {Array.isArray(selectedStep.messages) && selectedStep.messages.length > 0 &&
                        renderStringList(
                          "Messages",
                          selectedStep.messages.map((m) => JSON.stringify(m, null, 2)),
                          { language: "json", copyAllText: JSON.stringify(selectedStep.messages, null, 2) }
                        )}

                      {Array.isArray(selectedStep.tools) && selectedStep.tools.length > 0 &&
                        renderStringList(
                          "Tools",
                          selectedStep.tools.map((t) => JSON.stringify(t, null, 2)),
                          { language: "json", copyAllText: JSON.stringify(selectedStep.tools, null, 2) }
                        )}

                      {selectedStep.tool_choice !== undefined && (
                        <div className="space-y-2">
                          <Label>Tool Choice</Label>
                          <div className="rounded-md border bg-muted/30 p-2">
                            <CodeHighlighter
                              language="json"
                              style={theme === "dark" ? atomOneDark : github}
                              customStyle={{
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                              codeTagProps={{
                                style: {
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                },
                              }}
                            >
                              {JSON.stringify(selectedStep.tool_choice, null, 2)}
                            </CodeHighlighter>
                          </div>
                        </div>
                      )}

                      {selectedStep.llm_config && Object.keys(selectedStep.llm_config).length > 0 && (
                        <div className="space-y-2">
                          <Label>LLM Config</Label>
                          <div className="rounded-md border bg-muted/30 p-2">
                            <CodeHighlighter
                              language="json"
                              style={theme === "dark" ? atomOneDark : github}
                              customStyle={{
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                              codeTagProps={{
                                style: {
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                },
                              }}
                            >
                              {JSON.stringify(selectedStep.llm_config, null, 2)}
                            </CodeHighlighter>
                          </div>
                        </div>
                      )}

                      {selectedStep.extra_llm_parameters && Object.keys(selectedStep.extra_llm_parameters).length > 0 && (
                        <div className="space-y-2">
                          <Label>Extra LLM Parameters</Label>
                          <div className="rounded-md border bg-muted/30 p-2">
                            <CodeHighlighter
                              language="json"
                              style={theme === "dark" ? atomOneDark : github}
                              customStyle={{
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                              codeTagProps={{
                                style: {
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                },
                              }}
                            >
                              {JSON.stringify(selectedStep.extra_llm_parameters, null, 2)}
                            </CodeHighlighter>
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {selectedStep && (selectedStep.step_runner || selectedStep.step_runner_config) && (
                    <div className="space-y-4">
                      {selectedStep.step_runner && (
                        <div className="space-y-2">
                          <Label>Runner</Label>
                          <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                            {selectedStep.step_runner}
                          </div>
                        </div>
                      )}
                      {selectedStep.step_runner_config && (
                        <div className="space-y-2">
                          <Label>Runner Config</Label>
                          <div className="rounded-md border bg-muted/30 p-2">
                            <CodeHighlighter
                              language="json"
                              style={theme === "dark" ? atomOneDark : github}
                              customStyle={{
                                margin: 0,
                                background: "transparent",
                                fontSize: "0.75rem",
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                              codeTagProps={{
                                style: {
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                },
                              }}
                            >
                              {JSON.stringify(selectedStep.step_runner_config, null, 2)}
                            </CodeHighlighter>
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {selectedStep && (selectedStep.type === "parallel" || selectedStep.type === "parallel-steps") && Array.isArray(selectedStep.parallel_steps) && (
                    <div className="space-y-2">
                      <Label>Parallel Steps ({selectedStep.parallel_steps.length})</Label>
                      <div className="space-y-2">
                        {selectedStep.parallel_steps.map((ps) => {
                          let psYaml = "";
                          try {
                            psYaml = jsyaml.dump(ps, {
                              indent: 2,
                              lineWidth: -1,
                              noRefs: true,
                              quoteStyle: "double",
                            });
                          } catch {
                            psYaml = "";
                          }

                          return (
                            <div key={ps.name} className="rounded-md border p-3 space-y-2">
                              <div className="flex items-center gap-2">
                                {stepTypeIcons[ps.type]}
                                <span className="font-medium text-sm">{ps.name}</span>
                                <Badge variant="secondary" className="capitalize">
                                  {ps.type}
                                </Badge>
                              </div>

                              {psYaml && (
                                <div className="rounded-md border bg-muted/30 p-2">
                                  <CodeHighlighter
                                    language="yaml"
                                    style={theme === "dark" ? atomOneDark : github}
                                    customStyle={{
                                      margin: 0,
                                      background: "transparent",
                                      fontSize: "0.75rem",
                                      whiteSpace: "pre-wrap",
                                      wordBreak: "break-word",
                                    }}
                                    codeTagProps={{
                                      style: {
                                        whiteSpace: "pre-wrap",
                                        wordBreak: "break-word",
                                      },
                                    }}
                                    showLineNumbers
                                  >
                                    {psYaml.trim()}
                                  </CodeHighlighter>
                                </div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  )}

                  {selectedStep && selectedStep.type === "foreach" && (
                    <>
                      <div className="space-y-2">
                        <Label>Input File</Label>
                        <div className="rounded-md bg-muted p-3 font-mono text-xs">
                          {selectedStep.input}
                        </div>
                      </div>
                      <div className="space-y-2">
                        <Label>Variable Name</Label>
                        <div className="rounded-md bg-muted p-3 font-mono text-xs">
                          {selectedStep.variable}
                        </div>
                      </div>
                      {typeof selectedStep.threads === "number" && (
                        <div className="space-y-2">
                          <Label>Threads</Label>
                          <div className="rounded-md bg-muted p-3 font-mono text-xs">
                            {selectedStep.threads}
                          </div>
                        </div>
                      )}

                      {selectedStep.step && (
                        <div className="space-y-2">
                          <div className="flex items-center justify-between gap-2">
                            <Label>Foreach Step</Label>
                            {foreachStepYaml && (
                              <Button
                                className="rounded-md"
                                variant="outline"
                                size="icon"
                                onClick={async () => {
                                  try {
                                    await navigator.clipboard.writeText(foreachStepYaml);
                                    toast.success("Copied to clipboard");
                                  } catch {
                                    toast.error("Failed to copy");
                                  }
                                }}
                              >
                                <ClipboardIcon className="size-4" />
                                <span className="sr-only">Copy foreach step YAML</span>
                              </Button>
                            )}
                          </div>

                          <div className="rounded-md border p-2 text-sm">
                            <div className="flex items-center gap-2">
                              {stepTypeIcons[selectedStep.step.type]}
                              <span className="font-medium">{selectedStep.step.name}</span>
                              <Badge variant="secondary" className="capitalize">
                                {selectedStep.step.type}
                              </Badge>
                            </div>
                          </div>

                          {foreachStepYaml && (
                            <div className="rounded-md border bg-muted/30 p-2">
                              <CodeHighlighter
                                language="yaml"
                                style={theme === "dark" ? atomOneDark : github}
                                customStyle={{
                                  margin: 0,
                                  background: "transparent",
                                  fontSize: "0.75rem",
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                }}
                                codeTagProps={{
                                  style: {
                                    whiteSpace: "pre-wrap",
                                    wordBreak: "break-word",
                                  },
                                }}
                                showLineNumbers
                              >
                                {foreachStepYaml.trim()}
                              </CodeHighlighter>
                            </div>
                          )}
                        </div>
                      )}
                    </>
                  )}
                </div>

                {/* Timeout */}
                {selectedStep && selectedStep.timeout && (
                  <>
                    <Separator />
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <ClockIcon className="size-4" />
                      <span>Timeout: {selectedStep.timeout}s</span>
                    </div>
                  </>
                )}

                {/* Pre-condition */}
                {selectedStep && selectedStep.pre_condition && (
                  <>
                    <Separator />
                    <div className="space-y-2">
                      <Label className="text-muted-foreground">Pre-condition</Label>
                      <div className="rounded-md bg-muted p-3 font-mono text-xs">
                        {selectedStep.pre_condition}
                      </div>
                    </div>
                  </>
                )}

                {/* Exports */}
                {selectedStep && selectedStep.exports && Object.keys(selectedStep.exports).length > 0 && (
                  <>
                    <Separator />
                    <div className="space-y-2">
                      <Label>Exports</Label>
                      <div className="space-y-1">
                        {Object.entries(selectedStep.exports).map(([key, value]) => (
                          <div
                            key={key}
                            className="flex items-center gap-2 text-xs font-mono"
                          >
                            <Badge variant="outline" className="font-mono">
                              {key}
                            </Badge>
                            <span className="text-muted-foreground truncate">
                              = {value}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </>
                )}

                {/* Decision Rules */}
                {selectedStep?.decision && (
                  <>
                    <Separator />
                    {renderDecisionBlock(selectedStep.decision)}
                  </>
                )}

                {selectedStep && selectedStep.log && (
                  <>
                    <Separator />
                    <div className="space-y-2">
                      <Label>Log</Label>
                      <div className="rounded-md border bg-muted/30 p-2 font-mono text-xs whitespace-pre-wrap break-words">
                        {selectedStep.log}
                      </div>
                    </div>
                  </>
                )}
              </div>
            ) : (
              <div className="flex h-full items-center justify-center p-4">
                <p className="text-sm text-muted-foreground text-center">
                  Select a node to view its properties
                </p>
              </div>
            )}
          </ScrollArea>
        </TabsContent>

        <TabsContent value="items" className="flex-1 m-0 min-h-0">
          <ScrollArea className="h-full">
            <div className="p-3 space-y-2">
              {workflowKind === "flow" ? (
                <>
                  {allModules.length === 0 ? (
                    <div className="p-4 text-sm text-muted-foreground text-center">
                      No modules
                    </div>
                  ) : (
                    <div className="space-y-1">
                      {allModules.map((m) => {
                        const isSelected = selectionName === m.name;
                        return (
                          <Button
                            key={m.name}
                            type="button"
                            variant={isSelected ? "secondary" : "ghost"}
                            className="w-full justify-start"
                            onClick={() => {
                              onNavigateToNode?.(m.name);
                            }}
                          >
                            <div className="flex w-full items-center justify-between gap-3">
                              <span className="truncate font-mono text-xs">{m.name}</span>
                              <Badge variant="secondary" className="capitalize">
                                module
                              </Badge>
                            </div>
                          </Button>
                        );
                      })}
                    </div>
                  )}
                </>
              ) : (
                <>
                  {allSteps.length === 0 ? (
                    <div className="p-4 text-sm text-muted-foreground text-center">
                      No steps
                    </div>
                  ) : (
                    <div className="space-y-1">
                      {allSteps.map((s) => {
                        const isSelected = selectionName === s.name;
                        return (
                          <Button
                            key={s.name}
                            type="button"
                            variant={isSelected ? "secondary" : "ghost"}
                            className="w-full justify-start"
                            onClick={() => {
                              onNavigateToNode?.(s.name);
                            }}
                          >
                            <div className="flex w-full items-center justify-between gap-3">
                              <span className="truncate font-mono text-xs">{s.name}</span>
                              <Badge variant={badgeVariantForStepType(s.type)} className="capitalize">
                                {s.type}
                              </Badge>
                            </div>
                          </Button>
                        );
                      })}
                    </div>
                  )}
                </>
              )}
            </div>
          </ScrollArea>
        </TabsContent>

        <TabsContent value="yaml" className="flex-1 m-0 min-h-0">
          <div className="h-full min-h-0 overflow-y-auto p-3">
            <CodeHighlighter
              language="yaml"
              style={theme === "dark" ? atomOneDark : github}
              customStyle={{
                margin: 0,
                background: "transparent",
                fontSize: "0.75rem",
                maxHeight: "100%",
                overflowY: "auto",
                whiteSpace: wrapLongText ? "pre-wrap" : "pre",
                wordBreak: wrapLongText ? "break-word" : "normal",
              }}
              codeTagProps={{
                style: {
                  whiteSpace: wrapLongText ? "pre-wrap" : "pre",
                  wordBreak: wrapLongText ? "break-word" : "normal",
                },
              }}
              showLineNumbers
            >
              {yamlPreview}
            </CodeHighlighter>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
