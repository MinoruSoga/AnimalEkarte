import { AxiosError, AxiosHeaders, type AxiosResponse } from "axios";
import { describe, expect, it } from "vitest";
import { getRetryAfterMs } from "./get-revisit-intake";

function axiosError(status: number, headers: Record<string, string> = {}) {
  const response = {
    status,
    statusText: "",
    data: {},
    headers: new AxiosHeaders(headers),
    config: { headers: new AxiosHeaders() },
  } as AxiosResponse;
  return new AxiosError("x", undefined, response.config, undefined, response);
}

describe("getRetryAfterMs", () => {
  it("429/503 の Retry-After 秒をミリ秒へ変換する", () => {
    expect(getRetryAfterMs(axiosError(429, { "Retry-After": "7" }))).toBe(7000);
    expect(getRetryAfterMs(axiosError(503, { "Retry-After": "0" }))).toBe(0);
  });

  it("ヘッダ名の大文字小文字を区別しない", () => {
    expect(getRetryAfterMs(axiosError(429, { "retry-after": "3" }))).toBe(3000);
    expect(getRetryAfterMs(axiosError(503, { "RETRY-after": "4" }))).toBe(4000);
  });

  it("HTTP-date は現在時刻との差、過去日は 0", () => {
    const now = Date.parse("2026-10-21T00:00:00Z");
    const future = new Date(now + 30_000).toUTCString();
    const past = new Date(now - 30_000).toUTCString();
    expect(getRetryAfterMs(axiosError(503, { "Retry-After": future }), now)).toBe(30_000);
    expect(getRetryAfterMs(axiosError(429, { "Retry-After": past }), now)).toBe(0);
  });

  it("ヘッダ無し・解釈不能・対象外 status・非 axios は undefined", () => {
    expect(getRetryAfterMs(axiosError(429))).toBeUndefined();
    expect(getRetryAfterMs(axiosError(429, { "Retry-After": "soon" }))).toBeUndefined();
    expect(getRetryAfterMs(axiosError(500, { "Retry-After": "5" }))).toBeUndefined();
    expect(getRetryAfterMs(new Error("net"))).toBeUndefined();
  });
});
