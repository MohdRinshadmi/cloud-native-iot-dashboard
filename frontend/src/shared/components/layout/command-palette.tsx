import { useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Command } from 'cmdk';
import { Search } from 'lucide-react';
import { NAV_SECTIONS } from './nav-config';
import { useUIStore } from '@/stores/ui.store';
import { Dialog, DialogContent, DialogTitle } from '@/shared/components/ui/dialog';

/**
 * ⌘K command palette (cmdk inside a Radix dialog). Currently navigation-only;
 * device search and remote commands plug into the same surface in later phases.
 */
export function CommandPalette() {
  const open = useUIStore((s) => s.commandPaletteOpen);
  const setOpen = useUIStore((s) => s.setCommandPaletteOpen);
  const navigate = useNavigate();

  // Global shortcut: ⌘K (mac) / Ctrl+K.
  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen(!open);
      }
    };
    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [open, setOpen]);

  const go = (to: string) => {
    setOpen(false);
    void navigate({ to });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="overflow-hidden p-0 sm:max-w-xl [&>button]:hidden">
        <DialogTitle className="sr-only">Command palette</DialogTitle>
        <Command className="bg-transparent" label="Command palette">
          <div className="flex items-center gap-2 border-b border-border px-4">
            <Search className="h-4 w-4 shrink-0 text-muted-foreground" />
            <Command.Input
              placeholder="Jump to a page…"
              className="h-12 w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
            />
          </div>
          <Command.List className="max-h-80 overflow-y-auto p-2">
            <Command.Empty className="py-8 text-center text-sm text-muted-foreground">
              No results found.
            </Command.Empty>
            {NAV_SECTIONS.map((section) => (
              <Command.Group
                key={section.title}
                heading={section.title}
                className="mb-1 [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-[10px] [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:uppercase [&_[cmdk-group-heading]]:tracking-widest [&_[cmdk-group-heading]]:text-muted-foreground"
              >
                {section.items.map((item) => (
                  <Command.Item
                    key={item.to}
                    value={`${item.label} ${item.description}`}
                    onSelect={() => go(item.to)}
                    className="flex cursor-pointer items-center gap-3 rounded-md px-2 py-2.5 text-sm aria-selected:bg-accent aria-selected:text-accent-foreground"
                  >
                    <item.icon className="h-4 w-4 text-muted-foreground" />
                    <div>
                      <p>{item.label}</p>
                      <p className="text-xs text-muted-foreground">{item.description}</p>
                    </div>
                  </Command.Item>
                ))}
              </Command.Group>
            ))}
          </Command.List>
        </Command>
      </DialogContent>
    </Dialog>
  );
}
