import { useQuery } from '@tanstack/react-query';
import { getHealth } from '@/services/api/health';

/** Query keys are centralised per-feature to avoid string drift. */
export const healthKeys = {
  all: ['health'] as const,
};

/**
 * Polls backend readiness every 10s. Used by the connection indicator to show
 * live API status in the UI — the foundation's proof-of-wiring.
 */
export function useHealth() {
  return useQuery({
    queryKey: healthKeys.all,
    queryFn: getHealth,
    refetchInterval: 10_000,
    retry: 1,
    staleTime: 5_000,
  });
}
