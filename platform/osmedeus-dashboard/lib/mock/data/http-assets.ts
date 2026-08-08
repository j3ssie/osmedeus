import type { HttpAsset } from "@/lib/types/asset";

// Generate mock HTTP assets for testing pagination
function generateMockHttpAssets(workspaceId: string, count: number): HttpAsset[] {
  const domains = [
    "api", "www", "mail", "admin", "dashboard", "app", "cdn", "static",
    "dev", "staging", "test", "beta", "secure", "auth", "login", "portal",
    "shop", "store", "blog", "docs", "support", "help", "status", "monitor"
  ];

  const titles = [
    "Welcome to Our API",
    "Admin Dashboard",
    "Login Portal",
    "Documentation",
    "Help Center",
    "Status Page",
    "Application Home",
    "Secure Portal",
    undefined,
    "Blog",
    "Store",
    "Developer Portal",
  ];

  const statusCodes = [200, 200, 200, 200, 301, 302, 403, 404, 500, 200, 200, 200];

  const assets: HttpAsset[] = [];

  for (let i = 0; i < count; i++) {
    const domain = domains[i % domains.length];
    const statusCode = statusCodes[i % statusCodes.length];
    const url = `https://${domain}.example.com${i > 0 ? `/path-${i}` : ""}`;

    assets.push({
      id: `asset-${workspaceId}-${i.toString().padStart(4, "0")}`,
      workspace: workspaceId,
      assetValue: `${domain}.example.com`,
      url,
      input: `${domain}.example.com`,
      scheme: "https",
      method: "GET",
      path: i > 0 ? `/path-${i}` : "/",
      statusCode,
      contentType: statusCode === 200 ? "text/html; charset=utf-8" : "",
      contentLength: Math.floor(Math.random() * 500000) + 1000,
      title: statusCode === 200 ? titles[i % titles.length] : undefined,
      words: Math.floor(Math.random() * 5000) + 100,
      lines: Math.floor(Math.random() * 500) + 10,
      hostIp: `192.168.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`,
      aRecords: [`192.168.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`],
      tls: "TLS 1.3",
      assetType: "web",
      technologies: statusCode === 200
        ? ["nginx", "React", "Node.js"].slice(0, Math.floor(Math.random() * 3) + 1)
        : [],
      responseTime: `${Math.floor(Math.random() * 500) + 50}ms`,
      source: "httpx",
      createdAt: new Date(Date.now() - Math.floor(Math.random() * 7 * 24 * 60 * 60 * 1000)),
      updatedAt: new Date(Date.now() - Math.floor(Math.random() * 3 * 24 * 60 * 60 * 1000)),
      lastSeenAt: new Date(Date.now() - Math.floor(Math.random() * 2 * 24 * 60 * 60 * 1000)),
    });
  }

  return assets;
}

// Pre-generate assets for each workspace
export const mockHttpAssets: Record<string, HttpAsset[]> = {
  "ws-001": generateMockHttpAssets("ws-001", 856),
  "ws-002": generateMockHttpAssets("ws-002", 189),
  "ws-003": generateMockHttpAssets("ws-003", 1432),
  "ws-004": generateMockHttpAssets("ws-004", 312),
  "ws-005": generateMockHttpAssets("ws-005", 45),
  "ws-006": generateMockHttpAssets("ws-006", 4521),
  "ws-007": generateMockHttpAssets("ws-007", 167),
};
