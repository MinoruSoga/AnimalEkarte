import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";

import { useGetTrimmingsByPetId } from "./use-pet-trimmings";

const trimmingFixture = {
  id: 7,
  clinic_id: 1,
  status: "completed",
  start_time: "2026-01-03T10:00:00+09:00",
  style_request: "シャンプーカット",
  remarks: "嫌がりなし",
  pet: { id: 5, name: "モモ", owner: { id: 9, name: "飼主" } },
  staff: { id: 3, name: "鈴木" },
};

describe("useGetTrimmingsByPetId（NO32 昇格 shared hook）", () => {
  it("pet_id/page/limit をAPIへ転送し TrimmingUI[] を返す", async () => {
    let capturedUrl: URL | undefined;
    server.use(
      http.get("/api/v1/trimmings", ({ request }) => {
        capturedUrl = new URL(request.url);
        return HttpResponse.json({
          data: [trimmingFixture],
          total: 1,
          page: 1,
          limit: 100,
        });
      }),
    );

    const { result } = renderHook(() => useGetTrimmingsByPetId("5"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(capturedUrl?.searchParams.get("pet_id")).toBe("5");
    expect(capturedUrl?.searchParams.get("page")).toBe("1");
    expect(capturedUrl?.searchParams.get("limit")).toBe("100");
    expect(result.current.data?.[0]).toEqual(
      expect.objectContaining({
        id: "7",
        petId: "5",
        staff: "鈴木",
        styleRequest: "シャンプーカット",
        status: "完了",
      }),
    );
  });

  it("options.enabled=false のとき fetch せず pending のままにする（権限外抑止）", () => {
    let called = false;
    server.use(
      http.get("/api/v1/trimmings", () => {
        called = true;
        return HttpResponse.json({ data: [], total: 0, page: 1, limit: 100 });
      }),
    );

    const { result } = renderHook(() => useGetTrimmingsByPetId("5", { enabled: false }), {
      wrapper: createTestWrapper(),
    });

    expect(result.current.isPending).toBe(true);
    expect(result.current.fetchStatus).toBe("idle");
    expect(called).toBe(false);
  });

  it("petId 空でも fetch しない", () => {
    let called = false;
    server.use(
      http.get("/api/v1/trimmings", () => {
        called = true;
        return HttpResponse.json({ data: [], total: 0, page: 1, limit: 100 });
      }),
    );

    const { result } = renderHook(() => useGetTrimmingsByPetId(""), {
      wrapper: createTestWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
    expect(called).toBe(false);
  });
});
