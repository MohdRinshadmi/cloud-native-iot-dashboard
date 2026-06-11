import { useEffect, useState } from 'react';
import { RouterProvider } from '@tanstack/react-router';
import { Radar } from 'lucide-react';
import { router } from './router';
import { bootstrapAuth } from '@/services/api/client';

/**
 * App gates the router behind the auth bootstrap: one silent refresh attempt
 * resolves the session (httpOnly cookie → access token) BEFORE any route
 * guard runs, so guards never see an indeterminate auth state.
 */
export function App() {
  const [booted, setBooted] = useState(false);

  useEffect(() => {
    void bootstrapAuth().finally(() => setBooted(true));
  }, []);

  if (!booted) {
    return (
      <div className="grid min-h-screen place-items-center">
        <div className="flex flex-col items-center gap-3 text-muted-foreground">
          <span className="grid h-12 w-12 animate-pulse place-items-center rounded-xl bg-primary/15 text-primary">
            <Radar className="h-6 w-6" />
          </span>
          <p className="text-xs uppercase tracking-widest">Restoring session…</p>
        </div>
      </div>
    );
  }

  return <RouterProvider router={router} />;
}
