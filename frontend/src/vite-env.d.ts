/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  readonly VITE_SHOW_DEMO_ACCOUNTS?: string;
  /** Build-time Vercel environment: "preview" (STG) or "production". Unset is fail-closed. */
  readonly VITE_VERCEL_ENV?: string;
  /** Local Vite DEV only: shared staff-attach password for demo one-click login. Never commit real value. */
  readonly VITE_DEMO_LOGIN_PASSWORD?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

// M-10: vite.config.ts の `define` で埋め込む build-time 定数。
// frontend-deploy.yml が STG=preview / 本番=production を渡す。未設定は ""（fail-closed）。
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
