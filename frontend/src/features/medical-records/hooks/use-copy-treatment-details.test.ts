import { createElement } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { queryKeys } from "@/lib/query-keys";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";

import { createTreatment, getTreatments } from "../api/treatments";
import { buildMasterSelectionPayload } from "../lib/treatments-tab-model";
import { useCopyTreatmentDetails } from "./use-copy-treatment-details";
import type { Treatment } from "../types";

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

const SOURCE_ID = "101";
const DEST_ID = "200";

function sourceTreatment(overrides: Partial<Treatment> = {}): Treatment {
  return {
    id: "500",
    medical_record_id: SOURCE_ID,
    item_type: "consultation",
    unit_price: 1500,
    quantity: 1,
    is_selected: true,
    status: "pending",
    content: "初診料",
    memo: "",
    is_insurance: false,
    discount_rate: 0,
    discount_amount: 0,
    sort_order: 0,
    version: 1,
    created_at: "2026-02-01T00:00:00Z",
    updated_at: "2026-02-01T00:00:00Z",
    ...overrides,
  };
}

/** 複写に必要な 3 系統 + masters のハンドラを一括登録し、POST ボディを記録する。 */
function useCopyHandlers(input: {
  sourceRows: Treatment[];
  destRows?: Treatment[];
  consultations?: unknown[];
  procedures?: unknown[];
  medicines?: unknown[];
  postResponder?: (body: unknown, index: number) => HttpResponse;
  sourceResponder?: () => HttpResponse;
}) {
  const postBodies: unknown[] = [];
  server.use(
    http.get(`*/v1/medical-records/${SOURCE_ID}/treatments`, () =>
      input.sourceResponder ? input.sourceResponder() : HttpResponse.json(input.sourceRows),
    ),
    http.get(`*/v1/medical-records/${DEST_ID}/treatments`, () =>
      HttpResponse.json(input.destRows ?? []),
    ),
    http.get("*/v1/masters/consultations", () => HttpResponse.json(input.consultations ?? [])),
    http.get("*/v1/masters/procedures", () => HttpResponse.json(input.procedures ?? [])),
    http.get("*/v1/masters/medicines", () => HttpResponse.json(input.medicines ?? [])),
    http.post(`*/v1/medical-records/${DEST_ID}/treatments`, async ({ request }) => {
      const body = await request.json();
      postBodies.push(body);
      return input.postResponder
        ? input.postResponder(body, postBodies.length - 1)
        : HttpResponse.json({ id: postBodies.length }, { status: 201 });
    }),
  );
  return postBodies;
}

describe("useCopyTreatmentDetails", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("元カルテの明細を sort_order 昇順で逐次 POST し、現行マスタ価格へ再解決する", async () => {
    const postBodies = useCopyHandlers({
      // 降順で返して、クライアント側で昇順ソートされることを検証する
      sourceRows: [
        sourceTreatment({
          id: "502",
          item_type: "medicine",
          medicine_id: "42",
          consultation_id: undefined,
          content: "アモキシシリン",
          unit_price: 80,
          sort_order: 1,
        }),
        sourceTreatment({
          id: "501",
          item_type: "consultation",
          consultation_id: "7",
          content: "初診料",
          unit_price: 1500,
          sort_order: 0,
        }),
      ],
      // 複写先に既存行 sort_order=4 → offset は 5
      destRows: [
        sourceTreatment({
          id: "900",
          medical_record_id: DEST_ID,
          sort_order: 4,
        }),
      ],
      consultations: [{ id: 7, name: "初診料", price: 2500, is_active: true }],
      medicines: [{ id: 42, name: "アモキシシリン", price: 120, is_active: true }],
    });

    const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), {
      wrapper: createTestWrapper(),
    });

    await act(async () => {
      await result.current.copyTreatmentsFromRecord(SOURCE_ID);
    });

    expect(postBodies).toHaveLength(2);
    // 1件目: ソース sort_order=0 の consultation 行 → マスタ現行価格 2500 / sort_order 0+5
    const first = postBodies[0] as Record<string, unknown>;
    expect(first.item_type).toBe("consultation");
    expect(first.consultation_id).toBe(7);
    expect(first.unit_price).toBe(2500);
    expect(first.sort_order).toBe(5);
    expect(first.content).toBe("初診料");
    // 2件目: ソース sort_order=1 の medicine 行 → medicine_id は JSON number / sort_order 1+5
    const second = postBodies[1] as Record<string, unknown>;
    expect(second.item_type).toBe("medicine");
    expect(second.medicine_id).toBe(42);
    expect(typeof second.medicine_id).toBe("number");
    expect(second.unit_price).toBe(120);
    expect(second.sort_order).toBe(6);
    expect(toast.success).toHaveBeenCalledTimes(1);
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("複写先が空なら offset なしでソース sort_order を保持する", async () => {
    const postBodies = useCopyHandlers({
      // 参照を持たない手入力 other 行はソース値をそのまま複写する（offset=0 で sort_order 維持）
      sourceRows: [
        sourceTreatment({ item_type: "other", consultation_id: undefined, sort_order: 2 }),
      ],
    });

    const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID), {
      wrapper: createTestWrapper(),
    });

    await act(async () => {
      await result.current.copyTreatmentsFromRecord(SOURCE_ID);
    });

    expect(postBodies).toHaveLength(1);
    const body = postBodies[0] as Record<string, unknown>;
    expect(body.item_type).toBe("other");
    expect(body.sort_order).toBe(2);
    expect(body.unit_price).toBe(1500);
    expect(body.consultation_id).toBeUndefined();
  });

  it("現在の recordId が無いときは fetch/create とも行わず早期 return する", async () => {
    const postBodies = useCopyHandlers({ sourceRows: [sourceTreatment()] });

    const { result } = renderHook(() => useCopyTreatmentDetails(undefined), {
      wrapper: createTestWrapper(),
    });

    await act(async () => {
      await result.current.copyTreatmentsFromRecord(SOURCE_ID);
    });

    expect(postBodies).toHaveLength(0);
    expect(toast.success).not.toHaveBeenCalled();
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("複写元の明細 fetch が失敗したときエラートーストを出し create しない", async () => {
    const postBodies = useCopyHandlers({
      sourceRows: [],
      sourceResponder: () => HttpResponse.json({ message: "server error" }, { status: 500 }),
    });

    const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID), {
      wrapper: createTestWrapper(),
    });

    await act(async () => {
      await result.current.copyTreatmentsFromRecord(SOURCE_ID);
    });

    expect(postBodies).toHaveLength(0);
    expect(toast.error).toHaveBeenCalledTimes(1);
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("create が途中で失敗したとき後続行を POST せず中断する（作成済み行は残る）", async () => {
    const postBodies = useCopyHandlers({
      sourceRows: [
        sourceTreatment({ id: "501", sort_order: 0, consultation_id: "7" }),
        sourceTreatment({ id: "502", sort_order: 1, consultation_id: "7" }),
        sourceTreatment({ id: "503", sort_order: 2, consultation_id: "7" }),
      ],
      consultations: [{ id: 7, name: "初診料", price: 2500, is_active: true }],
      postResponder: (_body, index) =>
        index === 1
          ? HttpResponse.json({ message: "conflict" }, { status: 500 })
          : HttpResponse.json({ id: index + 1 }, { status: 201 }),
    });

    const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID), {
      wrapper: createTestWrapper(),
    });

    await act(async () => {
      await result.current.copyTreatmentsFromRecord(SOURCE_ID);
    });

    // 1件目成功・2件目失敗で中断 → 3件目は送信されない
    expect(postBodies).toHaveLength(2);
    await waitFor(() => expect(toast.error).toHaveBeenCalled());
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("複写元に明細が無いとき info トーストを出し create しない", async () => {
    const postBodies = useCopyHandlers({ sourceRows: [] });

    const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID), {
      wrapper: createTestWrapper(),
    });

    await act(async () => {
      await result.current.copyTreatmentsFromRecord(SOURCE_ID);
    });

    expect(postBodies).toHaveLength(0);
    expect(toast.info).toHaveBeenCalledTimes(1);
  });

  it("マスタ選択→保存→複写の wire 往復で、診察/処置/薬剤は現行マスタ価格へ再解決される", async () => {
    let sourceStore: Treatment[] = [];
    let destStore: Treatment[] = [];
    let sourcePostHeaders: Array<string | null> = [];
    let sourceGetHeaders: Array<string | null> = [];
    let destPostHeaders: Array<string | null> = [];
    let destGetHeaders: Array<string | null> = [];

    /** create 要求の数値参照を wire 応答の文字列参照へ変換する合成 wire 行の組立（実 backend 証明ではない）。 */
    const toWireRow = (recordId: string, body: Record<string, unknown>): Treatment => ({
      id: `${recordId}-row-${destStore.length + sourceStore.length}`,
      medical_record_id: recordId,
      item_type: body.item_type as Treatment["item_type"],
      consultation_id: body.consultation_id != null ? String(body.consultation_id) : undefined,
      procedure_id: body.procedure_id != null ? String(body.procedure_id) : undefined,
      medicine_id: body.medicine_id != null ? String(body.medicine_id) : undefined,
      unit_price: body.unit_price as number,
      quantity: (body.quantity as number) ?? 1,
      is_selected: true,
      status: "pending",
      content: (body.content as string) ?? "",
      memo: (body.memo as string) ?? "",
      is_insurance: false,
      discount_rate: 0,
      discount_amount: 0,
      sort_order: (body.sort_order as number) ?? 0,
      version: 1,
      created_at: "2026-10-01T00:00:00Z",
      updated_at: "2026-10-01T00:00:00Z",
    });

    server.use(
      http.post(`*/v1/medical-records/${SOURCE_ID}/treatments`, async ({ request }) => {
        sourcePostHeaders = [...sourcePostHeaders, request.headers.get("X-Clinic-ID")];
        const body = (await request.json()) as Record<string, unknown>;
        const row = toWireRow(SOURCE_ID, body);
        sourceStore = [...sourceStore, row];
        return HttpResponse.json(row, { status: 201 });
      }),
      http.post(`*/v1/medical-records/${DEST_ID}/treatments`, async ({ request }) => {
        destPostHeaders = [...destPostHeaders, request.headers.get("X-Clinic-ID")];
        const body = (await request.json()) as Record<string, unknown>;
        const row = toWireRow(DEST_ID, body);
        destStore = [...destStore, row];
        return HttpResponse.json(row, { status: 201 });
      }),
      http.get(`*/v1/medical-records/${SOURCE_ID}/treatments`, ({ request }) => {
        sourceGetHeaders = [...sourceGetHeaders, request.headers.get("X-Clinic-ID")];
        return HttpResponse.json(sourceStore);
      }),
      http.get(`*/v1/medical-records/${DEST_ID}/treatments`, ({ request }) => {
        destGetHeaders = [...destGetHeaders, request.headers.get("X-Clinic-ID")];
        return HttpResponse.json(destStore);
      }),
      http.get("*/v1/masters/consultations", () =>
        HttpResponse.json([{ id: 7, name: "初診", price: 3307, is_active: true }]),
      ),
      http.get("*/v1/masters/procedures", () =>
        HttpResponse.json([{ id: 8, name: "創傷処置", price: 3306, is_active: true }]),
      ),
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([{ id: 42, name: "アモキシシリン", price: 3303, is_active: true }]),
      ),
    );

    const selections: TreatmentMasterItem[] = [
      {
        id: "7",
        name: "初診",
        unitPrice: 3307,
        insuranceContract: "master",
        isNonInsurance: true,
        category: "診察",
      },
      {
        id: "8",
        name: "創傷処置",
        unitPrice: 3306,
        insuranceContract: "master",
        isNonInsurance: true,
        category: "処置",
      },
      {
        id: "42",
        name: "アモキシシリン",
        unitPrice: 3303,
        insuranceContract: "master",
        isNonInsurance: true,
        category: "薬剤",
        medicineId: "42",
      },
      { id: "7", name: "検査キット", unitPrice: 2000, isNonInsurance: true, category: "検査" },
    ];
    for (const [index, item] of selections.entries()) {
      await createTreatment(
        SOURCE_ID,
        buildMasterSelectionPayload({ item, quantity: 1, sortOrder: index }),
      );
    }
    expect(sourceStore).toHaveLength(4);
    expect(sourceStore[0]).toMatchObject({ consultation_id: "7", item_type: "consultation" });
    expect(sourceStore[1]).toMatchObject({ procedure_id: "8", item_type: "procedure" });
    expect(sourceStore[2]).toMatchObject({ medicine_id: "42", item_type: "medicine" });
    expect(sourceStore[3]).toMatchObject({ item_type: "other" });
    expect(sourceStore[3].consultation_id).toBeUndefined();

    const overriddenPrices = [710, 709, 703];
    sourceStore = sourceStore.map((row, index) =>
      index < overriddenPrices.length ? { ...row, unit_price: overriddenPrices[index] } : row,
    );

    const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), {
      wrapper: createTestWrapper(),
    });

    await act(async () => {
      await Promise.all([
        result.current.copyTreatmentsFromRecord(SOURCE_ID),
        result.current.copyTreatmentsFromRecord(SOURCE_ID),
      ]);
    });

    // 複写先マスタの GET は複写先 recordClinicId="3" でスコープされる（全3系統）
    // ※後続の clinic-scope describe で GET ヘッダを直接観測する

    const savedDestRows = await getTreatments(DEST_ID, "3");

    expect(savedDestRows).toMatchObject([
      {
        item_type: "consultation",
        consultation_id: "7",
        content: "初診",
        unit_price: 3307,
        sort_order: 0,
      },
      {
        item_type: "procedure",
        procedure_id: "8",
        content: "創傷処置",
        unit_price: 3306,
        sort_order: 1,
      },
      {
        item_type: "medicine",
        medicine_id: "42",
        content: "アモキシシリン",
        unit_price: 3303,
        sort_order: 2,
      },
      { item_type: "other", content: "検査キット", unit_price: 2000, sort_order: 3 },
    ]);
    expect(savedDestRows).toHaveLength(4);

    expect(sourcePostHeaders).toHaveLength(4);
    expect(sourcePostHeaders.every((h) => h !== "3")).toBe(true);
    expect(sourceGetHeaders.length).toBeGreaterThan(0);
    expect(sourceGetHeaders.every((h) => h !== "3")).toBe(true);
    expect(destPostHeaders).toHaveLength(4);
    expect(destPostHeaders.every((h) => h === "3")).toBe(true);
    expect(destGetHeaders.length).toBeGreaterThan(0);
    expect(destGetHeaders.every((h) => h === "3")).toBe(true);

    expect(sourceStore[0].unit_price).toBe(710);
    expect(sourceStore[1].unit_price).toBe(709);
    expect(sourceStore[2].unit_price).toBe(703);

    expect(toast.success).toHaveBeenCalledTimes(1);
    expect(toast.error).not.toHaveBeenCalled();
  });

  describe("PR670: 複写先マスタは recordClinicId でスコープし不一致は zero-write で拒否する", () => {
    it("recordClinicId='3' のとき3系統のマスタ GET に X-Clinic-ID=3 を送る", async () => {
      const masterGetHeaders: Array<string | null> = [];
      server.use(
        http.get(`*/v1/medical-records/${SOURCE_ID}/treatments`, () =>
          HttpResponse.json([sourceTreatment({ consultation_id: "7" })]),
        ),
        http.get(`*/v1/medical-records/${DEST_ID}/treatments`, () => HttpResponse.json([])),
        http.get("*/v1/masters/consultations", ({ request }) => {
          masterGetHeaders.push(`consultations:${request.headers.get("X-Clinic-ID")}`);
          return HttpResponse.json([{ id: 7, name: "初診料", price: 2500, is_active: true }]);
        }),
        http.get("*/v1/masters/procedures", ({ request }) => {
          masterGetHeaders.push(`procedures:${request.headers.get("X-Clinic-ID")}`);
          return HttpResponse.json([]);
        }),
        http.get("*/v1/masters/medicines", ({ request }) => {
          masterGetHeaders.push(`medicines:${request.headers.get("X-Clinic-ID")}`);
          return HttpResponse.json([]);
        }),
        http.post(`*/v1/medical-records/${DEST_ID}/treatments`, async () =>
          HttpResponse.json({ id: "x" }, { status: 201 }),
        ),
      );

      const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), {
        wrapper: createTestWrapper(),
      });
      await act(async () => {
        await result.current.copyTreatmentsFromRecord(SOURCE_ID);
      });

      expect(masterGetHeaders).toEqual(["consultations:3", "procedures:3", "medicines:3"]);
      expect(toast.success).toHaveBeenCalledTimes(1);
    });

    it("PR670: 拒否理由は安全な固定メッセージを toast.error へ出す（generic ではない）", async () => {
      const postBodies = useCopyHandlers({
        sourceRows: [sourceTreatment({ consultation_id: "999" })],
        consultations: [{ id: 7, name: "初診料", price: 2500, is_active: true }],
      });

      const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), {
        wrapper: createTestWrapper(),
      });
      await act(async () => {
        await result.current.copyTreatmentsFromRecord(SOURCE_ID);
      });

      expect(postBodies).toHaveLength(0);
      expect(toast.error).toHaveBeenCalledWith(
        "複写先のマスタを確認し、治療明細を再選択してください",
      );
      expect(toast.error).toHaveBeenCalledTimes(1);
      expect(toast.success).not.toHaveBeenCalled();
    });

    it("複写先マスタ不一致の typed 行がある場合、ペイロード構築段で例外 → POST 0件", async () => {
      const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
      const wrapper = ({ children }: { children: ReactNode }) =>
        createElement(QueryClientProvider, { client: queryClient }, children);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
      // 999 はグローバルキャッシュにだけ存在 → scoped dest では不一致のまま拒否される
      queryClient.setQueryData(queryKeys.masters.category("consultations"), [
        { id: "999", name: "旧診察", price: 1 },
      ]);
      const postBodies = useCopyHandlers({
        sourceRows: [
          sourceTreatment({ id: "501", sort_order: 0, consultation_id: "7" }),
          sourceTreatment({ id: "502", sort_order: 1, consultation_id: "999" }),
        ],
        consultations: [{ id: 7, name: "初診料", price: 2500, is_active: true }],
      });

      const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), { wrapper });
      await act(async () => {
        await result.current.copyTreatmentsFromRecord(SOURCE_ID);
      });

      // 先行の正当行(501)も POST されない: 全ペイロードを create 前に構築する
      expect(postBodies).toHaveLength(0);
      expect(invalidateSpy).not.toHaveBeenCalled();
      await waitFor(() => expect(toast.error).toHaveBeenCalled());
      expect(toast.error).toHaveBeenCalledTimes(1);
      expect(toast.success).not.toHaveBeenCalled();
    });

    it("cold scoped fetch: グローバル1キャッシュを毒しても dest=3 の値だけ使いヘッダを分離する", async () => {
      localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");
      try {
        const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        const wrapper = ({ children }: { children: ReactNode }) =>
          createElement(QueryClientProvider, { client: queryClient }, children);
        for (const cat of ["consultations", "procedures", "medicines"]) {
          queryClient.setQueryData(queryKeys.masters.category(cat), [
            { id: "999", name: "X", price: 11 },
          ]);
        }

        const sourceRows = [
          sourceTreatment({
            id: "503",
            item_type: "medicine",
            medicine_id: "42",
            consultation_id: undefined,
            unit_price: 1,
            sort_order: 2,
          }),
          sourceTreatment({
            id: "502",
            item_type: "procedure",
            procedure_id: "8",
            consultation_id: undefined,
            unit_price: 1,
            sort_order: 1,
          }),
          sourceTreatment({ id: "501", consultation_id: "7", unit_price: 1, sort_order: 0 }),
        ];
        const sourceSnapshot = structuredClone(sourceRows);
        const postBodies: Array<Record<string, unknown>> = [];
        const seen: string[] = [];

        server.use(
          http.get(`*/v1/medical-records/${SOURCE_ID}/treatments`, ({ request }) => {
            seen.push(`src:${request.headers.get("X-Clinic-ID")}`);
            return HttpResponse.json(sourceRows);
          }),
          http.get(`*/v1/medical-records/${DEST_ID}/treatments`, ({ request }) => {
            seen.push(`dstGet:${request.headers.get("X-Clinic-ID")}`);
            return HttpResponse.json([
              sourceTreatment({ id: "900", medical_record_id: DEST_ID, sort_order: 4 }),
            ]);
          }),
          http.get("*/v1/masters/consultations", ({ request }) => {
            seen.push(`cons:${request.headers.get("X-Clinic-ID")}`);
            return HttpResponse.json([{ id: 7, name: "初診料", price: 3307, is_active: true }]);
          }),
          http.get("*/v1/masters/procedures", ({ request }) => {
            seen.push(`proc:${request.headers.get("X-Clinic-ID")}`);
            return HttpResponse.json([{ id: 8, name: "処置X", price: 3306, is_active: true }]);
          }),
          http.get("*/v1/masters/medicines", ({ request }) => {
            seen.push(`med:${request.headers.get("X-Clinic-ID")}`);
            return HttpResponse.json([{ id: 42, name: "薬A", price: 3303, is_active: true }]);
          }),
          http.post(`*/v1/medical-records/${DEST_ID}/treatments`, async ({ request }) => {
            seen.push(`post:${request.headers.get("X-Clinic-ID")}`);
            const body = (await request.json()) as Record<string, unknown>;
            postBodies.push(body);
            return HttpResponse.json(body, { status: 201 });
          }),
        );

        const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), { wrapper });
        await act(async () => {
          await result.current.copyTreatmentsFromRecord(SOURCE_ID);
        });

        expect(seen.filter((x) => x.startsWith("src:"))).toEqual(["src:1"]);
        expect(seen.filter((x) => x.startsWith("dstGet:"))).toEqual(["dstGet:3"]);
        expect(seen.filter((x) => x === "post:3")).toHaveLength(3);
        expect(seen).toEqual(expect.arrayContaining(["cons:3", "proc:3", "med:3"]));
        expect(seen.every((x) => !x.endsWith(":999"))).toBe(true);

        expect(postBodies.map((b) => b.sort_order)).toEqual([5, 6, 7]);
        expect(postBodies[0].consultation_id).toBe(7);
        expect(postBodies[0].unit_price).toBe(3307);
        expect(postBodies[1].procedure_id).toBe(8);
        expect(postBodies[1].unit_price).toBe(3306);
        expect(postBodies[2].medicine_id).toBe(42);
        expect(postBodies[2].unit_price).toBe(3303);
        expect(postBodies.every((b) => b.consultation_id !== 999)).toBe(true);
        expect(sourceRows).toEqual(sourceSnapshot);
        expect(toast.success).toHaveBeenCalledTimes(1);
      } finally {
        localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
      }
    });

    it("複写先 scoped キャッシュが温まっている場合、マスタ GET は一切発行しない", async () => {
      const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
      const wrapper = ({ children }: { children: ReactNode }) =>
        createElement(QueryClientProvider, { client: queryClient }, children);

      for (const cat of ["consultations", "procedures", "medicines"]) {
        queryClient.setQueryData(queryKeys.masters.category(cat), [
          { id: "999", name: "X", price: 11 },
        ]);
      }
      queryClient.setQueryData(queryKeys.masters.categoryForClinic("consultations", "3"), [
        { id: "7", name: "初診料", price: 3307, isActive: true },
      ]);
      queryClient.setQueryData(queryKeys.masters.categoryForClinic("procedures", "3"), [
        { id: "8", name: "処置X", price: 3306, isActive: true },
      ]);
      queryClient.setQueryData(queryKeys.masters.categoryForClinic("medicines", "3"), [
        { id: "42", name: "薬A", price: 3303, isActive: true },
      ]);

      let masterRequests = 0;
      const postBodies = useCopyHandlers({
        sourceRows: [
          sourceTreatment({
            id: "503",
            item_type: "medicine",
            medicine_id: "42",
            consultation_id: undefined,
            unit_price: 1,
            sort_order: 2,
          }),
          sourceTreatment({
            id: "502",
            item_type: "procedure",
            procedure_id: "8",
            consultation_id: undefined,
            unit_price: 1,
            sort_order: 1,
          }),
          sourceTreatment({ id: "501", consultation_id: "7", unit_price: 1, sort_order: 0 }),
        ],
      });
      server.use(
        http.get("*/v1/masters/consultations", () => {
          masterRequests += 1;
          return HttpResponse.json([]);
        }),
        http.get("*/v1/masters/procedures", () => {
          masterRequests += 1;
          return HttpResponse.json([]);
        }),
        http.get("*/v1/masters/medicines", () => {
          masterRequests += 1;
          return HttpResponse.json([]);
        }),
      );

      const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), { wrapper });
      await act(async () => {
        await result.current.copyTreatmentsFromRecord(SOURCE_ID);
      });

      expect(masterRequests).toBe(0);
      expect(postBodies).toHaveLength(3);
      const [c, pr, m] = postBodies as Array<Record<string, unknown>>;
      expect(c.consultation_id).toBe(7);
      expect(c.unit_price).toBe(3307);
      expect(pr.procedure_id).toBe(8);
      expect(pr.unit_price).toBe(3306);
      expect(m.medicine_id).toBe(42);
      expect(m.unit_price).toBe(3303);
      expect(toast.success).toHaveBeenCalledTimes(1);
    });

    it("複写先マスタ GET が 403 でもグローバルキャッシュへフォールバックせず 0 POST", async () => {
      localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");
      try {
        const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        queryClient.setQueryData(queryKeys.masters.category("consultations"), [
          { id: "7", name: "初診料", price: 11 },
        ]);
        const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
        const wrapper = ({ children }: { children: ReactNode }) =>
          createElement(QueryClientProvider, { client: queryClient }, children);
        const postBodies = useCopyHandlers({
          sourceRows: [sourceTreatment({ consultation_id: "7" })],
        });
        server.use(
          http.get("*/v1/masters/consultations", () =>
            HttpResponse.json({ message: "forbidden" }, { status: 403 }),
          ),
        );

        const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), { wrapper });
        await act(async () => {
          await result.current.copyTreatmentsFromRecord(SOURCE_ID);
        });

        expect(postBodies).toHaveLength(0);
        expect(invalidateSpy).not.toHaveBeenCalled();
        await waitFor(() => expect(toast.error).toHaveBeenCalled());
        expect(toast.error).toHaveBeenCalledTimes(1);
        expect(toast.success).not.toHaveBeenCalled();
      } finally {
        localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
      }
    });

    it("recordClinicId 省略・空文字は従来のグローバルキー/暗黙ヘッダを維持する", async () => {
      const masterGetHeaders: Array<string | null> = [];
      server.use(
        http.get(`*/v1/medical-records/${SOURCE_ID}/treatments`, () =>
          HttpResponse.json([sourceTreatment({ consultation_id: "7" })]),
        ),
        http.get(`*/v1/medical-records/${DEST_ID}/treatments`, () => HttpResponse.json([])),
        http.get("*/v1/masters/consultations", ({ request }) => {
          masterGetHeaders.push(request.headers.get("X-Clinic-ID"));
          return HttpResponse.json([{ id: 7, name: "初診料", price: 2500, is_active: true }]);
        }),
        http.get("*/v1/masters/procedures", () => HttpResponse.json([])),
        http.get("*/v1/masters/medicines", () => HttpResponse.json([])),
        http.post(`*/v1/medical-records/${DEST_ID}/treatments`, async () =>
          HttpResponse.json({ id: "x" }, { status: 201 }),
        ),
      );

      localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");
      try {
        for (const clinicId of [undefined, ""]) {
          const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, clinicId), {
            wrapper: createTestWrapper(),
          });
          await act(async () => {
            await result.current.copyTreatmentsFromRecord(SOURCE_ID);
          });
        }
      } finally {
        localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
      }
      expect(masterGetHeaders.length).toBeGreaterThanOrEqual(2);
      expect(masterGetHeaders.every((h) => h === "1")).toBe(true);
    });

    it("ネットワーク途中失敗でも作成済み分は 1 回だけ invalidate する", async () => {
      const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
      const wrapper = ({ children }: { children: ReactNode }) =>
        createElement(QueryClientProvider, { client: queryClient }, children);
      const postBodies = useCopyHandlers({
        sourceRows: [
          sourceTreatment({ id: "501", sort_order: 0, consultation_id: "7" }),
          sourceTreatment({ id: "502", sort_order: 1, consultation_id: "7" }),
          sourceTreatment({ id: "503", sort_order: 2, consultation_id: "7" }),
        ],
        consultations: [{ id: 7, name: "初診料", price: 2500, is_active: true }],
        postResponder: (_body, index) =>
          index === 1
            ? HttpResponse.json({ message: "conflict" }, { status: 500 })
            : HttpResponse.json({ id: index + 1 }, { status: 201 }),
      });

      const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), { wrapper });
      await act(async () => {
        await result.current.copyTreatmentsFromRecord(SOURCE_ID);
      });

      expect(postBodies).toHaveLength(2);
      expect(invalidateSpy).toHaveBeenCalledTimes(1);
      expect(invalidateSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          queryKey: queryKeys.medicalRecords.treatments(DEST_ID),
        }),
      );
      await waitFor(() => expect(toast.error).toHaveBeenCalled());
    });
  });

  it("PR700 inventory: 複写先マスタの inventoryId=77 を採用しソース 999 は使わない", async () => {
    const postBodies = useCopyHandlers({
      sourceRows: [
        sourceTreatment({
          item_type: "medicine",
          medicine_id: "42",
          consultation_id: undefined,
          inventory_id: "999",
          unit_price: 80,
        }),
      ],
      medicines: [{ id: 42, name: "抗生剤A", price: 3303, is_active: true, inventory_id: 77 }],
    });

    const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), {
      wrapper: createTestWrapper(),
    });
    await act(async () => {
      await result.current.copyTreatmentsFromRecord(SOURCE_ID);
    });

    expect(postBodies).toHaveLength(1);
    const body = postBodies[0] as Record<string, unknown>;
    expect(body.inventory_id).toBe(77);
    expect(toast.success).toHaveBeenCalledTimes(1);
  });

  it.each([undefined, "0", "abc"])(
    "PR700 inventory: later invalid destination inventory %s is zero-write",
    async (inventoryId) => {
      const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
      const wrapper = ({ children }: { children: ReactNode }) =>
        createElement(QueryClientProvider, { client: queryClient }, children);
      const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
      const global = [{ id: "42", price: 80, isActive: true, inventoryId: "999" }];
      const destination = [{ id: "42", price: 3303, isActive: true, inventoryId }];
      queryClient.setQueryData(queryKeys.masters.category("medicines"), structuredClone(global));
      queryClient.setQueryData(
        queryKeys.masters.categoryForClinic("medicines", "3"),
        structuredClone(destination),
      );
      const sourceRows = [
        sourceTreatment({
          id: "first",
          item_type: "other",
          consultation_id: undefined,
          sort_order: 0,
        }),
        sourceTreatment({
          id: "second",
          item_type: "medicine",
          medicine_id: "42",
          consultation_id: undefined,
          inventory_id: "999",
          unit_price: 80,
          sort_order: 1,
        }),
      ];
      const original = structuredClone(sourceRows);
      const postBodies = useCopyHandlers({
        sourceRows,
        medicines: [{ id: 42, name: "抗生剤A", price: 3303, is_active: true }],
      });

      const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), {
        wrapper,
      });
      await act(async () => {
        await result.current.copyTreatmentsFromRecord(SOURCE_ID);
      });

      expect(postBodies).toHaveLength(0);
      expect(toast.error).toHaveBeenCalledWith(
        "複写先のマスタを確認し、治療明細を再選択してください",
      );
      expect(invalidateSpy).not.toHaveBeenCalled();
      expect(sourceRows).toEqual(original);
      expect(queryClient.getQueryData(queryKeys.masters.category("medicines"))).toEqual(global);
      expect(
        queryClient.getQueryData(queryKeys.masters.categoryForClinic("medicines", "3")),
      ).toEqual(destination);
    },
  );

  describe("PR700 inactive: 複写先マスタの inactive 行は再選択促進で拒否する", () => {
    const TYPES = [
      {
        key: "consultations" as const,
        refKey: "consultation_id" as const,
        refId: "7",
        itemType: "consultation" as const,
      },
      {
        key: "procedures" as const,
        refKey: "procedure_id" as const,
        refId: "8",
        itemType: "procedure" as const,
      },
      {
        key: "medicines" as const,
        refKey: "medicine_id" as const,
        refId: "42",
        itemType: "medicine" as const,
      },
    ];

    it.each(TYPES)(
      "PR700 inactive: %s cold GET (is_active:false) は一致扱いせず zero-write で拒否する",
      async ({ key, refId, itemType }) => {
        const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        const wrapper = ({ children }: { children: ReactNode }) =>
          createElement(QueryClientProvider, { client: queryClient }, children);
        const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
        // グローバルキャッシュに同名 active マスタを置いても scoped dest の inactive を拾い拒否する
        queryClient.setQueryData(queryKeys.masters.category(key), [
          { id: refId, name: "X", price: 3307, isActive: true },
        ]);
        const globalCacheBefore = structuredClone(
          queryClient.getQueryData(queryKeys.masters.category(key)),
        );

        const sourceRows = [
          sourceTreatment({
            item_type: itemType,
            consultation_id: undefined,
            procedure_id: undefined,
            medicine_id: undefined,
            ...(key === "consultations" ? { consultation_id: refId } : {}),
            ...(key === "procedures" ? { procedure_id: refId } : {}),
            ...(key === "medicines" ? { medicine_id: refId } : {}),
          }),
        ];
        const sourceSnapshot = structuredClone(sourceRows);
        const postBodies = useCopyHandlers({
          sourceRows,
          [key]: [
            {
              id: Number(refId),
              name: "X",
              price: 3307,
              is_active: false,
            },
          ],
        });

        const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), { wrapper });
        await act(async () => {
          await result.current.copyTreatmentsFromRecord(SOURCE_ID);
        });

        expect(postBodies).toHaveLength(0);
        expect(invalidateSpy).not.toHaveBeenCalled();
        expect(toast.error).toHaveBeenCalledWith(
          "複写先のマスタを確認し、治療明細を再選択してください",
        );
        expect(toast.error).toHaveBeenCalledTimes(1);
        expect(toast.success).not.toHaveBeenCalled();
        expect(sourceRows).toEqual(sourceSnapshot);
        expect(queryClient.getQueryData(queryKeys.masters.category(key))).toEqual(
          globalCacheBefore,
        );
      },
    );

    it.each(TYPES.flatMap((t) => [false, undefined].map((activity) => ({ ...t, activity }))))(
      "PR700 inactive: %s warm dest キャッシュ isActive=$activity は一致扱いせず拒否する",
      async ({ key, refId, itemType, activity }) => {
        const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        const wrapper = ({ children }: { children: ReactNode }) =>
          createElement(QueryClientProvider, { client: queryClient }, children);
        const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
        for (const cat of ["consultations", "procedures", "medicines"]) {
          queryClient.setQueryData(queryKeys.masters.categoryForClinic(cat, "3"), []);
        }
        queryClient.setQueryData(queryKeys.masters.categoryForClinic(key, "3"), [
          { id: refId, name: "X", price: 3307, isActive: activity },
        ]);
        const scopedCacheBefore = structuredClone(
          queryClient.getQueryData(queryKeys.masters.categoryForClinic(key, "3")),
        );

        const sourceRows = [
          sourceTreatment({
            item_type: itemType,
            consultation_id: undefined,
            procedure_id: undefined,
            medicine_id: undefined,
            ...(key === "consultations" ? { consultation_id: refId } : {}),
            ...(key === "procedures" ? { procedure_id: refId } : {}),
            ...(key === "medicines" ? { medicine_id: refId } : {}),
          }),
        ];
        const sourceSnapshot = structuredClone(sourceRows);
        const postBodies = useCopyHandlers({ sourceRows });

        const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), { wrapper });
        await act(async () => {
          await result.current.copyTreatmentsFromRecord(SOURCE_ID);
        });

        expect(postBodies).toHaveLength(0);
        expect(invalidateSpy).not.toHaveBeenCalled();
        expect(toast.error).toHaveBeenCalledWith(
          "複写先のマスタを確認し、治療明細を再選択してください",
        );
        expect(toast.error).toHaveBeenCalledTimes(1);
        expect(toast.success).not.toHaveBeenCalled();
        expect(sourceRows).toEqual(sourceSnapshot);
        // filter はキャッシュ配列を変更しない（コピー側のみ）
        expect(queryClient.getQueryData(queryKeys.masters.categoryForClinic(key, "3"))).toEqual(
          scopedCacheBefore,
        );
      },
    );
  });
});

it("stops copying unresolved insurance before any POST", async () => {
  vi.clearAllMocks();
  const postBodies = useCopyHandlers({ sourceRows: [sourceTreatment({ is_insurance: null })] });
  const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID), {
    wrapper: createTestWrapper(),
  });
  await act(async () => {
    await result.current.copyTreatmentsFromRecord(SOURCE_ID);
  });
  expect(postBodies).toHaveLength(0);
  expect(toast.error).toHaveBeenCalled();
});
