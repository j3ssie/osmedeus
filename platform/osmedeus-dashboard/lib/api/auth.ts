import { http } from "./http";
import { API_PREFIX } from "@/lib/api/prefix";
import { isDemoMode } from "./demo-mode";

export async function login(username: string, password: string): Promise<string> {
  if (isDemoMode()) {
    const token = "mock-" + Buffer.from(username).toString("base64");
    return token;
  }
  const res = await http.post(`${API_PREFIX}/login`, { username, password });
  return res.data?.token as string;
}
