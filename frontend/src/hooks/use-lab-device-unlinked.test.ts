import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

import { useAttachLabDeviceJob, useDetachLabDeviceJob } from "@/hooks/use-lab-device-unlinked";

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

const mockedPost = vi.mocked(axios.post);
const mockedHandleApiError = vi.mocked(handleApiError);

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

describe("use-lab-device-unlinked hooks (FE-RC-015 followup3)", () => {
  beforeEach(() => {
    mockedPost.mockReset();
    mockedHandleApiError.mockReset();
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
});
