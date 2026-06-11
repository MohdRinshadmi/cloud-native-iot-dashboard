import { createRoute, Outlet, redirect } from '@tanstack/react-router';
import { rootRoute } from './__root';
import { AppShell } from '@/shared/components/layout/app-shell';
import { useAuthStore } from '@/stores/auth.store';

/**
 * Pathless layout route: wraps every dashboard page in the persistent shell
 * (sidebar + topbar + command palette).
 *
 * This single beforeLoad guard protects the ENTIRE dashboard surface: the
 * auth bootstrap (silent refresh) has already resolved by the time the router
 * mounts, so the store status is authoritative here.
 */
export const appLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'app',
  beforeLoad: ({ location }) => {
    if (useAuthStore.getState().status !== 'authenticated') {
      throw redirect({ to: '/login', search: { redirect: location.href } });
    }
  },
  component: () => (
    <AppShell>
      <Outlet />
    </AppShell>
  ),
});
