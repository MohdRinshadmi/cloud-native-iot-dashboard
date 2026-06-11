import { Link, useRouter, type ErrorComponentProps } from '@tanstack/react-router';
import { AlertTriangle, ArrowLeft, Construction, RotateCcw } from 'lucide-react';
import { Button } from '@/shared/components/ui/button';
import { Badge } from '@/shared/components/ui/badge';
import { Card, CardContent } from '@/shared/components/ui/card';
import { Skeleton } from '@/shared/components/ui/skeleton';
import { ApiError } from '@/services/api/client';

/** Full-page skeleton shown while a lazy route chunk or loader is pending. */
export function PageSkeleton() {
  return (
    <div className="space-y-6" aria-busy="true" aria-label="Loading page">
      <div className="space-y-2">
        <Skeleton className="h-8 w-56" />
        <Skeleton className="h-4 w-80" />
      </div>
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 4 }, (_, i) => (
          <Skeleton key={i} className="h-28" />
        ))}
      </div>
      <Skeleton className="h-72" />
    </div>
  );
}

/** Route-level error boundary: structured ApiError details + retry. */
export function RouteError({ error }: ErrorComponentProps) {
  const router = useRouter();
  const apiError = error instanceof ApiError ? error : null;

  return (
    <div className="grid min-h-[60vh] place-items-center">
      <Card className="max-w-md text-center">
        <CardContent className="space-y-4 p-8">
          <span className="mx-auto grid h-12 w-12 place-items-center rounded-full bg-destructive/15 text-destructive">
            <AlertTriangle className="h-6 w-6" />
          </span>
          <div>
            <h2 className="text-lg font-semibold">Something went wrong</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {apiError ? apiError.message : 'An unexpected error occurred while rendering this page.'}
            </p>
            {apiError?.requestId && (
              <p className="mt-2 font-mono text-[10px] text-muted-foreground">
                request id: {apiError.requestId}
              </p>
            )}
          </div>
          <Button onClick={() => void router.invalidate()} variant="secondary">
            <RotateCcw /> Try again
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

/** 404 page. */
export function NotFound() {
  return (
    <div className="grid min-h-[60vh] place-items-center">
      <div className="text-center">
        <p className="font-mono text-7xl font-bold text-primary/40">404</p>
        <h2 className="mt-3 text-lg font-semibold">Page not found</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          The page you're looking for doesn't exist or was moved.
        </p>
        <Button asChild variant="secondary" className="mt-5">
          <Link to="/">
            <ArrowLeft /> Back to overview
          </Link>
        </Button>
      </div>
    </div>
  );
}

interface ComingSoonProps {
  title: string;
  description: string;
  phase: string;
}

/** Honest placeholder for features that land in a later phase. */
export function ComingSoon({ title, description, phase }: ComingSoonProps) {
  return (
    <Card className="border-dashed">
      <CardContent className="flex flex-col items-center gap-3 py-16 text-center">
        <span className="grid h-12 w-12 place-items-center rounded-full bg-primary/10 text-primary">
          <Construction className="h-6 w-6" />
        </span>
        <div>
          <h3 className="font-semibold">{title}</h3>
          <p className="mx-auto mt-1 max-w-sm text-sm text-muted-foreground">{description}</p>
        </div>
        <Badge variant="outline">{phase}</Badge>
      </CardContent>
    </Card>
  );
}
