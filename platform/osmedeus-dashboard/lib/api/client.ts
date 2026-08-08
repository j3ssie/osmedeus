/**
 * API Client Configuration
 *
 * This file provides the base configuration for API calls.
 * Currently uses mock data, but can be easily replaced with real API endpoints.
 *
 * To replace with real APIs:
 * 1. Update BASE_URL to your API endpoint
 * 2. Add authentication headers in getHeaders()
 * 3. Replace mock implementations in individual API modules
 */

export const API_CONFIG = {
  // Change this to your real API URL when ready
  baseUrl: process.env.BASE_API_URL || process.env.NEXT_PUBLIC_API_URL || "/osm/api",

  // Simulated network delay for mock data (ms)
  mockDelay: 300,
};

export function getHeaders(): HeadersInit {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  // Add auth token if available
  if (typeof window !== "undefined") {
    const session = localStorage.getItem("osmedeus_session");
    if (session) {
      // In real implementation, use a proper auth token
      headers["Authorization"] = `Bearer mock-token`;
    }
  }

  return headers;
}

export async function delay(ms: number = API_CONFIG.mockDelay): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
