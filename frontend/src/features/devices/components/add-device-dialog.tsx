import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { CirclePlus, Loader2 } from 'lucide-react';
import { useCreateDevice } from '@/hooks/use-devices';
import { ApiError } from '@/services/api/client';
import { Button } from '@/shared/components/ui/button';
import { Input } from '@/shared/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/shared/components/ui/dialog';

const deviceSchema = z.object({
  name: z.string().min(1, 'Name is required').max(120),
  model: z.string().max(120).optional(),
  firmware: z.string().max(60).optional(),
});

type DeviceForm = z.infer<typeof deviceSchema>;

/** "Add device" flow — RHF + Zod form inside a Radix dialog. */
export function AddDeviceDialog({ disabled }: { disabled?: boolean }) {
  const [open, setOpen] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);
  const createDevice = useCreateDevice();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<DeviceForm>({ resolver: zodResolver(deviceSchema) });

  const onSubmit = handleSubmit(async (values) => {
    setServerError(null);
    try {
      await createDevice.mutateAsync({
        name: values.name,
        ...(values.model ? { model: values.model } : {}),
        ...(values.firmware ? { firmware: values.firmware } : {}),
      });
      reset();
      setOpen(false);
    } catch (err) {
      setServerError(err instanceof ApiError ? err.message : 'Failed to create device');
    }
  });

  return (
    <Dialog open={open} onOpenChange={(o) => { setOpen(o); if (!o) { reset(); setServerError(null); } }}>
      <DialogTrigger asChild>
        <Button disabled={disabled} title={disabled ? 'Requires operator or admin role' : undefined}>
          <CirclePlus /> Add device
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Register a device</DialogTitle>
          <DialogDescription>
            The device starts in <span className="font-mono text-xs">provisioning</span> and goes
            online with its first heartbeat.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={(e) => void onSubmit(e)} className="space-y-4" noValidate>
          <div className="space-y-1.5">
            <label htmlFor="device-name" className="text-sm font-medium">Name</label>
            <Input id="device-name" placeholder="pump-station-03" {...register('name')} />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label htmlFor="device-model" className="text-sm font-medium">Model</label>
              <Input id="device-model" placeholder="PX-1000" {...register('model')} />
            </div>
            <div className="space-y-1.5">
              <label htmlFor="device-fw" className="text-sm font-medium">Firmware</label>
              <Input id="device-fw" placeholder="2.4.1" {...register('firmware')} />
            </div>
          </div>

          {serverError && (
            <p role="alert" className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive">
              {serverError}
            </p>
          )}

          <DialogFooter>
            <Button type="button" variant="ghost" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="animate-spin" />}
              Register device
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
