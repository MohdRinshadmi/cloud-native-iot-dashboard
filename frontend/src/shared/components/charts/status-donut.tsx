import { useMemo } from 'react';
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts';
import { STATUS_COLOR } from './chart-theme';
import { ChartTooltip } from './chart-tooltip';
import type { DeviceStatus } from '@/types/api';

interface StatusDonutProps {
  byStatus: Record<DeviceStatus, number>;
  total: number;
  height?: number;
}

const ORDER: DeviceStatus[] = ['online', 'degraded', 'offline', 'provisioning', 'decommissioned'];

/** Fleet status distribution as a donut, with a centered total (Recharts). */
export function StatusDonut({ byStatus, total, height = 240 }: StatusDonutProps) {
  const data = useMemo(
    () =>
      ORDER.map((status) => ({ name: status, value: byStatus[status] ?? 0 })).filter(
        (d) => d.value > 0,
      ),
    [byStatus],
  );

  return (
    <div className="relative" style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={data}
            dataKey="value"
            nameKey="name"
            innerRadius="62%"
            outerRadius="92%"
            paddingAngle={2}
            stroke="none"
            isAnimationActive={false}
          >
            {data.map((d) => (
              <Cell key={d.name} fill={STATUS_COLOR[d.name] ?? STATUS_COLOR.decommissioned} />
            ))}
          </Pie>
          <Tooltip content={<ChartTooltip />} />
        </PieChart>
      </ResponsiveContainer>
      <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
        <span className="font-mono text-3xl font-semibold">{total}</span>
        <span className="text-xs uppercase tracking-widest text-muted-foreground">devices</span>
      </div>
    </div>
  );
}
