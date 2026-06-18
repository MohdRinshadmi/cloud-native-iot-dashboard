import { useEffect, useRef, useState } from 'react';
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { useRealtimeEvent } from '@/providers/realtime-provider';
import { CHART } from '@/shared/components/charts/chart-theme';
import { ChartTooltip } from '@/shared/components/charts/chart-tooltip';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { Badge } from '@/shared/components/ui/badge';

const BUCKET_MS = 2000; // 2s buckets
const WINDOW = 60; // keep 60 buckets (~2 min)

interface Bucket {
  t: number;
  count: number;
}

/**
 * Live ingest throughput: telemetry events bucketed into 2s windows and
 * charted as a rolling area. Counts arriving WS frames directly, so it
 * reflects real pipeline volume this session.
 */
export function ThroughputCard() {
  const [buckets, setBuckets] = useState<Bucket[]>([]);
  const pending = useRef(0);

  useRealtimeEvent((e) => {
    if (e.type === 'telemetry') pending.current += 1;
  });

  useEffect(() => {
    const id = setInterval(() => {
      setBuckets((prev) => {
        const next = [...prev, { t: Date.now(), count: pending.current }];
        pending.current = 0;
        return next.slice(-WINDOW);
      });
    }, BUCKET_MS);
    return () => clearInterval(id);
  }, []);

  const perMin = buckets.slice(-30).reduce((s, b) => s + b.count, 0) * (60_000 / (30 * BUCKET_MS));

  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between space-y-0">
        <div className="space-y-1.5">
          <CardTitle>Ingest throughput</CardTitle>
          <CardDescription>Telemetry events received over WebSocket (2s buckets).</CardDescription>
        </div>
        <Badge variant="default">{Math.round(perMin)}/min</Badge>
      </CardHeader>
      <CardContent>
        {buckets.length < 2 ? (
          <div className="grid h-[200px] place-items-center rounded-lg border border-dashed border-border text-sm text-muted-foreground">
            Listening for telemetry… (run <code className="mx-1 rounded bg-muted px-1.5 py-0.5 font-mono text-xs">make sim</code>)
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={200}>
            <AreaChart data={buckets} margin={{ top: 8, right: 8, bottom: 0, left: -24 }}>
              <defs>
                <linearGradient id="grad-throughput" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor={CHART.primary} stopOpacity={0.4} />
                  <stop offset="100%" stopColor={CHART.primary} stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid stroke={CHART.grid} strokeDasharray="3 3" vertical={false} />
              <XAxis
                dataKey="t"
                tickFormatter={(t: number) =>
                  new Date(t).toLocaleTimeString([], { minute: '2-digit', second: '2-digit' })
                }
                stroke={CHART.axis}
                fontSize={11}
                tickLine={false}
                minTickGap={40}
              />
              <YAxis stroke={CHART.axis} fontSize={11} tickLine={false} allowDecimals={false} width={40} />
              <Tooltip
                content={<ChartTooltip />}
                labelFormatter={(t) => new Date(Number(t)).toLocaleTimeString()}
              />
              <Area
                type="monotone"
                dataKey="count"
                name="Events"
                stroke={CHART.primary}
                strokeWidth={2}
                fill="url(#grad-throughput)"
                isAnimationActive={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  );
}
