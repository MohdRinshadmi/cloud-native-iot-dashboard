import { useHealth } from '@/hooks/use-health';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/components/ui/tooltip';
import { cn } from '@/shared/lib/cn';

/**
 * Live API connectivity dot in the topbar — green/amber/red with details on
 * hover. Backed by the polling health query; upgraded to WebSocket state in
 * Phase 5.
 */
export function ConnectionIndicator() {
  const { data, isLoading, isError } = useHealth();

  const state = isLoading ? 'connecting' : isError || data?.status === 'down' ? 'down' : 'up';

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          type="button"
          aria-label={`API status: ${state}`}
          className="flex items-center gap-2 rounded-md border border-border px-2.5 py-1.5 text-xs text-muted-foreground transition-colors hover:bg-accent"
        >
          <span
            className={cn(
              'h-2 w-2 rounded-full',
              state === 'up' && 'bg-success animate-pulse-ring',
              state === 'down' && 'bg-destructive',
              state === 'connecting' && 'bg-warning animate-pulse',
            )}
          />
          <span className="hidden sm:inline">
            {state === 'up' ? 'Live' : state === 'down' ? 'Offline' : 'Connecting'}
          </span>
        </button>
      </TooltipTrigger>
      <TooltipContent side="bottom" className="space-y-1">
        <p className="font-medium">API {state === 'up' ? 'connected' : state === 'down' ? 'unreachable' : 'connecting…'}</p>
        {data && (
          <p className="font-mono text-[10px] text-muted-foreground">
            v{data.version} · up {Math.round(data.uptime_seconds)}s
          </p>
        )}
      </TooltipContent>
    </Tooltip>
  );
}
