import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useDeleteBugReport } from "./delete-bug-report";

const { axiosDeleteMock } = vi.hoisted(() => ({
  axiosDeleteMock: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    delete: axiosDeleteMock,
  },
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

beforeEach(() => {
  axiosDeleteMock.mockReset();
  axiosDeleteMock.mockResolvedValue({});
});

describe("useDeleteBugReport", () => {
  it("DELETE /v1/support/bug-reports/:id を呼ぶ", async () => {
    const { result } = renderHook(() => useDeleteBugReport(), {
      wrapper: createWrapper(),
    });

    result.current.mutate(3);

    await waitFor(() => expect(axiosDeleteMock).toHaveBeenCalledWith("/v1/support/bug-reports/3"));
  });
});
