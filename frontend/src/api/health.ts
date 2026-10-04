import type { HealthResponse } from '@/types/health';

export interface HealthCheckResult {
  data: HealthResponse;
  latencyMs: number;
}

export async function checkBackendHealth(): Promise<HealthCheckResult> {
  const startTime = performance.now();
  const response = await fetch('/api/v1/health', {
    headers: {
      Accept: 'application/json',
    },
    cache: 'no-cache',
  });

  const latencyMs = Math.round(performance.now() - startTime);

  if (!response.ok) {
    throw new Error(`Backend returned HTTP ${response.status}: ${response.statusText}`);
  }

  const data: HealthResponse = await response.json();
  return { data, latencyMs };
}
