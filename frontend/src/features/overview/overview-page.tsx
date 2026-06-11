import { useRef, useState } from 'react';
import { Activity, Bell, Cpu, Wifi } from 'lucide-react';
import { useRealtimeEvent } from '@/providers/realtime-provider';
import { useDevices } from '@/hooks/use-devices';
import { PageHeader } from '@/shared/components/layout/page-header';
import { Card, CardContent } from '@/shared/components/ui/card';
import { SystemStatusCard } from '@/features/system-status/components/system-status-card';
import { ActivityFeed } from './components/activity-feed';

/** Fleet overview: live KPIs + event stream + platform status. */
export function OverviewPage() {
  const { data: all } = useDevices({ limit: 1 });
  const { data: online } = useDevices({ status: 'online', limit: 1 });
  const telemetryPerMin = useTelemetryRate();

  const kpis = [
    {
      label: 'Total devices',
      icon: Cpu,
      value: format(all?.meta.total),
      hint: 'registered in this workspace',
    },
    {
      label: 'Online now',
      icon: Wifi,
      value: format(online?.meta.total),
      hint: 'heartbeat within 90s',
    },
    {
      label: 'Telemetry / min',
      icon: Activity,
      value: String(telemetryPerMin),
      hint: 'live WebSocket stream',
    },
    {
      label: 'Active alerts',
      icon: Bell,
      value: '—',
      hint: 'alert engine lands in Phase 9',
    },
  ];

  return (
    <>
      <PageHeader
        title="Overview"
        description="Fleet health, live telemetry and operational KPIs at a glance."
      />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {kpis.map((kpi) => (
          <Card key={kpi.label}>
            <CardContent className="flex items-start justify-between p-5">
              <div>
                <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {kpi.label}
                </p>
                <p className="mt-2 font-mono text-3xl font-semibold">{kpi.value}</p>
                <p className="mt-1 text-xs text-muted-foreground">{kpi.hint}</p>
              </div>
              <span className="grid h-9 w-9 place-items-center rounded-lg bg-primary/10 text-primary">
                <kpi.icon className="h-4 w-4" />
              </span>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="mt-6 grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <ActivityFeed />
        </div>
        <SystemStatusCard />
      </div>
    </>
  );
}

function format(n: number | undefined): string {
  return n === undefined ? '—' : String(n);
}

/** Counts telemetry events over a sliding 60s window, refreshed per event. */
function useTelemetryRate(): number {
  const stamps = useRef<number[]>([]);
  const [rate, setRate] = useState(0);

  useRealtimeEvent((e) => {
    if (e.type !== 'telemetry') return;
    const now = Date.now();
    stamps.current = [...stamps.current.filter((t) => now - t < 60_000), now];
    setRate(stamps.current.length);
  });

  return rate;
}
