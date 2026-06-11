import { createRouter } from '@tanstack/react-router';
import { rootRoute } from '@/routes/__root';
import { indexRoute } from '@/routes/index';

// Assemble the route tree. New feature routes register here (or move to
// file-based routing once the surface grows in Phase 6).
const routeTree = rootRoute.addChildren([indexRoute]);

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  scrollRestoration: true,
});

// Type-safe router: gives autocomplete + compile-time checks on every <Link>.
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
