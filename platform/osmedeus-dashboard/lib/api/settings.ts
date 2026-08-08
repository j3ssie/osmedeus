import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";

export async function getSettingsYaml(): Promise<string> {
  const res = await http.get(`${API_PREFIX}/settings/yaml`, { responseType: "text" });
  return res.data as string;
}

export async function updateSettingsYaml(yaml: string): Promise<{
  message: string;
  path: string;
  backup?: string;
}> {
  const res = await http.put(`${API_PREFIX}/settings/yaml`, yaml, {
    headers: { "Content-Type": "text/yaml" },
  });
  return res.data;
}
