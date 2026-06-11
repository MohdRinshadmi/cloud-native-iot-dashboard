import { PageHeader } from '@/shared/components/layout/page-header';
import { Badge } from '@/shared/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { Separator } from '@/shared/components/ui/separator';
import { env } from '@/shared/lib/env';

/** Workspace + environment settings. Auth/tenant management lands in Phase 4. */
export function SettingsPage() {
  return (
    <>
      <PageHeader title="Settings" description="Workspace, environment and platform configuration." />

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Environment</CardTitle>
            <CardDescription>Resolved client configuration (validated at boot).</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <ConfigRow label="API base URL" value={env.VITE_API_BASE_URL} />
            <Separator />
            <ConfigRow label="WebSocket URL" value={env.VITE_WS_URL} />
            <Separator />
            <ConfigRow label="Mode" value={import.meta.env.MODE} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Appearance</CardTitle>
            <CardDescription>Theme and display preferences.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Theme</span>
              <Badge variant="secondary">Dark (enterprise)</Badge>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Density</span>
              <Badge variant="outline">Comfortable</Badge>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}

function ConfigRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <span className="text-muted-foreground">{label}</span>
      <code className="truncate rounded bg-muted px-2 py-0.5 font-mono text-xs">{value}</code>
    </div>
  );
}
