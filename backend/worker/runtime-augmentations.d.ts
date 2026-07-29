// worker-configuration.d.ts was generated before these existing wrangler vars
// were added. Keep the Worker-local typecheck accurate until the next approved
// `wrangler types` regeneration refreshes the generated root declaration.
interface Env {
  DB_MAX_OPEN_CONNS: string;
  DB_MAX_IDLE_CONNS: string;
  S3_PUBLIC_BASE_URL: string;
  SCHEDULER_ENVIRONMENT: string;
  SCHEDULER_OPS_SECRET: string;
  SCHEDULER_ACCESS_TEAM_DOMAIN: string;
  SCHEDULER_ACCESS_AUDIENCE: string;
  SCHEDULER_ALERT_ALLOWED_HOST: string;
  SCHEDULER_ALERT_WEBHOOK_URL: string;
  SCHEDULER_ALERT_WEBHOOK_SECRET: string;
  // Shared secret for Worker→Container scheduled-jobs (X-Scheduler-Token).
  // Declare in wrangler secrets.required; value via `wrangler secret put` only.
  SCHEDULER_INTERNAL_TOKEN: string;
}

declare namespace Cloudflare {
  interface Env {
    DB_MAX_OPEN_CONNS: string;
    DB_MAX_IDLE_CONNS: string;
    S3_PUBLIC_BASE_URL: string;
    SCHEDULER_ENVIRONMENT: string;
    SCHEDULER_OPS_SECRET: string;
    SCHEDULER_ACCESS_TEAM_DOMAIN: string;
    SCHEDULER_ACCESS_AUDIENCE: string;
    SCHEDULER_ALERT_ALLOWED_HOST: string;
    SCHEDULER_ALERT_WEBHOOK_URL: string;
    SCHEDULER_ALERT_WEBHOOK_SECRET: string;
    SCHEDULER_INTERNAL_TOKEN: string;
  }
}

// Cloudflare Workers extends WebCrypto with a constant-time byte comparison.
interface SubtleCrypto {
  timingSafeEqual(left: BufferSource, right: BufferSource): boolean;
}
