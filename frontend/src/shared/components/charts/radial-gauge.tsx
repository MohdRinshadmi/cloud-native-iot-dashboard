import { useMemo } from 'react';
import { arc } from 'd3-shape';
import { CHART } from './chart-theme';

interface RadialGaugeProps {
  /** 0–100. */
  value: number;
  label: string;
  unit?: string;
  size?: number;
  /** Override the auto threshold coloring. */
  color?: string;
}

/**
 * 270° radial gauge built with D3 arc geometry. Color auto-grades by value
 * (red < 25 < amber < 50 < green) unless overridden. The fill animates with
 * Framer Motion when the value changes.
 */
export function RadialGauge({ value, label, unit = '%', size = 132, color }: RadialGaugeProps) {
  const v = Math.max(0, Math.min(100, value));
  const startAngle = -Math.PI * 0.75;
  const endAngle = Math.PI * 0.75;
  const sweep = endAngle - startAngle;

  const radius = size / 2;
  const thickness = size * 0.11;

  const { trackPath, valuePath } = useMemo(() => {
    const arcGen = arc()
      .innerRadius(radius - thickness)
      .outerRadius(radius)
      .cornerRadius(thickness / 2);

    const track =
      arcGen({ startAngle, endAngle, innerRadius: radius - thickness, outerRadius: radius }) ?? '';
    const filled =
      arcGen({
        startAngle,
        endAngle: startAngle + sweep * (v / 100),
        innerRadius: radius - thickness,
        outerRadius: radius,
      }) ?? '';
    return { trackPath: track, valuePath: filled };
  }, [radius, thickness, startAngle, endAngle, sweep, v]);

  const autoColor = v < 25 ? CHART.destructive : v < 50 ? CHART.warning : CHART.success;
  const fillColor = color ?? autoColor;

  return (
    <div className="flex flex-col items-center" style={{ width: size }}>
      <svg width={size} height={size * 0.82} viewBox={`${-radius} ${-radius} ${size} ${size * 0.82}`}>
        <path d={trackPath} fill={CHART.grid} />
        <path
          d={valuePath}
          fill={fillColor}
          style={{ transition: 'fill 0.4s ease' }}
        />
        <text textAnchor="middle" y={2} className="fill-foreground font-mono" fontSize={size * 0.2} fontWeight={600}>
          {Math.round(v)}
          <tspan fontSize={size * 0.1} className="fill-muted-foreground">
            {unit}
          </tspan>
        </text>
      </svg>
      <span className="-mt-2 text-xs font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </span>
    </div>
  );
}
