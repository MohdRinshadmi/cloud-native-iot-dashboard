/**
 * Wire types mirroring the Go backend DTOs (internal/interfaces/http/handler).
 * One file = one source of truth for the API contract on the client.
 */

export type Role = 'admin' | 'operator' | 'viewer';

export interface AuthUser {
  id: string;
  tenant_id: string;
  email: string;
  name: string;
  role: Role;
}

export interface SessionResponse {
  access_token: string;
  expires_in: number; // seconds
  user: AuthUser;
}

export type DeviceStatus =
  | 'provisioning'
  | 'online'
  | 'offline'
  | 'degraded'
  | 'decommissioned';

export interface Device {
  id: string;
  name: string;
  model: string;
  firmware: string;
  status: DeviceStatus;
  last_seen_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Paginated<T> {
  data: T[];
  meta: { total: number; limit: number; offset: number };
}
