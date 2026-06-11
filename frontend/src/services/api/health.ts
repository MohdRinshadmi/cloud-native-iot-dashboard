import { api } from './client';

/** Mirrors the backend health.Report shape (internal/application/health). */
export interface HealthReport {
  status: 'up' | 'down';
  version: string;
  uptime_seconds: number;
  components: Record<string, { status: 'up' | 'down'; error?: string }>;
}

/** Calls GET /api/v1/health on the Gin backend. */
export function getHealth(): Promise<HealthReport> {
  return api.get<HealthReport>('/health');
}
