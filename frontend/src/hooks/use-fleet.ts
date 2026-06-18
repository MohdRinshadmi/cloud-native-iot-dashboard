import { useQuery } from '@tanstack/react-query';
import { getFleetSummary, getTelemetryHistory } from '@/services/api/fleet';

export const fleetKeys = {
  summary: ['fleet', 'summary'] as const,
  history: (deviceId: string, minutes: number) =>
    ['fleet', 'history', deviceId, minutes] as const,
};

/** Fleet status distribution; refetched periodically + on status WS events. */
export function useFleetSummary() {
  return useQuery({
    queryKey: fleetKeys.summary,
    queryFn: getFleetSummary,
    refetchInterval: 15_000,
  });
}

/** Telemetry history for one device over the last `minutes`, oldest-first. */
export function useTelemetryHistory(deviceId: string, minutes: number) {
  return useQuery({
    queryKey: fleetKeys.history(deviceId, minutes),
    queryFn: () => getTelemetryHistory(deviceId, minutes),
    enabled: deviceId.length > 0,
    // Server returns newest-first; charts read left→right, so reverse once here.
    select: (rows) => [...rows].reverse(),
  });
}
