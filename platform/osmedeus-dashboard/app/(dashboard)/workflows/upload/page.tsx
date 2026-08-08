"use client";

import * as React from "react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { uploadWorkflowFile, uploadTargetsFile, type WorkflowUploadInfo, type UploadedFileInfo } from "@/lib/api/uploads";
import Link from "next/link";
import { toast } from "sonner";
import { LoaderIcon, FileCodeIcon, ArrowRightIcon } from "lucide-react";

export default function WorkflowUploadPage() {
  const [workflowFile, setWorkflowFile] = React.useState<File | null>(null);
  const [uploadingWorkflow, setUploadingWorkflow] = React.useState(false);
  const [workflowResult, setWorkflowResult] = React.useState<WorkflowUploadInfo | null>(null);
  const [targetsFile, setTargetsFile] = React.useState<File | null>(null);
  const [uploadingTargets, setUploadingTargets] = React.useState(false);
  const [targetsResult, setTargetsResult] = React.useState<UploadedFileInfo | null>(null);

  const submitWorkflow = async () => {
    if (!workflowFile) {
      toast.error("Please choose a YAML file");
      return;
    }
    setUploadingWorkflow(true);
    try {
      const info = await uploadWorkflowFile(workflowFile);
      setWorkflowResult(info);
      toast.success("Workflow uploaded", { description: `${info.name} (${info.kind})` });
    } catch (e) {
      toast.error("Upload failed", { description: e instanceof Error ? e.message : "" });
    } finally {
      setUploadingWorkflow(false);
    }
  };

  const submitTargets = async () => {
    if (!targetsFile) {
      toast.error("Please choose a targets file");
      return;
    }
    setUploadingTargets(true);
    try {
      const info = await uploadTargetsFile(targetsFile);
      setTargetsResult(info);
      toast.success("File uploaded", { description: info.filename });
    } catch (e) {
      toast.error("Upload failed", { description: e instanceof Error ? e.message : "" });
    } finally {
      setUploadingTargets(false);
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast.success("Copied to clipboard");
    } catch {
      toast.error("Failed to copy");
    }
  };

  return (
    <div className="space-y-4">
      
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Upload Workflow</CardTitle>
          <CardDescription>POST /osm/api/workflow-upload</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="workflow">Workflow YAML</Label>
            <Input id="workflow" type="file" accept=".yaml,.yml" onChange={(e) => setWorkflowFile(e.target.files?.[0] ?? null)} />
          </div>
          <Button onClick={submitWorkflow} disabled={uploadingWorkflow}>
            {uploadingWorkflow ? <LoaderIcon className="mr-2 size-4 animate-spin" /> : null}
            Upload
          </Button>
          {workflowResult ? (
            <div className="space-y-2 text-sm">
              <div className="flex items-center gap-2">
                <span>Uploaded:</span>
                <code className="px-1 rounded bg-muted">{workflowResult.name}</code>
                <span>({workflowResult.kind})</span>
                <Button variant="link" asChild>
                  <Link href={`/workflows-editor?workflow=${encodeURIComponent(workflowResult.name)}`}>
                    Open in Editor
                    <ArrowRightIcon className="ml-1 size-4" />
                  </Link>
                </Button>
              </div>
              {workflowResult.description ? (
                <div className="flex items-center gap-2">
                  <span>Description:</span>
                  <span className="text-muted-foreground">{workflowResult.description}</span>
                </div>
              ) : null}
              <div className="flex items-center gap-2">
                <span>Path:</span>
                <code className="px-1 rounded bg-muted break-all">{workflowResult.path}</code>
                <Button variant="secondary" size="sm" onClick={() => copyToClipboard(workflowResult.path)}>
                  Copy Path
                </Button>
              </div>
            </div>
          ) : null}
        </CardContent>
      </Card>
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Upload Input File</CardTitle>
          <CardDescription>POST /osm/api/upload-file</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="targets">Targets File</Label>
            <Input id="targets" type="file" accept=".txt" onChange={(e) => setTargetsFile(e.target.files?.[0] ?? null)} />
          </div>
          <Button onClick={submitTargets} disabled={uploadingTargets}>
            {uploadingTargets ? <LoaderIcon className="mr-2 size-4 animate-spin" /> : null}
            Upload
          </Button>
          {targetsResult ? (
            <div className="space-y-2 text-sm">
              <div className="flex items-center gap-2">
                <span>Uploaded:</span>
                <code className="px-1 rounded bg-muted">{targetsResult.filename}</code>
              </div>
              <div className="flex items-center gap-2">
                <span>Path:</span>
                <code className="px-1 rounded bg-muted break-all">{targetsResult.path}</code>
                <Button variant="secondary" size="sm" onClick={() => copyToClipboard(targetsResult.path)}>
                  Copy Path
                </Button>
              </div>
              <div className="flex items-center gap-4">
                {typeof targetsResult.size === "number" ? <span>Size: {targetsResult.size}</span> : null}
                {typeof targetsResult.lines === "number" ? <span>Lines: {targetsResult.lines}</span> : null}
              </div>
              <p className="text-muted-foreground">Use this path as target_file in scans.</p>
            </div>
          ) : null}
        </CardContent>
      </Card>
    </div>
  );
}
