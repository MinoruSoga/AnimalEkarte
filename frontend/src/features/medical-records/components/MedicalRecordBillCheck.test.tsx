import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
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
    onSelect: (item: {
      id: string;
      name: string;
      unitPrice: number;
      category: string;
      medicineId?: string;
    }) => void;
  }) => {
    mockDialogProps.current = props as unknown as Record<string, unknown>;
    return (
      <>
        <button
          type="button"
          onClick={() =>
            props.onSelect({ id: "7", name: "初診", unitPrice: 3307, category: "診察" })
          }
        >
          診察Aを選択
        </button>
        <button
          type="button"
          onClick={() =>
            props.onSelect({ id: "8", name: "処置X", unitPrice: 3306, category: "処置" })
          }
        >
          処置Xを選択
        </button>
        <button
          type="button"
          onClick={() =>
            props.onSelect({
              id: "42",
              name: "抗生剤A",
              unitPrice: 3303,
              category: "薬剤",
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

      await user.click(screen.getByRole("button", { name: "行を追加（検索）" }));
      await user.click(screen.getByRole("button", { name: label }));

      expect(mockCreateTreatmentMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          [refKey]: refId,
          content: expect.any(String),
          memo,
          is_insurance: true,
          quantity: 1,
          unit_price: price,
        }),
      );
      const call = mockCreateTreatmentMutate.mock.calls[0][0] as Record<string, unknown>;
      expect(typeof call[refKey]).toBe("number");
      expect(call.sort_order).toBe(1);
      expect(mockDialogProps.current.clinicId).toBe("2");
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
      const select = mockDialogProps.current.onSelect as ((item: unknown) => void) | undefined;
      expect(select).toBeDefined();
      select?.({ id: "7", name: "初診", unitPrice: 3307, category: "診察" });
      expect(mockCreateTreatmentMutate).not.toHaveBeenCalled();
      if (guard === "deceased") expect(toast.error).toHaveBeenCalled();
      view.unmount();
    },
  );
});

describe("MedicalRecordBillCheck FE-RC-027 item_type 判定", () => {
  it("マスタ検索で category=薬剤 を選択すると item_type=medicine で作成される", async () => {
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

    await user.click(screen.getByRole("button", { name: "行を追加（検索）" }));
    await user.click(screen.getByRole("button", { name: "薬剤Aを選択" }));

    expect(mockCreateTreatmentMutate).toHaveBeenCalledWith(
      expect.objectContaining({ item_type: "medicine", medicine_id: 42 }),
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
