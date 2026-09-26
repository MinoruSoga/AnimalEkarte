import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useCreateBugReport } from "./create-bug-report";
import type { CreateBugReportParams } from "./create-bug-report";

const { axiosPostMock } = vi.hoisted(() => ({
  axiosPostMock: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    post: axiosPostMock,
  },
}));

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

const baseParams: CreateBugReportParams = {
  title: "会計画面でエラー",
  detail: "確定ボタンを押すと真っ白になる",
  pageUrl: "http://localhost:3000/accounting",
  routePath: "/accounting",
  userAgent: "test-agent",
  viewport: "1280x720",
  appVersion: "",
  screenshot: null,
};

beforeEach(() => {
  axiosPostMock.mockReset();
  axiosPostMock.mockResolvedValue({ data: { id: 1, title: baseParams.title } });
});

describe("useCreateBugReport", () => {
  it("multipart/form-data で全フィールドを送信する", async () => {
    const { result } = renderHook(() => useCreateBugReport(), { wrapper: createWrapper() });

    result.current.mutate(baseParams);

    await waitFor(() => expect(axiosPostMock).toHaveBeenCalledTimes(1));
    const [url, body, config] = axiosPostMock.mock.calls[0] as [
      string,
      FormData,
      { headers: Record<string, string> },
    ];
    expect(url).toBe("/v1/support/bug-reports");
    expect(config.headers["Content-Type"]).toBe("multipart/form-data");
    expect(body).toBeInstanceOf(FormData);
    expect(body.get("title")).toBe(baseParams.title);
    expect(body.get("detail")).toBe(baseParams.detail);
    expect(body.get("page_url")).toBe(baseParams.pageUrl);
    expect(body.get("route_path")).toBe(baseParams.routePath);
    expect(body.get("user_agent")).toBe(baseParams.userAgent);
    expect(body.get("viewport")).toBe(baseParams.viewport);
    expect(body.get("screenshot")).toBeNull();
  });

  it("screenshot 指定時はファイルパートを付ける", async () => {
    const file = new File([new Uint8Array([0x89, 0x50])], "shot.png", { type: "image/png" });
    const { result } = renderHook(() => useCreateBugReport(), { wrapper: createWrapper() });

    result.current.mutate({ ...baseParams, screenshot: file });

    await waitFor(() => expect(axiosPostMock).toHaveBeenCalledTimes(1));
    const body = axiosPostMock.mock.calls[0][1] as FormData;
    const part = body.get("screenshot");
    expect(part).toBeInstanceOf(File);
    expect((part as File).name).toBe("shot.png");
  });
});
