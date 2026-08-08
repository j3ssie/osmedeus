export interface EventLogItem {
  id: string;
  topic: string;
  eventId: string;
  name: string;
  source: string;
  dataType?: string;
  data?: string;
  workspace?: string;
  runId?: string;
  workflowName?: string;
  processed: boolean;
  createdAt: Date;
  processedAt?: Date;
  error?: string;
}

