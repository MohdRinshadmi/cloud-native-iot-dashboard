import { api } from './client';
import type { DeviceStatus, TelemetryReading } from '@/types/api';

export interface FleetSummary {
  total: number;
  online: number;
  offline: number;
  degraded: number;
  other: number;
  by_status: Record<DeviceStatus, number>;
}

export function getFleetSummary(): Promise<FleetSummary> {
  return api.get<FleetSummary>('/fleet/summary');
}

/** Telemetry history endpoint returns `{ data: [...] }`, newest-first. */
export function getTelemetryHistory(
  deviceId: string,
  minutes: number,
  limit = 1000,
): Promise<TelemetryReading[]> {
  return api
    .get<{ data: TelemetryReading[] }>(
      `/devices/${encodeURIComponent(deviceId)}/telemetry?minutes=${minutes}&limit=${limit}`,
    )
    .then((r) => r.data);
}
