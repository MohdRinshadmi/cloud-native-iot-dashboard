import { PageHeader } from '@/shared/components/layout/page-header';
import { ComingSoon } from '@/shared/components/layout/route-states';

export function InsightsPage() {
  return (
    <>
      <PageHeader
        title="AI Insights"
        description="Anomaly detection, predictive maintenance and fleet intelligence."
      />
      <ComingSoon
        title="AI analytics layer"
        description="Rule-based anomaly detection first, with a pluggable model interface for health scoring, risk scoring and predictive maintenance."
        phase="Phase 9"
      />
    </>
  );
}
