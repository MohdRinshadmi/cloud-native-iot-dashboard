import type { ReactNode } from 'react';
import { QueryProvider } from './query-provider';
import { TooltipProvider } from '@/shared/components/ui/tooltip';

/**
 * Composition point for all cross-cutting providers. Order matters: outer
 * providers wrap inner ones. Theme is currently fixed to dark (the `dark`
 * class on <html>); a ThemeProvider slots in here when light mode lands.
 */
export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <QueryProvider>
      <TooltipProvider delayDuration={200}>{children}</TooltipProvider>
    </QueryProvider>
  );
}
