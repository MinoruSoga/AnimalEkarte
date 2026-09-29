import { act, renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";

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
    expect(first.consultation_id).toBe("7");
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
    expect(body.consultation_id).toBe("999");
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
});
