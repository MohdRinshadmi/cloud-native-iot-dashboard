import { PageHeader } from '@/shared/components/layout/page-header';
import { ComingSoon } from '@/shared/components/layout/route-states';

export function AnalyticsPage() {
  return (
    <>
      <PageHeader
        title="Analytics"
        description="Historical trends, fleet distributions and operational metrics."
      />
      <ComingSoon
        title="Analytics workbench"
        description="Recharts + D3 visualizations over historical telemetry: trend lines, distributions, percentile bands and fleet comparisons."
        phase="Phase 6"
      />
    </>
  );
}
