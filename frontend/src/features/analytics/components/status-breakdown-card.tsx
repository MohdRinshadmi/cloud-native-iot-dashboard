import { useFleetSummary } from '@/hooks/use-fleet';
import { StatusDonut } from '@/shared/components/charts/status-donut';
import { STATUS_COLOR } from '@/shared/components/charts/chart-theme';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { Skeleton } from '@/shared/components/ui/skeleton';
import type { DeviceStatus } from '@/types/api';

const LEGEND: DeviceStatus[] = ['online', 'degraded', 'offline', 'provisioning', 'decommissioned'];

/** Fleet status distribution: donut + labelled legend with live counts. */
export function StatusBreakdownCard() {
  const { data, isLoading } = useFleetSummary();

  return (
    <Card>
      <CardHeader>
        <CardTitle>Fleet status</CardTitle>
        <CardDescription>Distribution across all registered devices.</CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading || !data ? (
          <Skeleton className="h-[240px] w-full" />
        ) : (
          <>
            <StatusDonut byStatus={data.by_status} total={data.total} />
            <ul className="mt-4 grid grid-cols-2 gap-2 text-sm">
              {LEGEND.map((s) => (
                <li key={s} className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: STATUS_COLOR[s] }} />
                  <span className="capitalize text-muted-foreground">{s}</span>
                  <span className="ml-auto font-mono font-medium">{data.by_status[s] ?? 0}</span>
                </li>
              ))}
            </ul>
          </>
        )}
      </CardContent>
    </Card>
  );
}
