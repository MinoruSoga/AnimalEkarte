import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import {
  MedicineCalculationTypePerWeight,
  MedicineUnitPerTablet,
} from "@/types/generated/models";
import { TreatmentRow } from "./TreatmentRow";
import type { Treatment, UpdateTreatmentInput } from "../../types";
import type { MedicineDoseContext } from "../../api/medicine-dose-lookup";

// FE-refactor 残件1: 保存済み dose_* スナップショットの read-only 表示（保存値ありのみ表示・
// なしは非表示）。乖離警告・再計算プレビューとの整合はスコープ外（別チケット）。

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

const baseTreatment: Treatment = {
  id: "1",
  medical_record_id: "10",
  item_type: "consultation",
  unit_price: 1000,
  quantity: 1,
  is_selected: true,
  status: "pending",
  content: "診察",
  memo: "",
  is_insurance: false,
  discount_rate: 0,
  discount_amount: 0,
  sort_order: 0,
  created_at: "2026-07-12T00:00:00Z",
  updated_at: "2026-07-12T00:00:00Z",
};

const noopDoseContext: MedicineDoseContext = {
  medicines: undefined,
  petSpecies: null,
  weightKg: null,
};

const noop = () => {};

function renderRow(
  treatment: Treatment,
  {
    onUpdate = noop,
    doseContext = noopDoseContext,
  }: {
    onUpdate?: (treatmentId: string, input: UpdateTreatmentInput) => void;
    doseContext?: MedicineDoseContext;
  } = {}
) {
  render(
    <table>
      <tbody>
        <TreatmentRow
          treatment={treatment}
          isFirst={true}
          isLast={true}
          onUpdate={onUpdate}
          onDelete={noop}
          onMoveUp={noop}
          onMoveDown={noop}
          isUpdating={false}
          doseContext={doseContext}
        />
      </tbody>
    </table>,
    { wrapper: createWrapper() }
  );
}

const blockingDoseContext: MedicineDoseContext = {
  medicines: [
    {
      id: "5",
      name: "テスト薬",
      parentId: undefined,
      dosageForm: undefined,
      medicineUnit: MedicineUnitPerTablet,
      price: 100,
      defaultQuantity: undefined,
      inventoryId: undefined,
      description: "",
      isActive: true,
      sortOrder: 0,
      taxType: "excluded",
      taxRate: 0.1,
      isNonInsurance: false,
      createdAt: "",
      updatedAt: "",
      calculationType: MedicineCalculationTypePerWeight,
      strength: 10,
      frequencyPerDay: undefined,
      defaultDurationDays: undefined,
      doseParams: [],
    },
  ],
  petSpecies: "dog",
  weightKg: 4,
};

describe("TreatmentRow — 保存済み投与量スナップショット read-only 表示", () => {
  beforeEach(() => {
    // 保存済みスナップショット表示は dose-params API と独立（BE 保存時点で確定済みの値を
    // そのまま出すだけ）。medicine 行では useMedicineDoseParams がマウント時に発火するため、
    // 未処理リクエスト警告を避けるためだけの空応答スタブ。
    server.use(
      http.get("*/v1/masters/medicines/:id/dose-params", () => HttpResponse.json([]))
    );
  });

  it("dose_amount_mg が保存済みの行では保存時 mg・体重を表示する", () => {
    renderRow({
      ...baseTreatment,
      item_type: "medicine",
      medicine_id: "5",
      dose_weight_kg: 4.2,
      dose_weight_source: "vital_records:123",
      dose_amount_mg: 12.345,
      dose_amount_unit: "mg",
      dose_param_snapshot: { species: "dog", dose_per_kg: 2.5 },
    });

    expect(screen.getByText(/保存時12\.345mg/)).toBeInTheDocument();
    expect(screen.getByText(/体重4\.2kg/)).toBeInTheDocument();
  });

  it("dose_amount_mg が未計算の行では保存時表示を出さない", () => {
    renderRow(baseTreatment);

    expect(screen.queryByText(/保存時/)).not.toBeInTheDocument();
  });

  it("dose_amount_mg が null の行では保存時表示を出さない（明示 null）", () => {
    renderRow({
      ...baseTreatment,
      item_type: "medicine",
      medicine_id: "5",
      dose_amount_mg: null,
    });

    expect(screen.queryByText(/保存時/)).not.toBeInTheDocument();
  });

  it("3つの44px actionが収まる操作列幅を確保する", () => {
    renderRow(baseTreatment);

    expect(screen.getByTitle("上に移動").closest("td")).toHaveClass("w-36");
  });
});

describe("TreatmentRow — 絶対上限超過の物理ブロック", () => {
  beforeEach(() => {
    server.use(
      http.get("*/v1/masters/medicines/5/dose-params", () =>
        HttpResponse.json([
          {
            id: 1,
            clinic_id: 1,
            medicine_id: 5,
            species: "dog",
            dose_basis: "per_administration",
            dose_per_kg: 5,
            max_mg_per_kg: 5,
            notes: "",
            created_at: "2026-07-15T00:00:00Z",
            updated_at: "2026-07-15T00:00:00Z",
          },
        ])
      )
    );
  });

  it("上限超過 quantity は更新せず inline で保存不可理由を表示する", async () => {
    const onUpdate = vi.fn();
    const user = userEvent.setup();
    renderRow(
      {
        ...baseTreatment,
        item_type: "medicine",
        medicine_id: "5",
        quantity: 2,
      },
      { onUpdate, doseContext: blockingDoseContext }
    );

    await screen.findByText(/推奨2/);
    await user.click(screen.getByRole("button", { name: "2" }));
    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    await user.clear(quantityInput);
    await user.type(quantityInput, "5");
    await user.keyboard("{Enter}");

    expect(onUpdate).not.toHaveBeenCalled();
    expect(screen.queryByText("投与量を確認してください")).not.toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent(/上限.*保存できません/);
  });
});
