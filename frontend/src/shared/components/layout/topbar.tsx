import { LogOut, Search, Settings, UserRound } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { Breadcrumbs } from './breadcrumbs';
import { ConnectionIndicator } from './connection-indicator';
import { useUIStore } from '@/stores/ui.store';
import { useAuthStore } from '@/stores/auth.store';
import { Kbd } from '@/shared/components/ui/kbd';
import { Separator } from '@/shared/components/ui/separator';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/shared/components/ui/dropdown-menu';

/** Top chrome: breadcrumbs · search trigger · live status · user menu. */
export function Topbar() {
  const setPaletteOpen = useUIStore((s) => s.setCommandPaletteOpen);
  const user = useAuthStore((s) => s.user);
  const navigate = useNavigate();

  return (
    <header className="sticky top-0 z-20 flex h-14 items-center justify-between gap-4 border-b border-border bg-background/80 px-6 backdrop-blur">
      <Breadcrumbs />

      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={() => setPaletteOpen(true)}
          className="hidden items-center gap-3 rounded-md border border-border px-3 py-1.5 text-xs text-muted-foreground transition-colors hover:bg-accent md:flex"
        >
          <Search className="h-3.5 w-3.5" />
          <span>Search…</span>
          <Kbd>⌘K</Kbd>
        </button>

        <ConnectionIndicator />
        <Separator orientation="vertical" className="h-6" />

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              type="button"
              aria-label="User menu"
              className="grid h-8 w-8 place-items-center rounded-full border border-border bg-secondary text-secondary-foreground transition-colors hover:bg-accent"
            >
              <UserRound className="h-4 w-4" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-52">
            <DropdownMenuLabel>
              {user ? (
                <div>
                  <p className="text-sm font-medium text-foreground">{user.name}</p>
                  <p className="text-xs font-normal">{user.email}</p>
                </div>
              ) : (
                'Guest session'
              )}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={() => void navigate({ to: '/settings' })}>
              <Settings /> Settings
            </DropdownMenuItem>
            <DropdownMenuItem disabled>
              <LogOut /> Sign out (Phase 4)
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
