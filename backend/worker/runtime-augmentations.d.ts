// worker-configuration.d.ts was generated before these existing wrangler vars
// were added. Keep the Worker-local typecheck accurate until the next approved
// `wrangler types` regeneration refreshes the generated root declaration.
interface Env {
  DB_MAX_OPEN_CONNS: string;
  DB_MAX_IDLE_CONNS: string;
  S3_PUBLIC_BASE_URL: string;
}

declare namespace Cloudflare {
  interface Env {
    DB_MAX_OPEN_CONNS: string;
    DB_MAX_IDLE_CONNS: string;
    S3_PUBLIC_BASE_URL: string;
  }
}

// Cloudflare Workers extends WebCrypto with a constant-time byte comparison.
interface SubtleCrypto {
  timingSafeEqual(left: BufferSource, right: BufferSource): boolean;
}
