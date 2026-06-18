import { useMemo } from 'react';
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { CHART, METRIC_THEME, type MetricKey } from './chart-theme';
import { ChartTooltip } from './chart-tooltip';
import type { TelemetryReading } from '@/types/api';

interface TelemetryChartProps {
  readings: TelemetryReading[];
  metrics: MetricKey[];
  height?: number;
}

/**
 * Multi-series area chart over a telemetry window (Recharts). Gradient fills,
 * shared dark tooltip, time-formatted X axis. Pure presentational — the parent
 * owns data fetching + live-append, so this re-renders cheaply on new points.
 */
export function TelemetryChart({ readings, metrics, height = 280 }: TelemetryChartProps) {
  const data = useMemo(
    () =>
      readings.map((r) => ({
        ts: new Date(r.ts).getTime(),
        temperature: r.temperature,
        battery: r.battery,
        voltage: r.voltage,
        cpu: r.cpu,
        memory: r.memory,
        signal: r.signal,
      })),
    [readings],
  );

  if (data.length === 0) {
    return (
      <div
        style={{ height }}
        className="grid place-items-center rounded-lg border border-dashed border-border text-sm text-muted-foreground"
      >
        No telemetry in this window — start the simulator (make sim).
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={height}>
      <AreaChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: -16 }}>
        <defs>
          {metrics.map((m) => (
            <linearGradient key={m} id={`grad-${m}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={METRIC_THEME[m].color} stopOpacity={0.35} />
              <stop offset="100%" stopColor={METRIC_THEME[m].color} stopOpacity={0} />
            </linearGradient>
          ))}
        </defs>
        <CartesianGrid stroke={CHART.grid} strokeDasharray="3 3" vertical={false} />
        <XAxis
          dataKey="ts"
          type="number"
          scale="time"
          domain={['dataMin', 'dataMax']}
          tickFormatter={(t: number) =>
            new Date(t).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
          }
          stroke={CHART.axis}
          fontSize={11}
          tickLine={false}
          minTickGap={48}
        />
        <YAxis stroke={CHART.axis} fontSize={11} tickLine={false} width={44} />
        <Tooltip
          content={<ChartTooltip />}
          labelFormatter={(t) => new Date(Number(t)).toLocaleTimeString()}
        />
        {metrics.map((m) => (
          <Area
            key={m}
            type="monotone"
            dataKey={m}
            name={METRIC_THEME[m].label}
            unit={METRIC_THEME[m].unit}
            stroke={METRIC_THEME[m].color}
            strokeWidth={2}
            fill={`url(#grad-${m})`}
            connectNulls
            isAnimationActive={false}
            dot={false}
          />
        ))}
      </AreaChart>
    </ResponsiveContainer>
  );
}
