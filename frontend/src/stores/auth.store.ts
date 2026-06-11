import { create } from 'zustand';
import type { AuthUser, Role } from '@/types/api';

/**
 * Auth/session state.
 *
 * Security model:
 * - The ACCESS token lives in memory only (never localStorage) — an XSS
 *   payload can't exfiltrate it from storage, and it dies with the tab.
 * - The REFRESH token is an httpOnly cookie owned by the backend; JS can
 *   never read it. On reload we silently call /auth/refresh to re-establish
 *   the session (see bootstrapAuth in services/api/client).
 */
interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  status: 'unknown' | 'authenticated' | 'unauthenticated';
  setSession: (user: AuthUser, accessToken: string) => void;
  clear: () => void;
  hasRole: (...roles: Role[]) => boolean;
}

export const useAuthStore = create<AuthState>()((set, get) => ({
  user: null,
  accessToken: null,
  status: 'unknown', // resolved by the silent-refresh bootstrap at startup
  setSession: (user, accessToken) => set({ user, accessToken, status: 'authenticated' }),
  clear: () => set({ user: null, accessToken: null, status: 'unauthenticated' }),
  hasRole: (...roles) => {
    const role = get().user?.role;
    return role !== undefined && roles.includes(role);
  },
}));
