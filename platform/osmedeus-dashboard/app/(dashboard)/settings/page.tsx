"use client";

import * as React from "react";
import { Light as SyntaxHighlighter } from "react-syntax-highlighter";
import yamlLang from "react-syntax-highlighter/dist/esm/languages/hljs/yaml";
import github from "react-syntax-highlighter/dist/esm/styles/hljs/github";
import atomOneDark from "react-syntax-highlighter/dist/esm/styles/hljs/atom-one-dark";
import { useTheme } from "next-themes";
import { Card, CardContent } from "@/components/ui/card";
import { SectionCardHeader } from "@/components/shared/section-card-header";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { Select, SelectTrigger, SelectContent, SelectItem, SelectValue } from "@/components/ui/select";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import {
  LoaderIcon,
  ClipboardIcon,
  SaveIcon,
  SettingsIcon,
  DatabaseIcon,
  PlugIcon,
  PaletteIcon,
  ChevronDownIcon,
} from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { getSettingsYaml } from "@/lib/api/settings";
import { presets } from "@/theme-presets";
import { defaultThemeState } from "@/config/theme";
import { isDemoMode, setDemoMode } from "@/lib/api/demo-mode";
import { API_PREFIX } from "@/lib/api/prefix";

export default function SettingsPage() {
  const [endpoint, setEndpoint] = React.useState("");
  const [token, setToken] = React.useState("");
  const [prefixKey, setPrefixKey] = React.useState("");
  const [demoModeEnabled, setDemoModeEnabled] = React.useState(false);
  const [preset, setPreset] = React.useState("default");
  const [sidebarCollapsedByDefault, setSidebarCollapsedByDefault] = React.useState(false);
  const [settingsYaml, setSettingsYaml] = React.useState("");
  const [loadingYaml, setLoadingYaml] = React.useState(false);
  // `resolvedTheme` so a "system" preference still picks the dark code style.
  const { resolvedTheme: theme } = useTheme();

  const [serverConfigCliOpen, setServerConfigCliOpen] = React.useState(false);

  const serverConfigCliUsage = `▷ Examples\n  osmedeus config set server.port 9000\n  osmedeus config set server.username admin\n  osmedeus config set scan_tactic.default 20\n  osmedeus config set global_vars.github_token ghp_xxx\n  osmedeus config set notification.enabled true\n\n▷ Available Keys\n  base_folder                    Base directory path\n  server.host                    Server bind host\n  server.port                    Server port number\n  server.username                Auth username\n  server.password                Auth password\n  server.ui_path                 UI static files path\n  database.db_engine             sqlite or postgresql\n  database.host                  Database host\n  database.port                  Database port\n  scan_tactic.aggressive         Aggressive mode threads\n  scan_tactic.default            Default mode threads\n  scan_tactic.gently             Gentle mode threads\n  redis.host                     Redis host\n  redis.port                     Redis port\n  global_vars.<name>             Set a global variable\n  notification.enabled           Enable notifications (true/false)\n  notification.telegram.bot_token Telegram bot token\n  environments.binaries_path     Binaries directory\n  storage.enabled                Enable cloud storage (true/false)`;

  const currentPreset = React.useMemo(() => {
    return preset === "default" ? null : presets[preset];
  }, [preset]);

  // Register YAML language for syntax highlighter
  React.useEffect(() => {
    SyntaxHighlighter.registerLanguage("yaml", yamlLang);
  }, []);

  // Load saved settings from localStorage on mount
  React.useEffect(() => {
    if (typeof window !== "undefined") {
      const savedEndpoint = localStorage.getItem("osmedeus_api_endpoint") || "";
      const savedToken = localStorage.getItem("osmedeus_token") || "";
      const savedPrefixKey = localStorage.getItem("osmedeus_workspace_prefix_key") || "";
      const savedPreset = localStorage.getItem("osmedeus_theme_preset");
      const savedSidebarCollapsedByDefault =
        localStorage.getItem("osmedeus_sidebar_collapsed_by_default") === "true";
      setEndpoint(savedEndpoint);
      setToken(savedToken);
      setPrefixKey(savedPrefixKey);
      if (savedPreset) setPreset(savedPreset);
      setSidebarCollapsedByDefault(savedSidebarCollapsedByDefault);
      // Load demo mode preference
      setDemoModeEnabled(isDemoMode());
    }
  }, []);

  const saveSidebarCollapsedByDefault = (checked: boolean) => {
    setSidebarCollapsedByDefault(checked);
    if (typeof window !== "undefined") {
      localStorage.setItem("osmedeus_sidebar_collapsed_by_default", checked ? "true" : "false");
      window.dispatchEvent(
        new CustomEvent("osmedeus-sidebar-collapsed-by-default-changed", { detail: checked })
      );
      toast.success("Appearance updated");
    }
  };

  // Auto-load YAML on mount (works in mock mode without credentials)
  React.useEffect(() => {
    loadYaml();
  }, []);

  const saveApiConfig = () => {
    if (typeof window !== "undefined") {
      const trimmed = endpoint.trim().replace(/\/+$/, "");
      const normalized =
        trimmed === API_PREFIX
          ? window.location.origin
          : trimmed.endsWith(API_PREFIX)
            ? trimmed.slice(0, Math.max(0, trimmed.length - API_PREFIX.length)).replace(/\/+$/, "")
            : trimmed;
      const cleanEndpoint = normalized.startsWith("/") ? window.location.origin : normalized;

      localStorage.setItem("osmedeus_api_endpoint", cleanEndpoint);
      localStorage.setItem("osmedeus_token", token);
      localStorage.setItem("osmedeus_workspace_prefix_key", prefixKey.trim());
      toast.success("API configuration saved");
    }
  };

  const handleDemoModeChange = (checked: boolean) => {
    setDemoModeEnabled(checked);
    setDemoMode(checked);
    toast.success(checked ? "Demo mode enabled" : "Demo mode disabled", {
      description: "Page will reload to apply changes",
    });
    setTimeout(() => window.location.reload(), 500);
  };

  const loadYaml = async () => {
    setLoadingYaml(true);
    try {
      const data = await getSettingsYaml();
      setSettingsYaml(data);
      toast.success("Configuration loaded");
    } catch {
      toast.error("Failed to load configuration");
    } finally {
      setLoadingYaml(false);
    }
  };

  const copyYaml = async () => {
    try {
      await navigator.clipboard.writeText(settingsYaml);
      toast.success("Copied to clipboard");
    } catch {
      toast.error("Failed to copy");
    }
  };

  const savePreset = (value: string) => {
    setPreset(value);
    if (typeof window !== "undefined") {
      localStorage.setItem("osmedeus_theme_preset", value);
      localStorage.removeItem("osmedeus_theme_light_primary");
      localStorage.removeItem("osmedeus_theme_light_secondary");
      localStorage.removeItem("osmedeus_theme_dark_primary");
      localStorage.removeItem("osmedeus_theme_dark_secondary");
      window.dispatchEvent(new Event("osmedeus-theme-colors-updated"));
      toast.success("Theme applied", { description: `Using ${value} preset` });
    }
  };

  return (
    <div className="space-y-4">
      {/* Data Source */}
      <Card className="overflow-hidden">
        <SectionCardHeader
          icon={DatabaseIcon}
          title="Data Source"
          description="Choose between demo data or real API"
        />
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <label className="text-sm font-medium">Demo Mode</label>
              <p className="text-sm text-muted-foreground">
                Use sample data instead of connecting to the backend API
              </p>
            </div>
            <Switch
              checked={demoModeEnabled}
              onCheckedChange={handleDemoModeChange}
            />
          </div>
          {demoModeEnabled && (
            <div className="rounded-control border border-warning/25 bg-warning-soft p-3 text-sm text-warning">
              Demo mode is active. Data shown is sample data, not from your backend.
            </div>
          )}
          {!demoModeEnabled && (
            <div className="rounded-control border border-success/25 bg-success-soft p-3 text-sm text-success">
              Real API mode. Configure your backend endpoint and token below.
            </div>
          )}
        </CardContent>
      </Card>

      {/* API Configuration */}
      <Card className="overflow-hidden">
        <SectionCardHeader
          icon={PlugIcon}
          title="API Configuration"
          description="Configure the backend endpoint and authentication token"
        />
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <label className="text-sm font-medium">API Endpoint</label>
              <Input
                placeholder="http://localhost:8002"
                value={endpoint}
                onChange={(e) => setEndpoint(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Token</label>
              <Input
                type="password"
                placeholder="Bearer token"
                value={token}
                onChange={(e) => setToken(e.target.value)}
              />
            </div>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Workspace Prefix Key</label>
            <Input
              type="password"
              placeholder="workspace_prefix_key"
              value={prefixKey}
              onChange={(e) => setPrefixKey(e.target.value)}
            />
            <p className="text-sm text-muted-foreground">
              Secret key from the server config (server.workspace_prefix_key). Enables
              opening raw artifacts in a new tab via the static{" "}
              <code className="font-mono text-xs">/ws/&lt;key&gt;/…</code> route.
            </p>
            <p className="text-sm text-muted-foreground">
              Get it from the server with{" "}
              <button
                type="button"
                onClick={async () => {
                  try {
                    await navigator.clipboard.writeText(
                      "osmedeus config view server.workspace_prefix_key --force"
                    );
                    toast.success("Command copied to clipboard");
                  } catch {
                    toast.error("Failed to copy");
                  }
                }}
                title="Click to copy"
                className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-xs cursor-pointer transition-colors hover:bg-muted/80"
              >
                osmedeus config view server.workspace_prefix_key --force
                <ClipboardIcon className="size-3" />
              </button>
            </p>
          </div>
          <Button onClick={saveApiConfig}>
            <SaveIcon className="mr-2 size-4" />
            Save API Configuration
          </Button>
        </CardContent>
      </Card>

      {/* Appearance */}
      <Card className="overflow-hidden">
        <SectionCardHeader
          icon={PaletteIcon}
          title="Appearance"
          description="Customize the dashboard theme"
          actions={
            <div className="flex items-center gap-2">
              <label className="text-sm font-medium">Theme Presets</label>
              <Select value={preset} onValueChange={savePreset}>
                <SelectTrigger className="min-w-[22rem] w-[22rem]">
                  <SelectValue placeholder="Select a preset" />
                </SelectTrigger>
                <SelectContent className="min-w-[22rem] max-h-64">
                  <SelectItem value="default">
                    <div className="flex items-center gap-2">
                      <span>Default</span>
                      <div className="flex items-center gap-0.5">
                        <span
                          className="size-2.5 rounded-full border"
                          style={{ backgroundColor: defaultThemeState.light?.primary }}
                        />
                        <span
                          className="size-2.5 rounded-full border"
                          style={{ backgroundColor: defaultThemeState.light?.secondary }}
                        />
                        <span
                          className="size-2.5 rounded-full border"
                          style={{ backgroundColor: defaultThemeState.dark?.primary }}
                        />
                        <span
                          className="size-2.5 rounded-full border"
                          style={{ backgroundColor: defaultThemeState.dark?.secondary }}
                        />
                      </div>
                    </div>
                  </SelectItem>
                  {Object.keys(presets).map((name) => (
                    <SelectItem key={name} value={name}>
                      <div className="flex items-center gap-2">
                        <span className="capitalize">{name}</span>
                        <div className="flex items-center gap-0.5">
                          <span
                            className="size-2.5 rounded-full border"
                            style={{ backgroundColor: presets[name].light?.primary }}
                          />
                          <span
                            className="size-2.5 rounded-full border"
                            style={{ backgroundColor: presets[name].light?.secondary }}
                          />
                          <span
                            className="size-2.5 rounded-full border"
                            style={{ backgroundColor: presets[name].dark?.primary }}
                          />
                          <span
                            className="size-2.5 rounded-full border"
                            style={{ backgroundColor: presets[name].dark?.secondary }}
                          />
                        </div>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          }
        />
        <CardContent className="space-y-4">
          {currentPreset && (
            <div className="flex flex-wrap gap-4">
              <div className="space-y-1">
                <span className="text-xs text-muted-foreground">Light</span>
                <div className="flex gap-1">
                  <div
                    className="size-8 rounded border"
                    style={{ backgroundColor: currentPreset.light?.primary }}
                    title="Primary"
                  />
                  <div
                    className="size-8 rounded border"
                    style={{ backgroundColor: currentPreset.light?.secondary }}
                    title="Secondary"
                  />
                </div>
              </div>
              <div className="space-y-1">
                <span className="text-xs text-muted-foreground">Dark</span>
                <div className="flex gap-1">
                  <div
                    className="size-8 rounded border"
                    style={{ backgroundColor: currentPreset.dark?.primary }}
                    title="Primary"
                  />
                  <div
                    className="size-8 rounded border"
                    style={{ backgroundColor: currentPreset.dark?.secondary }}
                    title="Secondary"
                  />
                </div>
              </div>
            </div>
          )}

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <label className="text-sm font-medium">Collapse sidebar by default</label>
              <p className="text-sm text-muted-foreground">
                Start with the sidebar collapsed when opening the dashboard
              </p>
            </div>
            <Switch
              checked={sidebarCollapsedByDefault}
              onCheckedChange={saveSidebarCollapsedByDefault}
            />
          </div>
        </CardContent>
      </Card>

      {/* Server Configuration */}
      <Card className="overflow-hidden">
        <SectionCardHeader
          icon={SettingsIcon}
          title="Server Configuration"
          description="View the server YAML configuration (read-only)"
          actions={
            <Button variant="outline" size="sm" onClick={copyYaml} disabled={!settingsYaml}>
              <ClipboardIcon className="mr-2 size-4" />
              Copy
            </Button>
          }
        />
        <CardContent className="space-y-4">
          <Collapsible open={serverConfigCliOpen} onOpenChange={setServerConfigCliOpen}>
            <div className="rounded-md border bg-muted/30 p-3">
              <CollapsibleTrigger asChild>
                <button type="button" className="w-full text-left">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <div className="text-sm font-medium">Update via CLI</div>
                      <div className="mt-1 text-sm text-muted-foreground">
                        Use osmedeus config CLI to change the configuration.
                      </div>
                    </div>
                    <ChevronDownIcon
                      className={`mt-0.5 size-4 shrink-0 text-muted-foreground transition-transform ${
                        serverConfigCliOpen ? "rotate-180" : ""
                      }`}
                    />
                  </div>
                </button>
              </CollapsibleTrigger>

              <CollapsibleContent>
                <pre className="mt-3 whitespace-pre-wrap font-mono text-xs leading-relaxed">
                  {serverConfigCliUsage}
                </pre>
              </CollapsibleContent>
            </div>
          </Collapsible>

          {loadingYaml ? (
            <div className="py-8 text-center text-sm text-muted-foreground">
              <LoaderIcon className="mx-auto mb-2 size-6 animate-spin" />
              Loading configuration...
            </div>
          ) : !settingsYaml ? (
            <div className="py-8 text-center text-sm text-muted-foreground">
              No configuration loaded.
            </div>
          ) : (
            <div className="overflow-auto rounded-md border bg-muted/30 p-4">
              <SyntaxHighlighter
                language="yaml"
                style={theme === "dark" ? atomOneDark : github}
                customStyle={{
                  margin: 0,
                  padding: 0,
                  background: "transparent",
                  fontSize: "0.8rem",
                }}
                showLineNumbers
              >
                {settingsYaml}
              </SyntaxHighlighter>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
