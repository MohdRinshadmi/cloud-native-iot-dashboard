import { api } from './client';
import type { Device, DeviceStatus, Paginated } from '@/types/api';

export interface DeviceListParams {
  q?: string;
  status?: DeviceStatus | '';
  limit?: number;
  offset?: number;
}

export function listDevices(params: DeviceListParams = {}): Promise<Paginated<Device>> {
  const qs = new URLSearchParams();
  if (params.q) qs.set('q', params.q);
  if (params.status) qs.set('status', params.status);
  qs.set('limit', String(params.limit ?? 50));
  qs.set('offset', String(params.offset ?? 0));
  return api.get<Paginated<Device>>(`/devices?${qs.toString()}`);
}

export function getDevice(id: string): Promise<Device> {
  return api.get<Device>(`/devices/${encodeURIComponent(id)}`);
}

export interface CreateDeviceInput {
  name: string;
  model?: string;
  firmware?: string;
}

export function createDevice(input: CreateDeviceInput): Promise<Device> {
  return api.post<Device>('/devices', { body: input });
}

export function deleteDevice(id: string): Promise<void> {
  return api.delete<void>(`/devices/${encodeURIComponent(id)}`);
}
