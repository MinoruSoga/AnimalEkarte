import { describe, expect, it, vi } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import {
  isNonDisclosureReadStatus,
  resolveEntityReadResult,
} from "./entity-read-result";

function axiosError(status: number | undefined): AxiosError {
  const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
  if (status === undefined) {
    // Network failure: no response
    return new AxiosError(
      "Network Error",
      AxiosError.ERR_NETWORK,
      config,
      undefined,
      undefined,
    );
  }
  return new AxiosError(
    "request failed",
    AxiosError.ERR_BAD_RESPONSE,
    config,
    undefined,
    {
      config,
      data: { error: "not found" },
      headers: new AxiosHeaders(),
      status,
      statusText: "Error",
    },
  );
}

describe("resolveEntityReadResult", () => {
  it("id なし → idle（create route）", () => {
    expect(
      resolveEntityReadResult({
        id: undefined,
        data: undefined,
        isLoading: false,
        isError: false,
        error: null,
      }),
    ).toEqual({ status: "idle" });
  });

  it("loading + data なし → loading", () => {
    expect(
      resolveEntityReadResult({
        id: "1",
        data: undefined,
        isLoading: true,
        isError: false,
        error: null,
      }),
    ).toEqual({ status: "loading" });
  });

  it("data あり → found", () => {
    const data = { id: "1" };
    expect(
      resolveEntityReadResult({
        id: "1",
        data,
        isLoading: false,
        isError: false,
        error: null,
      }),
    ).toEqual({ status: "found", data });
  });

  it("404 → notFound", () => {
    expect(
      resolveEntityReadResult({
        id: "999",
        data: undefined,
        isLoading: false,
        isError: true,
        error: axiosError(404),
      }),
    ).toEqual({ status: "notFound" });
  });

  it("403 → forbiddenOrHidden（クライアント非開示・404 と UI 同等）", () => {
    expect(
      resolveEntityReadResult({
        id: "999",
        data: undefined,
        isLoading: false,
        isError: true,
        error: axiosError(403),
      }),
    ).toEqual({ status: "forbiddenOrHidden" });
  });

  it("network error → error かつ retry を保持（404 へ偽装しない）", () => {
    const refetch = vi.fn();
    const error = axiosError(undefined);
    const result = resolveEntityReadResult({
      id: "999",
      data: undefined,
      isLoading: false,
      isError: true,
      error,
      refetch,
    });
    expect(result.status).toBe("error");
    if (result.status !== "error") return;
    expect(result.error).toBe(error);
    expect(result.retry).toBeTypeOf("function");
    result.retry?.();
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it("settled で data なし → notFound", () => {
    expect(
      resolveEntityReadResult({
        id: "1",
        data: undefined,
        isLoading: false,
        isError: false,
        error: null,
      }),
    ).toEqual({ status: "notFound" });
  });
});

describe("isNonDisclosureReadStatus", () => {
  it("notFound と forbiddenOrHidden のみ true", () => {
    expect(isNonDisclosureReadStatus("notFound")).toBe(true);
    expect(isNonDisclosureReadStatus("forbiddenOrHidden")).toBe(true);
    expect(isNonDisclosureReadStatus("error")).toBe(false);
    expect(isNonDisclosureReadStatus("found")).toBe(false);
    expect(isNonDisclosureReadStatus("loading")).toBe(false);
    expect(isNonDisclosureReadStatus("idle")).toBe(false);
  });
});
