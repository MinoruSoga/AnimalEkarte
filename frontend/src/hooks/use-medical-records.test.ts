import { afterEach, describe, expect, it } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";
import {
  useGetMedicalRecords,
  getMedicalRecords,
  transformToHistoryItem,
  useGetPetMedicalHistory,
  type MedicalRecordInterviewHistoryItem,
} from "./use-medical-records";
import type { MedicalRecordResponse } from "@/types/generated/medicalrecord-responses";

describe("use-medical-records (FE-RC-015 elevation)", () => {
  it("exports the list hook and query function from hooks (no features import)", () => {
    expect(typeof useGetMedicalRecords).toBe("function");
    expect(typeof getMedicalRecords).toBe("function");
  });
});

const minimalResponse: MedicalRecordResponse = {
  id: 1,
  clinic_id: 1,
  record_no: "MR-001",
  date: "2026-03-25T00:00:00Z",
  status: "finalized",
  version: 1,
  visit_count: 0,
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("useGetPetMedicalHistory の transformToHistoryItem（EMR-182 前回複写）", () => {
  it("inquiry の複写可能値を copySource へマップする", () => {
    const result = transformToHistoryItem({
      ...minimalResponse,
      inquiry: {
        id: 1,
        chief_complaint: "元気がない",
        notes: "安静と投薬",
        chief_complaint_type_id: 5,
      },
    });
    expect(result.copySource).toEqual({
      recordId: "1",
      chiefComplaint: "元気がない",
      treatmentPolicy: "安静と投薬",
      chiefComplaintTypeId: 5,
    });
  });

  it("copySource.recordId にレコード id の文字列を保持する（EMR-219 明細複写の起点）", () => {
    const result = transformToHistoryItem({
      ...minimalResponse,
      id: 42,
      inquiry: { id: 1, chief_complaint: "元気がない", notes: "" },
    });
    expect(result.copySource?.recordId).toBe("42");
  });

  it("copySource は空文字・未設定の項目を含めない", () => {
    const result = transformToHistoryItem({
      ...minimalResponse,
      inquiry: { id: 1, chief_complaint: "", notes: "", chief_complaint_type_id: 2 },
    });
    expect(result.copySource).toEqual({ recordId: "1", chiefComplaintTypeId: 2 });
  });

  it("inquiry 未設定でも copySource は recordId を保持する（EMR-219）", () => {
    expect(transformToHistoryItem({ ...minimalResponse, inquiry: undefined }).copySource).toEqual({
      recordId: "1",
    });
  });

  it("複写可能な問診値が無くても copySource は recordId を保持する（EMR-219: 明細のみの複写を可能に）", () => {
    const result = transformToHistoryItem({
      ...minimalResponse,
      inquiry: { id: 1, chief_complaint: "", notes: "" },
    });
    expect(result.copySource).toEqual({ recordId: "1" });
  });

  it("id が無く複写可能値も無いとき copySource は undefined", () => {
    const result = transformToHistoryItem({
      ...minimalResponse,
      id: undefined as unknown as number,
      inquiry: { id: 1, chief_complaint: "", notes: "" },
    });
    expect(result.copySource).toBeUndefined();
  });
});

// EMR-356: 体重投影の wire フィールドは additive optional。生成型は触らず、
// この test では拡張 fixture を構造的部分型として渡す（any 不使用）。
type MedicalRecordResponseWithWeights = MedicalRecordResponse & {
  weight_measurements?: {
    id: number;
    value: number;
    unit: "Kg" | "g";
    recorded_at: string;
  }[];
};

type HistoryItemWithWeights = MedicalRecordInterviewHistoryItem & {
  weightMeasurements?: {
    id: string;
    value: number;
    unit: "Kg" | "g";
    measuredAt: string;
  }[];
};

describe("transformToHistoryItem — 体重測定の投影（EMR-356）", () => {
  it("weight_measurements を値・単位・測定日時つきで投影し、複数測定を保持する", () => {
    const record: MedicalRecordResponseWithWeights = {
      ...minimalResponse,
      weight_measurements: [
        { id: 11, value: 4.2, unit: "Kg", recorded_at: "2026-03-01T09:30:00+09:00" },
        { id: 12, value: 4300, unit: "g", recorded_at: "2026-03-01T15:00:00+09:00" },
      ],
    };

    const item = transformToHistoryItem(record) as HistoryItemWithWeights;

    expect(item.weightMeasurements).toEqual([
      { id: "11", value: 4.2, unit: "Kg", measuredAt: "2026-03-01T09:30:00+09:00" },
      { id: "12", value: 4300, unit: "g", measuredAt: "2026-03-01T15:00:00+09:00" },
    ]);
  });

  it("weight_measurements 未設定の行は体重を持たない（他カルテの最新値を代理表示しない）", () => {
    const item = transformToHistoryItem(minimalResponse) as HistoryItemWithWeights;
    expect(item.weightMeasurements ?? []).toHaveLength(0);
  });

  it("copySource に体重測定を混入しない（複写ペイロードは問診値のみ）", () => {
    const record: MedicalRecordResponseWithWeights = {
      ...minimalResponse,
      weight_measurements: [
        { id: 11, value: 4.2, unit: "Kg", recorded_at: "2026-03-01T09:30:00+09:00" },
      ],
    };
    const item = transformToHistoryItem(record);
    expect(item.copySource).toEqual({ recordId: "1" });
  });
});

describe("useGetPetMedicalHistory — 取得エラーと0件の区別（EMR-356）", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("履歴取得エラーは isError で区別できる（記録なしと混同しない）", async () => {
    server.use(
      http.get("*/v1/medical-records", () => HttpResponse.json({ error: "boom" }, { status: 500 })),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      // EMR-356: 現行 hook は isError を返さず常に falsy のため waitFor timeout で RED。
      // 実装後は query の isError=true が返り待機が完了する。
      const current: ReturnType<typeof useGetPetMedicalHistory> & { isError?: boolean } =
        result.current;
      expect(current.isError).toBe(true);
    });
    expect(result.current.historyItems).toHaveLength(0);
  });

  it("履歴0件は isError=false の空配列（エラーとは区別される）", async () => {
    server.use(
      http.get("*/v1/medical-records", () =>
        HttpResponse.json({ data: [], total: 0, page: 1, limit: 50 }),
      ),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });
    expect(result.current.historyItems).toHaveLength(0);
    const current: ReturnType<typeof useGetPetMedicalHistory> & { isError?: boolean } =
      result.current;
    expect(current.isError).toBe(false);
  });
});

describe("useGetPetMedicalHistory — clinic コンテキストと行フィルタ（EMR-356）", () => {
  it("clinicId 指定時は X-Clinic-ID ヘッダを送る（別医院の履歴キャッシュを使い回さない）", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedClinicHeader: string | null = null;
    server.use(
      http.get("*/v1/medical-records", ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json({ data: [], total: 0, page: 1, limit: 50 });
      }),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "2"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(receivedClinicHeader).toBe("2");
  });

  it("clinicId 省略時は2引数呼び出しと同じ従来挙動（グローバル選択クリニックへフォールバック）", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    let receivedClinicHeader: string | null = null;
    server.use(
      http.get("*/v1/medical-records", ({ request }) => {
        receivedClinicHeader = request.headers.get("X-Clinic-ID");
        return HttpResponse.json({ data: [], total: 0, page: 1, limit: 50 });
      }),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    // axios インターセプタがグローバル選択 "1" を付与する（hook は明示ヘッダを送らない）
    expect(receivedClinicHeader).toBe("1");
  });

  it("応答行は要求した pet/clinic に一致するものだけを履歴へ投影する", async () => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");

    server.use(
      http.get("*/v1/medical-records", () =>
        HttpResponse.json({
          data: [
            { ...minimalResponse, id: 1, pet_id: 9, clinic_id: 1 },
            { ...minimalResponse, id: 2, pet_id: 8, clinic_id: 1 },
            { ...minimalResponse, id: 3, pet_id: 9, clinic_id: 2 },
          ],
          total: 3,
          page: 1,
          limit: 50,
        }),
      ),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "1"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.historyItems.map((item) => item.id)).toEqual(["1"]);
  });
});

// EMR-356 review-fix (EMR356-2): page=1&limit=50 は全履歴ではない。
// total/page/limit を使って全ページを取得するが、暴走防止の上限到達・
// 進捗なし・total 不整合は「未完」として明示し、断定的な「測定なし」を許さない。
describe("useGetPetMedicalHistory — 全ページ取得と未完の明示（EMR-356 review-fix）", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  const rows = (startId: number, count: number): MedicalRecordResponse[] =>
    Array.from({ length: count }, (_, i) => ({
      ...minimalResponse,
      id: startId + i,
      pet_id: 9,
      clinic_id: 1,
    }));

  const asPagedResult = (current: ReturnType<typeof useGetPetMedicalHistory>) =>
    current as ReturnType<typeof useGetPetMedicalHistory> & {
      historyIncomplete?: boolean;
      currentRecordDate?: string;
    };

  it.each(["repeated", "wrong-page", "changed-total"])(
    "%s は進捗不整合として未完を明示する",
    async (failure) => {
      const requestedPages: number[] = [];
      server.use(
        http.get("*/v1/medical-records", ({ request }) => {
          const page = Number(new URL(request.url).searchParams.get("page"));
          requestedPages.push(page);
          return HttpResponse.json({
            data: rows(failure === "repeated" ? 100 : 100 * page, 50),
            total: failure === "changed-total" && page === 2 ? 101 : 100,
            page: failure === "wrong-page" ? 1 : page,
            limit: 50,
          });
        }),
      );
      const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "1"), {
        wrapper: createTestWrapper(),
      });
      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.historyIncomplete).toBe(true);
      expect(requestedPages).toEqual([1, 2]);
    },
  );

  it("同一ページの重複も全件取得と断定しない", async () => {
    server.use(
      http.get("*/v1/medical-records", () =>
        HttpResponse.json({
          data: [...rows(100, 1), ...rows(100, 1)],
          total: 2,
          page: 1,
          limit: 50,
        }),
      ),
    );
    const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "1"), {
      wrapper: createTestWrapper(),
    });
    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.historyIncomplete).toBe(true);
    expect(result.current.historyItems).toHaveLength(1);
  });

  it("51件超: page2 の行も履歴へ届き、行ローカル weightMeasurements を保持する", async () => {
    const requestedPages: string[] = [];
    server.use(
      http.get("*/v1/medical-records", ({ request }) => {
        const page = new URL(request.url).searchParams.get("page") ?? "1";
        requestedPages.push(page);
        if (page === "1") {
          return HttpResponse.json({ data: rows(100, 50), total: 51, page: 1, limit: 50 });
        }
        return HttpResponse.json({
          data: [
            {
              ...minimalResponse,
              id: 150,
              pet_id: 9,
              clinic_id: 1,
              weight_measurements: [
                { id: 11, value: 4.2, unit: "Kg", recorded_at: "2026-01-05T09:30:00+09:00" },
              ],
            },
          ],
          total: 51,
          page: 2,
          limit: 50,
        });
      }),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "1"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(asPagedResult(result.current).historyIncomplete).toBe(false);
    expect(requestedPages).toContain("2");
    expect(result.current.historyItems).toHaveLength(51);
    const weighted = result.current.historyItems.find((item) => item.id === "150");
    expect(weighted?.weightMeasurements).toEqual([
      { id: "11", value: 4.2, unit: "Kg", measuredAt: "2026-01-05T09:30:00+09:00" },
    ]);
  });

  it("bounded: total が上限ページを超えるときは打ち切り historyIncomplete=true で明示する", async () => {
    const requestedPages: string[] = [];
    server.use(
      http.get("*/v1/medical-records", ({ request }) => {
        const page = Number(new URL(request.url).searchParams.get("page") ?? "1");
        requestedPages.push(String(page));
        return HttpResponse.json({
          data: rows(1000 + page * 100, 50),
          total: 5000,
          page,
          limit: 50,
        });
      }),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "1"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(requestedPages).toHaveLength(20);
    expect(asPagedResult(result.current).historyIncomplete).toBe(true);
  });

  it("total と矛盾する空ページは履歴未完を明示する", async () => {
    server.use(
      http.get("*/v1/medical-records", ({ request }) => {
        const page = new URL(request.url).searchParams.get("page") ?? "1";
        if (page === "1") {
          return HttpResponse.json({ data: rows(100, 50), total: 51, page: 1, limit: 50 });
        }
        return HttpResponse.json({ data: [], total: 51, page: 2, limit: 50 });
      }),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "1"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(asPagedResult(result.current).historyIncomplete).toBe(true);
    expect(result.current.historyItems).toHaveLength(50);
  });

  it("total 未提供でも返却行が limit 未満なら完全（追加リクエストなし）", async () => {
    const requestedPages: string[] = [];
    server.use(
      http.get("*/v1/medical-records", ({ request }) => {
        const page = new URL(request.url).searchParams.get("page") ?? "1";
        requestedPages.push(page);
        return HttpResponse.json({ data: rows(100, 30), page: 1, limit: 50 });
      }),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "1"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(asPagedResult(result.current).historyIncomplete).toBe(false);
    expect(requestedPages).toEqual(["1"]);
  });

  it("total 未提供で満ページのときは次ページを probe し、短いページで完了する", async () => {
    server.use(
      http.get("*/v1/medical-records", ({ request }) => {
        const page = new URL(request.url).searchParams.get("page") ?? "1";
        if (page === "1") {
          return HttpResponse.json({ data: rows(100, 50), page: 1, limit: 50 });
        }
        return HttpResponse.json({ data: rows(200, 10), page: 2, limit: 50 });
      }),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "1"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(asPagedResult(result.current).historyIncomplete).toBe(false);
    expect(result.current.historyItems).toHaveLength(60);
  });

  it("2ページ目以降の取得エラーは isError（部分履歴を「記録なし」としない）", async () => {
    server.use(
      http.get("*/v1/medical-records", ({ request }) => {
        const page = new URL(request.url).searchParams.get("page") ?? "1";
        if (page === "1") {
          return HttpResponse.json({ data: rows(100, 50), total: 51, page: 1, limit: 50 });
        }
        return HttpResponse.json({ error: "boom" }, { status: 500 });
      }),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "1"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(result.current.historyItems).toHaveLength(0);
  });

  it("表示中カルテの行は履歴から除外しつつ wire record.date を currentRecordDate で返す", async () => {
    server.use(
      http.get("*/v1/medical-records", () =>
        HttpResponse.json({
          data: [
            {
              ...minimalResponse,
              id: 42,
              pet_id: 9,
              clinic_id: 1,
              date: "2026-03-01T14:00:00+09:00",
            },
            {
              ...minimalResponse,
              id: 41,
              pet_id: 9,
              clinic_id: 1,
              date: "2026-02-01T00:00:00+09:00",
            },
          ],
          total: 2,
          page: 1,
          limit: 50,
        }),
      ),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9", "42", "1"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.historyItems.map((item) => item.id)).toEqual(["41"]);
    // 元の日付表現を保持し、呼び出し側でカルテ日付のJST開始へ正規化する。
    expect(asPagedResult(result.current).currentRecordDate).toBe("2026-03-01T14:00:00+09:00");
  });

  it("page2 の行が weight_measurements キー未提供でも undefined を維持する（legacy 投影との区別）", async () => {
    server.use(
      http.get("*/v1/medical-records", ({ request }) => {
        const page = new URL(request.url).searchParams.get("page") ?? "1";
        if (page === "1") {
          return HttpResponse.json({ data: rows(100, 50), total: 51, page: 1, limit: 50 });
        }
        return HttpResponse.json({
          data: [{ ...minimalResponse, id: 150, pet_id: 9, clinic_id: 1 }],
          total: 51,
          page: 2,
          limit: 50,
        });
      }),
    );

    const { result } = renderHook(() => useGetPetMedicalHistory("9", undefined, "1"), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    const legacy = result.current.historyItems.find((item) => item.id === "150");
    expect(legacy).toBeDefined();
    expect(legacy?.weightMeasurements).toBeUndefined();
  });
});
