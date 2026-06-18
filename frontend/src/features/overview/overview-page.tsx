import { useEffect, useRef, useState } from 'react';
import { Activity, Bell, Cpu, Wifi } from 'lucide-react';
import { useRealtimeEvent } from '@/providers/realtime-provider';
import { useFleetSummary } from '@/hooks/use-fleet';
import { PageHeader } from '@/shared/components/layout/page-header';
import { Card, CardContent } from '@/shared/components/ui/card';
import { Sparkline } from '@/shared/components/charts/sparkline';
import { CHART } from '@/shared/components/charts/chart-theme';
import { SystemStatusCard } from '@/features/system-status/components/system-status-card';
import { ActivityFeed } from './components/activity-feed';

/** Fleet overview: live KPIs + event stream + platform status. */
export function OverviewPage() {
  const { data: summary } = useFleetSummary();
  const { perMin, trend } = useTelemetryRate();

  const kpis = [
    {
      label: 'Total devices',
      icon: Cpu,
      value: fmt(summary?.total),
      hint: 'registered in this workspace',
    },
    {
      label: 'Online now',
      icon: Wifi,
      value: fmt(summary?.online),
      hint: 'heartbeat within 90s',
      accent: 'text-success',
    },
    {
      label: 'Telemetry / min',
      icon: Activity,
      value: String(perMin),
      hint: 'live WebSocket stream',
      spark: trend,
    },
    {
      label: 'Degraded / offline',
      icon: Bell,
      value: summary ? String(summary.degraded + summary.offline) : '—',
      hint: 'needs attention',
      accent: summary && summary.degraded + summary.offline > 0 ? 'text-warning' : undefined,
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
              <div className="min-w-0">
                <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {kpi.label}
                </p>
                <p className={`mt-2 font-mono text-3xl font-semibold ${kpi.accent ?? ''}`}>
                  {kpi.value}
                </p>
                {kpi.spark && kpi.spark.length > 1 ? (
                  <div className="mt-2">
                    <Sparkline values={kpi.spark} color={CHART.primary} width={120} height={28} />
                  </div>
                ) : (
                  <p className="mt-1 text-xs text-muted-foreground">{kpi.hint}</p>
                )}
              </div>
              <span className="grid h-9 w-9 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
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

function fmt(n: number | undefined): string {
  return n === undefined ? '—' : String(n);
}

/**
 * Telemetry rate over a sliding 60s window, plus a per-2s-bucket trend series
 * for the sparkline. Counts WS frames directly.
 */
function useTelemetryRate(): { perMin: number; trend: number[] } {
  const stamps = useRef<number[]>([]);
  const pending = useRef(0);
  const [perMin, setPerMin] = useState(0);
  const [trend, setTrend] = useState<number[]>([]);

  useRealtimeEvent((e) => {
    if (e.type !== 'telemetry') return;
    pending.current += 1;
    const now = Date.now();
    stamps.current = [...stamps.current.filter((t) => now - t < 60_000), now];
    setPerMin(stamps.current.length);
  });

  useEffect(() => {
    const id = setInterval(() => {
      setTrend((prev) => [...prev, pending.current].slice(-40));
      pending.current = 0;
    }, 2000);
    return () => clearInterval(id);
  }, []);

  return { perMin, trend };
}
