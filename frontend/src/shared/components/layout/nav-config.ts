import {
  Activity,
  Bell,
  BrainCircuit,
  Cpu,
  LayoutDashboard,
  Map,
  Settings,
  type LucideIcon,
} from 'lucide-react';

/**
 * Single source of truth for primary navigation. The sidebar, breadcrumbs and
 * command palette all derive from this — add a route once, it appears
 * everywhere.
 */
export interface NavItem {
  label: string;
  to: string;
  icon: LucideIcon;
  description: string;
}

export interface NavSection {
  title: string;
  items: NavItem[];
}

export const NAV_SECTIONS: NavSection[] = [
  {
    title: 'Monitor',
    items: [
      { label: 'Overview', to: '/', icon: LayoutDashboard, description: 'Fleet KPIs and live status' },
      { label: 'Devices', to: '/devices', icon: Cpu, description: 'Device inventory and health' },
      { label: 'Map', to: '/map', icon: Map, description: 'Live GPS tracking and geofences' },
      { label: 'Alerts', to: '/alerts', icon: Bell, description: 'Anomalies, thresholds and history' },
    ],
  },
  {
    title: 'Intelligence',
    items: [
      { label: 'Analytics', to: '/analytics', icon: Activity, description: 'Historical and operational analytics' },
      { label: 'AI Insights', to: '/insights', icon: BrainCircuit, description: 'Anomaly detection and predictions' },
    ],
  },
  {
    title: 'System',
    items: [
      { label: 'Settings', to: '/settings', icon: Settings, description: 'Workspace and platform settings' },
    ],
  },
];

export const ALL_NAV_ITEMS: NavItem[] = NAV_SECTIONS.flatMap((s) => s.items);

/** Resolve a pathname to its nav item (longest-prefix match, '/' exact). */
export function findNavItem(pathname: string): NavItem | undefined {
  if (pathname === '/') return ALL_NAV_ITEMS.find((i) => i.to === '/');
  return ALL_NAV_ITEMS.filter((i) => i.to !== '/')
    .sort((a, b) => b.to.length - a.to.length)
    .find((i) => pathname.startsWith(i.to));
}
