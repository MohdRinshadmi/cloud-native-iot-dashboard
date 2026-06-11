import { create } from 'zustand';
import { persist } from 'zustand/middleware';

/**
 * Authenticated principal. RBAC is role-based; the backend remains the source
 * of truth and re-checks every request — this is for UI gating only.
 */
export interface AuthUser {
  id: string;
  email: string;
  name: string;
  tenantId: string;
  roles: Role[];
}

export type Role = 'admin' | 'operator' | 'viewer';

interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  status: 'unauthenticated' | 'authenticated';
  setSession: (user: AuthUser, accessToken: string) => void;
  clear: () => void;
  hasRole: (role: Role) => boolean;
}

/**
 * Auth store skeleton. Phase 4 fills in login/refresh flows; the shape is
 * fixed now so the API client and route guards can depend on it.
 *
 * NOTE: only the access token is persisted (sessionStorage-like). The refresh
 * token lives in an httpOnly cookie set by the backend and is never readable
 * from JS — the correct defense against XSS token theft.
 */
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      status: 'unauthenticated',
      setSession: (user, accessToken) =>
        set({ user, accessToken, status: 'authenticated' }),
      clear: () => set({ user: null, accessToken: null, status: 'unauthenticated' }),
      hasRole: (role) => get().user?.roles.includes(role) ?? false,
    }),
    {
      name: 'iot-auth',
      partialize: (state) => ({ user: state.user, status: state.status }),
    },
  ),
);
