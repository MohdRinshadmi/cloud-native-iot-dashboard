import { useMemo } from 'react';
import { scaleLinear } from 'd3-scale';
import { area, line, curveMonotoneX } from 'd3-shape';
import { extent } from 'd3-array';

interface SparklineProps {
  values: number[];
  width?: number;
  height?: number;
  color?: string;
  /** Render a soft gradient fill under the line. */
  fill?: boolean;
}

/**
 * Compact trend line. D3 owns the math (scales + path generation); React owns
 * the DOM (one declarative <svg>). This "D3 for geometry, React for render"
 * split avoids imperative DOM mutation and stays reconciler-friendly.
 */
export function Sparkline({
  values,
  width = 120,
  height = 36,
  color = 'hsl(190 96% 55%)',
  fill = true,
}: SparklineProps) {
  const { linePath, areaPath } = useMemo(() => {
    if (values.length < 2) return { linePath: '', areaPath: '' };

    const pad = 2;
    const x = scaleLinear()
      .domain([0, values.length - 1])
      .range([pad, width - pad]);
    const [lo, hi] = extent(values) as [number, number];
    const y = scaleLinear()
      .domain([lo, hi === lo ? lo + 1 : hi])
      .range([height - pad, pad]);

    const lineGen = line<number>()
      .x((_, i) => x(i))
      .y((d) => y(d))
      .curve(curveMonotoneX);
    const areaGen = area<number>()
      .x((_, i) => x(i))
      .y0(height)
      .y1((d) => y(d))
      .curve(curveMonotoneX);

    return { linePath: lineGen(values) ?? '', areaPath: areaGen(values) ?? '' };
  }, [values, width, height]);

  if (!linePath) {
    return <div style={{ width, height }} aria-hidden />;
  }

  const gradId = `spark-${color.replace(/[^a-z0-9]/gi, '')}`;
  return (
    <svg width={width} height={height} className="overflow-visible" role="img" aria-label="trend">
      {fill && (
        <>
          <defs>
            <linearGradient id={gradId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={color} stopOpacity={0.3} />
              <stop offset="100%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <path d={areaPath} fill={`url(#${gradId})`} />
        </>
      )}
      <path d={linePath} fill="none" stroke={color} strokeWidth={1.5} strokeLinecap="round" />
    </svg>
  );
}
