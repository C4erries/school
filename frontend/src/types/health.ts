export interface HealthResponse {
  status: 'ok' | string;
  timestamp: string;
  service: string;
  version?: string;
}

export type HealthStatus = 'idle' | 'checking' | 'healthy' | 'unhealthy';

export interface HealthState {
  status: HealthStatus;
  data: HealthResponse | null;
  error: string | null;
  lastCheckedAt: Date | null;
  latencyMs: number | null;
}
