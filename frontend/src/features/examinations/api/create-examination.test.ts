import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

import { useCreateExamination } from "./create-examination";

vi.mock("@/lib/axios", () => ({
  axios: {
    post: vi.fn(),
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

describe("useCreateExamination (FE-RC-005)", () => {
  beforeEach(() => {
    mockedPost.mockReset();
    mockedHandleApiError.mockReset();
  });

  it("失敗時に onError で handleApiError を 1 回呼ぶ（通知は api 層）", async () => {
    const apiError = new Error("API error");
    mockedPost.mockRejectedValueOnce(apiError);
    const { result } = renderHook(() => useCreateExamination(), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      result.current.mutate({
        pet_id: 1,
        exam_type_id: 2,
        doctor_id: 3,
        date: "2026-09-03",
      });
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(mockedHandleApiError).toHaveBeenCalledTimes(1);
    expect(mockedHandleApiError).toHaveBeenCalledWith(apiError, "検査作成");
  });
});
