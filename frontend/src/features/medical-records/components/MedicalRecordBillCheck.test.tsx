import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { MedicalRecordBillCheck } from "./MedicalRecordBillCheck";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: true, canEdit: true, canDelete: true }),
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

vi.mock("@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog", () => ({
  TreatmentSearchDialog: ({
    onSelect,
  }: {
    onSelect: (item: { id: string; name: string; unitPrice: number; category: string }) => void;
  }) => (
    <button
      type="button"
      onClick={() => onSelect({ id: "med-1", name: "抗生剤A", unitPrice: 500, category: "薬剤" })}
    >
      薬剤Aを選択
    </button>
  ),
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
});

afterEach(() => {
  server.resetHandlers();
});

// FE-RC-027: マスタ検索から選択した明細の item_type 判定は "薬品"（誤字。マスタの実カテゴリは
// "薬剤"）と比較していたため常に一致せず "other" にフォールバックしていた。
// TreatmentsTab と同じ共有 resolveItemTypeFromCategory に統一し、"薬剤" → "medicine" を保証する。
describe("MedicalRecordBillCheck FE-RC-027 item_type 判定", () => {
  it("マスタ検索で category=薬剤 を選択すると item_type=medicine で作成される", async () => {
    server.use(
      http.get("*/v1/medical-records/:id/billing-confirmation", () =>
        HttpResponse.json({ id: "1", medical_record_id: MEDICAL_RECORD_ID, status: "pending" }),
      ),
    );

    const user = userEvent.setup();
    render(<MedicalRecordBillCheck medicalRecordId={MEDICAL_RECORD_ID} petId="1" />, {
      wrapper: createTestWrapper(),
    });

    await user.click(screen.getByRole("button", { name: "行を追加（検索）" }));
    await user.click(screen.getByRole("button", { name: "薬剤Aを選択" }));

    expect(mockCreateTreatmentMutate).toHaveBeenCalledWith(
      expect.objectContaining({ item_type: "medicine" }),
    );
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

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledTimes(1);
    });
  });
});
