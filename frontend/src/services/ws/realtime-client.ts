import { env } from '@/shared/lib/env';
import type { RealtimeEvent } from '@/types/api';

export type RealtimeStatus = 'connecting' | 'open' | 'closed';

type EventListener = (e: RealtimeEvent) => void;
type StatusListener = (s: RealtimeStatus) => void;

/**
 * WebSocket abstraction with production behaviors:
 *
 * - RECONNECTION: exponential backoff with jitter (1s → 15s cap), reset on a
 *   successful open. The token getter is re-read on every attempt so a
 *   refreshed access token is picked up automatically.
 * - FAN-OUT: components subscribe/unsubscribe listeners; one socket serves the
 *   whole app.
 * - LIFECYCLE: close() stops reconnection (sign-out, unmount).
 */
export class RealtimeClient {
  private ws: WebSocket | null = null;
  private attempts = 0;
  private closedByUser = false;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  private eventListeners = new Set<EventListener>();
  private statusListeners = new Set<StatusListener>();
  private _status: RealtimeStatus = 'closed';

  constructor(private readonly getToken: () => string | null) {}

  get status(): RealtimeStatus {
    return this._status;
  }

  connect(): void {
    this.closedByUser = false;
    this.open();
  }

  close(): void {
    this.closedByUser = true;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.ws?.close();
    this.ws = null;
    this.setStatus('closed');
  }

  onEvent(fn: EventListener): () => void {
    this.eventListeners.add(fn);
    return () => this.eventListeners.delete(fn);
  }

  onStatus(fn: StatusListener): () => void {
    this.statusListeners.add(fn);
    fn(this._status);
    return () => this.statusListeners.delete(fn);
  }

  // ---- internals -------------------------------------------------------------

  private open(): void {
    const token = this.getToken();
    if (!token || this.closedByUser) return;

    this.setStatus('connecting');
    const ws = new WebSocket(`${env.VITE_WS_URL}?token=${encodeURIComponent(token)}`);
    this.ws = ws;

    ws.onopen = () => {
      this.attempts = 0;
      this.setStatus('open');
    };

    ws.onmessage = (msg: MessageEvent<string>) => {
      try {
        const event = JSON.parse(msg.data) as RealtimeEvent;
        this.eventListeners.forEach((fn) => fn(event));
      } catch {
        // Malformed frame — ignore rather than tear down the socket.
      }
    };

    ws.onclose = () => {
      this.setStatus('closed');
      this.scheduleReconnect();
    };

    ws.onerror = () => {
      // onclose always follows onerror; reconnect is handled there.
      ws.close();
    };
  }

  private scheduleReconnect(): void {
    if (this.closedByUser) return;
    this.attempts += 1;
    // Exponential backoff with full jitter, capped at 15s.
    const base = Math.min(1000 * 2 ** this.attempts, 15_000);
    const delay = Math.random() * base;
    this.reconnectTimer = setTimeout(() => this.open(), delay);
  }

  private setStatus(s: RealtimeStatus): void {
    if (this._status === s) return;
    this._status = s;
    this.statusListeners.forEach((fn) => fn(s));
  }
}
