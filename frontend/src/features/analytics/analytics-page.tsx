import { PageHeader } from '@/shared/components/layout/page-header';
import { StatusBreakdownCard } from './components/status-breakdown-card';
import { ModelDistributionCard } from './components/model-distribution-card';
import { ThroughputCard } from './components/throughput-card';

/** Fleet-level analytics: composition, status distribution and live throughput. */
export function AnalyticsPage() {
  return (
    <>
      <PageHeader
        title="Analytics"
        description="Fleet composition, status distribution and real-time ingest throughput."
      />

      <div className="grid gap-6 lg:grid-cols-2">
        <StatusBreakdownCard />
        <ModelDistributionCard />
      </div>

      <div className="mt-6">
        <ThroughputCard />
      </div>
    </>
  );
}
