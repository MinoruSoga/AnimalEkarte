import { describe, it, expect, afterEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";
import { useGetTreatments, useCreateTreatment } from "./treatments";

afterEach(() => {
  localStorage.clear();
});

describe("treatments API クリニックスコープ (#186 review P2-15)", () => {
  it("グローバル選択クリニックと異なる clinicId を渡した場合、X-Clinic-ID は clinicId 引数の値を送信する", async () => {
    // グローバル選択クリニックは "1"
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedClinicHeader: string | null = null;
    server.use(
      http.get("*/v1/medical-records/:id/treatments", ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json([]);
      }),
    );

    // 対象カルテのクリニックは "2"（グローバル選択とは異なる拠点横断ケース）
    const { result } = renderHook(() => useGetTreatments("42", "2"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedClinicHeader).toBe("2");
  });

  it("clinicId 未指定の場合はグローバル選択クリニックにフォールバックする（同一クリニック時の従来挙動を維持）", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedClinicHeader: string | null = null;
    server.use(
      http.get("*/v1/medical-records/:id/treatments", ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json([]);
      }),
    );

    const { result } = renderHook(() => useGetTreatments("42"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedClinicHeader).toBe("1");
  });

  it("clinicId が undefined→解決値へ変わると、別クエリとして解決後の clinicId で再フェッチする（レコード取得前後のキャッシュレース対策）", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    const receivedHeaders: Array<string | null> = [];
    server.use(
      http.get("*/v1/medical-records/:id/treatments", ({ request }) => {
        receivedHeaders.push(request.headers.get("X-Clinic-ID"));
        return HttpResponse.json([]);
      }),
    );

    // 初回描画時は record.clinicId 未解決（undefined）— MedicalRecordForm の
    // トップレベル eager fetch が currentRecord 解決前に走る状況を模す。
    const { result, rerender } = renderHook(
      ({ clinicId }: { clinicId?: string }) => useGetTreatments("42", clinicId),
      { wrapper: createTestWrapper(), initialProps: { clinicId: undefined } },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedHeaders[0]).toBe("1"); // グローバル選択クリニックで初回フェッチ

    // currentRecord 解決 → 拠点横断レコードの実クリニックが判明
    rerender({ clinicId: "2" });

    await waitFor(() => expect(receivedHeaders.length).toBe(2));
    expect(receivedHeaders[1]).toBe("2"); // クエリキーが変わり自動的に正しいクリニックへ再フェッチされる
  });

  it("mutation にも同一の clinicId が X-Clinic-ID として送信される", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedClinicHeader: string | null = null;
    server.use(
      http.post("*/v1/medical-records/:id/treatments", ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json({ id: "1" }, { status: 201 });
      }),
    );

    const { result } = renderHook(() => useCreateTreatment("42", "2"), {
      wrapper: createTestWrapper(),
    });

    result.current.mutate({
      item_type: "other",
      content: "test",
      unit_price: 0,
      quantity: 1,
      is_selected: true,
      is_insurance: false,
      discount_amount: 0,
      sort_order: 0,
    });

    await waitFor(() => expect(receivedClinicHeader).toBe("2"));
  });
});
