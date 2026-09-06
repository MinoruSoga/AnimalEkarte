const ALLOWED_ORIGINS = new Set([
  "http://localhost:3003",
  "http://127.0.0.1:3003",
  "http://host.docker.internal:3003",
]);

export function assertClinicalAppEnv(appEnv: string | undefined): void {
  if ((appEnv ?? "").trim().toLowerCase() !== "test") {
    throw new Error("clinical e2e requires APP_ENV=test");
  }
}

export function assertClinicalBaseURL(rawURL: string | undefined): string {
  const fallback = "http://localhost:3003";
  const parsed = new URL(rawURL && rawURL.trim() !== "" ? rawURL : fallback);
  if (!ALLOWED_ORIGINS.has(parsed.origin)) {
    throw new Error("clinical e2e base URL is not a local compose frontend");
  }
  return parsed.origin;
}

export function assertClinicalTeardownRegistered(flag: string | undefined): void {
  if (flag !== "registered") {
    throw new Error("clinical e2e teardown is not registered");
  }
}
