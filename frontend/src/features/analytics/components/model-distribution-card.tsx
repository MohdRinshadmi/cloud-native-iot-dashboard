import { useMemo } from 'react';
import { Bar, BarChart, CartesianGrid, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { useDevices } from '@/hooks/use-devices';
import { CHART } from '@/shared/components/charts/chart-theme';
import { ChartTooltip } from '@/shared/components/charts/chart-tooltip';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { Skeleton } from '@/shared/components/ui/skeleton';

const BAR_COLORS = [CHART.primary, CHART.violet, CHART.success, CHART.warning, 'hsl(190 80% 50%)'];

/** Devices grouped by hardware model (computed client-side from the inventory). */
export function ModelDistributionCard() {
  const { data, isLoading } = useDevices({ limit: 200 });

  const bars = useMemo(() => {
    const counts = new Map<string, number>();
    data?.data.forEach((d) => {
      const model = d.model || 'unknown';
      counts.set(model, (counts.get(model) ?? 0) + 1);
    });
    return [...counts.entries()]
      .map(([model, count]) => ({ model, count }))
      .sort((a, b) => b.count - a.count);
  }, [data]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Devices by model</CardTitle>
        <CardDescription>Hardware composition of the fleet.</CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-[240px] w-full" />
        ) : (
          <ResponsiveContainer width="100%" height={240}>
            <BarChart data={bars} margin={{ top: 8, right: 8, bottom: 0, left: -20 }}>
              <CartesianGrid stroke={CHART.grid} strokeDasharray="3 3" vertical={false} />
              <XAxis dataKey="model" stroke={CHART.axis} fontSize={11} tickLine={false} />
              <YAxis stroke={CHART.axis} fontSize={11} tickLine={false} allowDecimals={false} />
              <Tooltip content={<ChartTooltip />} cursor={{ fill: 'hsl(217 33% 18% / 0.4)' }} />
              <Bar dataKey="count" name="Devices" radius={[4, 4, 0, 0]} isAnimationActive={false}>
                {bars.map((_, i) => (
                  <Cell key={i} fill={BAR_COLORS[i % BAR_COLORS.length]} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  );
}
