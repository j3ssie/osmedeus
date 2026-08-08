import { NextResponse } from "next/server";

export const dynamic = "force-static";

type MockVulnerability = {
  id: number;
  workspace: string;
  vuln_info: string;
  vuln_title: string;
  vuln_desc: string;
  vuln_poc: string;
  severity: string;
  confidence?: string;
  asset_type: string;
  asset_value: string;
  tags: string[];
  detail_http_request?: string;
  detail_http_response?: string;
  raw_vuln_json?: string;
  created_at: string;
  updated_at: string;
};

export let mockVulnerabilities: MockVulnerability[] = [
  {
    id: 1,
    workspace: "example.com",
    vuln_info: "CVE-2024-1234",
    vuln_title: "SQL Injection in Login Form",
    vuln_desc: "The login form is vulnerable to SQL injection via the username parameter.",
    vuln_poc: "username=' OR '1'='1' --&password=test",
    severity: "critical",
    confidence: "Certain",
    asset_type: "web",
    asset_value: "https://example.com/login",
    tags: ["sqli", "owasp-top10", "authentication"],
    detail_http_request: "POST /login HTTP/1.1\nHost: example.com\n...",
    detail_http_response: "HTTP/1.1 200 OK\n...",
    raw_vuln_json: "{\"template\":\"sqli-login.yaml\"}",
    created_at: "2025-01-15T10:30:00Z",
    updated_at: "2025-01-15T10:30:00Z",
  },
  {
    id: 2,
    workspace: "example.com",
    vuln_info: "CVE-2024-5678",
    vuln_title: "Cross-Site Scripting (XSS) in Search",
    vuln_desc: "Reflected XSS vulnerability in the search functionality.",
    vuln_poc: "<script>alert('XSS')</script>",
    severity: "high",
    confidence: "Firm",
    asset_type: "web",
    asset_value: "https://example.com/search",
    tags: ["xss", "owasp-top10"],
    detail_http_request: "GET /search?q=<script>alert(1)</script> HTTP/1.1\n...",
    detail_http_response: "HTTP/1.1 200 OK\n...",
    raw_vuln_json: "{\"template\":\"xss-reflected.yaml\"}",
    created_at: "2025-01-15T10:31:00Z",
    updated_at: "2025-01-15T10:31:00Z",
  },
  {
    id: 3,
    workspace: "api.example.com",
    vuln_info: "",
    vuln_title: "Missing Security Headers",
    vuln_desc: "The response is missing recommended security headers.",
    vuln_poc: "curl -I https://api.example.com",
    severity: "low",
    confidence: "Tentative",
    asset_type: "api",
    asset_value: "https://api.example.com",
    tags: ["headers", "misconfiguration"],
    created_at: "2025-01-15T10:32:00Z",
    updated_at: "2025-01-15T10:32:00Z",
  },
  {
    id: 4,
    workspace: "staging.example.com",
    vuln_info: "",
    vuln_title: "Open Redirect",
    vuln_desc: "Unvalidated redirect parameter allows redirect to attacker-controlled site.",
    vuln_poc: "https://staging.example.com/redirect?next=https://evil.example",
    severity: "medium",
    confidence: "Manual Review Required",
    asset_type: "web",
    asset_value: "https://staging.example.com/redirect",
    tags: ["redirect", "owasp-top10"],
    created_at: "2025-01-15T10:33:00Z",
    updated_at: "2025-01-15T10:33:00Z",
  },
];

export function filterMockVulnerabilities(input: {
  workspace?: string | null;
  severity?: string | null;
  confidence?: string | null;
  assetValue?: string | null;
}): MockVulnerability[] {
  const normalizeConfidenceKey = (value: string): string =>
    value
      .trim()
      .toLowerCase()
      .replace(/[_-]+/g, " ")
      .replace(/\s+/g, " ");

  const workspace = input.workspace?.trim() || "";
  const severity = input.severity?.trim() || "";
  const confidence = input.confidence?.trim() || "";
  const assetValue = input.assetValue?.trim() || "";

  const severities = severity
    ? severity
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean)
    : [];

  const confidenceKeys = confidence
    ? confidence
        .split(",")
        .map((c) => normalizeConfidenceKey(c))
        .filter(Boolean)
    : [];
  const confidenceKeySet = new Set(confidenceKeys);

  return mockVulnerabilities.filter((v) => {
    if (workspace && v.workspace !== workspace) return false;
    if (severities.length && !severities.includes(v.severity)) return false;
    if (confidenceKeySet.size) {
      const vConfidenceKey = v.confidence ? normalizeConfidenceKey(v.confidence) : "";
      if (!confidenceKeySet.has(vConfidenceKey)) return false;
    }
    if (assetValue && !v.asset_value.includes(assetValue)) return false;
    return true;
  });
}

export async function GET() {
  const offset = 0;
  const limit = 20;

  const filtered = filterMockVulnerabilities({});

  const sliced = filtered.slice(offset, offset + limit);

  return NextResponse.json({
    data: sliced,
    pagination: {
      total: filtered.length,
      offset,
      limit,
    },
  });
}

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as Record<string, unknown> | null;
  const workspace = typeof body?.workspace === "string" ? body.workspace.trim() : "";
  if (!workspace) {
    return NextResponse.json(
      { error: true, message: "Workspace is required" },
      { status: 400 }
    );
  }

  const now = new Date().toISOString();
  const nextId =
    mockVulnerabilities.length === 0
      ? 1
      : Math.max(...mockVulnerabilities.map((v) => v.id)) + 1;

  const created: MockVulnerability = {
    id: nextId,
    workspace,
    vuln_info: typeof body?.vuln_info === "string" ? body.vuln_info : "",
    vuln_title: typeof body?.vuln_title === "string" ? body.vuln_title : "",
    vuln_desc: typeof body?.vuln_desc === "string" ? body.vuln_desc : "",
    vuln_poc: typeof body?.vuln_poc === "string" ? body.vuln_poc : "",
    severity: typeof body?.severity === "string" ? body.severity : "info",
    confidence: typeof body?.confidence === "string" ? body.confidence : undefined,
    asset_type: typeof body?.asset_type === "string" ? body.asset_type : "",
    asset_value: typeof body?.asset_value === "string" ? body.asset_value : "",
    tags: Array.isArray(body?.tags) ? (body?.tags as unknown[]).filter((t) => typeof t === "string") as string[] : [],
    detail_http_request:
      typeof body?.detail_http_request === "string" ? body.detail_http_request : undefined,
    detail_http_response:
      typeof body?.detail_http_response === "string" ? body.detail_http_response : undefined,
    raw_vuln_json: typeof body?.raw_vuln_json === "string" ? body.raw_vuln_json : undefined,
    created_at: now,
    updated_at: now,
  };

  mockVulnerabilities = [created, ...mockVulnerabilities];

  return NextResponse.json(
    { data: created, message: "Vulnerability created successfully" },
    { status: 201 }
  );
}
