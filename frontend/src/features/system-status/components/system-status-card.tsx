import { motion } from 'framer-motion';
import { Activity, CheckCircle2, Loader2, XCircle } from 'lucide-react';
import { useHealth } from '@/hooks/use-health';
import { cn } from '@/shared/lib/cn';

/**
 * Live system-status card — the foundation's vertical slice. It exercises the
 * full stack: TanStack Query -> API client -> Gin /health -> rendered state.
 * When you see "API · up" here, every wiring layer is proven end-to-end.
 */
export function SystemStatusCard() {
  const { data, isLoading, isError } = useHealth();

  const state = isLoading
    ? 'loading'
    : isError || data?.status === 'down'
      ? 'down'
      : 'up';

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, ease: 'easeOut' }}
      className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-panel"
    >
      <div className="flex items-center gap-3">
        <span
          className={cn(
            'grid h-10 w-10 place-items-center rounded-lg',
            state === 'up' && 'bg-success/15 text-success animate-pulse-ring',
            state === 'down' && 'bg-destructive/15 text-destructive',
            state === 'loading' && 'bg-muted text-muted-foreground',
          )}
        >
          <Activity className="h-5 w-5" />
        </span>
        <div>
          <h2 className="text-sm font-medium text-muted-foreground">Backend connectivity</h2>
          <p className="text-lg font-semibold">Gin API · {label(state)}</p>
        </div>
      </div>

      <div className="mt-5 space-y-2 text-sm">
        <Row label="Status" value={<StatusBadge state={state} />} />
        <Row label="Version" value={<span className="font-mono">{data?.version ?? '—'}</span>} />
        <Row
          label="Uptime"
          value={
            <span className="font-mono">
              {data ? `${Math.round(data.uptime_seconds)}s` : '—'}
            </span>
          }
        />
      </div>
    </motion.div>
  );
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between border-b border-border/50 pb-2 last:border-0">
      <span className="text-muted-foreground">{label}</span>
      {value}
    </div>
  );
}

function StatusBadge({ state }: { state: 'up' | 'down' | 'loading' }) {
  if (state === 'loading')
    return (
      <span className="inline-flex items-center gap-1.5 text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" /> Checking
      </span>
    );
  if (state === 'down')
    return (
      <span className="inline-flex items-center gap-1.5 text-destructive">
        <XCircle className="h-4 w-4" /> Down
      </span>
    );
  return (
    <span className="inline-flex items-center gap-1.5 text-success">
      <CheckCircle2 className="h-4 w-4" /> Operational
    </span>
  );
}

function label(state: 'up' | 'down' | 'loading') {
  return state === 'up' ? 'connected' : state === 'down' ? 'unreachable' : 'connecting';
}
