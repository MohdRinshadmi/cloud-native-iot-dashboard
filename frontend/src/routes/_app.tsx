import { createRoute, Outlet } from '@tanstack/react-router';
import { rootRoute } from './__root';
import { AppShell } from '@/shared/components/layout/app-shell';

/**
 * Pathless layout route: wraps every dashboard page in the persistent shell
 * (sidebar + topbar + command palette). Phase 4 adds the auth guard here —
 * one `beforeLoad` protects the entire app surface.
 */
export const appLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'app',
  component: () => (
    <AppShell>
      <Outlet />
    </AppShell>
  ),
});
