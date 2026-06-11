import { createRoute, lazyRouteComponent } from '@tanstack/react-router';
import { appLayoutRoute } from './_app';

export const devicesRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/devices',
  component: lazyRouteComponent(() => import('@/features/devices/devices-page'), 'DevicesPage'),
});

export const deviceDetailRoute = createRoute({
  getParentRoute: () => appLayoutRoute,
  path: '/devices/$deviceId',
  component: lazyRouteComponent(
    () => import('@/features/devices/device-detail-page'),
    'DeviceDetailPage',
  ),
});
