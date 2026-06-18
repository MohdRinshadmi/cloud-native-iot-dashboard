import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { RealtimeClient, type RealtimeStatus } from '@/services/ws/realtime-client';
import { useAuthStore } from '@/stores/auth.store';
import { deviceKeys } from '@/hooks/use-devices';
import type { RealtimeEvent } from '@/types/api';

interface RealtimeContextValue {
  status: RealtimeStatus;
  /** Subscribe to the raw event stream; returns an unsubscribe fn. */
  subscribe: (fn: (e: RealtimeEvent) => void) => () => void;
}

const RealtimeContext = createContext<RealtimeContextValue | null>(null);

/**
 * Owns the app's single WebSocket connection. Mounted inside the
 * authenticated layout, so a token always exists when it connects.
 *
 * Cross-cutting cache sync lives here: device_status events invalidate the
 * devices queries (throttled) so tables/detail views reflect reality without
 * each component wiring its own handler.
 */
export function RealtimeProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<RealtimeStatus>('closed');
  const queryClient = useQueryClient();
  const clientRef = useRef<RealtimeClient | null>(null);

  if (clientRef.current === null) {
    clientRef.current = new RealtimeClient(() => useAuthStore.getState().accessToken);
  }
  const client = clientRef.current;

  useEffect(() => {
    client.connect();
    const offStatus = client.onStatus(setStatus);

    // Throttled cache invalidation on status transitions.
    let lastInvalidate = 0;
    const offEvents = client.onEvent((e) => {
      if (e.type !== 'device_status') return;
      const now = Date.now();
      if (now - lastInvalidate > 2000) {
        lastInvalidate = now;
        void queryClient.invalidateQueries({ queryKey: deviceKeys.all });
        void queryClient.invalidateQueries({ queryKey: ['fleet', 'summary'] });
      }
    });

    return () => {
      offStatus();
      offEvents();
      client.close();
    };
  }, [client, queryClient]);

  const value = useMemo<RealtimeContextValue>(
    () => ({ status, subscribe: (fn) => client.onEvent(fn) }),
    [status, client],
  );

  return <RealtimeContext.Provider value={value}>{children}</RealtimeContext.Provider>;
}

/** Access the realtime connection (must be inside RealtimeProvider). */
export function useRealtime(): RealtimeContextValue {
  const ctx = useContext(RealtimeContext);
  if (!ctx) throw new Error('useRealtime must be used within RealtimeProvider');
  return ctx;
}

/** Subscribe to realtime events for the lifetime of the component. */
export function useRealtimeEvent(handler: (e: RealtimeEvent) => void): void {
  const { subscribe } = useRealtime();
  const handlerRef = useRef(handler);
  handlerRef.current = handler;

  useEffect(() => subscribe((e) => handlerRef.current(e)), [subscribe]);
}
