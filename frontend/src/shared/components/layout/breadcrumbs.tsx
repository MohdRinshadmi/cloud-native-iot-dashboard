import { Link, useRouterState } from '@tanstack/react-router';
import { ChevronRight } from 'lucide-react';
import { findNavItem } from './nav-config';
import { cn } from '@/shared/lib/cn';

/**
 * Path-derived breadcrumbs: section item from nav config, plus the trailing
 * dynamic segment (e.g. a device id) when present.
 */
export function Breadcrumbs() {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const navItem = findNavItem(pathname);

  if (!navItem) return null;

  // Anything beyond the nav item's path is a detail segment (id, etc.).
  const rest = pathname === navItem.to ? '' : pathname.slice(navItem.to.length).replace(/^\//, '');
  const detail = rest ? decodeURIComponent(rest.split('/')[0] ?? '') : '';

  return (
    <nav aria-label="Breadcrumb" className="flex items-center gap-1.5 text-sm">
      <Link
        to={navItem.to}
        className={cn(
          'font-medium transition-colors',
          detail ? 'text-muted-foreground hover:text-foreground' : 'text-foreground',
        )}
      >
        {navItem.label}
      </Link>
      {detail && (
        <>
          <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
          <span className="max-w-48 truncate font-mono text-xs text-foreground">{detail}</span>
        </>
      )}
    </nav>
  );
}
