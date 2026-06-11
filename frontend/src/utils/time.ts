const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });

const UNITS: Array<[Intl.RelativeTimeFormatUnit, number]> = [
  ['year', 365 * 24 * 3600],
  ['month', 30 * 24 * 3600],
  ['day', 24 * 3600],
  ['hour', 3600],
  ['minute', 60],
  ['second', 1],
];

/** "3 minutes ago" / "in 2 hours" from an ISO timestamp. */
export function formatRelative(iso: string | null | undefined): string {
  if (!iso) return 'never';
  const deltaSec = (new Date(iso).getTime() - Date.now()) / 1000;
  const abs = Math.abs(deltaSec);

  for (const [unit, sec] of UNITS) {
    if (abs >= sec || unit === 'second') {
      return rtf.format(Math.round(deltaSec / sec), unit);
    }
  }
  return 'now';
}

/** Locale-aware absolute timestamp for tooltips/detail views. */
export function formatAbsolute(iso: string | null | undefined): string {
  if (!iso) return '—';
  return new Date(iso).toLocaleString();
}
