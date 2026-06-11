import { createRoute, lazyRouteComponent } from '@tanstack/react-router';
import { appLayoutRoute } from './_app';

/** Overview — the default landing surface. Code-split like every page. */
export const overviewRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/',
  component: lazyRouteComponent(() => import('@/features/overview/overview-page'), 'OverviewPage'),
});
