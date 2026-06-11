import { createRootRoute, Outlet } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';

// Router devtools are dev-only and code-split so they never ship to prod.
const RouterDevtools = import.meta.env.PROD
  ? () => null
  : lazy(() =>
      import('@tanstack/router-devtools').then((m) => ({ default: m.TanStackRouterDevtools })),
    );

/**
 * Root route — the application shell. Every page renders inside <Outlet/>.
 * The persistent chrome (sidebar, topbar) lands here in Phase 6.
 */
export const rootRoute = createRootRoute({
  component: () => (
    <>
      <Outlet />
      <Suspense>
        <RouterDevtools />
      </Suspense>
    </>
  ),
});
