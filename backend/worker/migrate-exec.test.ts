// G10-6: unit tests for the pure functions migrate-exec.ts was split out for
// ("unit test容易性のため分離" per its own file header) but shipped without any.
// Runs under @cloudflare/vitest-pool-workers (workerd runtime) rather than plain
// Node vitest because timingSafeEqual depends on crypto.subtle.timingSafeEqual,
// a Cloudflare-specific SubtleCrypto extension that does not exist in Node's
// WebCrypto implementation — a Node-run test would either fail to resolve the
// call or silently exercise a different code path than production.
import { describe, expect, it } from "vitest";
import { isAuthorizedMigrateRequest, timingSafeEqual, toMigrateResponse, attachLoginSeedMigrateEnv } from "./migrate-exec";

function utf8ByteLength(value: string): number {
  return new TextEncoder().encode(value).length;
}

function dummySecret(utf8Bytes: number): string {
  return "t".repeat(utf8Bytes);
}

// Dummy fixture only — not a real credential. Must stay >= 32 UTF-8 bytes.
const SECRET = dummySecret(48);

function requestWithAuth(header: string | null): Request {
  const headers = new Headers();
  if (header !== null) {
    headers.set("Authorization", header);
  }
  return new Request("https://example.com/_internal/migrate", { method: "POST", headers });
}

describe("timingSafeEqual", () => {
  it("returns true for identical strings", () => {
    expect(timingSafeEqual("abc123", "abc123")).toBe(true);
  });

  it("returns false for different strings of the same length", () => {
    expect(timingSafeEqual("abc123", "abc124")).toBe(false);
  });

  it("returns false when lengths differ", () => {
    expect(timingSafeEqual("short", "much-longer-value")).toBe(false);
  });

  it("returns false against an empty string", () => {
    expect(timingSafeEqual("abc123", "")).toBe(false);
  });
});

describe("isAuthorizedMigrateRequest", () => {
  it("returns false when secret is undefined (fail-closed)", () => {
    const req = requestWithAuth(`Bearer ${SECRET}`);
    expect(isAuthorizedMigrateRequest(req, undefined)).toBe(false);
  });

  it("returns false when secret is an empty string (fail-closed)", () => {
    const req = requestWithAuth(`Bearer ${SECRET}`);
    expect(isAuthorizedMigrateRequest(req, "")).toBe(false);
  });

  it("returns false when the Authorization header is missing", () => {
    const req = requestWithAuth(null);
    expect(isAuthorizedMigrateRequest(req, SECRET)).toBe(false);
  });

  it("returns false when the secret does not match", () => {
    const req = requestWithAuth("Bearer wrong-secret");
    expect(isAuthorizedMigrateRequest(req, SECRET)).toBe(false);
  });

  it("returns false when the Bearer prefix is missing", () => {
    const req = requestWithAuth(SECRET);
    expect(isAuthorizedMigrateRequest(req, SECRET)).toBe(false);
  });

  it("returns true for a correct `Bearer <secret>` header", () => {
    expect(utf8ByteLength(SECRET)).toBeGreaterThanOrEqual(32);
    const req = requestWithAuth(`Bearer ${SECRET}`);
    expect(isAuthorizedMigrateRequest(req, SECRET)).toBe(true);
  });

  it("returns false when a matching Bearer secret is 31 UTF-8 bytes", () => {
    const secret = dummySecret(31);
    expect(utf8ByteLength(secret)).toBe(31);
    const req = requestWithAuth(`Bearer ${secret}`);
    expect(isAuthorizedMigrateRequest(req, secret)).toBe(false);
  });

  it("returns true when a matching Bearer secret is exactly 32 UTF-8 bytes", () => {
    const secret = dummySecret(32);
    expect(utf8ByteLength(secret)).toBe(32);
    const req = requestWithAuth(`Bearer ${secret}`);
    expect(isAuthorizedMigrateRequest(req, secret)).toBe(true);
  });

  it("returns true when a matching Bearer secret is longer than 32 UTF-8 bytes", () => {
    const secret = dummySecret(33);
    expect(utf8ByteLength(secret)).toBe(33);
    const req = requestWithAuth(`Bearer ${secret}`);
    expect(isAuthorizedMigrateRequest(req, secret)).toBe(true);
  });

  it("uses UTF-8 byte length, not JS string length", () => {
    // 30 ASCII + U+00E9 (2 UTF-8 bytes) = 31 JS chars, 32 UTF-8 bytes.
    const secret = `${dummySecret(30)}\u00e9`;
    expect(secret.length).toBe(31);
    expect(utf8ByteLength(secret)).toBe(32);
    const req = requestWithAuth(`Bearer ${secret}`);
    expect(isAuthorizedMigrateRequest(req, secret)).toBe(true);
  });
});

describe("toMigrateResponse", () => {
  it("maps exitCode 0 to HTTP 200", async () => {
    const res = toMigrateResponse({ exitCode: 0, stdout: "ok", stderr: "" });
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body).toEqual({ exitCode: 0, stdout: "ok", stderr: "" });
  });

  it("maps a non-zero exitCode to HTTP 500", async () => {
    const res = toMigrateResponse({ exitCode: 1, stdout: "", stderr: "boom" });
    expect(res.status).toBe(500);
    const body = await res.json();
    expect(body).toEqual({ exitCode: 1, stdout: "", stderr: "boom" });
  });

  it("sets Content-Type: application/json", () => {
    const res = toMigrateResponse({ exitCode: 0, stdout: "", stderr: "" });
    expect(res.headers.get("Content-Type")).toBe("application/json");
  });
});

describe("attachLoginSeedMigrateEnv", () => {
  const dbEnv = {
    DB_HOST: "db.example.test",
    DB_PORT: "5432",
  };

  it("adds APP_ENV when set", () => {
    expect(attachLoginSeedMigrateEnv(dbEnv, "staging")).toEqual({
      DB_HOST: "db.example.test",
      DB_PORT: "5432",
      APP_ENV: "staging",
    });
  });

  it("omits APP_ENV when empty so Go fail-closes", () => {
    expect(attachLoginSeedMigrateEnv(dbEnv, "")).toEqual({
      DB_HOST: "db.example.test",
      DB_PORT: "5432",
    });
  });
});
