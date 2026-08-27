/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  readonly VITE_SHOW_DEMO_ACCOUNTS?: string;
  /** Local Vite DEV only: shared staff-attach password for demo one-click login. Never commit real value. */
  readonly VITE_DEMO_LOGIN_PASSWORD?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

// M-10: vite.config.ts の `define` で埋め込む build-time 定数。
// Vercel がビルド時に自動注入する VERCEL_ENV（"production"/"preview"/"development"）を
// クライアントバンドルへ焼き込み、本番判定を Dashboard の環境変数設定ミスから守る。
declare const __VERCEL_ENV__: string;

// Ambient declarations for lucide-react direct ESM subpath imports.
// lucide-react v0.487.0 ships .js files per icon but no per-icon .d.ts.
// This declaration maps `lucide-react/dist/esm/icons/*` to LucideIcon so
// TypeScript accepts the direct imports required by bundle-barrel-imports rule.
declare module "lucide-react/dist/esm/icons/*" {
  import type { LucideIcon } from "lucide-react";
  const Icon: LucideIcon;
  export default Icon;
}
