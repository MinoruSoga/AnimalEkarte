import { describe, expect, it } from "vitest";
import axios, { AxiosError } from "axios";
import { resolveFetchError } from "./handle-fetch-error";

function makeAxiosError(status: number): AxiosError {
  const err = new AxiosError("Request failed");
  err.response = {
    status,
    data: {},
    statusText: "",
    headers: {},
    config: {} as never,
  };
  return err;
}

class FakeLiffApiError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = "LiffApiError";
  }
}

describe("resolveFetchError（R-F22: line-reserve/liff 共通のステータス別エラー解決）", () => {
  it("axios の 401 エラーは再ログインメッセージを返し canRetry=false", () => {
    const result = resolveFetchError(makeAxiosError(401), "コースの取得");

    expect(result.status).toBe(401);
    expect(result.canRetry).toBe(false);
    expect(result.message).toContain("LINEアプリを再起動して開き直してください");
  });

  it("status を持つ fetch 系エラー（LiffApiError 等）の 401 も同様に扱う", () => {
    const result = resolveFetchError(new FakeLiffApiError(401, "unauthorized"), "健康記録の取得");

    expect(result.status).toBe(401);
    expect(result.canRetry).toBe(false);
  });

  it("axios の 5xx エラーはサーバーエラーメッセージを返し canRetry=true", () => {
    const result = resolveFetchError(makeAxiosError(500), "コースの取得");

    expect(result.status).toBe(500);
    expect(result.canRetry).toBe(true);
    expect(result.message).toContain("サーバーエラーが発生しました");
  });

  it("fetch 系エラーの 503 もサーバーエラーメッセージを返す", () => {
    const result = resolveFetchError(new FakeLiffApiError(503, "unavailable"), "コースの取得");

    expect(result.status).toBe(503);
    expect(result.canRetry).toBe(true);
    expect(result.message).toContain("サーバーエラーが発生しました");
  });

  it("4xx（401以外）はコンテキスト付きの汎用メッセージを返し canRetry=true", () => {
    const result = resolveFetchError(makeAxiosError(404), "コースの取得");

    expect(result.status).toBe(404);
    expect(result.canRetry).toBe(true);
    expect(result.message).toBe("コースの取得に失敗しました。");
  });

  it("ステータス不明（ネットワーク断・非HTTPエラー）は canRetry=true のフォールバックメッセージを返す", () => {
    const result = resolveFetchError(new Error("Network Error"), "コースの取得");

    expect(result.status).toBeUndefined();
    expect(result.canRetry).toBe(true);
    expect(result.message).toContain("コースの取得に失敗しました");
  });

  it("axios.isAxiosError が true でも response が無い場合はステータス不明として扱う", () => {
    const err = new AxiosError("timeout");
    expect(axios.isAxiosError(err)).toBe(true);

    const result = resolveFetchError(err, "コースの取得");

    expect(result.status).toBeUndefined();
    expect(result.canRetry).toBe(true);
  });
});
