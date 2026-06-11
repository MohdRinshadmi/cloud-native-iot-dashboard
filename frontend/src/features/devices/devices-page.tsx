import { useState } from 'react';
import { Link } from '@tanstack/react-router';
import { ChevronLeft, ChevronRight, Cpu, Search, Trash2 } from 'lucide-react';
import { useDevices, useDeleteDevice } from '@/hooks/use-devices';
import { useDebounce } from '@/hooks/use-debounce';
import { useAuthStore } from '@/stores/auth.store';
import { formatRelative } from '@/utils/time';
import { PageHeader } from '@/shared/components/layout/page-header';
import { Button } from '@/shared/components/ui/button';
import { Input } from '@/shared/components/ui/input';
import { Card } from '@/shared/components/ui/card';
import { Skeleton } from '@/shared/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/shared/components/ui/table';
import { DeviceStatusPill } from './components/device-status-pill';
import { AddDeviceDialog } from './components/add-device-dialog';
import type { DeviceStatus } from '@/types/api';
import { cn } from '@/shared/lib/cn';

const PAGE_SIZE = 10;

const STATUS_FILTERS: Array<{ label: string; value: DeviceStatus | '' }> = [
  { label: 'All', value: '' },
  { label: 'Online', value: 'online' },
  { label: 'Degraded', value: 'degraded' },
  { label: 'Offline', value: 'offline' },
  { label: 'Provisioning', value: 'provisioning' },
];

/** Live device inventory backed by GET /api/v1/devices. */
export function DevicesPage() {
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState<DeviceStatus | ''>('');
  const [offset, setOffset] = useState(0);
  const q = useDebounce(search);

  const { data, isLoading } = useDevices({ q, status, limit: PAGE_SIZE, offset });
  const deleteDevice = useDeleteDevice();
  const canManage = useAuthStore((s) => s.hasRole('admin', 'operator'));
  const isAdmin = useAuthStore((s) => s.hasRole('admin'));

  const total = data?.meta.total ?? 0;
  const from = total === 0 ? 0 : offset + 1;
  const to = Math.min(offset + PAGE_SIZE, total);

  return (
    <>
      <PageHeader
        title="Devices"
        description="Inventory, connectivity and health for every registered device."
        actions={<AddDeviceDialog disabled={!canManage} />}
      />

      <div className="mb-4 flex flex-wrap items-center gap-3">
        <div className="relative max-w-sm flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Filter by name or model…"
            className="pl-9"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setOffset(0);
            }}
          />
        </div>
        <div className="flex items-center gap-1 rounded-lg border border-border p-1">
          {STATUS_FILTERS.map((f) => (
            <button
              key={f.label}
              type="button"
              onClick={() => {
                setStatus(f.value);
                setOffset(0);
              }}
              className={cn(
                'rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
                status === f.value
                  ? 'bg-primary/15 text-primary'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
              )}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Device</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Model</TableHead>
              <TableHead>Firmware</TableHead>
              <TableHead>Last seen</TableHead>
              {isAdmin && <TableHead className="text-right">Actions</TableHead>}
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading &&
              Array.from({ length: 5 }, (_, i) => (
                <TableRow key={i}>
                  <TableCell colSpan={isAdmin ? 6 : 5}>
                    <Skeleton className="h-6 w-full" />
                  </TableCell>
                </TableRow>
              ))}

            {!isLoading && data?.data.length === 0 && (
              <TableRow>
                <TableCell colSpan={isAdmin ? 6 : 5}>
                  <div className="flex flex-col items-center gap-3 py-14 text-center">
                    <span className="grid h-12 w-12 place-items-center rounded-full bg-muted text-muted-foreground">
                      <Cpu className="h-6 w-6" />
                    </span>
                    <div>
                      <p className="font-medium">No devices found</p>
                      <p className="mt-1 text-sm text-muted-foreground">
                        {q || status
                          ? 'Try adjusting your search or status filter.'
                          : 'Register your first device to start streaming telemetry.'}
                      </p>
                    </div>
                  </div>
                </TableCell>
              </TableRow>
            )}

            {data?.data.map((d) => (
              <TableRow key={d.id}>
                <TableCell>
                  <Link
                    to="/devices/$deviceId"
                    params={{ deviceId: d.id }}
                    className="group block"
                  >
                    <p className="font-medium transition-colors group-hover:text-primary">
                      {d.name}
                    </p>
                    <p className="font-mono text-[11px] text-muted-foreground">
                      {d.id.slice(0, 8)}
                    </p>
                  </Link>
                </TableCell>
                <TableCell>
                  <DeviceStatusPill status={d.status} />
                </TableCell>
                <TableCell className="text-muted-foreground">{d.model || '—'}</TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {d.firmware || '—'}
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {formatRelative(d.last_seen_at)}
                </TableCell>
                {isAdmin && (
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={`Delete ${d.name}`}
                      className="h-8 w-8 text-muted-foreground hover:text-destructive"
                      disabled={deleteDevice.isPending}
                      onClick={() => {
                        if (window.confirm(`Delete device "${d.name}"? This cannot be undone.`)) {
                          deleteDevice.mutate(d.id);
                        }
                      }}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </TableCell>
                )}
              </TableRow>
            ))}
          </TableBody>
        </Table>

        <div className="flex items-center justify-between border-t border-border px-4 py-3 text-sm text-muted-foreground">
          <span>
            {from}–{to} of {total}
          </span>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={offset === 0}
              onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
            >
              <ChevronLeft /> Prev
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={to >= total}
              onClick={() => setOffset(offset + PAGE_SIZE)}
            >
              Next <ChevronRight />
            </Button>
          </div>
        </div>
      </Card>
    </>
  );
}
