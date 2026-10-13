import { createElement } from "react";
import type { ReactNode } from "react";
import { describe, it, expect, afterEach, vi } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
import { useGetVitals, useCreateVital, useUpdateVital, useDeleteVital } from "./vitals";
import type { CreateVitalInput } from "../types";

const CREATE_INPUT: CreateVitalInput = {
  recorded_at: "2026-07-16T00:00:00Z",
  temperature: null,
  heart_rate: null,
  respiration_rate: null,
  weight: 4.5,
  weight_unit: "Kg",
  notes: null,
};

afterEach(() => {
  localStorage.clear();
});

describe("vitals API クリニックスコープ (#186 review P2-15)", () => {
  it("グローバル選択クリニックと異なる clinicId を渡した場合、X-Clinic-ID は clinicId 引数の値を送信する", async () => {
    // グローバル選択クリニックは "1"
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedClinicHeader: string | null = null;
    server.use(
      http.get("*/v1/medical-records/:id/vitals", ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json([]);
      }),
    );

    // 対象カルテのクリニックは "2"（グローバル選択とは異なる拠点横断ケース）
    const { result } = renderHook(() => useGetVitals("42", "2"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedClinicHeader).toBe("2");
  });

  it("clinicId 未指定の場合はグローバル選択クリニックにフォールバックする（同一クリニック時の従来挙動を維持）", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedClinicHeader: string | null = null;
    server.use(
      http.get("*/v1/medical-records/:id/vitals", ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json([]);
      }),
    );

    const { result } = renderHook(() => useGetVitals("42"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedClinicHeader).toBe("1");
  });

  it("clinicId が undefined→解決値へ変わると、別クエリとして解決後の clinicId で再フェッチする（レコード取得前後のキャッシュレース対策）", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    const receivedHeaders: Array<string | null> = [];
    server.use(
      http.get("*/v1/medical-records/:id/vitals", ({ request }) => {
        receivedHeaders.push(request.headers.get("X-Clinic-ID"));
        return HttpResponse.json([]);
      }),
    );

    const { result, rerender } = renderHook(
      ({ clinicId }: { clinicId?: string }) => useGetVitals("42", clinicId),
      { wrapper: createTestWrapper(), initialProps: { clinicId: undefined as string | undefined } },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(receivedHeaders[0]).toBe("1");

    rerender({ clinicId: "2" });

    await waitFor(() => expect(receivedHeaders.length).toBe(2));
    expect(receivedHeaders[1]).toBe("2");
  });

  it("mutation にも同一の clinicId が X-Clinic-ID として送信される", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedClinicHeader: string | null = null;
    server.use(
      http.post("*/v1/medical-records/:id/vitals", ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json({ id: "1" }, { status: 201 });
      }),
    );

    const { result } = renderHook(() => useCreateVital("42", "2"), {
      wrapper: createTestWrapper(),
    });

    result.current.mutate({
      recorded_at: "2026-07-16T00:00:00Z",
      temperature: 38,
      heart_rate: null,
      respiration_rate: null,
      weight: null,
      weight_unit: "Kg",
      notes: null,
    });

    await waitFor(() => expect(receivedClinicHeader).toBe("2"));
  });

  it("create mutation はメモを `notes` キーで送り、legacy `note` キーは送らない（EMR-68）", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedBody: Record<string, unknown> | null = null;
    server.use(
      http.post("*/v1/medical-records/:id/vitals", async ({ request }) => {
        receivedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json({ id: "1" }, { status: 201 });
      }),
    );

    const { result } = renderHook(() => useCreateVital("42"), {
      wrapper: createTestWrapper(),
    });

    result.current.mutate({
      recorded_at: "2026-07-16T00:00:00Z",
      temperature: 38,
      heart_rate: null,
      respiration_rate: null,
      weight: null,
      weight_unit: "Kg",
      notes: "食欲やや低下",
    });

    await waitFor(() => expect(receivedBody).not.toBeNull());
    expect(receivedBody).toMatchObject({ notes: "食欲やや低下" });
    expect(receivedBody).not.toHaveProperty("note");
  });
});

/**
 * EMR-356: 問診画面の体重パネルが canView=false のとき、DOM を隠すだけでなく
 * 実際の vitals GET も発行しないこと（enabled オプションの後方互換追加）。
 */
describe("useGetVitals enabled オプション（EMR-356）", () => {
  it("enabled=false のとき実際の GET リクエストを発行しない", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let fetchCount = 0;
    server.use(
      http.get("*/v1/medical-records/:id/vitals", () => {
        fetchCount += 1;
        return HttpResponse.json([]);
      }),
    );

    const { result } = renderHook(() => useGetVitals("42", "2", { enabled: false }), {
      wrapper: createTestWrapper(),
    });

    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(result.current.fetchStatus).toBe("idle");
    expect(result.current.data).toBeUndefined();
    expect(fetchCount).toBe(0);
  });

  it("enabled 省略時は従来通りフェッチする（後方互換）", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let fetchCount = 0;
    server.use(
      http.get("*/v1/medical-records/:id/vitals", () => {
        fetchCount += 1;
        return HttpResponse.json([]);
      }),
    );

    const { result } = renderHook(() => useGetVitals("42", "2"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(fetchCount).toBe(1);
  });
});

/**
 * EMR-356: 問診履歴の体重投影は一覧 API（["medical-records", "history", petId]）
 * 経由なので、バイタル mutation 成功時に履歴/一覧キャッシュも invalidate する必要がある。
 * 失敗時は何も invalidate しない（擬似的な成功表示もしない）。
 */
describe("vitals mutation のキャッシュ無効化（EMR-356）", () => {
  function createSpiedWrapper() {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
    // .ts ファイルのため JSX は使えない（parser error）。createElement で構築する。
    const wrapper = ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client: queryClient }, children);
    return { invalidateSpy, wrapper };
  }

  it("create 成功で現在の vitals と medicalRecords 履歴/一覧の両方を invalidate する", async () => {
    server.use(
      http.post("*/v1/medical-records/:id/vitals", () =>
        HttpResponse.json({ id: "1" }, { status: 201 }),
      ),
    );
    const { invalidateSpy, wrapper } = createSpiedWrapper();
    const { result } = renderHook(() => useCreateVital("42", "2"), { wrapper });

    await result.current.mutateAsync(CREATE_INPUT);

    expect(invalidateSpy).toHaveBeenCalledWith(
      expect.objectContaining({ queryKey: queryKeys.medicalRecords.vitals("42") }),
    );
    expect(invalidateSpy).toHaveBeenCalledWith(
      expect.objectContaining({ queryKey: queryKeys.medicalRecords.all() }),
    );
  });

  it("update/delete 成功時も medicalRecords 履歴/一覧を invalidate する", async () => {
    server.use(
      http.patch("*/v1/medical-records/:id/vitals/:vitalId", () => HttpResponse.json({ id: "9" })),
      http.delete("*/v1/medical-records/:id/vitals/:vitalId", () =>
        HttpResponse.json({}, { status: 204 }),
      ),
    );
    const { invalidateSpy, wrapper } = createSpiedWrapper();
    const { result: updateResult } = renderHook(() => useUpdateVital("42", "2"), { wrapper });
    const { result: deleteResult } = renderHook(() => useDeleteVital("42", "2"), { wrapper });

    await updateResult.current.mutateAsync({ vitalId: "9", input: { weight: 4.6 } });
    await deleteResult.current.mutateAsync("9");

    expect(invalidateSpy).toHaveBeenCalledWith(
      expect.objectContaining({ queryKey: queryKeys.medicalRecords.all() }),
    );
    // update と delete の双方が呼んでいること（最低2回）
    expect(
      invalidateSpy.mock.calls.filter(([arg]) => JSON.stringify(arg).includes('"medical-records"'))
        .length,
    ).toBeGreaterThanOrEqual(2);
  });

  it("mutation 失敗時は履歴キャッシュを invalidate しない（成功偽装しない）", async () => {
    server.use(
      http.post("*/v1/medical-records/:id/vitals", () =>
        HttpResponse.json({ error: "boom" }, { status: 500 }),
      ),
    );
    const { invalidateSpy, wrapper } = createSpiedWrapper();
    const { result } = renderHook(() => useCreateVital("42"), { wrapper });

    await expect(result.current.mutateAsync(CREATE_INPUT)).rejects.toThrow();
    expect(invalidateSpy).not.toHaveBeenCalled();
  });
});
