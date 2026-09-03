import { Link } from '@tanstack/react-router';
import { PanelLeftClose, PanelLeftOpen, Radar } from 'lucide-react';
import { NAV_SECTIONS } from './nav-config';
import { useUIStore } from '@/stores/ui.store';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/components/ui/tooltip';
import { Button } from '@/shared/components/ui/button';
import { cn } from '@/shared/lib/cn';

/**
 * Collapsible primary navigation. Expanded: labels + section titles.
 * Collapsed: icon rail with tooltips. State persists across reloads.
 */
export function Sidebar() {
  const collapsed = useUIStore((s) => s.sidebarCollapsed);
  const toggle = useUIStore((s) => s.toggleSidebar);

  return (
    <aside
      className={cn(
        'sticky top-0 z-30 flex h-screen shrink-0 flex-col border-r border-border bg-card/60 backdrop-blur-xl transition-[width] duration-200',
        collapsed ? 'w-[4.25rem]' : 'w-60',
      )}
    >
      {/* Brand */}
      <div className={cn('flex h-14 items-center gap-2.5 border-b border-border px-4', collapsed && 'justify-center px-0')}>
        <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-primary/15 text-primary shadow-glow ring-1 ring-primary/30">
          <Radar className="h-5 w-5" />
        </span>
        {!collapsed && (
          <div className="leading-tight">
            <p className="text-sm font-semibold tracking-tight">Fleet Command</p>
            <p className="text-[10px] uppercase tracking-[0.18em] text-muted-foreground">IoT Analytics</p>
          </div>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-5 overflow-y-auto px-2 py-4">
        {NAV_SECTIONS.map((section) => (
          <div key={section.title}>
            {!collapsed && (
              <p className="mb-1.5 px-2 text-[10px] font-medium uppercase tracking-widest text-muted-foreground">
                {section.title}
              </p>
            )}
            <ul className="space-y-0.5">
              {section.items.map((item) => {
                const link = (
                  <Link
                    to={item.to}
                    activeOptions={{ exact: item.to === '/' }}
                    className={cn(
                      'group relative flex items-center gap-3 rounded-md px-2.5 py-2 text-sm text-muted-foreground transition-colors',
                      'before:absolute before:left-0 before:top-1/2 before:h-4 before:w-[3px] before:-translate-y-1/2 before:rounded-r-full before:bg-primary before:opacity-0 before:transition-opacity',
                      'hover:bg-accent hover:text-foreground',
                      '[&.active]:bg-primary/10 [&.active]:text-primary [&.active]:font-medium [&.active]:before:opacity-100',
                      collapsed && 'justify-center px-0',
                    )}
                  >
                    <item.icon className="h-[18px] w-[18px] shrink-0" />
                    {!collapsed && <span>{item.label}</span>}
                  </Link>
                );

                return (
                  <li key={item.to}>
                    {collapsed ? (
                      <Tooltip>
                        <TooltipTrigger asChild>{link}</TooltipTrigger>
                        <TooltipContent side="right">{item.label}</TooltipContent>
                      </Tooltip>
                    ) : (
                      link
                    )}
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </nav>

      {/* Collapse toggle */}
      <div className={cn('border-t border-border p-2', collapsed && 'flex justify-center')}>
        <Button
          variant="ghost"
          size={collapsed ? 'icon' : 'sm'}
          onClick={toggle}
          className={cn('text-muted-foreground', !collapsed && 'w-full justify-start gap-3 px-2.5')}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? <PanelLeftOpen /> : <PanelLeftClose />}
          {!collapsed && <span className="text-sm">Collapse</span>}
        </Button>
      </div>
    </aside>
  );
}
