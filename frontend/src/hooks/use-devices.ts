import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createDevice,
  deleteDevice,
  getDevice,
  listDevices,
  type CreateDeviceInput,
  type DeviceListParams,
} from '@/services/api/devices';

/** Centralised query keys for the devices feature. */
export const deviceKeys = {
  all: ['devices'] as const,
  list: (params: DeviceListParams) => ['devices', 'list', params] as const,
  detail: (id: string) => ['devices', 'detail', id] as const,
};

export function useDevices(params: DeviceListParams) {
  return useQuery({
    queryKey: deviceKeys.list(params),
    queryFn: () => listDevices(params),
    // Keep the previous page on-screen while the next one loads — no flicker.
    placeholderData: keepPreviousData,
  });
}

export function useDevice(id: string) {
  return useQuery({
    queryKey: deviceKeys.detail(id),
    queryFn: () => getDevice(id),
    enabled: id.length > 0,
  });
}

export function useCreateDevice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateDeviceInput) => createDevice(input),
    onSuccess: () => void qc.invalidateQueries({ queryKey: deviceKeys.all }),
  });
}

export function useDeleteDevice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteDevice(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: deviceKeys.all }),
  });
}
