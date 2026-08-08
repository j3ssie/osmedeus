/**
 * Demo mode utility for runtime switching between mock and real API
 *
 * Priority:
 * 1. localStorage.osmedeus_demo_mode (runtime toggle)
 * 2. process.env.NEXT_PUBLIC_USE_MOCK (build-time fallback)
 */

/**
 * Check if demo mode is enabled (runtime check)
 * Call this in API functions instead of checking env var directly
 */
export function isDemoMode(): boolean {
  // First check localStorage (runtime toggle)
  if (typeof window !== "undefined") {
    const stored = localStorage.getItem("osmedeus_demo_mode");
    if (stored !== null) {
      return stored === "true";
    }
  }
  // Fallback to env var (build-time setting)
  return process.env.NEXT_PUBLIC_USE_MOCK === "true";
}

/**
 * Set demo mode preference
 * Note: Page reload is recommended after calling this
 */
export function setDemoMode(enabled: boolean): void {
  if (typeof window !== "undefined") {
    localStorage.setItem("osmedeus_demo_mode", String(enabled));
  }
}

/**
 * Get raw demo mode preference from localStorage
 * Returns null if not set (use for display purposes)
 */
export function getDemoModePreference(): boolean | null {
  if (typeof window !== "undefined") {
    const stored = localStorage.getItem("osmedeus_demo_mode");
    if (stored !== null) {
      return stored === "true";
    }
  }
  return null;
}

/**
 * Clear demo mode preference (will fall back to env var)
 */
export function clearDemoModePreference(): void {
  if (typeof window !== "undefined") {
    localStorage.removeItem("osmedeus_demo_mode");
  }
}
