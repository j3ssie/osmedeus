export interface Artifact {
  id: string;
  runId: string;
  workspace: string;
  name: string;
  artifactPath: string;
  artifactType: string;
  contentType: string;
  sizeBytes: number;
  lineCount: number;
  description: string;
  createdAt: Date;
}

