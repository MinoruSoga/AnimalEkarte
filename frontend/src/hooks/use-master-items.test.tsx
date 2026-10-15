import type { ReactNode } from "react";
import { describe, it, expect, afterEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { useGetMasterItems } from "./use-master-items";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

// EMR-306: transformGenericMasterItem は wire の merged_into_cage_id を
// MasterItem.mergedIntoCageId（optional string）へ写す。フィールド欠落・null も許容。
describe("useGetMasterItems transformGenericMasterItem (EMR-306)", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("merged_into_cage_id → mergedIntoCageId を String() で写し、null/欠落は undefined にする", async () => {
    server.use(
      http.get("*/v1/masters/cages", () =>
        HttpResponse.json([
          {
            id: 1,
            name: "ケージA",
            is_active: true,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
            merged_into_cage_id: 5,
          },
          {
            id: 2,
            name: "ケージB",
            is_active: true,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
            merged_into_cage_id: null,
          },
          {
            id: 3,
            name: "ケージC",
            is_active: true,
            created_at: "2026-01-01T00:00:00Z",
            updated_at: "2026-01-01T00:00:00Z",
          },
        ]),
      ),
    );

    const { result } = renderHook(() => useGetMasterItems("cage"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.data).toHaveLength(3));
    expect(result.current.data[0].mergedIntoCageId).toBe("5");
    // 欠落・null は transformCage と同じ null sentinel に正規化される（EMR-306 統一）
    expect(result.current.data[1].mergedIntoCageId).toBeNull();
    expect(result.current.data[2].mergedIntoCageId).toBeNull();
    // 既存フィールドの transform が壊れていないこと
    expect(result.current.data[0].id).toBe("1");
    expect(result.current.data[0].status).toBe("active");
  });
});

// EMR-412 R4: 共有フックはクエリの失敗状態を握り潰さず isError を additive に公開する。
// （HospitalizationList 側の props 偽装ではなく、実クエリの失敗が surface すること）
describe("useGetMasterItems error state (EMR-412 R4)", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("マスタ取得が 500 で失敗すると isError=true を公開する", async () => {
    server.use(http.get("*/v1/masters/cages", () => new HttpResponse(null, { status: 500 })));

    const { result } = renderHook(() => useGetMasterItems("cage"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.isLoading).toBe(false);
  });

  it("マスタ取得成功時は isError=false を公開する", async () => {
    server.use(http.get("*/v1/masters/cages", () => HttpResponse.json([])));

    const { result } = renderHook(() => useGetMasterItems("cage"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.isError).toBe(false);
  });
});
