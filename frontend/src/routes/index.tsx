import { createRoute } from '@tanstack/react-router';
import { rootRoute } from './__root';
import { SystemStatusCard } from '@/features/system-status/components/system-status-card';

export const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: LandingPage,
});

function LandingPage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-10 px-6">
      <header className="text-center">
        <p className="mb-2 inline-flex items-center rounded-full border border-primary/30 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
          Phase 1 · Foundation online
        </p>
        <h1 className="bg-gradient-to-b from-foreground to-muted-foreground bg-clip-text text-4xl font-bold tracking-tight text-transparent sm:text-5xl">
          Cloud-Native IoT Analytics
        </h1>
        <p className="mt-3 max-w-xl text-balance text-muted-foreground">
          Real-time telemetry, AI insights and fleet monitoring engineered for 1,000,000+ devices.
        </p>
      </header>

      <SystemStatusCard />

      <footer className="text-xs text-muted-foreground">
        React 19 · Vite · TanStack Router/Query · Zustand · Go + Gin · Postgres · Redis · MQTT
      </footer>
    </main>
  );
}
