import { useHealth } from '@/hooks/use-health';
import { useRealtime } from '@/providers/realtime-provider';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/components/ui/tooltip';
import { cn } from '@/shared/lib/cn';

/**
 * Connection truth in the topbar, two layers deep:
 *  - "Live"      → WebSocket open (streaming events)
 *  - "Degraded"  → WS down but the REST API answers (polling still works)
 *  - "Offline"   → nothing reachable
 */
export function ConnectionIndicator() {
  const ws = useRealtime();
  const { data, isError } = useHealth();

  const apiUp = !isError && data?.status === 'up';
  const state = ws.status === 'open' ? 'live' : apiUp ? 'degraded' : 'offline';

  const labels = {
    live: { text: 'Live', detail: 'WebSocket stream connected' },
    degraded: { text: 'Degraded', detail: 'REST reachable, realtime stream reconnecting…' },
    offline: { text: 'Offline', detail: 'API unreachable' },
  } as const;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          type="button"
          aria-label={`Connection: ${labels[state].text}`}
          className="flex items-center gap-2 rounded-md border border-border bg-elevated/40 px-2.5 py-1.5 text-xs text-muted-foreground transition-colors hover:border-border/70 hover:bg-accent hover:text-foreground"
        >
          <span
            className={cn(
              'h-2 w-2 rounded-full',
              state === 'live' && 'bg-success animate-pulse-ring',
              state === 'degraded' && 'bg-warning animate-pulse',
              state === 'offline' && 'bg-destructive',
            )}
          />
          <span className="hidden sm:inline">{labels[state].text}</span>
        </button>
      </TooltipTrigger>
      <TooltipContent side="bottom" className="space-y-1">
        <p className="font-medium">{labels[state].detail}</p>
        {data && (
          <p className="font-mono text-[10px] text-muted-foreground">
            api v{data.version} · up {Math.round(data.uptime_seconds)}s
          </p>
        )}
      </TooltipContent>
    </Tooltip>
  );
}
