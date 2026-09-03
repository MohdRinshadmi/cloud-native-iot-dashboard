import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/shared/lib/cn';

const badgeVariants = cva(
  'inline-flex items-center gap-1.5 rounded-full border-transparent px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset transition-colors',
  {
    variants: {
      variant: {
        default: 'bg-primary/12 text-primary ring-primary/25',
        secondary: 'bg-secondary text-secondary-foreground ring-white/5',
        outline: 'bg-transparent text-muted-foreground ring-border',
        success: 'bg-success/12 text-success ring-success/25',
        warning: 'bg-warning/12 text-warning ring-warning/25',
        destructive: 'bg-destructive/12 text-destructive ring-destructive/25',
      },
    },
    defaultVariants: { variant: 'default' },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant }), className)} {...props} />;
}

export { Badge, badgeVariants };
