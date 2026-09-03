import { LogOut, Search, Settings } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { Breadcrumbs } from './breadcrumbs';
import { ConnectionIndicator } from './connection-indicator';
import { useUIStore } from '@/stores/ui.store';
import { useAuthStore } from '@/stores/auth.store';
import { logout } from '@/services/api/auth';
import { Badge } from '@/shared/components/ui/badge';
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

  const handleSignOut = async () => {
    await logout();
    void navigate({ to: '/login' });
  };

  return (
    <header className="sticky top-0 z-20 flex h-14 items-center justify-between gap-4 border-b border-border bg-background/70 px-6 backdrop-blur-xl">
      <Breadcrumbs />

      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={() => setPaletteOpen(true)}
          className="hidden items-center gap-3 rounded-md border border-border bg-elevated/40 px-3 py-1.5 text-xs text-muted-foreground transition-colors hover:border-border/70 hover:bg-accent hover:text-foreground md:flex"
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
              className="grid h-8 w-8 place-items-center rounded-full bg-primary/15 text-sm font-semibold text-primary ring-1 ring-primary/30 transition-all hover:bg-primary/25 hover:ring-primary/50"
            >
              {user?.name.charAt(0).toUpperCase() ?? '?'}
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuLabel>
              {user ? (
                <div className="space-y-1">
                  <p className="text-sm font-medium text-foreground">{user.name}</p>
                  <p className="text-xs font-normal">{user.email}</p>
                  <Badge variant="outline" className="capitalize">
                    {user.role}
                  </Badge>
                </div>
              ) : (
                'Guest session'
              )}
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={() => void navigate({ to: '/settings' })}>
              <Settings /> Settings
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => void handleSignOut()}>
              <LogOut /> Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
