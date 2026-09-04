/// <reference types="vitest" />
import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react-swc";
import tailwindcss from "@tailwindcss/vite";
import path, { resolve } from "path";

/** dev サーバーで /line-reserve/* の HTML ナビゲーションを line-reserve/index.html にリライトする */
function lineReserveDevPlugin(): Plugin {
  return {
    name: "line-reserve-dev",
    configureServer(server) {
      server.middlewares.use((req, _res, next) => {
        if (!req.url?.startsWith("/line-reserve")) {
          next();
          return;
        }

        // line-reserve/index.html は <script src="./src/main.tsx"> と相対パスで
        // main.tsx を読み込む。ブラウザはこれをアドレスバーの実URL基準で解決するため、
        // /line-reserve/{clinicId}/ を開くと /line-reserve/{clinicId}/src/... という
        // 実ファイルの位置より1階層深いパスをリクエストしてしまう(BUG-402)。
        // 実ファイルの位置(/line-reserve/src/...)へ正規化する。
        const nestedAssetMatch = req.url.match(/^\/line-reserve\/[^/]+\/(src\/.+)$/);
        if (nestedAssetMatch) {
          req.url = `/line-reserve/${nestedAssetMatch[1]}`;
          next();
          return;
        }

        if (!req.url.includes(".")) {
          // .tsx, .ts, .css 等のアセットリクエストは除外
          req.url = "/line-reserve/index.html";
        }
        next();
      });
    },
  };
}

/** dev サーバーで /liff/* の HTML ナビゲーションを liff/index.html にリライトする */
function liffDevPlugin(): Plugin {
  return {
    name: "liff-dev",
    configureServer(server) {
      server.middlewares.use((req, _res, next) => {
        if (!req.url?.startsWith("/liff")) {
          next();
          return;
        }

        // line-reserve と同じく /liff/{clinicId}/src/main.tsx が 503 になる (BUG-017 / S12)。
        const nestedAssetMatch = req.url.match(/^\/liff\/[^/]+\/(src\/.+)$/);
        if (nestedAssetMatch) {
          req.url = `/liff/${nestedAssetMatch[1]}`;
          next();
          return;
        }

        if (!req.url.includes(".")) {
          req.url = "/liff/index.html";
        }
        next();
      });
    },
  };
}

const manualChunkGroups = [
  // React core
  ["vendor-react", ["react", "react-dom", "react-router"]],
  // Data fetching
  ["vendor-query", ["@tanstack/react-query", "axios"]],
  // Radix UI primitives in use
  [
    "vendor-radix",
    [
      "@radix-ui/react-alert-dialog",
      "@radix-ui/react-checkbox",
      "@radix-ui/react-dialog",
      "@radix-ui/react-dropdown-menu",
      "@radix-ui/react-label",
      "@radix-ui/react-popover",
      "@radix-ui/react-radio-group",
      "@radix-ui/react-scroll-area",
      "@radix-ui/react-select",
      "@radix-ui/react-separator",
      "@radix-ui/react-slot",
      "@radix-ui/react-switch",
      "@radix-ui/react-tabs",
      "@radix-ui/react-toggle-group",
    ],
  ],
  // Animation and DnD
  ["vendor-motion", ["motion"]],
  ["vendor-dnd", ["@dnd-kit/core", "@dnd-kit/sortable", "@dnd-kit/utilities"]],
  // Date handling
  ["vendor-date", ["date-fns"]],
  // Charts
  ["vendor-charts", ["recharts"]],
  // Icons
  ["vendor-icons", ["lucide-react"]],
  // Misc utilities
  ["vendor-misc", ["sonner", "clsx", "tailwind-merge", "cmdk"]],
  // LIFF SDK
  ["vendor-liff", ["@line/liff"]],
] as const;

const devAllowedHosts = [
  "localhost",
  "127.0.0.1",
  "0.0.0.0",
  "frontend",
  "host.docker.internal",
  "animalekarte-frontend-1",
  ".noah-karte.com",
];

function resolveManualChunk(id: string): string | undefined {
  if (!id.includes("node_modules")) return undefined;

  for (const [chunkName, packageNames] of manualChunkGroups) {
    if (packageNames.some((packageName) => id.includes(`/node_modules/${packageName}/`))) {
      return chunkName;
    }
  }

  return undefined;
}

const vercelEnv = process.env.VERCEL_ENV ?? "";
const STG_API_URL = "https://api.stg.noah-karte.com/api";
const PROD_API_URL = "https://api.noah-karte.com/api";
const define: Record<string, string> = {
  __VERCEL_ENV__: JSON.stringify(vercelEnv),
};
if (vercelEnv === "preview" || vercelEnv === "production") {
  define["import.meta.env.VITE_DEMO_LOGIN_PASSWORD"] = JSON.stringify(
    vercelEnv === "preview" ? (process.env.DEMO_LOGIN_PASSWORD ?? "") : "",
  );
  // loadEnv は frontend/.env.production を読む。AWS ALB が入っていると CSP
  // connect-src（api.stg.noah-karte.com / api.noah-karte.com）と食い違い、
  // ログインが「接続できません」になる。preview/production はホストを define で固定する。
  define["import.meta.env.VITE_API_URL"] = JSON.stringify(
    vercelEnv === "preview" ? STG_API_URL : PROD_API_URL,
  );
}

export default defineConfig({
  // M-10: Vercel が自動注入する VERCEL_ENV をビルド時定数として埋め込む。
  // frontend/src/features/auth/lib/show-demo-accounts.ts 参照。
  define,
  plugins: [react(), tailwindcss(), lineReserveDevPlugin(), liffDevPlugin()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    host: "0.0.0.0",
    port: 3000,
    allowedHosts: devAllowedHosts,
    proxy: {
      "/api": {
        target: "http://backend:8080",
        changeOrigin: true,
      },
    },
  },
  build: {
    target: "esnext",
    outDir: "dist",
    rollupOptions: {
      input: {
        main: resolve(__dirname, "index.html"),
        "line-reserve": resolve(__dirname, "line-reserve/index.html"),
        liff: resolve(__dirname, "liff/index.html"),
      },
      output: {
        manualChunks: resolveManualChunk,
      },
    },
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./src/testing/setup.ts",
    css: true,
    // scripts/**: node:test ベースのガードレール回帰テスト（例: check-eslint-disable-rationale.test.mjs）は
    // 専用 CI ジョブが `node --test` で直接実行する。Vitest のデフォルト include glob (*.test.mjs) にも
    // 一致してしまい「No test suite found」で Test (with coverage) step を落とすため除外する。
    exclude: ["**/node_modules/**", "**/dist/**", "tests/**", "e2e/**", "scripts/**"],
    coverage: {
      provider: "v8",
      // json-summary: coverage-policy.md Phase 1 ratchet（scripts/coverage-ratchet.mjs）が
      // total.statements.pct を読むために必要
      reporter: ["text", "json", "json-summary", "html"],
      exclude: [
        "src/types/generated/**", // tygo 生成型定義 — テスト対象外
        "src/testing/**", // テストセットアップ・MSW モック — テスト対象外
      ],
    },
  },
});
