import { useParams } from '@tanstack/react-router';
import { useDevice } from '@/hooks/use-devices';
import { formatAbsolute, formatRelative } from '@/utils/time';
import { PageHeader } from '@/shared/components/layout/page-header';
import { ComingSoon, PageSkeleton } from '@/shared/components/layout/route-states';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/components/ui/card';
import { Separator } from '@/shared/components/ui/separator';
import { DeviceStatusPill } from './components/device-status-pill';
import { LiveTelemetryPanel } from './components/live-telemetry-panel';
import { TelemetryHistorySection } from './components/telemetry-history-section';

/** Single-device drill-down backed by GET /api/v1/devices/:id. */
export function DeviceDetailPage() {
  const params = useParams({ strict: false });
  const deviceId = params.deviceId ?? '';
  const { data: device, isLoading } = useDevice(deviceId);

  if (isLoading) return <PageSkeleton />;
  if (!device) return null; // error boundary handles failures

  return (
    <>
      <PageHeader
        title={device.name}
        description="Telemetry, health score, command history and configuration for this device."
        actions={<DeviceStatusPill status={device.status} />}
      />

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Identity</CardTitle>
            <CardDescription>Registration and hardware metadata.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <DetailRow label="Device ID" value={<code className="font-mono text-xs">{device.id}</code>} />
            <Separator />
            <DetailRow label="Model" value={device.model || '—'} />
            <Separator />
            <DetailRow label="Firmware" value={<code className="font-mono text-xs">{device.firmware || '—'}</code>} />
            <Separator />
            <DetailRow label="Registered" value={formatAbsolute(device.created_at)} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Connectivity</CardTitle>
            <CardDescription>Heartbeat and link state.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <DetailRow label="Status" value={<DeviceStatusPill status={device.status} />} />
            <Separator />
            <DetailRow
              label="Last seen"
              value={
                <span title={formatAbsolute(device.last_seen_at)}>
                  {formatRelative(device.last_seen_at)}
                </span>
              }
            />
            <Separator />
            <DetailRow label="Updated" value={formatAbsolute(device.updated_at)} />
          </CardContent>
        </Card>
      </div>

      <div className="mt-6">
        <LiveTelemetryPanel deviceId={device.id} />
      </div>

      <div className="mt-6">
        <TelemetryHistorySection deviceId={device.id} />
      </div>

      <div className="mt-6">
        <ComingSoon
          title="Remote commands & OTA"
          description="The command console (reboot, config push) and over-the-air firmware update simulation attach here."
          phase="Phase 7"
        />
      </div>
    </>
  );
}

function DetailRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <span className="text-muted-foreground">{label}</span>
      <span>{value}</span>
    </div>
  );
}
