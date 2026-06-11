import { createRouter } from '@tanstack/react-router';
import { rootRoute } from '@/routes/__root';
import { appLayoutRoute } from '@/routes/_app';
import { overviewRoute } from '@/routes/index';
import { devicesRoute, deviceDetailRoute } from '@/routes/devices';
import {
  alertsRoute,
  analyticsRoute,
  insightsRoute,
  mapRoute,
  settingsRoute,
} from '@/routes/pages';
import { NotFound, PageSkeleton, RouteError } from '@/shared/components/layout/route-states';

// Route tree: a single pathless layout route owns the shell; every page is a
// code-split child. Adding a page = one route def + one nav-config entry.
const routeTree = rootRoute.addChildren([
  appLayoutRoute.addChildren([
    overviewRoute,
    devicesRoute,
    deviceDetailRoute,
    analyticsRoute,
    mapRoute,
    alertsRoute,
    insightsRoute,
    settingsRoute,
  ]),
]);

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  scrollRestoration: true,
  defaultPendingComponent: PageSkeleton,
  defaultErrorComponent: RouteError,
  defaultNotFoundComponent: NotFound,
});

// Type-safe router: gives autocomplete + compile-time checks on every <Link>.
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
