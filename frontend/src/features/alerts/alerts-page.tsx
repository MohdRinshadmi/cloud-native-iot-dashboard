import { PageHeader } from '@/shared/components/layout/page-header';
import { ComingSoon } from '@/shared/components/layout/route-states';

export function AlertsPage() {
  return (
    <>
      <PageHeader
        title="Alerts"
        description="Anomaly and threshold alerts with severity, history and audit trail."
      />
      <ComingSoon
        title="Alert center"
        description="Severity-graded alert feed, acknowledgement workflow, notification rules and a full audit log."
        phase="Phase 9"
      />
    </>
  );
}
