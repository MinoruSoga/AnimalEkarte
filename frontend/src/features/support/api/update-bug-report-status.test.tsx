import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useUpdateBugReportStatus } from "./update-bug-report-status";

const { axiosPatchMock } = vi.hoisted(() => ({
  axiosPatchMock: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    patch: axiosPatchMock,
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
  axiosPatchMock.mockReset();
  axiosPatchMock.mockResolvedValue({ data: { id: 3, status: "resolved" } });
});

describe("useUpdateBugReportStatus", () => {
  it("PATCH /v1/support/bug-reports/:id/status に status を送る", async () => {
    const { result } = renderHook(() => useUpdateBugReportStatus(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ id: 3, status: "resolved" });

    await waitFor(() =>
      expect(axiosPatchMock).toHaveBeenCalledWith("/v1/support/bug-reports/3/status", {
        status: "resolved",
      }),
    );
  });
});
