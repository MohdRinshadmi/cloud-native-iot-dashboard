import { useParams } from '@tanstack/react-router';
import { PageHeader } from '@/shared/components/layout/page-header';
import { ComingSoon } from '@/shared/components/layout/route-states';
import { Badge } from '@/shared/components/ui/badge';

/** Single-device drill-down; hydrates with telemetry + charts in Phases 3–6. */
export function DeviceDetailPage() {
  const params = useParams({ strict: false });
  const deviceId = params.deviceId ?? 'unknown-device';
  return (
    <>
      <PageHeader
        title={deviceId}
        description="Telemetry, health score, command history and configuration for this device."
        actions={<Badge variant="outline">provisioning</Badge>}
      />
      <ComingSoon
        title="Device detail"
        description="Live telemetry panels, health gauges, command console and audit trail render here once the device and telemetry APIs are wired."
        phase="Phases 3–6"
      />
    </>
  );
}
