import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { motion } from 'framer-motion';
import { Loader2, Radar, ShieldCheck } from 'lucide-react';
import { login } from '@/services/api/auth';
import { ApiError } from '@/services/api/client';
import { Button } from '@/shared/components/ui/button';
import { Input } from '@/shared/components/ui/input';
import { Card, CardContent } from '@/shared/components/ui/card';

const loginSchema = z.object({
  email: z.string().email('Enter a valid email address'),
  password: z.string().min(1, 'Password is required'),
});

type LoginForm = z.infer<typeof loginSchema>;

/** Sign-in screen. Lives outside the app shell; redirects back after auth. */
export function LoginPage() {
  const navigate = useNavigate();
  const search = useSearch({ strict: false }) as { redirect?: string };
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({ resolver: zodResolver(loginSchema) });

  const onSubmit = handleSubmit(async (values) => {
    setServerError(null);
    try {
      await login(values.email, values.password);
      void navigate({ to: search.redirect ?? '/', replace: true });
    } catch (err) {
      setServerError(
        err instanceof ApiError && err.status !== 0
          ? err.message
          : 'Cannot reach the API — is the backend running?',
      );
    }
  });

  return (
    <main className="grid min-h-screen place-items-center px-6">
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: 'easeOut' }}
        className="w-full max-w-sm"
      >
        <div className="mb-8 text-center">
          <span className="mx-auto mb-4 grid h-12 w-12 place-items-center rounded-xl bg-primary/15 text-primary">
            <Radar className="h-6 w-6" />
          </span>
          <h1 className="text-2xl font-semibold tracking-tight">Fleet Command</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Sign in to your IoT analytics workspace
          </p>
        </div>

        <Card>
          <CardContent className="p-6">
            <form onSubmit={(e) => void onSubmit(e)} className="space-y-4" noValidate>
              <div className="space-y-1.5">
                <label htmlFor="email" className="text-sm font-medium">
                  Email
                </label>
                <Input
                  id="email"
                  type="email"
                  autoComplete="email"
                  placeholder="you@company.com"
                  {...register('email')}
                />
                {errors.email && <p className="text-xs text-destructive">{errors.email.message}</p>}
              </div>

              <div className="space-y-1.5">
                <label htmlFor="password" className="text-sm font-medium">
                  Password
                </label>
                <Input
                  id="password"
                  type="password"
                  autoComplete="current-password"
                  placeholder="••••••••••••"
                  {...register('password')}
                />
                {errors.password && (
                  <p className="text-xs text-destructive">{errors.password.message}</p>
                )}
              </div>

              {serverError && (
                <p
                  role="alert"
                  className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive"
                >
                  {serverError}
                </p>
              )}

              <Button type="submit" className="w-full" disabled={isSubmitting}>
                {isSubmitting ? <Loader2 className="animate-spin" /> : <ShieldCheck />}
                {isSubmitting ? 'Signing in…' : 'Sign in'}
              </Button>
            </form>
          </CardContent>
        </Card>

        <div className="mt-4 rounded-lg border border-dashed border-border px-4 py-3 text-center text-xs text-muted-foreground">
          <p className="font-medium text-foreground">Demo workspace</p>
          <p className="mt-1 font-mono">admin@demo.local · Password123!</p>
          <p className="font-mono">operator@ / viewer@ for RBAC testing</p>
        </div>
      </motion.div>
    </main>
  );
}
