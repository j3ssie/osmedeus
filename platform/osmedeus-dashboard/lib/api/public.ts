import { http } from "./http";

export async function getServerInfo(): Promise<any> {
  const res = await http.get("/");
  return res.data;
}

export async function getHealth(): Promise<any> {
  const res = await http.get("/health");
  return res.data;
}

export async function getReadiness(): Promise<any> {
  const res = await http.get("/health/ready");
  return res.data;
}

