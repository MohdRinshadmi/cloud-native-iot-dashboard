import { env } from '@/shared/lib/env';
import { useAuthStore } from '@/stores/auth.store';
import type { SessionResponse } from '@/types/api';

/**
 * ApiError is the single error type surfaced by the client. Callers (and the
 * TanStack Query error boundary) can switch on `status` / `code` without
 * parsing strings.
 */
export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
    public readonly requestId?: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
  /** Skip Authorization header + 401 retry (used by auth endpoints). */
  skipAuth?: boolean;
}

/**
 * Thin, typed wrapper around fetch. Centralising transport here means auth
 * headers, silent token refresh, tracing and error shaping are defined once —
 * feature code only ever calls `api.get<T>(...)`.
 *
 * 401 handling: one transparent refresh attempt (deduped across concurrent
 * requests), then a single retry. A failed refresh clears the session and the
 * route guard bounces the user to /login.
 */
class ApiClient {
  constructor(private readonly baseUrl: string) {}

  private async request<T>(
    method: string,
    path: string,
    opts: RequestOptions = {},
    isRetry = false,
  ): Promise<T> {
    const { body, skipAuth, headers, ...rest } = opts;

    const finalHeaders = new Headers(headers);
    finalHeaders.set('Accept', 'application/json');
    if (body !== undefined) finalHeaders.set('Content-Type', 'application/json');
    if (!skipAuth) {
      const token = useAuthStore.getState().accessToken;
      if (token) finalHeaders.set('Authorization', `Bearer ${token}`);
    }

    const init: RequestInit = { method, headers: finalHeaders, ...rest };
    if (body !== undefined) init.body = JSON.stringify(body);

    let res: Response;
    try {
      res = await fetch(`${this.baseUrl}${path}`, init);
    } catch {
      throw new ApiError(0, 'NETWORK', 'Network request failed', undefined);
    }

    // Transparent session renewal on expiry.
    if (res.status === 401 && !skipAuth && !isRetry) {
      const renewed = await refreshSession();
      if (renewed) return this.request<T>(method, path, opts, true);
    }

    const requestId = res.headers.get('X-Request-ID') ?? undefined;

    if (res.status === 204) return undefined as T;

    const isJson = res.headers.get('Content-Type')?.includes('application/json');
    const payload: unknown = isJson ? await res.json() : await res.text();

    if (!res.ok) {
      const err = extractError(payload);
      throw new ApiError(res.status, err.code, err.message, requestId);
    }

    return payload as T;
  }

  get<T>(path: string, opts?: RequestOptions) {
    return this.request<T>('GET', path, opts);
  }
  post<T>(path: string, opts?: RequestOptions) {
    return this.request<T>('POST', path, opts);
  }
  put<T>(path: string, opts?: RequestOptions) {
    return this.request<T>('PUT', path, opts);
  }
  patch<T>(path: string, opts?: RequestOptions) {
    return this.request<T>('PATCH', path, opts);
  }
  delete<T>(path: string, opts?: RequestOptions) {
    return this.request<T>('DELETE', path, opts);
  }
}

/** Best-effort extraction of the backend's `{ error: { code, message } }` shape. */
function extractError(payload: unknown): { code: string; message: string } {
  if (payload && typeof payload === 'object' && 'error' in payload) {
    const e = (payload as { error: unknown }).error;
    if (e && typeof e === 'object') {
      const obj = e as Record<string, unknown>;
      return {
        code: typeof obj.code === 'string' ? obj.code : 'UNKNOWN',
        message: typeof obj.message === 'string' ? obj.message : 'Request failed',
      };
    }
  }
  return { code: 'UNKNOWN', message: 'Request failed' };
}

// ---- silent refresh ----------------------------------------------------------

let refreshInFlight: Promise<boolean> | null = null;

/**
 * Exchange the httpOnly refresh cookie for a new access token. Deduplicated:
 * concurrent 401s share one refresh call. Uses raw fetch (not ApiClient) to
 * avoid recursion.
 */
export function refreshSession(): Promise<boolean> {
  refreshInFlight ??= (async () => {
    try {
      const res = await fetch(`${env.VITE_API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: { Accept: 'application/json' },
      });
      if (!res.ok) {
        useAuthStore.getState().clear();
        return false;
      }
      const session = (await res.json()) as SessionResponse;
      useAuthStore.getState().setSession(session.user, session.access_token);
      return true;
    } catch {
      useAuthStore.getState().clear();
      return false;
    } finally {
      refreshInFlight = null;
    }
  })();
  return refreshInFlight;
}

/**
 * Resolve the initial auth state at app boot: try a silent refresh once.
 * Returns when the store's status is no longer 'unknown'.
 */
export async function bootstrapAuth(): Promise<void> {
  await refreshSession();
}

export const api = new ApiClient(env.VITE_API_BASE_URL);
