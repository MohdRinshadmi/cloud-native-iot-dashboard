import { Badge } from '@/shared/components/ui/badge';
import type { DeviceStatus } from '@/types/api';
import { cn } from '@/shared/lib/cn';

const STATUS_CONFIG: Record<
  DeviceStatus,
  { label: string; variant: 'success' | 'warning' | 'destructive' | 'secondary' | 'outline'; dot: string }
> = {
  online: { label: 'Online', variant: 'success', dot: 'bg-success' },
  degraded: { label: 'Degraded', variant: 'warning', dot: 'bg-warning' },
  offline: { label: 'Offline', variant: 'destructive', dot: 'bg-destructive' },
  provisioning: { label: 'Provisioning', variant: 'secondary', dot: 'bg-muted-foreground' },
  decommissioned: { label: 'Decommissioned', variant: 'outline', dot: 'bg-muted-foreground' },
};

/** Status badge with a live-feeling pulse dot for online devices. */
export function DeviceStatusPill({ status }: { status: DeviceStatus }) {
  const cfg = STATUS_CONFIG[status];
  return (
    <Badge variant={cfg.variant}>
      <span className={cn('h-1.5 w-1.5 rounded-full', cfg.dot, status === 'online' && 'animate-pulse')} />
      {cfg.label}
    </Badge>
  );
}
