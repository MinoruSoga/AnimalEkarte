import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, render, screen, waitFor, within } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";

import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { queryKeys } from "@/lib/query-keys";
import {
  DOSE_PARAMS_LOOKUP_FAILED_MESSAGE,
  medicineDoseParamsQueryKey,
} from "../api/medicine-dose-lookup";
import { MEDICAL_RECORD_FIXED_ACTION_LAYOUT } from "@/constants/medical-record-fixed-action-layout";
import { MedicalRecordBillCheck } from "./MedicalRecordBillCheck";

const { mockPermission } = vi.hoisted(() => ({
  mockPermission: {
    current: { canView: true, canCreate: true, canEdit: true, canDelete: true },
  },
}));
vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mockPermission.current,
}));

// usePermission をモックしても useClinicTaxRates は独立に useAuth() を呼ぶため最小モックが必要
// （TreatmentsTab.test.tsx と同パターン）。
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: { clinic: {} } }),
}));

const { mockCreateTreatmentMutate } = vi.hoisted(() => ({
  mockCreateTreatmentMutate: vi.fn(),
}));

const MOCK_TREATMENT = {
  id: "1",
  medical_record_id: "10",
  item_type: "medicine",
  unit_price: 100,
  quantity: 1,
  is_selected: true,
  status: "active",
  content: "テスト処置",
  memo: "",
  is_insurance: true,
  discount_rate: 0,
  discount_amount: 0,
  sort_order: 0,
  created_at: "2026-01-01T00:00:00+09:00",
  updated_at: "2026-01-01T00:00:00+09:00",
};

vi.mock("../api/treatments", () => ({
  // ガードで早期 return させないため明細1件ありの状態にする
  useGetTreatments: () => ({ data: [MOCK_TREATMENT] }),
  useCreateTreatment: () => ({ mutate: mockCreateTreatmentMutate, isPending: false }),
  useUpdateTreatment: () => ({ mutate: vi.fn() }),
  useDeleteTreatment: () => ({ mutate: vi.fn() }),
}));

vi.mock("../api/get-record-examinations", () => ({
  useGetRecordExaminations: () => ({ data: undefined }),
}));

vi.mock("../api/get-pet-vaccinations", () => ({
  useGetPetVaccinations: () => ({ data: [] }),
}));

const { mockDialogProps } = vi.hoisted(() => ({
  mockDialogProps: { current: {} as Record<string, unknown> },
}));

vi.mock("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog", () => ({
  TreatmentSearchDialog: (props: {
    clinicId?: string;
    onSelect: (item: TreatmentMasterItem) => void;
  }) => {
    mockDialogProps.current = props as unknown as Record<string, unknown>;
    return (
      <>
        <button
          type="button"
          onClick={() =>
            props.onSelect({
              id: "7",
              name: "初診",
              unitPrice: 3307,
              category: "診察",
              insuranceContract: "master",
              isNonInsurance: false,
            })
          }
        >
          診察Aを選択
        </button>
        <button
          type="button"
          onClick={() =>
            props.onSelect({
              id: "8",
              name: "処置X",
              unitPrice: 3306,
              category: "処置",
              insuranceContract: "master",
              isNonInsurance: false,
            })
          }
        >
          処置Xを選択
        </button>
        <button
          type="button"
          onClick={() =>
            props.onSelect({
              id: "9",
              name: "自費診察",
              unitPrice: 5000,
              category: "診察",
              insuranceContract: "master",
              isNonInsurance: true,
            })
          }
        >
          保険対象外診察を選択
        </button>
        <button
          type="button"
          onClick={() =>
            props.onSelect({
              id: "42",
              name: "抗生剤A",
              unitPrice: 3303,
              category: "薬剤",
              insuranceContract: "master",
              isNonInsurance: false,
              medicineId: "42",
            })
          }
        >
          薬剤Aを選択
        </button>
      </>
    );
  },
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

const MEDICAL_RECORD_ID = "10";

beforeEach(() => {
  vi.mocked(toast.success).mockClear();
  vi.mocked(toast.error).mockClear();
  mockCreateTreatmentMutate.mockClear();
  mockDialogProps.current = {};
  mockPermission.current = { canView: true, canCreate: true, canEdit: true, canDelete: true };
});

afterEach(() => {
  server.resetHandlers();
});

// FE-RC-027: マスタ検索から選択した明細の item_type 判定は "薬品"（誤字。マスタの実カテゴリは
// "薬剤"）と比較していたため常に一致せず "other" にフォールバックしていた。
// TreatmentsTab と同じ共有 resolveItemTypeFromCategory に統一し、"薬剤" → "medicine" を保証する。
describe("MedicalRecordBillCheck PR670 マスタ選択", () => {
  it.each(["予防", "検査"])("未採用 %s は caller の保険対象=true を維持する", async (category) => {
    server.use(
      http.get("*/v1/medical-records/:id/clinical-plan", () => HttpResponse.json({})),
      http.get("*/v1/medical-records/:id/billing-confirmation", () =>
        HttpResponse.json({ id: "1", medical_record_id: "10", status: "pending" }),
      ),
      http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])),
    );
    const user = userEvent.setup();
    render(
      <MedicalRecordBillCheck medicalRecordId={MEDICAL_RECORD_ID} petId="1" recordClinicId="2" />,
      { wrapper: createTestWrapper() },
    );
    await user.click(screen.getByRole("button", { name: "行を追加（検索）" }));
    const select = mockDialogProps.current.onSelect as (item: TreatmentMasterItem) => void;
    await act(async () => {
      select({ id: "71", name: "legacy master", unitPrice: 1000, category });
    });
    await waitFor(() =>
      expect(mockCreateTreatmentMutate).toHaveBeenCalledWith(
        expect.objectContaining({ item_type: "other", memo: category, is_insurance: true }),
      ),
    );
  });

  it("EMR-454: master が明示的に保険対象外の品目は is_insurance=false で作成する", async () => {
    server.use(
      http.get("*/v1/medical-records/:id/billing-confirmation", () =>
        HttpResponse.json({ id: "1", medical_record_id: MEDICAL_RECORD_ID, status: "pending" }),
      ),
      http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])),
    );
    const user = userEvent.setup();
    render(
      <MedicalRecordBillCheck medicalRecordId={MEDICAL_RECORD_ID} petId="1" recordClinicId="2" />,
      { wrapper: createTestWrapper() },
    );

    await user.click(screen.getByRole("button", { name: "行を追加（検索）" }));
    await user.click(screen.getByRole("button", { name: "保険対象外診察を選択" }));

    await waitFor(() =>
      expect(mockCreateTreatmentMutate).toHaveBeenCalledWith(
        expect.objectContaining({ consultation_id: 9, is_insurance: false }),
      ),
    );
  });

  it.each([
    { label: "診察Aを選択", refKey: "consultation_id", refId: 7, memo: "診察", price: 3307 },
    { label: "処置Xを選択", refKey: "procedure_id", refId: 8, memo: "処置", price: 3306 },
    { label: "薬剤Aを選択", refKey: "medicine_id", refId: 42, memo: "薬剤", price: 3303 },
  ])(
    "$label → 実ビルダーで数値参照・memo=カテゴリ・is_insurance=true・nextOrder で作成する",
    async ({ label, refKey, refId, memo, price }) => {
      server.use(
        http.get("*/v1/medical-records/:id/billing-confirmation", () =>
          HttpResponse.json({ id: "1", medical_record_id: MEDICAL_RECORD_ID, status: "pending" }),
        ),
      );

      const user = userEvent.setup();
      render(
        <MedicalRecordBillCheck medicalRecordId={MEDICAL_RECORD_ID} petId="1" recordClinicId="2" />,
        { wrapper: createTestWrapper() },
      );

      server.use(
        http.get("*/v1/masters/medicines", () =>
          HttpResponse.json([
            { id: 42, name: "抗生剤A", price: 3303, is_active: true, tax_rate: "0.1" },
          ]),
        ),
        http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])),
      );

      await user.click(screen.getByRole("button", { name: "行を追加（検索）" }));
      await user.click(screen.getByRole("button", { name: label }));

      await waitFor(() =>
        expect(mockCreateTreatmentMutate).toHaveBeenCalledWith(
          expect.objectContaining({
            [refKey]: refId,
            content: expect.any(String),
            memo,
            is_insurance: true,
            quantity: 1,
            unit_price: price,
          }),
        ),
      );
      const call = mockCreateTreatmentMutate.mock.calls[0][0] as Record<string, unknown>;
      expect(typeof call[refKey]).toBe("number");
      expect(call.sort_order).toBe(1);
      expect(mockDialogProps.current.clinicId).toBe("2");
    },
  );

  it.each(["lock", "permission", "deceased", "record", "clinic", "species"] as const)(
    "PR700 deferred: %s during lookup prevents create",
    async (change) => {
      let releaseGet: (() => void) | undefined;
      let getSeen = false;
      server.use(
        http.get("*/v1/medical-records/10/clinical-plan", () => HttpResponse.json({})),
        http.get("*/v1/medical-records/:id/billing-confirmation", () =>
          HttpResponse.json({ id: "1", medical_record_id: "10", status: "pending" }),
        ),
        http.get("*/v1/masters/medicines", () => {
          getSeen = true;
          const medicine = { id: 42, name: "薬A", price: 3303, is_active: true, tax_rate: "0.1" };
          return new Promise<ReturnType<typeof HttpResponse.json>>((resolve) => {
            releaseGet = () => resolve(HttpResponse.json([medicine]));
          });
        }),
        http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])),
      );
      const { rerender } = render(
        <MedicalRecordBillCheck medicalRecordId={MEDICAL_RECORD_ID} petId="1" recordClinicId="2" />,
        { wrapper: createTestWrapper() },
      );
      let select: ((item: unknown) => void) | undefined;
      await waitFor(() => {
        select = mockDialogProps.current.onSelect as ((item: unknown) => void) | undefined;
        expect(select).toBeDefined();
      });
      let resolveDone = false;
      const promise = Promise.resolve(
        select!({
          id: "42",
          name: "薬A",
          unitPrice: 3303,
          category: "薬剤",
          insuranceContract: "master",
          isNonInsurance: false,
          medicineId: "42",
        }),
      ).finally(() => {
        resolveDone = true;
      });
      await waitFor(() => expect(getSeen).toBe(true), { timeout: 3000 });
      if (change === "permission") mockPermission.current.canEdit = false;
      rerender(
        <MedicalRecordBillCheck
          medicalRecordId={change === "record" ? "11" : MEDICAL_RECORD_ID}
          petId="1"
          recordClinicId={change === "clinic" ? "3" : "2"}
          petSpecies={change === "species" ? "猫" : undefined}
          isPetDeceased={change === "deceased"}
          isWriteBlocked={change === "lock"}
        />,
      );
      await act(async () => {
        releaseGet!();
        await promise;
      });
      expect(resolveDone).toBe(true);
      expect(mockCreateTreatmentMutate).not.toHaveBeenCalled();
    },
  );

  it.each(["deceased", "permission"] as const)(
    "ガード (%s) はキャプチャした選択コールバックでも create を拒否する",
    async (guard) => {
      server.use(
        http.get("*/v1/medical-records/:id/billing-confirmation", () =>
          HttpResponse.json({ id: "1", medical_record_id: MEDICAL_RECORD_ID, status: "pending" }),
        ),
      );
      if (guard === "permission") mockPermission.current.canEdit = false;
      const view = render(
        <MedicalRecordBillCheck
          medicalRecordId={MEDICAL_RECORD_ID}
          petId="1"
          recordClinicId="2"
          isPetDeceased={guard === "deceased"}
        />,
        { wrapper: createTestWrapper() },
      );
      await waitFor(() => expect(mockDialogProps.current.onSelect).toBeDefined());
      const select = mockDialogProps.current.onSelect as ((item: unknown) => void) | undefined;
      expect(select).toBeDefined();
      select?.({
        id: "7",
        name: "初診",
        unitPrice: 3307,
        category: "診察",
        insuranceContract: "master",
        isNonInsurance: false,
      });
      expect(mockCreateTreatmentMutate).not.toHaveBeenCalled();
      if (guard === "deceased") expect(toast.error).toHaveBeenCalled();
      view.unmount();
    },
  );
});

describe("PR700 BillCheck real dose selection", () => {
  it.each(["recommended", "cap", "reason", "http", "missing-weight", "missing-params"] as const)(
    "%s uses scoped authority and the actual dose gate",
    async (mode) => {
      const headers: string[] = [];
      const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
      const globalMasters = [{ id: "42", isActive: false, price: 1 }];
      queryClient.setQueryData(queryKeys.masters.category("medicines"), globalMasters);
      server.use(
        http.get("*/v1/medical-records/:id/billing-confirmation", () =>
          HttpResponse.json({ id: "1", medical_record_id: MEDICAL_RECORD_ID, status: "pending" }),
        ),
        http.get("*/v1/masters/medicines", ({ request }) => {
          headers.push(`medicine:${request.headers.get("X-Clinic-ID")}`);
          return HttpResponse.json([
            {
              id: 42,
              name: "薬A",
              price: 3303,
              is_active: true,
              tax_rate: "0.1",
              calculation_type: "per_weight",
              medicine_unit: "per_tablet",
              strength: "100",
            },
          ]);
        }),
        http.get("*/v1/medical-records/10/vitals", ({ request }) => {
          headers.push(`vitals:${request.headers.get("X-Clinic-ID")}`);
          return HttpResponse.json(
            mode === "missing-weight"
              ? []
              : [
                  {
                    id: "v1",
                    medical_record_id: "10",
                    recorded_at: "2026-10-18T12:00:00Z",
                    weight: 10,
                    weight_unit: "Kg",
                  },
                ],
          );
        }),
        http.get("*/v1/masters/medicines/42/dose-params", ({ request }) => {
          headers.push(`dose:${request.headers.get("X-Clinic-ID")}`);
          if (mode === "http") return HttpResponse.json({ error: "private" }, { status: 500 });
          return HttpResponse.json(
            mode === "missing-params"
              ? []
              : [
                  {
                    species: "dog",
                    dose_basis: "per_dose",
                    dose_per_kg: 20,
                    min_mg_per_kg: mode === "reason" ? 20 : 1,
                    max_mg_per_kg: 100,
                    absolute_max_dose: mode === "cap" ? 50 : mode === "reason" ? 150 : 1000,
                    rounding_step: 1,
                    rounding_mode: "up",
                  },
                ],
          );
        }),
      );
      render(
        <QueryClientProvider client={queryClient}>
          <MedicalRecordBillCheck
            medicalRecordId={MEDICAL_RECORD_ID}
            petId="1"
            recordClinicId="2"
            petSpecies="犬"
          />
        </QueryClientProvider>,
      );
      const user = userEvent.setup();
      await user.click(screen.getByRole("button", { name: "行を追加（検索）" }));
      await user.click(await screen.findByRole("button", { name: "薬剤Aを選択" }));
      if (mode === "cap" || mode === "reason" || mode === "http") {
        await waitFor(() => expect(toast.error).toHaveBeenCalled());
        expect(mockCreateTreatmentMutate).not.toHaveBeenCalled();
        expect(toast.error).toHaveBeenCalledWith(
          mode === "http"
            ? DOSE_PARAMS_LOOKUP_FAILED_MESSAGE
            : expect.stringContaining(mode === "cap" ? "絶対上限" : "推奨値付近で追加"),
        );
      } else {
        await waitFor(() =>
          expect(mockCreateTreatmentMutate).toHaveBeenCalledWith(
            expect.objectContaining({
              medicine_id: 42,
              quantity: mode === "recommended" ? 2 : 1,
              memo: "薬剤",
              is_insurance: true,
              sort_order: 1,
            }),
          ),
        );
      }
      expect(headers).toEqual(expect.arrayContaining(["medicine:2", "vitals:2"]));
      if (mode !== "missing-weight") expect(headers).toContain("dose:2");
      expect(queryClient.getQueryData(queryKeys.masters.category("medicines"))).toEqual(
        globalMasters,
      );
      expect(
        queryClient.getQueryData(queryKeys.masters.categoryForClinic("medicines", "2")),
      ).toBeDefined();
      if (mode === "recommended")
        expect(queryClient.getQueryData(medicineDoseParamsQueryKey("42", "2"))).toBeDefined();
    },
  );
});

describe("MedicalRecordBillCheck FE-RC-027 item_type 判定", () => {
  it("マスタ検索で category=薬剤 を選択すると item_type=medicine で作成される", async () => {
    server.use(
      http.get("*/v1/medical-records/:id/billing-confirmation", () =>
        HttpResponse.json({ id: "1", medical_record_id: MEDICAL_RECORD_ID, status: "pending" }),
      ),
      http.get("*/v1/masters/medicines", () =>
        HttpResponse.json([
          { id: 42, name: "抗生剤A", price: 3303, is_active: true, tax_rate: "0.1" },
        ]),
      ),
      http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])),
    );

    const user = userEvent.setup();
    render(
      <MedicalRecordBillCheck medicalRecordId={MEDICAL_RECORD_ID} petId="1" recordClinicId="2" />,
      { wrapper: createTestWrapper() },
    );

    await user.click(screen.getByRole("button", { name: "行を追加（検索）" }));
    await user.click(screen.getByRole("button", { name: "薬剤Aを選択" }));

    await waitFor(() =>
      expect(mockCreateTreatmentMutate).toHaveBeenCalledWith(
        expect.objectContaining({ item_type: "medicine", medicine_id: 42 }),
      ),
    );
    expect(mockDialogProps.current.clinicId).toBe("2");
  });
});

// EMR-432: 会計（医師確認）タブも同じ右下固定ボタン群を描くため、スクロール末尾の明細表が
// ボタン群に隠れない下余白（共有レイアウト定数）をスクロール領域へ付ける。
describe("MedicalRecordBillCheck EMR-432 右下固定ボタンの下余白", () => {
  it("スクロール領域と固定ボタン群が共有レイアウト定数を使う", () => {
    server.use(
      http.get("*/v1/medical-records/:id/billing-confirmation", () =>
        HttpResponse.json({ id: "1", medical_record_id: MEDICAL_RECORD_ID, status: "pending" }),
      ),
    );

    const { container } = render(
      <MedicalRecordBillCheck medicalRecordId={MEDICAL_RECORD_ID} petId="1" />,
      { wrapper: createTestWrapper() },
    );

    const scrollers = Array.from(container.querySelectorAll<HTMLElement>(".overflow-y-auto"));
    expect(scrollers.length).toBeGreaterThan(0);
    expect(
      scrollers.some((el) =>
        el.className.includes(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.scrollBottomClearance),
      ),
    ).toBe(true);

    const group = screen.getByRole("button", { name: "チェック完了" }).parentElement as HTMLElement;
    expect(group).toHaveClass(...MEDICAL_RECORD_FIXED_ACTION_LAYOUT.position.split(" "));
  });
});

// FE-RC-005: useCreateBillingConfirmation / useCreateBillingReturn は hook 側で
// onError → handleApiError（toast.error）を持つ。呼び出し元がさらに handleApiError を
// 呼ぶと、react-query が hook 側 + 呼び出し側の両方のコールバックを実行するため
// 失敗時に toast.error が二重発火する（二重トースト回帰）。
describe("MedicalRecordBillCheck FE-RC-005 billing-confirmation 二重トースト回帰", () => {
  it("チェック完了の失敗時、エラートーストは1回だけ表示する", async () => {
    server.use(
      http.get("*/v1/medical-records/:id/billing-confirmation", () =>
        HttpResponse.json({ id: "1", medical_record_id: MEDICAL_RECORD_ID, status: "pending" }),
      ),
      http.post("*/v1/medical-records/:id/billing-confirmation/confirm", () =>
        HttpResponse.json({ error: "internal error" }, { status: 500 }),
      ),
    );

    const user = userEvent.setup();
    render(<MedicalRecordBillCheck medicalRecordId={MEDICAL_RECORD_ID} petId="1" />, {
      wrapper: createTestWrapper(),
    });

    await user.click(screen.getByRole("button", { name: /チェック完了/ }));
    const confirmDialog = await screen.findByRole("alertdialog");
    await user.click(within(confirmDialog).getByRole("button", { name: "チェック完了" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledTimes(1);
    });
  });

  it("差戻の失敗時、エラートーストは1回だけ表示する", async () => {
    server.use(
      http.get("*/v1/medical-records/:id/billing-confirmation", () =>
        HttpResponse.json({ id: "1", medical_record_id: MEDICAL_RECORD_ID, status: "confirmed" }),
      ),
      http.post("*/v1/medical-records/:id/billing-confirmation/return", () =>
        HttpResponse.json({ error: "internal error" }, { status: 500 }),
      ),
    );

    const user = userEvent.setup();
    render(<MedicalRecordBillCheck medicalRecordId={MEDICAL_RECORD_ID} petId="1" />, {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /確認を取り消す/ })).toBeInTheDocument();
    });
    await user.click(screen.getByRole("button", { name: /確認を取り消す/ }));
    const returnDialog = await screen.findByRole("alertdialog");
    await user.click(within(returnDialog).getByRole("button", { name: "差し戻す" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledTimes(1);
    });
  });
});
