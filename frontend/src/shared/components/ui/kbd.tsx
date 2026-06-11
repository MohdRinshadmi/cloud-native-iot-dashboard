import { cn } from '@/shared/lib/cn';

/** Keyboard-shortcut chip, e.g. <Kbd>⌘K</Kbd>. */
function Kbd({ className, ...props }: React.HTMLAttributes<HTMLElement>) {
  return (
    <kbd
      className={cn(
        'pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border border-border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground',
        className,
      )}
      {...props}
    />
  );
}

export { Kbd };
