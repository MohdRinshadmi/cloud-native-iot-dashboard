import { Activity, Bell, Cpu, Wifi } from 'lucide-react';
import { PageHeader } from '@/shared/components/layout/page-header';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { Badge } from '@/shared/components/ui/badge';
import { SystemStatusCard } from '@/features/system-status/components/system-status-card';

/**
 * Fleet overview. KPI tiles are structural placeholders until Phase 6 wires
 * real aggregates; the layout, hierarchy and motion are final.
 */
const KPI_TILES = [
  { label: 'Total devices', icon: Cpu, value: '—', hint: 'Awaiting device API (Phase 3)' },
  { label: 'Online now', icon: Wifi, value: '—', hint: 'Heartbeat pipeline (Phase 5)' },
  { label: 'Telemetry / min', icon: Activity, value: '—', hint: 'MQTT ingest (Phase 5)' },
  { label: 'Active alerts', icon: Bell, value: '—', hint: 'Alert engine (Phase 9)' },
];

export function OverviewPage() {
  return (
    <>
      <PageHeader
        title="Overview"
        description="Fleet health, live telemetry and operational KPIs at a glance."
        actions={<Badge variant="success">All systems nominal</Badge>}
      />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {KPI_TILES.map((kpi) => (
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
                <kpi.icon className="h-4.5 w-4.5" />
              </span>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="mt-6 grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <Card className="h-full">
            <CardHeader>
              <CardTitle>Live activity</CardTitle>
              <CardDescription>
                Real-time device events stream here once the WebSocket pipeline lands in Phase 5.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid h-48 place-items-center rounded-lg border border-dashed border-border text-sm text-muted-foreground">
                Waiting for telemetry…
              </div>
            </CardContent>
          </Card>
        </div>
        <SystemStatusCard />
      </div>
    </>
  );
}
