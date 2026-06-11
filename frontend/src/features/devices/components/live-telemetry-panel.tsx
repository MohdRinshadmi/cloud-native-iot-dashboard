import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import {
  Battery,
  Cpu,
  Gauge,
  MemoryStick,
  Signal,
  Thermometer,
  type LucideIcon,
} from 'lucide-react';
import { useRealtimeEvent } from '@/providers/realtime-provider';
import { api } from '@/services/api/client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { Badge } from '@/shared/components/ui/badge';
import type { TelemetryReading } from '@/types/api';

interface MetricSpec {
  key: keyof Pick<TelemetryReading, 'temperature' | 'battery' | 'voltage' | 'cpu' | 'memory' | 'signal'>;
  label: string;
  unit: string;
  icon: LucideIcon;
  decimals: number;
}

const METRICS: MetricSpec[] = [
  { key: 'temperature', label: 'Temperature', unit: '°C', icon: Thermometer, decimals: 1 },
  { key: 'battery', label: 'Battery', unit: '%', icon: Battery, decimals: 0 },
  { key: 'voltage', label: 'Voltage', unit: 'V', icon: Gauge, decimals: 2 },
  { key: 'cpu', label: 'CPU', unit: '%', icon: Cpu, decimals: 0 },
  { key: 'memory', label: 'Memory', unit: '%', icon: MemoryStick, decimals: 0 },
  { key: 'signal', label: 'Signal', unit: 'dBm', icon: Signal, decimals: 0 },
];

/**
 * Real-time metric tiles for one device: seeded from the Redis-backed
 * `latest` endpoint, then updated in place by WebSocket telemetry frames.
 */
export function LiveTelemetryPanel({ deviceId }: { deviceId: string }) {
  const { data: seed } = useQuery({
    queryKey: ['telemetry', 'latest', deviceId],
    queryFn: () => api.get<TelemetryReading>(`/devices/${deviceId}/telemetry/latest`),
    retry: false, // 404 simply means "no data yet"
  });

  const [live, setLive] = useState<TelemetryReading | null>(null);
  useRealtimeEvent((e) => {
    if (e.type === 'telemetry' && e.data.device_id === deviceId) setLive(e.data);
  });

  const reading = live ?? seed ?? null;
  const isStreaming = live !== null;

  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between space-y-0">
        <div className="space-y-1.5">
          <CardTitle>Live telemetry</CardTitle>
          <CardDescription>
            {reading
              ? `Last reading ${new Date(reading.ts).toLocaleTimeString()}`
              : 'No telemetry yet — start the simulator (make sim).'}
          </CardDescription>
        </div>
        {isStreaming && <Badge variant="success">Streaming</Badge>}
      </CardHeader>
      <CardContent>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {METRICS.map((m) => {
            const value = reading?.[m.key];
            return (
              <div key={m.key} className="rounded-lg border border-border bg-background/40 p-4">
                <div className="flex items-center gap-2 text-muted-foreground">
                  <m.icon className="h-3.5 w-3.5" />
                  <span className="text-xs font-medium uppercase tracking-wider">{m.label}</span>
                </div>
                <motion.p
                  key={value} // re-animate when the value changes
                  initial={{ opacity: 0.4 }}
                  animate={{ opacity: 1 }}
                  className="mt-2 font-mono text-2xl font-semibold"
                >
                  {value !== undefined ? value.toFixed(m.decimals) : '—'}
                  <span className="ml-1 text-sm font-normal text-muted-foreground">{m.unit}</span>
                </motion.p>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
