import { CirclePlus, Cpu, Search } from 'lucide-react';
import { PageHeader } from '@/shared/components/layout/page-header';
import { Button } from '@/shared/components/ui/button';
import { Input } from '@/shared/components/ui/input';
import { Card, CardContent } from '@/shared/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/shared/components/ui/table';

/**
 * Device inventory. The toolbar + table shell is final; rows hydrate from the
 * device API in Phase 3 (list/create) and Phase 7 (groups, firmware, OTA).
 */
export function DevicesPage() {
  return (
    <>
      <PageHeader
        title="Devices"
        description="Inventory, connectivity and health for every registered device."
        actions={
          <Button disabled title="Device registration API lands in Phase 3">
            <CirclePlus /> Add device
          </Button>
        }
      />

      <div className="mb-4 flex items-center gap-3">
        <div className="relative max-w-sm flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder="Filter by name, id or model…" className="pl-9" disabled />
        </div>
      </div>

      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Device</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Model</TableHead>
              <TableHead>Firmware</TableHead>
              <TableHead>Last seen</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell colSpan={6}>
                <CardContent className="flex flex-col items-center gap-3 py-14 text-center">
                  <span className="grid h-12 w-12 place-items-center rounded-full bg-muted text-muted-foreground">
                    <Cpu className="h-6 w-6" />
                  </span>
                  <div>
                    <p className="font-medium">No devices yet</p>
                    <p className="mt-1 max-w-sm text-sm text-muted-foreground">
                      The device registry API arrives in Phase 3 — this table hydrates with live
                      inventory, status pills and row actions.
                    </p>
                  </div>
                </CardContent>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </Card>
    </>
  );
}
