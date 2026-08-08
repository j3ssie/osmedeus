import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";

export interface UploadedFileInfo {
  message: string;
  filename: string;
  path: string;
  size?: number;
  lines?: number;
}

export async function uploadTargetsFile(file: File): Promise<UploadedFileInfo> {
  if (isDemoMode()) {
    return {
      message: "Demo mode: file upload is a stub",
      filename: file.name,
      path: `/tmp/demo/${file.name}`,
      size: file.size,
    };
  }
  const form = new FormData();
  form.append("file", file);
  const res = await http.post(`${API_PREFIX}/upload-file`, form, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return res.data;
}

export interface WorkflowUploadInfo {
  message: string;
  name: string;
  kind: "flow" | "module";
  description?: string;
  path: string;
}

export async function uploadWorkflowFile(file: File): Promise<WorkflowUploadInfo> {
  if (isDemoMode()) {
    return {
      message: "Demo mode: workflow upload is a stub",
      name: file.name.replace(/\.(ya?ml)$/i, ""),
      kind: "flow",
      path: `/tmp/demo/${file.name}`,
    };
  }
  const form = new FormData();
  form.append("file", file);
  const res = await http.post(`${API_PREFIX}/workflow-upload`, form, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return res.data;
}
