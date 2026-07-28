// G10-6: minimal Vitest 4 Workers config for backend/worker/*.test.ts.
//
// Points at vitest.wrangler.jsonc (test-only) rather than ../wrangler.jsonc
// (the deploy config): that file declares containers/durable_objects/
// hyperdrive bindings which the pure-function tests in migrate-exec.test.ts
// neither need nor can exercise in the vitest-pool-workers sandbox. Only
// compatibility_date/flags are needed to get a real workerd runtime (for
// crypto.subtle.timingSafeEqual — see migrate-exec.test.ts's file header for
// why that matters).
import { fileURLToPath } from "node:url";

import { cloudflareTest } from "@cloudflare/vitest-pool-workers";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [
    cloudflareTest({
      wrangler: {
        configPath: fileURLToPath(
          new URL("./vitest.wrangler.jsonc", import.meta.url),
        ),
      },
    }),
  ],
});
