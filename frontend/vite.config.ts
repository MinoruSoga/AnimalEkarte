/// <reference types="vitest" />
import { defineConfig, type Plugin } from 'vite';
import react from '@vitejs/plugin-react-swc';
import tailwindcss from '@tailwindcss/vite';
import fs from 'fs';
import path, { resolve } from 'path';

/** dev サーバーで /line-reserve/* を line-reserve/index.html にリライトする */
function lineReserveDevPlugin(): Plugin {
  return {
    name: 'line-reserve-dev',
    configureServer(server) {
      server.middlewares.use((req, _res, next) => {
        if (req.url?.startsWith('/line-reserve')) {
          const htmlPath = resolve(__dirname, 'line-reserve/index.html');
          if (fs.existsSync(htmlPath)) {
            req.url = '/line-reserve/index.html';
          }
        }
        next();
      });
    },
  };
}

export default defineConfig({
  plugins: [react(), tailwindcss(), lineReserveDevPlugin()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 3000,
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
      },
      output: {
        manualChunks: {
          // React コア
          'vendor-react': ['react', 'react-dom', 'react-router'],
          // データフェッチ
          'vendor-query': ['@tanstack/react-query', 'axios'],
          // Radix UI プリミティブ群（全パッケージ）
          'vendor-radix': [
            '@radix-ui/react-accordion',
            '@radix-ui/react-alert-dialog',
            '@radix-ui/react-aspect-ratio',
            '@radix-ui/react-avatar',
            '@radix-ui/react-checkbox',
            '@radix-ui/react-collapsible',
            '@radix-ui/react-context-menu',
            '@radix-ui/react-dialog',
            '@radix-ui/react-dropdown-menu',
            '@radix-ui/react-hover-card',
            '@radix-ui/react-label',
            '@radix-ui/react-menubar',
            '@radix-ui/react-navigation-menu',
            '@radix-ui/react-popover',
            '@radix-ui/react-progress',
            '@radix-ui/react-radio-group',
            '@radix-ui/react-scroll-area',
            '@radix-ui/react-select',
            '@radix-ui/react-separator',
            '@radix-ui/react-slider',
            '@radix-ui/react-slot',
            '@radix-ui/react-switch',
            '@radix-ui/react-tabs',
            '@radix-ui/react-toggle',
            '@radix-ui/react-toggle-group',
            '@radix-ui/react-tooltip',
          ],
          // アニメーション・DnD（重量ライブラリを隔離）
          'vendor-motion': ['motion/react'],
          'vendor-dnd': ['@dnd-kit/core', '@dnd-kit/sortable', '@dnd-kit/utilities'],
          // 日付操作
          'vendor-date': ['date-fns'],
          // アイコン（全体で最大の外部 JS 資産になりやすい）
          'vendor-icons': ['lucide-react'],
          // その他ユーティリティ
          'vendor-misc': ['zustand', 'sonner', 'clsx', 'tailwind-merge', 'cmdk'],
        },
      },
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/testing/setup.ts',
    css: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
    },
  },
});
