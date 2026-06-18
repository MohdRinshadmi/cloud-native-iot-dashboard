import type { TooltipProps } from 'recharts';
import { cn } from '@/shared/lib/cn';

/** Shared dark tooltip for Recharts charts; matches the card surface. */
export function ChartTooltip({ active, payload, label }: TooltipProps<number, string>) {
  if (!active || !payload?.length) return null;
  return (
    <div className="rounded-lg border border-border bg-card/95 px-3 py-2 text-xs shadow-xl backdrop-blur">
      {label !== undefined && (
        <p className="mb-1 font-medium text-foreground">{String(label)}</p>
      )}
      <ul className="space-y-0.5">
        {payload.map((entry) => (
          <li key={entry.dataKey} className="flex items-center gap-2">
            <span
              className={cn('h-2 w-2 rounded-full')}
              style={{ backgroundColor: entry.color }}
            />
            <span className="text-muted-foreground">{entry.name}</span>
            <span className="ml-auto font-mono font-medium text-foreground">
              {typeof entry.value === 'number' ? entry.value.toFixed(1) : entry.value}
              {entry.unit as string}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}
