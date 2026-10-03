import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useCreatePlaneTicket } from "./create-plane-ticket";

const { axiosPostMock } = vi.hoisted(() => ({
  axiosPostMock: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    post: axiosPostMock,
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
  axiosPostMock.mockReset();
  axiosPostMock.mockResolvedValue({
    data: { id: 3, plane_issue_url: "https://app.plane.so/ws/browse/EMR-9/" },
  });
});

describe("useCreatePlaneTicket", () => {
  it("POST /v1/support/bug-reports/:id/plane-ticket を呼ぶ", async () => {
    const { result } = renderHook(() => useCreatePlaneTicket(), {
      wrapper: createWrapper(),
    });

    result.current.mutate(3);

    await waitFor(() =>
      expect(axiosPostMock).toHaveBeenCalledWith("/v1/support/bug-reports/3/plane-ticket"),
    );
  });
});
