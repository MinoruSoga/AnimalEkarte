import { act, renderHook, waitFor } from "@testing-library/react";
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
      consultations: [{ id: 7, name: "初診料", price: 2500 }],
      medicines: [{ id: 42, name: "アモキシシリン", price: 120 }],
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
      sourceRows: [sourceTreatment({ consultation_id: "999", sort_order: 2 })],
    });

    const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID), {
      wrapper: createTestWrapper(),
    });

    await act(async () => {
      await result.current.copyTreatmentsFromRecord(SOURCE_ID);
    });

    // マスタ不一致 (999) → 当時価格は参照維持だが「マスタ不一致行のみ」。
    // sort_order は offset=0 でソース値 2 のまま。
    expect(postBodies).toHaveLength(1);
    const body = postBodies[0] as Record<string, unknown>;
    expect(body.sort_order).toBe(2);
    expect(body.unit_price).toBe(1500);
    expect(body.consultation_id).toBe(999);
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
        sourceTreatment({ id: "501", sort_order: 0 }),
        sourceTreatment({ id: "502", sort_order: 1 }),
        sourceTreatment({ id: "503", sort_order: 2 }),
      ],
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
      { id: "7", name: "初診", unitPrice: 3307, category: "診察" },
      { id: "8", name: "創傷処置", unitPrice: 3306, category: "処置" },
      { id: "42", name: "アモキシシリン", unitPrice: 3303, category: "薬剤", medicineId: "42" },
      { id: "7", name: "検査キット", unitPrice: 2000, category: "検査" },
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
    sourceStore = [
      ...sourceStore,
      {
        ...toWireRow(SOURCE_ID, {
          item_type: "consultation",
          consultation_id: 999,
          content: "初診",
          unit_price: 1500,
          sort_order: 4,
        }),
        id: `${SOURCE_ID}-row-legacy`,
      },
    ];

    const { result } = renderHook(() => useCopyTreatmentDetails(DEST_ID, "3"), {
      wrapper: createTestWrapper(),
    });

    await act(async () => {
      await Promise.all([
        result.current.copyTreatmentsFromRecord(SOURCE_ID),
        result.current.copyTreatmentsFromRecord(SOURCE_ID),
      ]);
    });

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
      {
        item_type: "consultation",
        consultation_id: "999",
        content: "初診",
        unit_price: 1500,
        sort_order: 4,
      },
    ]);
    expect(savedDestRows).toHaveLength(5);

    expect(sourcePostHeaders).toHaveLength(4);
    expect(sourcePostHeaders.every((h) => h !== "3")).toBe(true);
    expect(sourceGetHeaders.length).toBeGreaterThan(0);
    expect(sourceGetHeaders.every((h) => h !== "3")).toBe(true);
    expect(destPostHeaders).toHaveLength(5);
    expect(destPostHeaders.every((h) => h === "3")).toBe(true);
    expect(destGetHeaders.length).toBeGreaterThan(0);
    expect(destGetHeaders.every((h) => h === "3")).toBe(true);

    expect(sourceStore[0].unit_price).toBe(710);
    expect(sourceStore[1].unit_price).toBe(709);
    expect(sourceStore[2].unit_price).toBe(703);

    expect(toast.success).toHaveBeenCalledTimes(1);
    expect(toast.error).not.toHaveBeenCalled();
  });
});
