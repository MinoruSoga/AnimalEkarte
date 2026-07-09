// G10-6: minimal vitest-pool-workers config for backend/worker/*.test.ts.
//
// Points at vitest.wrangler.jsonc (test-only) rather than ../wrangler.jsonc
// (the deploy config): that file declares containers/durable_objects/
// hyperdrive bindings which the pure-function tests in migrate-exec.test.ts
// neither need nor can exercise in the vitest-pool-workers sandbox. Only
// compatibility_date/flags are needed to get a real workerd runtime (for
// crypto.subtle.timingSafeEqual — see migrate-exec.test.ts's file header for
// why that matters).
import { defineWorkersConfig } from "@cloudflare/vitest-pool-workers/config";

export default defineWorkersConfig({
  test: {
    poolOptions: {
      workers: {
        wrangler: { configPath: "./vitest.wrangler.jsonc" },
      },
    },
  },
});
