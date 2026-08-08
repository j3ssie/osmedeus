export type EventStatus = "running" | "completed" | "failed";

export interface EventLog {
  id: string;
  workflowName: string;
  workflowKind?: "flow" | "module";
  target: string;
  status: EventStatus;
  createdAt?: Date;
  startedAt?: Date;
  completedAt?: Date;
  output?: string;
  error?: string;
}

