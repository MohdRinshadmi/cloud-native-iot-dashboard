import { createRoute, lazyRouteComponent, redirect } from '@tanstack/react-router';
import { z } from 'zod';
import { rootRoute } from './__root';
import { useAuthStore } from '@/stores/auth.store';

const loginSearchSchema = z.object({
  redirect: z.string().optional(),
});

/** Sign-in route — outside the app shell. Already-authed users bounce home. */
export const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  validateSearch: loginSearchSchema,
  beforeLoad: () => {
    if (useAuthStore.getState().status === 'authenticated') {
      throw redirect({ to: '/' });
    }
  },
  component: lazyRouteComponent(() => import('@/features/auth/login-page'), 'LoginPage'),
});
