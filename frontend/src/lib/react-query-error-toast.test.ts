import { describe, it, expect, vi, beforeEach } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { toast } from "sonner";

import { queryClient } from "./react-query";
import { queryKeys } from "./query-keys";

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

// EMR-339: handle-api-error.test.ts の axiosError helper と同一実装。
function axiosError(status: number, data: Record<string, unknown>): AxiosError {
  const config = {
    headers: new AxiosHeaders(),
  } as InternalAxiosRequestConfig;
  return new AxiosError("request failed", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
    config,
    data,
    headers: new AxiosHeaders(),
    status,
    statusText: "Internal Server Error",
  });
}

const SERVER_ERROR_MESSAGE = "サーバーエラーが発生しました。しばらく経ってから再度お試しください。";

// EMR-339: QueryCache.onError → 実物の handleApiError → toast.error の実経路で
// 同一表示文言のトーストが同一 id を共有し、silentError が無音のままであることを検証する。
describe("QueryCache onError のトースト ID（EMR-339）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("表示文言が同じ複数 query の失敗は同じ toast id を共有する（同時 500 失敗の重複抑止）", async () => {
    await Promise.all([
      queryClient
        .fetchQuery({
          queryKey: queryKeys.accountings.detail("emr339-500-a"),
          queryFn: () => Promise.reject(axiosError(500, {})),
          retry: 0,
        })
        .catch(() => {}),
      queryClient
        .fetchQuery({
          queryKey: queryKeys.accountings.detail("emr339-500-b"),
          queryFn: () => Promise.reject(axiosError(500, {})),
          retry: 0,
        })
        .catch(() => {}),
    ]);

    expect(toast.error).toHaveBeenCalledTimes(2);
    const [, firstOptions] = vi.mocked(toast.error).mock.calls[0];
    const [, secondOptions] = vi.mocked(toast.error).mock.calls[1];
    expect(firstOptions).toEqual({ id: SERVER_ERROR_MESSAGE });
    expect(secondOptions).toEqual(firstOptions);
    expect(toast.error).toHaveBeenNthCalledWith(1, SERVER_ERROR_MESSAGE, {
      id: SERVER_ERROR_MESSAGE,
    });
  });

  it("表示文言が異なる query 失敗は別の toast id を持つ", async () => {
    await Promise.all([
      queryClient
        .fetchQuery({
          queryKey: queryKeys.accountings.detail("emr339-500"),
          queryFn: () => Promise.reject(axiosError(500, {})),
          retry: 0,
        })
        .catch(() => {}),
      queryClient
        .fetchQuery({
          queryKey: queryKeys.accountings.detail("emr339-403"),
          queryFn: () => Promise.reject(axiosError(403, {})),
          retry: 0,
          meta: { errorContext: "削除" },
        })
        .catch(() => {}),
    ]);

    expect(toast.error).toHaveBeenCalledTimes(2);
    const [firstMessage, firstOptions] = vi.mocked(toast.error).mock.calls[0];
    const [secondMessage, secondOptions] = vi.mocked(toast.error).mock.calls[1];
    expect(firstMessage).toBe(SERVER_ERROR_MESSAGE);
    expect(secondMessage).toBe("削除の権限がありません。");
    expect(firstOptions).toEqual({ id: firstMessage });
    expect(secondOptions).toEqual({ id: secondMessage });
    expect(secondOptions).not.toEqual(firstOptions);
  });

  it("meta.silentError の query 失敗はトーストを出さない", async () => {
    await queryClient
      .fetchQuery({
        queryKey: queryKeys.accountings.detail("emr339-silent"),
        queryFn: () => Promise.reject(axiosError(500, {})),
        retry: 0,
        meta: { silentError: true },
      })
      .catch(() => {});

    expect(toast.error).not.toHaveBeenCalled();
  });
});
