import axios, { AxiosInstance, AxiosResponse } from "axios";
import { isDemoMode } from "./demo-mode";
import { getActiveOrgUUID, isOrgScopedPath } from "./active-org";
import { API_PREFIX } from "@/lib/api/prefix";

const MOCK_API_PREFIX = "/api/mock/api";

function resolveBaseURL(): string {
  if (isDemoMode()) return "";

  const normalize = (raw: string): string => {
    const trimmed = raw.trim().replace(/\/+$/, "");
    if (!trimmed) return "";
    if (trimmed === API_PREFIX) return "";
    if (trimmed.endsWith(API_PREFIX)) {
      return trimmed.slice(0, Math.max(0, trimmed.length - API_PREFIX.length)).replace(/\/+$/, "");
    }
    return trimmed;
  };

  if (typeof window !== "undefined") {
    const saved = localStorage.getItem("osmedeus_api_endpoint");
    if (saved) {
      const normalized = normalize(saved);
      if (normalized.startsWith("/")) return window.location.origin;
      return normalized;
    }
  }
  const envUrl = process.env.BASE_API_URL || process.env.NEXT_PUBLIC_API_URL;
  if (envUrl) {
    const normalized = normalize(envUrl);
    if (typeof window !== "undefined" && normalized.startsWith("/")) return window.location.origin;
    return normalized;
  }
  if (typeof window !== "undefined") return window.location.origin;
  return "";
}

export const http: AxiosInstance = axios.create({
  baseURL: resolveBaseURL(),
  headers: { "Content-Type": "application/json" },
  withCredentials: false,
});

http.interceptors.request.use((config) => {
  if (isDemoMode()) {
    config.baseURL = "";
    if (typeof config.url === "string") {
      if (config.url === API_PREFIX) {
        config.url = MOCK_API_PREFIX;
      } else if (config.url.startsWith(`${API_PREFIX}/`)) {
        config.url = `${MOCK_API_PREFIX}${config.url.slice(API_PREFIX.length)}`;
      }
    }
  }
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("osmedeus_token");
    if (token) {
      config.headers = config.headers || {};
      (config.headers as Record<string, string>)["Authorization"] = `Bearer ${token}`;
    }
  }

  // Scope org-aware list endpoints to the active org. Doing it here rather than
  // in each API module means every existing list page picks the filter up with
  // no change. An explicit `org` on the call always wins, and no active org
  // means no param at all - the response then spans every org.
  if (typeof config.url === "string" && config.url.startsWith(`${API_PREFIX}/`)) {
    const params = (config.params || {}) as Record<string, unknown>;
    if (typeof params.org === "undefined" && isOrgScopedPath(config.url.slice(API_PREFIX.length))) {
      const orgUUID = getActiveOrgUUID();
      if (orgUUID) {
        config.params = { ...params, org: orgUUID };
      }
    }
  }
  return config;
});

http.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    const status = error?.response?.status;
    const message =
      error?.response?.data?.message ||
      error?.message ||
      "Request failed";
    if (status === 401 && typeof window !== "undefined") {
      const msg = String(message).toLowerCase();
      if (msg.includes("api key")) {
        return Promise.reject(new Error(`${status || 0}:${message}`));
      }
      localStorage.setItem("osmedeus_force_logged_out", "true");
      localStorage.removeItem("osmedeus_token");
      localStorage.removeItem("osmedeus_session");
      try {
        document.cookie = "osmedeus_cookie=; Path=/; Max-Age=0; SameSite=Lax";
        document.cookie = "osmedeus_token=; Path=/; Max-Age=0; SameSite=Lax";
        document.cookie = "osmedeus_session=; Path=/; Max-Age=0; SameSite=Lax";
      } catch {}
      if (window.location.pathname !== "/login") {
        window.location.href = "/login";
      }
    }
    return Promise.reject(new Error(`${status || 0}:${message}`));
  }
);

export function getHttpBaseURL(): string {
  return (http.defaults.baseURL as string) || "";
}
