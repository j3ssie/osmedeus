export type RouteMeta = { title: string; description?: string };

export function getRouteMeta(pathname: string): RouteMeta {
  const path = pathname || "/";
  if (path === "/") {
    return {
      title: "Statistics",
      description: "Overview of your security reconnaissance operations",
    };
  }
  if (path.startsWith("/assets/workspaces/")) {
    return {
      title: "Workspace",
      description: "Browse and manage assets in the selected workspace",
    };
  }
  if (path.startsWith("/workflows/")) {
    return {
      title: "Visualize and Manage your workflow",
    };
  }
  const map: Record<string, RouteMeta> = {
    "/workflows-editor": {
      title: "Workflow Editor",
      description: "Select and edit your workflows",
    },
    "/registry": {
      title: "Binary Registry",
      description: "View available tools and install from registry",
    },
    "/assets": {
      title: "Assets",
      description: "Browse and manage your discovered assets by workspace",
    },
    "/inventory": {
      title: "Inventory",
      description: "Explore workspaces and assets",
    },
    "/inventory/orgs": {
      title: "Orgs",
      description: "Group workspaces under an org to query across all of them",
    },
    "/inventory/workspaces": {
      title: "Workspaces Inventory",
      description: "Browse and manage workspaces",
    },
    "/inventory/assets": {
      title: "Assets Inventory",
      description: "Assets across workspaces",
    },
    "/inventory/artifacts": {
      title: "Artifacts Inventory",
      description: "Artifacts across workspaces",
    },
    "/schedules": {
      title: "Schedules",
      description: "Manage scheduled workflow executions",
    },
    "/events": {
      title: "Event Logs",
      description: "View running and completed task events",
    },
    "/utilities": {
      title: "Utilities Functions",
      description: "Browse and execute utility functions",
    },
    "/llm": {
      title: "LLM Playground",
      description: "LLM chat completions, embeddings, and tool calling",
    },
    "/workflows": {
      title: "Workflows",
      description: "Browse and edit your workflows",
    },
    "/scans": {
      title: "Scans",
      description: "View and manage your scans",
    },
    "/scans/new": {
      title: "New Scan",
      description: "Configure and start a new scan",
    },
    "/settings": {
      title: "Settings",
      description: "Manage profile and API configuration",
    },
    "/vulnerabilities": {
      title: "Vulnerabilities",
      description: "View and manage discovered vulnerabilities",
    },
  };
  if (path.startsWith("/inventory/workspaces/")) {
    return {
      title: "Workspace",
      description: "Browse and manage assets in the selected workspace",
    };
  }
  return map[path] ?? { title: "Osmedeus Dashboard" };
}
