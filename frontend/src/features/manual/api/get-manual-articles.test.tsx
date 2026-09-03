import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useGetManualArticleOverrides } from "./get-manual-articles";

const { axiosGetMock } = vi.hoisted(() => ({
  axiosGetMock: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    get: axiosGetMock,
  },
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

beforeEach(() => {
  axiosGetMock.mockReset();
  axiosGetMock.mockResolvedValue({ data: { data: [] } });
});

describe("useGetManualArticleOverrides", () => {
  it("enabled=falseならoverride APIを呼ばない", () => {
    const { result } = renderHook(() => useGetManualArticleOverrides(false), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
    expect(axiosGetMock).not.toHaveBeenCalled();
  });

  it("enabled=trueならoverride APIを呼ぶ", async () => {
    renderHook(() => useGetManualArticleOverrides(true), { wrapper: createWrapper() });

    await waitFor(() => expect(axiosGetMock).toHaveBeenCalledWith("/v1/manual/articles"));
  });
});
