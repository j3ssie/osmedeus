export interface SystemStats {
  workflows: {
    total: number;
    flows: number;
    modules: number;
  };
  scans: {
    total: number;
    completed: number;
    running: number;
    failed: number;
  };
  workspaces: {
    total: number;
  };
  assets: {
    total: number;
  };
  vulnerabilities: {
    total: number;
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
  schedules: {
    total: number;
    enabled: number;
  };
}
