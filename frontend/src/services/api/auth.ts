import { api } from './client';
import { useAuthStore } from '@/stores/auth.store';
import type { AuthUser, SessionResponse } from '@/types/api';

/**
 * Auth API service. All session endpoints send credentials so the httpOnly
 * refresh cookie travels with them; skipAuth avoids a pointless 401-retry
 * loop on the credential endpoints themselves.
 */

export async function login(email: string, password: string): Promise<SessionResponse> {
  const session = await api.post<SessionResponse>('/auth/login', {
    body: { email, password },
    credentials: 'include',
    skipAuth: true,
  });
  useAuthStore.getState().setSession(session.user, session.access_token);
  return session;
}

export async function logout(): Promise<void> {
  try {
    await api.post<void>('/auth/logout', { credentials: 'include' });
  } finally {
    // Local session dies even if the network call fails.
    useAuthStore.getState().clear();
  }
}

export function getMe(): Promise<AuthUser> {
  return api.get<AuthUser>('/auth/me');
}
