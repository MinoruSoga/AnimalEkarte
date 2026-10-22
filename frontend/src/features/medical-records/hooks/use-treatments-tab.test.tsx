import type { ReactNode } from "react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import { useTreatmentsTab } from "./use-treatments-tab";
import type { Treatment, UpdateTreatmentInput } from "../types";
import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import { queryKeys } from "@/lib/query-keys";
import {
  DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
  medicineDoseParamsQueryKey,
} from "../api/medicine-dose-lookup";

// PR679 F5 カバレッジ補完: TreatmentsTab 統合テストで covered な確定/死亡経路に加え、
// フックの最新値ゲート（権限喪失・キャプチャ済みコールバック・await 後再検査）と
// 集計/追加フォーム/並び替えの直接挙動を固定する。

// usePermission は useAuth.hasPermission を読む。権限をテスト中に喪失させるため
// hoisted な可変 map で制御する（TreatmentsTab.test.tsx の最小モックと同パターン）。
const permissions = vi.hoisted(() => ({ map: {} as Record<string, boolean> }));
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }] },
    currentClinicId: "1",
    hasPermission: (resource: string, action: string) =>
      permissions.map[`${resource}:${action}`] ?? true,
  }),
}));

vi.mock("@/hooks/use-clinic-tax-rates", () => ({
  useClinicTaxRates: () => ({ standardTaxRate: 0.1 }),
}));

const toastErrorSpy = vi.hoisted(() => vi.fn());
vi.mock("sonner", () => ({
  toast: { error: toastErrorSpy, success: vi.fn() },
}));

const MEDICAL_RECORD_ID = "10";
const RECORD_CLINIC_ID = "9";

function makeTreatment(overrides: Partial<Treatment> = {}): Treatment {
  return {
    id: "t-1",
    medical_record_id: MEDICAL_RECORD_ID,
    item_type: "consultation",
    unit_price: 1000,
    quantity: 1,
    is_selected: true,
    status: "active",
    content: "診察",
    memo: "",
    is_insurance: true,
    discount_rate: 0,
    discount_amount: 0,
    sort_order: 0,
    version: 3,
    created_at: "2026-10-01T00:00:00Z",
    updated_at: "2026-10-01T00:00:00Z",
    ...overrides,
  };
}

const ROW_A = makeTreatment({ id: "1", sort_order: 0, unit_price: 1000, is_selected: true });
const ROW_B = makeTreatment({
  id: "2",
  sort_order: 1,
  unit_price: 2000,
  quantity: 2,
  is_selected: true,
  content: "処置",
});
const ROW_C = makeTreatment({
  id: "3",
  sort_order: 2,
  unit_price: 500,
  is_selected: false,
  content: "未選択",
});

const SAFE_MASTER_ITEM: TreatmentMasterItem = {
  id: "cons-1",
  name: "一般診察",
  unitPrice: 500,
  category: "診察",
  insuranceContract: "master",
  isNonInsurance: true,
};

// dose lookup を await させるため medicineId 持ちの選択肢。dose-params を空配列で解決し
// doseCalcInput=null（missing）→ hard/deviation gate 不発の通常 create 経路にする。
const MED_MASTER_ITEM: TreatmentMasterItem = {
  id: "med-1",
  name: "薬A",
  unitPrice: 100,
  category: "薬剤",
  insuranceContract: "master",
  isNonInsurance: true,
  medicineId: "1",
};

type Counts = { post: number; patch: number; del: number; put: number };

function installHandlers(rows: Treatment[], counts: Counts) {
  let patchBody: Record<string, unknown> | null = null;
  let patchClinicId: string | null = null;
  let putBody: { treatments: Array<{ id: number; sort_order: number }> } | null = null;
  server.use(
    http.get("*/v1/medical-records/:id/treatments", () => HttpResponse.json(rows)),
    http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])),
    http.get("*/v1/masters/medicines", () => HttpResponse.json([])),
    http.post("*/v1/medical-records/:id/treatments", async ({ request }) => {
      counts.post += 1;
      const body = (await request.json()) as Record<string, unknown>;
      return HttpResponse.json(makeTreatment({ id: "new-1", ...(body as Partial<Treatment>) }));
    }),
    http.patch("*/v1/medical-records/:id/treatments/:tid", async ({ request }) => {
      counts.patch += 1;
      patchBody = (await request.json()) as Record<string, unknown>;
      patchClinicId = request.headers.get("X-Clinic-ID");
      return HttpResponse.json(makeTreatment());
    }),
    http.delete("*/v1/medical-records/:id/treatments/:tid", () => {
      counts.del += 1;
      return new HttpResponse(null, { status: 204 });
    }),
    http.put("*/v1/medical-records/:id/treatments", async ({ request }) => {
      counts.put += 1;
      putBody = (await request.json()) as { treatments: Array<{ id: number; sort_order: number }> };
      return new HttpResponse(null, { status: 200 });
    }),
  );
  return {
    getPatchBody: () => patchBody,
    getPatchClinicId: () => patchClinicId,
    getPutBody: () => putBody,
  };
}

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

function renderTreatmentsTab(props: { isWriteBlocked?: boolean } = {}) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const hook = renderHook(
    ({ isWriteBlocked = false }) =>
      useTreatmentsTab({
        medicalRecordId: MEDICAL_RECORD_ID,
        ownerDiscountRate: 10,
        petSpecies: "dog",
        recordClinicId: RECORD_CLINIC_ID,
        isWriteBlocked,
      }),
    { wrapper: createWrapper(queryClient), initialProps: props },
  );
  return { ...hook, queryClient };
}

async function waitForLoaded(result: { current: { isLoading: boolean } }) {
  await waitFor(() => expect(result.current.isLoading).toBe(false));
}

describe("useTreatmentsTab", () => {
  let counts: Counts;
  let probes: ReturnType<typeof installHandlers>;

  beforeEach(() => {
    permissions.map = {};
    counts = { post: 0, patch: 0, del: 0, put: 0 };
    probes = installHandlers([ROW_A, ROW_B, ROW_C], counts);
  });

  afterEach(() => {
    server.resetHandlers();
  });

  it("PR700 scoped doseContext and retry use recordClinicId, never the global dose cache", async () => {
    const headers: Array<string | null> = [];
    server.use(
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([
          {
            id: 1,
            name: "薬A",
            price: 100,
            medicine_unit: "per_tablet",
            is_active: true,
            tax_rate: "0.1",
            calculation_type: "per_weight",
          },
        ]),
      ),
      http.get("*/v1/medical-records/:id/vitals", () =>
        HttpResponse.json([
          { id: "1", recorded_at: "2026-10-18T00:00:00Z", weight: 4, weight_unit: "Kg" },
        ]),
      ),
      http.get("*/v1/masters/medicines/1/dose-params", ({ request }) => {
        headers.push(request.headers.get("X-Clinic-ID"));
        return HttpResponse.json({ private: "not shown" }, { status: 500 });
      }),
    );
    const { result, queryClient } = renderTreatmentsTab();
    queryClient.setQueryData(medicineDoseParamsQueryKey("1"), []);
    const resetSpy = vi.spyOn(queryClient, "resetQueries");
    await waitForLoaded(result);
    await act(async () => {
      await result.current.handleSelectFromMaster(MED_MASTER_ITEM);
    });
    expect(result.current.masterDoseBlockReason).toBe(DOSE_PARAMS_LOOKUP_FAILED_MESSAGE);
    expect(result.current.doseContext.clinicId).toBe(RECORD_CLINIC_ID);
    expect(counts.post).toBe(0);
    server.use(
      http.get("*/v1/masters/medicines/1/dose-params", ({ request }) => {
        headers.push(request.headers.get("X-Clinic-ID"));
        return HttpResponse.json([]);
      }),
    );
    act(() => result.current.handleRetryMasterDoseLookup());
    await waitFor(() => expect(counts.post).toBe(1));
    expect(resetSpy).toHaveBeenCalledWith({
      queryKey: medicineDoseParamsQueryKey("1", RECORD_CLINIC_ID),
    });
    expect(headers).toEqual([RECORD_CLINIC_ID, RECORD_CLINIC_ID]);
    expect(queryClient.getQueryData(medicineDoseParamsQueryKey("1"))).toEqual([]);
    expect(
      queryClient.getQueryData(queryKeys.masters.categoryForClinic("medicines", RECORD_CLINIC_ID)),
    ).toBeDefined();
  });

  it("sort_order 昇順で並べ、選択行だけの小計/件数と飼主割引+税込み合計を返す", async () => {
    const { result } = renderTreatmentsTab();
    await waitForLoaded(result);

    expect(result.current.sortedTreatments.map((t) => t.id)).toEqual(["1", "2", "3"]);
    // selected: 1000*1 + 2000*2 = 5000 / 全行小計: 5500
    expect(result.current.selectedSubtotal).toBe(5000);
    expect(result.current.selectedCount).toBe(2);
    expect(result.current.totalSubtotal).toBe(5500);
    // 飼主割引10% → 4500、税10% → 4950
    expect(result.current.finalTotal).toBe(4950);
  });

  it("handleUpdate は読取済み version を CAS として同送し、recordClinicId をヘッダに付ける", async () => {
    const { result } = renderTreatmentsTab();
    await waitForLoaded(result);

    const input: UpdateTreatmentInput = { content: "更新", quantity: 2 };
    act(() => result.current.handleUpdate("2", input));

    await waitFor(() => expect(counts.patch).toBe(1));
    expect(probes.getPatchBody()).toMatchObject({ content: "更新", quantity: 2, version: 3 });
    expect(probes.getPatchClinicId()).toBe(RECORD_CLINIC_ID);
  });

  it("対象行の version が number でない場合は fail-closed で拒否し PATCH を送らない", async () => {
    server.use(
      http.get("*/v1/medical-records/:id/treatments", () =>
        HttpResponse.json([{ ...makeTreatment({ id: "9" }), version: undefined }]),
      ),
    );
    const { result } = renderTreatmentsTab();
    await waitForLoaded(result);

    act(() => result.current.handleUpdate("9", { content: "更新" }));

    expect(counts.patch).toBe(0);
    expect(toastErrorSpy).toHaveBeenCalledWith("治療明細の読み込みが完了してから保存してください");
  });

  it("handleDelete/handleMoveUp/handleMoveDown は権限ありで API を呼ぶ", async () => {
    const { result } = renderTreatmentsTab();
    await waitForLoaded(result);

    act(() => result.current.handleDelete("1"));
    await waitFor(() => expect(counts.del).toBe(1));

    act(() => result.current.handleMoveUp("2"));
    await waitFor(() => expect(counts.put).toBe(1));
    // 2(sort1) が上へ → 行順は維持したまま sort_order を入替（1:1, 2:0, 3:2）。
    // BE 契約どおり id は number 変換される
    expect(probes.getPutBody()?.treatments).toEqual([
      { id: 1, sort_order: 1 },
      { id: 2, sort_order: 0 },
      { id: 3, sort_order: 2 },
    ]);

    act(() => result.current.handleMoveDown("2"));
    await waitFor(() => expect(counts.put).toBe(2));
    expect(probes.getPutBody()?.treatments).toEqual([
      { id: 1, sort_order: 0 },
      { id: 2, sort_order: 2 },
      { id: 3, sort_order: 1 },
    ]);
  });

  it("先頭行の moveUp と最終行の moveDown は境界で no-op（API 非送信）", async () => {
    const { result } = renderTreatmentsTab();
    await waitForLoaded(result);

    act(() => result.current.handleMoveUp("1"));
    act(() => result.current.handleMoveDown("3"));

    expect(counts.put).toBe(0);
  });

  it("手入力追加は内容があれば create し、成功時にフォームをリセットして最終行フォーカスする", async () => {
    const { result } = renderTreatmentsTab();
    await waitForLoaded(result);

    act(() => result.current.handleStartAdding());
    expect(result.current.isAdding).toBe(true);
    act(() => result.current.setAddContent("  点滴  "));
    act(() => result.current.handleAddSubmit());

    await waitFor(() => expect(counts.post).toBe(1));
    await waitFor(() => expect(result.current.isAdding).toBe(false));
    expect(result.current.addContent).toBe("");
    expect(result.current.focusLastRow).toBe(true);

    act(() => result.current.handleAutoFocusDone());
    expect(result.current.focusLastRow).toBe(false);
  });

  it("手入力追加は空内容なら create せず、cancel はフォームを閉じて内容を消す", async () => {
    const { result } = renderTreatmentsTab();
    await waitForLoaded(result);

    act(() => result.current.handleStartAdding());
    act(() => result.current.setAddContent("   "));
    act(() => result.current.handleAddSubmit());
    expect(counts.post).toBe(0);
    expect(result.current.isAdding).toBe(true);

    act(() => result.current.setAddContent("メモ"));
    act(() => result.current.handleAddCancel());
    expect(result.current.isAdding).toBe(false);
    expect(result.current.addContent).toBe("");
  });

  it("isWriteBlocked 中は検索・手入力を開かず、全 mutation ハンドラが API を送らない", async () => {
    const { result, queryClient } = renderTreatmentsTab({ isWriteBlocked: true });
    await waitForLoaded(result);

    act(() => result.current.handleOpenSearch());
    act(() => result.current.handleStartAdding());
    act(() => result.current.handleUpdate("1", { content: "x" }));
    act(() => result.current.handleDelete("1"));
    act(() => result.current.handleMoveUp("2"));
    act(() => result.current.handleMoveDown("1"));
    act(() => {
      result.current.setAddContent("点滴");
    });
    act(() => result.current.handleAddSubmit());
    await act(async () => {
      await result.current.handleSelectFromMaster(SAFE_MASTER_ITEM);
    });

    expect(result.current.isSearchOpen).toBe(false);
    expect(result.current.isAdding).toBe(false);
    expect(queryClient.getMutationCache().getAll()).toHaveLength(0);
    expect(counts).toEqual({ post: 0, patch: 0, del: 0, put: 0 });
  });

  it("isWriteBlocked への途中遷移後も、キャプチャ済みコールバックは最新値で弾かれる", async () => {
    const { result, rerender, queryClient } = renderTreatmentsTab({ isWriteBlocked: false });
    await waitForLoaded(result);

    act(() => result.current.setAddContent("点滴"));
    // draft 時点のハンドラ参照を捕捉 → 確定/ロックへ遷移 → 旧参照から呼んでも止まる
    const captured = {
      update: result.current.handleUpdate,
      del: result.current.handleDelete,
      moveUp: result.current.handleMoveUp,
      addSubmit: result.current.handleAddSubmit,
      select: result.current.handleSelectFromMaster,
    };
    rerender({ isWriteBlocked: true });
    act(() => {
      result.current.setAddContent("点滴");
    });

    act(() => captured.update("1", { content: "x" }));
    act(() => captured.del("1"));
    act(() => captured.moveUp("2"));
    act(() => captured.addSubmit());
    await act(async () => {
      await captured.select(SAFE_MASTER_ITEM);
    });

    expect(queryClient.getMutationCache().getAll()).toHaveLength(0);
    expect(counts).toEqual({ post: 0, patch: 0, del: 0, put: 0 });
  });

  it("isWriteBlocked で canEditDiscount は fail-closed に折り畳まれる", async () => {
    const { result, rerender } = renderTreatmentsTab({ isWriteBlocked: false });
    await waitForLoaded(result);
    expect(result.current.canEditDiscount).toBe(true);

    rerender({ isWriteBlocked: true });
    expect(result.current.canEditDiscount).toBe(false);
  });

  it("権限喪失後に rerender するとキャプチャ済みハンドラも最新権限で弾かれる", async () => {
    const { result, rerender, queryClient } = renderTreatmentsTab({ isWriteBlocked: false });
    await waitForLoaded(result);

    const captured = { del: result.current.handleDelete, update: result.current.handleUpdate };
    permissions.map["medical-records:delete"] = false;
    permissions.map["medical-records:edit"] = false;
    rerender({ isWriteBlocked: false });

    act(() => captured.del("1"));
    act(() => captured.update("1", { content: "x" }));

    expect(queryClient.getMutationCache().getAll()).toHaveLength(0);
    expect(counts.del).toBe(0);
    expect(counts.patch).toBe(0);
  });

  it("dose-params 応答を待つ間に create 権限を喪失した場合、await 後の create は送らない", async () => {
    let resolveDoseParams: ((response: Response) => void) | null = null;
    server.use(
      http.get("*/v1/medical-records/:id/vitals", () =>
        HttpResponse.json([
          {
            id: "1",
            medical_record_id: MEDICAL_RECORD_ID,
            recorded_at: "2026-07-15T00:00:00Z",
            weight: 4,
            weight_unit: "Kg",
            created_at: "2026-07-15T00:00:00Z",
            updated_at: "2026-07-15T00:00:00Z",
          },
        ]),
      ),
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([
          {
            id: 1,
            name: "薬A",
            medicine_unit: "per_tablet",
            price: 100,
            is_active: true,
            tax_rate: "0.1",
            sort_order: 0,
            calculation_type: "per_weight",
          },
        ]),
      ),
      http.get(
        "*/v1/masters/medicines/:id/dose-params",
        () =>
          new Promise<Response>((resolve) => {
            resolveDoseParams = resolve;
          }),
      ),
    );

    const { result, rerender } = renderTreatmentsTab({ isWriteBlocked: false });
    await waitForLoaded(result);

    await waitFor(() => expect(result.current.doseContext.weightKg).toBe(4));
    await waitFor(() => expect(result.current.doseContext.medicines).toHaveLength(1));

    let selectPromise: Promise<void> | undefined;
    act(() => {
      selectPromise = result.current.handleSelectFromMaster(MED_MASTER_ITEM);
    });
    // fetchQuery 中に権限喪失 → コミットで最新値が ref に同期される
    permissions.map["medical-records:create"] = false;
    rerender({ isWriteBlocked: false });

    // fetch が MSW ハンドラへ届くのを待ってから応答を解決する（未着なら解決漏れでハングする）
    await waitFor(() => expect(resolveDoseParams).not.toBeNull());
    await act(async () => {
      resolveDoseParams?.(HttpResponse.json([]));
      await selectPromise;
    });

    expect(counts.post).toBe(0);
    expect(result.current.focusLastRow).toBe(false);
  });

  it("dose-params 応答後も権限が残っていれば master 選択は create して最終行フォーカスする", async () => {
    let resolveDoseParams: ((response: Response) => void) | null = null;
    server.use(
      http.get("*/v1/medical-records/:id/vitals", () =>
        HttpResponse.json([
          {
            id: "1",
            medical_record_id: MEDICAL_RECORD_ID,
            recorded_at: "2026-07-15T00:00:00Z",
            weight: 4,
            weight_unit: "Kg",
            created_at: "2026-07-15T00:00:00Z",
            updated_at: "2026-07-15T00:00:00Z",
          },
        ]),
      ),
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([
          {
            id: 1,
            name: "薬A",
            medicine_unit: "per_tablet",
            price: 100,
            is_active: true,
            tax_rate: "0.1",
            sort_order: 0,
            calculation_type: "per_weight",
          },
        ]),
      ),
      http.get(
        "*/v1/masters/medicines/:id/dose-params",
        () =>
          new Promise<Response>((resolve) => {
            resolveDoseParams = resolve;
          }),
      ),
    );

    const { result } = renderTreatmentsTab({ isWriteBlocked: false });
    await waitForLoaded(result);

    await waitFor(() => expect(result.current.doseContext.weightKg).toBe(4));
    await waitFor(() => expect(result.current.doseContext.medicines).toHaveLength(1));

    let selectPromise: Promise<void> | undefined;
    act(() => {
      selectPromise = result.current.handleSelectFromMaster(MED_MASTER_ITEM);
    });
    await waitFor(() => expect(resolveDoseParams).not.toBeNull());
    await act(async () => {
      resolveDoseParams?.(HttpResponse.json([]));
      await selectPromise;
    });

    await waitFor(() => expect(counts.post).toBe(1));
    expect(result.current.focusLastRow).toBe(true);
  });

  it("medicineId 無しの master 選択は dose 待ち無しで即時 create する", async () => {
    const { result } = renderTreatmentsTab();
    await waitForLoaded(result);

    await act(async () => {
      await result.current.handleSelectFromMaster(SAFE_MASTER_ITEM);
    });

    await waitFor(() => expect(counts.post).toBe(1));
    // 既存3行（最大 sort_order=2）の次は 3
    expect(result.current.focusLastRow).toBe(true);
  });

  it.each([null, undefined])(
    "master insurance %s blocks before create with an actionable message",
    async (isNonInsurance) => {
      const { result } = renderTreatmentsTab();
      await waitForLoaded(result);
      await act(async () => {
        await result.current.handleSelectFromMaster({ ...SAFE_MASTER_ITEM, isNonInsurance });
      });
      expect(toastErrorSpy).toHaveBeenCalledWith(
        "保険設定が未設定または未取得です。マスタの設定を確認してください",
      );
      expect(counts.post).toBe(0);
    },
  );
});
