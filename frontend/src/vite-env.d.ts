/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  readonly VITE_SHOW_DEMO_ACCOUNTS?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

// Ambient declarations for lucide-react direct ESM subpath imports.
// lucide-react v0.487.0 ships .js files per icon but no per-icon .d.ts.
// This declaration maps `lucide-react/dist/esm/icons/*` to LucideIcon so
// TypeScript accepts the direct imports required by bundle-barrel-imports rule.
declare module "lucide-react/dist/esm/icons/*" {
  import type { LucideIcon } from "lucide-react";
  const Icon: LucideIcon;
  export default Icon;
}
