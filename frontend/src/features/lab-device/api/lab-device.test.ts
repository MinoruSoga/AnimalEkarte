import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

import {
  useAttachLabDeviceJob,
  useClearLabDeviceWait,
  useDetachLabDeviceJob,
  usePutLabDeviceWait,
  useReceiveLabDeviceFrames,
} from "./lab-device";

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

const mockedPut = vi.mocked(axios.put);
const mockedPost = vi.mocked(axios.post);
const mockedDelete = vi.mocked(axios.delete);
const mockedHandleApiError = vi.mocked(handleApiError);

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

describe("lab-device API mutations (FE-RC-012)", () => {
  beforeEach(() => {
    mockedPut.mockReset();
    mockedPost.mockReset();
    mockedDelete.mockReset();
    mockedHandleApiError.mockReset();
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
      await expect(
        result.current.mutateAsync({ jobId: "job-1", petId: 1 }),
      ).rejects.toThrow();
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
