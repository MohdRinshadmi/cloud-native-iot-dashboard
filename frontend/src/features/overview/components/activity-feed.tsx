import { useMemo, useState } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import { Activity, Radio, WifiOff } from 'lucide-react';
import { useRealtimeEvent } from '@/providers/realtime-provider';
import { useDevices } from '@/hooks/use-devices';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { cn } from '@/shared/lib/cn';
import type { RealtimeEvent } from '@/types/api';

interface FeedEntry {
  id: number;
  kind: 'telemetry' | 'online' | 'offline';
  text: string;
  detail: string;
  at: Date;
}

const MAX_ENTRIES = 25;
let entrySeq = 0;

/** Live event stream — every WS frame becomes a human-readable feed row. */
export function ActivityFeed() {
  const [entries, setEntries] = useState<FeedEntry[]>([]);

  // Device names for friendly labels (events carry ids).
  const { data: devicePage } = useDevices({ limit: 200 });
  const nameById = useMemo(() => {
    const m = new Map<string, string>();
    devicePage?.data.forEach((d) => m.set(d.id, d.name));
    return m;
  }, [devicePage]);

  useRealtimeEvent((e: RealtimeEvent) => {
    const entry = toEntry(e, nameById);
    if (!entry) return;
    setEntries((prev) => [entry, ...prev].slice(0, MAX_ENTRIES));
  });

  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle>Live activity</CardTitle>
        <CardDescription>Telemetry and status events streaming over WebSocket.</CardDescription>
      </CardHeader>
      <CardContent>
        {entries.length === 0 ? (
          <div className="grid h-64 place-items-center rounded-lg border border-dashed border-border text-sm text-muted-foreground">
            Waiting for telemetry… (run <code className="mx-1 rounded bg-muted px-1.5 py-0.5 font-mono text-xs">make sim</code>)
          </div>
        ) : (
          <ul className="max-h-72 space-y-1 overflow-y-auto pr-1">
            <AnimatePresence initial={false}>
              {entries.map((entry) => (
                <motion.li
                  key={entry.id}
                  initial={{ opacity: 0, y: -8 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.2 }}
                  className="flex items-center gap-3 rounded-md px-2 py-1.5 text-sm hover:bg-accent/40"
                >
                  <span
                    className={cn(
                      'grid h-7 w-7 shrink-0 place-items-center rounded-md',
                      entry.kind === 'telemetry' && 'bg-primary/10 text-primary',
                      entry.kind === 'online' && 'bg-success/10 text-success',
                      entry.kind === 'offline' && 'bg-destructive/10 text-destructive',
                    )}
                  >
                    {entry.kind === 'telemetry' ? (
                      <Activity className="h-3.5 w-3.5" />
                    ) : entry.kind === 'online' ? (
                      <Radio className="h-3.5 w-3.5" />
                    ) : (
                      <WifiOff className="h-3.5 w-3.5" />
                    )}
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium">{entry.text}</p>
                    <p className="truncate text-xs text-muted-foreground">{entry.detail}</p>
                  </div>
                  <span className="shrink-0 font-mono text-[10px] text-muted-foreground">
                    {entry.at.toLocaleTimeString()}
                  </span>
                </motion.li>
              ))}
            </AnimatePresence>
          </ul>
        )}
      </CardContent>
    </Card>
  );
}

function toEntry(e: RealtimeEvent, names: Map<string, string>): FeedEntry | null {
  if (e.type === 'telemetry') {
    const name = names.get(e.data.device_id) ?? e.data.device_id.slice(0, 8);
    const parts: string[] = [];
    if (e.data.temperature !== undefined) parts.push(`${e.data.temperature.toFixed(1)}°C`);
    if (e.data.battery !== undefined) parts.push(`batt ${e.data.battery.toFixed(0)}%`);
    if (e.data.cpu !== undefined) parts.push(`cpu ${e.data.cpu.toFixed(0)}%`);
    if (e.data.signal !== undefined) parts.push(`${e.data.signal.toFixed(0)} dBm`);
    return {
      id: ++entrySeq,
      kind: 'telemetry',
      text: name,
      detail: parts.join(' · ') || 'telemetry received',
      at: new Date(),
    };
  }
  if (e.type === 'device_status') {
    const name = names.get(e.data.device_id) ?? e.data.device_id.slice(0, 8);
    const online = e.data.status === 'online';
    return {
      id: ++entrySeq,
      kind: online ? 'online' : 'offline',
      text: `${name} went ${e.data.status}`,
      detail: online ? 'heartbeat resumed' : 'connection lost',
      at: new Date(),
    };
  }
  return null;
}
