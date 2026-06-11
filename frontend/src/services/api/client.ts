import { env } from '@/shared/lib/env';

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
  /** Bearer token; injected by the auth layer (Phase 4). */
  token?: string;
}

/**
 * Thin, typed wrapper around fetch. Centralising transport here means retries,
 * auth headers, tracing and error shaping are defined once — feature code only
 * ever calls `api.get<T>(...)`.
 */
class ApiClient {
  constructor(private readonly baseUrl: string) {}

  private async request<T>(method: string, path: string, opts: RequestOptions = {}): Promise<T> {
    const { body, token, headers, ...rest } = opts;

    const finalHeaders = new Headers(headers);
    finalHeaders.set('Accept', 'application/json');
    if (body !== undefined) finalHeaders.set('Content-Type', 'application/json');
    if (token) finalHeaders.set('Authorization', `Bearer ${token}`);

    const init: RequestInit = { method, headers: finalHeaders, ...rest };
    if (body !== undefined) init.body = JSON.stringify(body);

    let res: Response;
    try {
      res = await fetch(`${this.baseUrl}${path}`, init);
    } catch {
      throw new ApiError(0, 'NETWORK', 'Network request failed', undefined);
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

export const api = new ApiClient(env.VITE_API_BASE_URL);
