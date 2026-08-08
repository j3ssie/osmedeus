export interface Schedule {
  id: string;
  name: string;
  workflowName: string;
  workflowPath?: string;
  triggerName?: string;
  triggerType?: "cron" | "event" | "watch" | "manual";
  schedule?: string;
  eventTopic?: string;
  watchPath?: string;
  inputConfig?: Record<string, any>;
  isEnabled: boolean;
  lastRun?: Date;
  nextRun?: Date;
  runCount?: number;
  createdAt?: Date;
  updatedAt?: Date;
}
