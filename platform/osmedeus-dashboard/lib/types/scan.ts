export type ScanStatus = "pending" | "running" | "completed" | "failed" | "cancelled";
export type ScanPriority = "low" | "medium" | "high" | string;

export interface Scan {
  id: string;
  runId: string;
  runUuid: string;
  workflowName: string;
  workflowKind: "flow" | "module";
  target: string;
  params?: Record<string, unknown>;
  status: ScanStatus;
  priority?: ScanPriority;
  workspace?: string;
  workspacePath?: string;
  startedAt?: Date;
  completedAt?: Date;
  totalSteps: number;
  completedSteps: number;
  triggerType: string;
  triggerName?: string;
  runGroupId?: string;
  errorMessage?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface NewScanInput {
  workflowId: string;
  workflowKind?: "flow" | "module";
  workspaceId?: string;
  target?: string;
  empty_target?: boolean;
  targets?: string[];
  target_file?: string;
  concurrency?: number;
  threads_hold?: number;
  heuristics_check?: "basic" | "advanced" | string;
  repeat?: boolean;
  repeat_wait_time?: string;
  params?: Record<string, string>;
  priority?: "low" | "medium" | "high";
  timeout?: string;
  runner_type?: "local" | "docker" | "ssh";
  docker_image?: string;
  ssh_host?: string;
  schedule?: string;
  run_mode?: "local" | "distributed" | "cloud";
  cloud_provider?: "aws" | "gcp" | "digitalocean" | "linode" | "azure" | "hetzner";
  cloud_instances?: number;
  cloud_instance_type?: string;
  cloud_region?: string;
  cloud_auto_destroy?: boolean;
  cloud_reuse_infra?: string;
  cloud_use_spot?: boolean;
}

export type StepResultStatus =
  | "pending"
  | "running"
  | "completed"
  | "failed"
  | "skipped"
  | "success"
  | "error"
  | "unknown";

export interface StepResult {
  id: string;
  runId?: number | string;
  runUuid?: string;
  stepName: string;
  stepType: string;
  status: StepResultStatus | string;
  command?: string;
  output?: string;
  error?: string;
  durationMs?: number;
  startedAt?: Date;
  completedAt?: Date;
  createdAt?: Date;
}

export interface StepResultsResponse {
  data: StepResult[];
  pagination: {
    limit: number;
    offset: number;
    total: number;
  };
}
