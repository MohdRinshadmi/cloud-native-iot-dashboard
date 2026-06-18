import { useMemo, useState } from 'react';
import { useTelemetryHistory } from '@/hooks/use-fleet';
import { useRealtimeEvent } from '@/providers/realtime-provider';
import { TelemetryChart } from '@/shared/components/charts/telemetry-chart';
import { METRIC_THEME, type MetricKey } from '@/shared/components/charts/chart-theme';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { Skeleton } from '@/shared/components/ui/skeleton';
import { cn } from '@/shared/lib/cn';
import type { TelemetryReading } from '@/types/api';

const RANGES = [
  { label: '15m', minutes: 15 },
  { label: '1h', minutes: 60 },
  { label: '6h', minutes: 360 },
  { label: '24h', minutes: 1440 },
];

const SELECTABLE: MetricKey[] = ['temperature', 'battery', 'voltage', 'cpu', 'memory', 'signal'];

/**
 * Telemetry time-series with a range picker and metric toggles. The fetched
 * history is merged with a live buffer of WebSocket readings, so the chart
 * keeps extending in real time without refetching.
 */
export function TelemetryHistorySection({ deviceId }: { deviceId: string }) {
  const [minutes, setMinutes] = useState(60);
  const [selected, setSelected] = useState<MetricKey[]>(['temperature', 'cpu']);
  const { data: history, isLoading } = useTelemetryHistory(deviceId, minutes);

  // Live buffer: readings that arrived since this view mounted.
  const [liveBuffer, setLiveBuffer] = useState<TelemetryReading[]>([]);
  useRealtimeEvent((e) => {
    if (e.type === 'telemetry' && e.data.device_id === deviceId) {
      setLiveBuffer((prev) => [...prev.slice(-600), e.data]);
    }
  });

  const readings = useMemo(() => mergeByTs(history ?? [], liveBuffer), [history, liveBuffer]);

  const toggle = (m: MetricKey) =>
    setSelected((prev) =>
      prev.includes(m) ? (prev.length > 1 ? prev.filter((x) => x !== m) : prev) : [...prev, m],
    );

  return (
    <Card>
      <CardHeader className="flex-row flex-wrap items-center justify-between gap-3 space-y-0">
        <CardTitle>Telemetry history</CardTitle>
        <div className="flex items-center gap-1 rounded-lg border border-border p-1">
          {RANGES.map((r) => (
            <button
              key={r.label}
              type="button"
              onClick={() => setMinutes(r.minutes)}
              className={cn(
                'rounded-md px-2.5 py-1 text-xs font-medium transition-colors',
                minutes === r.minutes
                  ? 'bg-primary/15 text-primary'
                  : 'text-muted-foreground hover:bg-accent',
              )}
            >
              {r.label}
            </button>
          ))}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap gap-2">
          {SELECTABLE.map((m) => {
            const on = selected.includes(m);
            return (
              <button
                key={m}
                type="button"
                onClick={() => toggle(m)}
                className={cn(
                  'flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs transition-colors',
                  on ? 'border-transparent text-foreground' : 'border-border text-muted-foreground',
                )}
                style={on ? { backgroundColor: `${METRIC_THEME[m].color}22` } : undefined}
              >
                <span
                  className="h-2 w-2 rounded-full"
                  style={{ backgroundColor: on ? METRIC_THEME[m].color : 'currentColor' }}
                />
                {METRIC_THEME[m].label}
              </button>
            );
          })}
        </div>

        {isLoading ? (
          <Skeleton className="h-[280px] w-full" />
        ) : (
          <TelemetryChart readings={readings} metrics={selected} />
        )}
      </CardContent>
    </Card>
  );
}

/** Merge fetched history with the live buffer, de-duped + sorted by timestamp. */
function mergeByTs(history: TelemetryReading[], live: TelemetryReading[]): TelemetryReading[] {
  if (live.length === 0) return history;
  const seen = new Set(history.map((r) => r.ts));
  const merged = [...history];
  for (const r of live) {
    if (!seen.has(r.ts)) merged.push(r);
  }
  return merged.sort((a, b) => new Date(a.ts).getTime() - new Date(b.ts).getTime());
}
