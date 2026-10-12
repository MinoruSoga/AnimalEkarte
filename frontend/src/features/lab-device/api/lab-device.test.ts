import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryCache, QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

import {
  useAttachLabDeviceJob,
  useClearLabDeviceWait,
  useDetachLabDeviceJob,
  useGetLabDeviceAgentConsumer,
  useGetLabDeviceUnlinked,
  usePutLabDeviceWait,
  useReceiveLabDeviceFrames,
} from "./lab-device";
import { useGetLabDeviceUnlinked as hooksUseGetLabDeviceUnlinked } from "@/hooks/use-lab-device-unlinked";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
    put: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

const mockedGet = vi.mocked(axios.get);
const mockedPut = vi.mocked(axios.put);
const mockedPost = vi.mocked(axios.post);
const mockedDelete = vi.mocked(axios.delete);
const mockedHandleApiError = vi.mocked(handleApiError);

function createWrapper(queryClient?: QueryClient) {
  const client =
    queryClient ??
    new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client }, children);
  };
}

// src/lib/react-query.ts の QueryCache と同じ契約を持つ client。
// silentError 非対象の失敗は引き続きグローバル handleApiError へ届くことを実証する。
function createReportingWrapper() {
  return createWrapper(
    new QueryClient({
      // Observe every state transition in renderHook, including refetch errors
      // that retain the same cached data. The route reads isError during render.
      defaultOptions: { queries: { retry: false, notifyOnChangeProps: "all" } },
      queryCache: new QueryCache({
        onError: (error, query) => {
          if (query.meta?.silentError) return;
          handleApiError(error, query.meta?.errorContext ?? "データ取得");
        },
      }),
    }),
  );
}

function axiosErrorWithStatus(status: number): AxiosError {
  const config = {
    headers: new AxiosHeaders(),
  } as InternalAxiosRequestConfig;
  return new AxiosError("request failed", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
    config,
    data: {},
    headers: new AxiosHeaders(),
    status,
    statusText: "",
  });
}

describe("lab-device API mutations (FE-RC-012)", () => {
  beforeEach(() => {
    mockedPut.mockReset();
    mockedPost.mockReset();
    mockedDelete.mockReset();
    mockedHandleApiError.mockReset();
  });

  it("re-exports useGetLabDeviceUnlinked from @/hooks (FE-RC-015 followup3)", () => {
    expect(useGetLabDeviceUnlinked).toBe(hooksUseGetLabDeviceUnlinked);
  });

  it("usePutLabDeviceWait: 失敗時に onError で handleApiError を呼び、握り潰さない", async () => {
    const apiError = new Error("network down");
    mockedPut.mockRejectedValueOnce(apiError);
    const { result } = renderHook(() => usePutLabDeviceWait(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      result.current.mutate(1);
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(mockedHandleApiError).toHaveBeenCalledWith(apiError, "受診中ペットの選択");
  });

  it("useClearLabDeviceWait: 失敗時に onError で handleApiError を呼ぶ", async () => {
    const apiError = new Error("network down");
    mockedDelete.mockRejectedValueOnce(apiError);
    const { result } = renderHook(() => useClearLabDeviceWait(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      result.current.mutate();
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(mockedHandleApiError).toHaveBeenCalledWith(apiError, "待機の解除");
  });

  it("useAttachLabDeviceJob: 失敗時に onError で handleApiError を呼ぶ（サイレント失敗防止）", async () => {
    const apiError = new Error("conflict");
    mockedPost.mockRejectedValueOnce(apiError);
    const { result } = renderHook(() => useAttachLabDeviceJob(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      await expect(result.current.mutateAsync({ jobId: "job-1", petId: 1 })).rejects.toThrow();
    });

    expect(mockedHandleApiError).toHaveBeenCalledWith(apiError, "検査結果の紐付け");
  });

  it("useDetachLabDeviceJob: 失敗時に onError で handleApiError を呼ぶ（サイレント失敗防止）", async () => {
    const apiError = new Error("conflict");
    mockedPost.mockRejectedValueOnce(apiError);
    const { result } = renderHook(() => useDetachLabDeviceJob(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      result.current.mutate("job-1");
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(mockedHandleApiError).toHaveBeenCalledWith(apiError, "検査結果の紐付け解除");
  });

  it("useReceiveLabDeviceFrames: 失敗時に onError で handleApiError を呼ぶ（サイレント失敗防止・FE-RC-012 followup）", async () => {
    const apiError = new Error("bad frame");
    mockedPost.mockRejectedValueOnce(apiError);
    const { result } = renderHook(() => useReceiveLabDeviceFrames(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      await expect(
        result.current.mutateAsync({ payloadBase64: "AA==", deviceHint: "auto" }),
      ).rejects.toThrow();
    });

    // LabDeviceBoard.onFrame は 400 系ならこの通知で足りるため再通知せず、
    // 401/500 系のみ機器向けの具体的な案内に差し替える（二重トースト回避、FE-RC-005）。
    expect(mockedHandleApiError).toHaveBeenCalledWith(apiError, "検査機器電文の受信");
  });
});

describe("useGetLabDeviceAgentConsumer (EMR-335)", () => {
  beforeEach(() => {
    mockedGet.mockReset();
    mockedHandleApiError.mockReset();
  });

  it("GET /v1/lab-device/agent-consumer 200 は発行済み capability token を返す", async () => {
    mockedGet.mockResolvedValueOnce({
      data: { agent_consumer_token: "cap-token-dummy", expires_in: 300 },
    });
    const { result } = renderHook(() => useGetLabDeviceAgentConsumer(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedGet).toHaveBeenCalledWith("/v1/lab-device/agent-consumer");
    expect(result.current.data).toBe("cap-token-dummy");
  });

  it("503（秘密未設定）は reject せず null を返し、グローバルエラーハンドラも呼ばない", async () => {
    // 503 は「未設定」という期待状態のため fetch 内で null 解決し、route が
    // 「検査機器エージェントが未設定です」を表示する。汎用 5xx トーストにしない。
    mockedGet.mockRejectedValueOnce(axiosErrorWithStatus(503));
    const { result } = renderHook(() => useGetLabDeviceAgentConsumer(), {
      wrapper: createReportingWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeNull();
    expect(mockedHandleApiError).not.toHaveBeenCalled();
  });

  it("500 は従来どおり reject してグローバル QueryCache 経由でハンドラへ届く", async () => {
    const apiError = axiosErrorWithStatus(500);
    mockedGet.mockRejectedValueOnce(apiError);
    const { result } = renderHook(() => useGetLabDeviceAgentConsumer(), {
      wrapper: createReportingWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(mockedHandleApiError).toHaveBeenCalledWith(apiError, "データ取得");
  });

  it("ネットワークエラー（非 Axios）も reject してグローバルハンドラへ届く", async () => {
    const apiError = new Error("network down");
    mockedGet.mockRejectedValueOnce(apiError);
    const { result } = renderHook(() => useGetLabDeviceAgentConsumer(), {
      wrapper: createReportingWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(mockedHandleApiError).toHaveBeenCalledWith(apiError, "データ取得");
  });

  it("成功から再取得500・未設定503・回復までquery状態を保ち、500だけを通知する", async () => {
    const serverError = axiosErrorWithStatus(500);
    mockedGet
      .mockResolvedValueOnce({ data: { agent_consumer_token: "old-capability", expires_in: 300 } })
      .mockRejectedValueOnce(serverError)
      .mockRejectedValueOnce(axiosErrorWithStatus(503))
      .mockResolvedValueOnce({ data: { agent_consumer_token: "new-capability", expires_in: 300 } });
    const { result } = renderHook(() => useGetLabDeviceAgentConsumer(), {
      wrapper: createReportingWrapper(),
    });
    await waitFor(() => expect(result.current.data).toBe("old-capability"));

    await act(async () => {
      await result.current.refetch();
    });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.data).toBe("old-capability");
    expect(mockedHandleApiError).toHaveBeenCalledWith(serverError, "データ取得");

    await act(async () => {
      await result.current.refetch();
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeNull();
    await act(async () => {
      await result.current.refetch();
    });
    expect(result.current.isSuccess).toBe(true);
    await waitFor(() => expect(result.current.data).toBe("new-capability"));
    expect(mockedHandleApiError).toHaveBeenCalledTimes(1);
  });
});
