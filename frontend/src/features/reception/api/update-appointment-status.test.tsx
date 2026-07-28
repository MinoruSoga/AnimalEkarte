import type { ReactNode } from "react";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

import { useUpdateAppointmentStatus } from "./update-appointment-status";

vi.mock("@/lib/axios", () => ({
  axios: { patch: vi.fn() },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

const mockedPatch = vi.mocked(axios.patch);
const mockedHandleApiError = vi.mocked(handleApiError);

function createWrapper(queryClient: QueryClient) {
  return function QueryClientWrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

describe("useUpdateAppointmentStatus", () => {
  beforeEach(() => {
    mockedPatch.mockReset();
    mockedHandleApiError.mockReset();
  });

  it.each([
    {
      label: "成功",
      arrange: () => mockedPatch.mockResolvedValueOnce({ data: undefined }),
      expectError: false,
    },
    {
      label: "失敗",
      arrange: () => mockedPatch.mockRejectedValueOnce(new Error("boom")),
      expectError: true,
    },
  ])("$label時に reception query を exactly once 再同期する", async ({
    arrange,
    expectError,
  }) => {
    arrange();
    const queryClient = new QueryClient({
      defaultOptions: {
        mutations: { retry: false },
        queries: { retry: false },
      },
    });
    const invalidateSpy = vi
      .spyOn(queryClient, "invalidateQueries")
      .mockResolvedValue(undefined);
    const { result } = renderHook(() => useUpdateAppointmentStatus(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current
        .mutateAsync({ id: "101", status: "checked_in" })
        .catch(() => undefined);
    });

    expect(invalidateSpy).toHaveBeenCalledTimes(1);
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.reception.all(),
    });
    if (expectError) {
      expect(mockedHandleApiError).toHaveBeenCalledWith(
        expect.any(Error),
        "受付ステータスの更新",
      );
    } else {
      expect(mockedHandleApiError).not.toHaveBeenCalled();
    }
  });
});
