import { createRoute, lazyRouteComponent } from '@tanstack/react-router';
import { appLayoutRoute } from './_app';

/** Single-page routes — wiring only; implementations live in features/. */

export const analyticsRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/analytics',
  component: lazyRouteComponent(() => import('@/features/analytics/analytics-page'), 'AnalyticsPage'),
});

export const mapRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/map',
  component: lazyRouteComponent(() => import('@/features/map/map-page'), 'MapPage'),
});

export const alertsRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/alerts',
  component: lazyRouteComponent(() => import('@/features/alerts/alerts-page'), 'AlertsPage'),
});

export const insightsRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/insights',
  component: lazyRouteComponent(() => import('@/features/insights/insights-page'), 'InsightsPage'),
});

export const settingsRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/settings',
  component: lazyRouteComponent(() => import('@/features/settings/settings-page'), 'SettingsPage'),
});
