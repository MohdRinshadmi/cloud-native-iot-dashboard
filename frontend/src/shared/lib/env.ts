import { z } from 'zod';

/**
 * Validate runtime configuration ONCE at module load. A misconfigured deploy
 * should fail loudly at boot, not as a confusing runtime error three clicks
 * into the app. Vite inlines `import.meta.env.VITE_*` at build time.
 */
const envSchema = z.object({
  VITE_API_BASE_URL: z.string().url().default('http://localhost:8080/api/v1'),
  VITE_WS_URL: z.string().url().default('ws://localhost:8080/api/v1/ws'),
});

const parsed = envSchema.safeParse(import.meta.env);

if (!parsed.success) {
  console.error('❌ Invalid environment configuration:', parsed.error.flatten().fieldErrors);
  throw new Error('Invalid environment configuration');
}

export const env = parsed.data;
