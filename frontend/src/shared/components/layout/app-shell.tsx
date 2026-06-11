import type { ReactNode } from 'react';
import { Sidebar } from './sidebar';
import { Topbar } from './topbar';
import { CommandPalette } from './command-palette';

/**
 * The persistent dashboard chrome: sidebar rail + topbar + scrollable content
 * region. Every authenticated page renders inside this shell.
 */
export function AppShell({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <Topbar />
        <main className="flex-1 px-6 py-6">{children}</main>
      </div>
      <CommandPalette />
    </div>
  );
}
