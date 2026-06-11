/// <reference types="vitest" />
import { defineConfig, type Plugin } from 'vite';
import react from '@vitejs/plugin-react-swc';
import tailwindcss from '@tailwindcss/vite';
import path, { resolve } from 'path';

/** dev サーバーで /line-reserve/* の HTML ナビゲーションを line-reserve/index.html にリライトする */
function lineReserveDevPlugin(): Plugin {
  return {
    name: 'line-reserve-dev',
    configureServer(server) {
      server.middlewares.use((req, _res, next) => {
        if (
          req.url?.startsWith('/line-reserve') &&
          !req.url.includes('.') // .tsx, .ts, .css 等のアセットリクエストは除外
        ) {
          req.url = '/line-reserve/index.html';
        }
        next();
      });
    },
  };
}

/** dev サーバーで /liff/* の HTML ナビゲーションを liff/index.html にリライトする */
function liffDevPlugin(): Plugin {
  return {
    name: 'liff-dev',
    configureServer(server) {
      server.middlewares.use((req, _res, next) => {
        if (req.url?.startsWith('/liff') && !req.url.includes('.')) {
          req.url = '/liff/index.html';
        }
        next();
      });
    },
  };
}

const manualChunkGroups = [
  // React core
  ['vendor-react', ['react', 'react-dom', 'react-router']],
  // Data fetching
  ['vendor-query', ['@tanstack/react-query', 'axios']],
  // Radix UI primitives in use
  [
    'vendor-radix',
    [
      '@radix-ui/react-alert-dialog',
      '@radix-ui/react-checkbox',
      '@radix-ui/react-dialog',
      '@radix-ui/react-dropdown-menu',
      '@radix-ui/react-label',
      '@radix-ui/react-popover',
      '@radix-ui/react-radio-group',
      '@radix-ui/react-scroll-area',
      '@radix-ui/react-select',
      '@radix-ui/react-separator',
      '@radix-ui/react-slot',
      '@radix-ui/react-switch',
      '@radix-ui/react-tabs',
      '@radix-ui/react-toggle-group',
    ],
  ],
  // Animation and DnD
  ['vendor-motion', ['motion']],
  ['vendor-dnd', ['@dnd-kit/core', '@dnd-kit/sortable', '@dnd-kit/utilities']],
  // Date handling
  ['vendor-date', ['date-fns']],
  // Charts
  ['vendor-charts', ['recharts']],
  // Icons
  ['vendor-icons', ['lucide-react']],
  // Misc utilities
  ['vendor-misc', ['sonner', 'clsx', 'tailwind-merge', 'cmdk']],
  // LIFF SDK
  ['vendor-liff', ['@line/liff']],
] as const;

function resolveManualChunk(id: string): string | undefined {
  if (!id.includes('node_modules')) return undefined;

  for (const [chunkName, packageNames] of manualChunkGroups) {
    if (packageNames.some((packageName) => id.includes(`/node_modules/${packageName}/`))) {
      return chunkName;
    }
  }

  return undefined;
}

export default defineConfig({
  plugins: [react(), tailwindcss(), lineReserveDevPlugin(), liffDevPlugin()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 3000,
    allowedHosts: true,
    proxy: {
      '/api': {
        target: 'http://backend:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'esnext',
    outDir: 'dist',
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
        'line-reserve': resolve(__dirname, 'line-reserve/index.html'),
        liff: resolve(__dirname, 'liff/index.html'),
      },
      output: {
        manualChunks: resolveManualChunk,
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/testing/setup.ts',
    css: true,
    exclude: ['**/node_modules/**', '**/dist/**', 'tests/**', 'e2e/**'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
    },
  },
});
