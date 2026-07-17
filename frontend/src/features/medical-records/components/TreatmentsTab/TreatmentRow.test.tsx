import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import { TreatmentRow } from "./TreatmentRow";
import type { Treatment } from "../../types";
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

function renderRow(treatment: Treatment) {
  render(
    <table>
      <tbody>
        <TreatmentRow
          treatment={treatment}
          isFirst={true}
          isLast={true}
          onUpdate={noop}
          onDelete={noop}
          onMoveUp={noop}
          onMoveDown={noop}
          isUpdating={false}
          doseContext={noopDoseContext}
        />
      </tbody>
    </table>,
    { wrapper: createWrapper() }
  );
}

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
});
