/**
 * Chart palette mirroring the CSS design tokens in index.css. Recharts/D3 can't
 * read CSS variables at render time, so the canonical hsl() values live here —
 * one place to keep charts on-brand with the dark enterprise theme.
 */
export const CHART = {
  primary: 'hsl(199 89% 56%)',
  success: 'hsl(142 71% 45%)',
  warning: 'hsl(38 92% 50%)',
  destructive: 'hsl(0 72% 51%)',
  violet: 'hsl(262 83% 65%)',
  muted: 'hsl(215 20% 55%)',
  grid: 'hsl(217 33% 20%)',
  axis: 'hsl(215 20% 50%)',
  surface: 'hsl(222 44% 10%)',
  border: 'hsl(217 33% 22%)',
} as const;

/** Per-metric color + label + unit, shared across every telemetry chart. */
export const METRIC_THEME = {
  temperature: { color: CHART.warning, label: 'Temperature', unit: '°C' },
  battery: { color: CHART.success, label: 'Battery', unit: '%' },
  voltage: { color: CHART.violet, label: 'Voltage', unit: 'V' },
  cpu: { color: CHART.primary, label: 'CPU', unit: '%' },
  memory: { color: 'hsl(190 80% 50%)', label: 'Memory', unit: '%' },
  signal: { color: CHART.muted, label: 'Signal', unit: 'dBm' },
} as const;

export type MetricKey = keyof typeof METRIC_THEME;

/** Status → color for donuts/pills/legends. */
export const STATUS_COLOR: Record<string, string> = {
  online: CHART.success,
  degraded: CHART.warning,
  offline: CHART.destructive,
  provisioning: CHART.primary,
  decommissioned: CHART.muted,
};
