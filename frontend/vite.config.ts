import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'node:path';

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    host: true, // listen on 0.0.0.0 so the dev server works inside containers
  },
  build: {
    target: 'es2022',
    sourcemap: true,
    rollupOptions: {
      output: {
        // Split heavy, rarely-changing vendor code into its own chunk so app
        // updates don't bust the framework cache.
        manualChunks(id) {
          if (!id.includes('node_modules')) return undefined;
          if (id.includes('@tanstack')) return 'tanstack';
          // Charting libs are heavy and only used on chart routes — isolate
          // them so the base bundle stays lean and they cache independently.
          if (
            id.includes('/recharts/') ||
            id.includes('/d3-') ||
            id.includes('/victory-vendor/') ||
            id.includes('/react-smooth/')
          ) {
            return 'charts';
          }
          // Match ONLY the React core packages (not react-is/react-smooth/etc,
          // which would create circular chunk references).
          if (/\/node_modules\/(react|react-dom|scheduler)\//.test(id)) {
            return 'react-vendor';
          }
          return 'vendor';
        },
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    css: false,
  },
});
