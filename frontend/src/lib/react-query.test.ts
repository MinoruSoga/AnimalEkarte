import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

import { handleApiError } from "@/lib/handle-api-error";
import { queryClient } from "./react-query";

const FAILED_QUERY_KEY = ["fe5-16-fail"] as const;
const SILENT_QUERY_KEY = ["fe5-16-silent"] as const;

describe("QueryCache onError（FE5-16: グローバルエラー可視化）", () => {
  beforeEach(() => {
    vi.mocked(handleApiError).mockClear();
  });

  it("QueryCache onError: query 失敗時に handleApiError が呼ばれる", async () => {
    await queryClient
      .fetchQuery({
        queryKey: FAILED_QUERY_KEY,
        queryFn: () => Promise.reject(new Error("boom")),
        retry: 0,
      })
      .catch(() => {});

    expect(handleApiError).toHaveBeenCalledTimes(1);
  });

  it("QueryCache onError: meta.silentError 指定時は呼ばれない", async () => {
    await queryClient
      .fetchQuery({
        queryKey: SILENT_QUERY_KEY,
        queryFn: () => Promise.reject(new Error("boom")),
        retry: 0,
        meta: { silentError: true },
      })
      .catch(() => {});

    expect(handleApiError).not.toHaveBeenCalled();
  });
});
