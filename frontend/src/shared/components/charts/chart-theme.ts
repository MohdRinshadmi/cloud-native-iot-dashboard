/**
 * Chart palette mirroring the CSS design tokens in index.css. Recharts/D3 can't
 * read CSS variables at render time, so the canonical hsl() values live here —
 * one place to keep charts on-brand with the dark enterprise theme.
 */
export const CHART = {
  primary: 'hsl(190 96% 55%)',
  success: 'hsl(152 58% 46%)',
  warning: 'hsl(35 92% 55%)',
  destructive: 'hsl(358 70% 56%)',
  violet: 'hsl(258 84% 70%)',
  muted: 'hsl(216 15% 58%)',
  grid: 'hsl(218 16% 20%)',
  axis: 'hsl(216 14% 52%)',
  surface: 'hsl(220 20% 9%)',
  border: 'hsl(218 16% 22%)',
} as const;

/** Per-metric color + label + unit, shared across every telemetry chart. */
export const METRIC_THEME = {
  temperature: { color: CHART.warning, label: 'Temperature', unit: '°C' },
  battery: { color: CHART.success, label: 'Battery', unit: '%' },
  voltage: { color: CHART.violet, label: 'Voltage', unit: 'V' },
  cpu: { color: CHART.primary, label: 'CPU', unit: '%' },
  memory: { color: 'hsl(168 68% 48%)', label: 'Memory', unit: '%' },
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
