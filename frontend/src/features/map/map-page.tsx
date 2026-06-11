import { PageHeader } from '@/shared/components/layout/page-header';
import { ComingSoon } from '@/shared/components/layout/route-states';

export function MapPage() {
  return (
    <>
      <PageHeader
        title="Map"
        description="Live GPS tracking, clustering and geofences on OpenStreetMap."
      />
      <ComingSoon
        title="Fleet map"
        description="Leaflet + OpenStreetMap with live device markers, marker clustering at scale, geofence drawing and location analytics."
        phase="Phase 8"
      />
    </>
  );
}
