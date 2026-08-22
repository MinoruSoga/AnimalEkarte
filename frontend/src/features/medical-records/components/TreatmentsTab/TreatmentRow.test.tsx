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

    expect(onUpdate).not.toHaveBeenCalled();
    expect(screen.queryByText("投与量を確認してください")).not.toBeInTheDocument();
    // 無効値が入力中の間は保存不可理由を隠さない
    expect(await screen.findByRole("alert")).toHaveTextContent(/上限.*保存できません/);

    await user.keyboard("{Enter}");
    expect(onUpdate).not.toHaveBeenCalled();
  });

  it("上限超過 → 正常値復帰でエラー消滅", async () => {
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
    // 無効値が入力中の間は保存不可理由を隠さない
    expect(await screen.findByRole("alert")).toHaveTextContent(/上限.*保存できません/);

    await user.keyboard("{Enter}");

    expect(onUpdate).not.toHaveBeenCalled();
    // 数量は正常値に戻っている。sticky エラーを F5 なしで消す
    expect(screen.getByRole("button", { name: "2" })).toBeInTheDocument();
    expect(screen.queryByText(/上限.*保存できません/)).not.toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});

// #201 TASK-025: useMedicineDoseParams の error を捨てず、technical failure 中は onUpdate 0 回。
describe("TreatmentRow — dose-params technical failure (TASK-025)", () => {
  const UPSTREAM_LEAK = "database timeout; SQLSTATE 57P01 internal stack";

  const safeDoseContext: MedicineDoseContext = {
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

  const safeDoseParam = {
    id: 1,
    clinic_id: 1,
    medicine_id: 5,
    species: "dog",
    dose_basis: "per_administration",
    // dose_per_kg 5 × weight 4kg / strength 10mg = 推奨 2。1→2 は安全域内・乖離なし。
    dose_per_kg: 5,
    max_mg_per_kg: 100,
    notes: "",
    created_at: "2026-07-15T00:00:00Z",
    updated_at: "2026-07-15T00:00:00Z",
  };

  it("dose-params 取得失敗時に visible error を role=alert で表示し onUpdate を呼ばない", async () => {
    server.use(
      http.get("*/v1/masters/medicines/5/dose-params", () =>
        HttpResponse.json({ message: UPSTREAM_LEAK }, { status: 500 })
      )
    );
    const onUpdate = vi.fn();
    const user = userEvent.setup();
    renderRow(
      {
        ...baseTreatment,
        item_type: "medicine",
        medicine_id: "5",
        quantity: 1,
      },
      { onUpdate, doseContext: safeDoseContext }
    );

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(/投与量パラメータを取得できなかった/);
    expect(alert).not.toHaveTextContent(UPSTREAM_LEAK);
    expect(screen.queryByText(UPSTREAM_LEAK)).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "1" }));
    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    await user.clear(quantityInput);
    await user.type(quantityInput, "2");
    await user.keyboard("{Enter}");

    expect(onUpdate).not.toHaveBeenCalled();
  });

  it("retry 後に dose-params が成功したら安全域内 quantity の onUpdate が復元する", async () => {
    let doseParamsCalls = 0;
    server.use(
      http.get("*/v1/masters/medicines/5/dose-params", () => {
        doseParamsCalls += 1;
        if (doseParamsCalls === 1) {
          return HttpResponse.json({ message: UPSTREAM_LEAK }, { status: 500 });
        }
        return HttpResponse.json([safeDoseParam]);
      })
    );

    const onUpdate = vi.fn();
    const user = userEvent.setup();
    renderRow(
      {
        ...baseTreatment,
        item_type: "medicine",
        medicine_id: "5",
        quantity: 1,
      },
      { onUpdate, doseContext: safeDoseContext }
    );

    expect(await screen.findByRole("alert")).toHaveTextContent(
      /投与量パラメータを取得できなかった/
    );

    await user.click(
      screen.getByRole("button", { name: /投与量パラメータの取得を再試行/ })
    );

    // dose_per_kg=5 × 4kg / strength10 = 推奨2。プレビュー文言で ready を待つ。
    await screen.findByText(/推奨2/);

    await user.click(screen.getByRole("button", { name: "1" }));
    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    await user.clear(quantityInput);
    await user.type(quantityInput, "2");
    await user.keyboard("{Enter}");

    expect(onUpdate).toHaveBeenCalledWith("1", { quantity: 2 });
  });

  it("体重欠落では technical failure にせず従来どおり onUpdate 可能（missing data）", async () => {
    server.use(
      http.get("*/v1/masters/medicines/5/dose-params", () =>
        HttpResponse.json([safeDoseParam])
      )
    );
    const onUpdate = vi.fn();
    const user = userEvent.setup();
    renderRow(
      {
        ...baseTreatment,
        item_type: "medicine",
        medicine_id: "5",
        quantity: 1,
      },
      {
        onUpdate,
        doseContext: { ...safeDoseContext, weightKg: null },
      }
    );

    // 失敗 alert が無いことを少し待つ（query 成功で empty gate）
    await new Promise((r) => setTimeout(r, 50));
    expect(screen.queryByText(/投与量パラメータを取得できなかった/)).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "1" }));
    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    await user.clear(quantityInput);
    await user.type(quantityInput, "3");
    await user.keyboard("{Enter}");

    expect(onUpdate).toHaveBeenCalledWith("1", { quantity: 3 });
  });

  it("species 欠落では technical failure にせず従来どおり onUpdate 可能（missing data）", async () => {
    server.use(
      http.get("*/v1/masters/medicines/5/dose-params", () =>
        HttpResponse.json([safeDoseParam])
      )
    );
    const onUpdate = vi.fn();
    const user = userEvent.setup();
    renderRow(
      {
        ...baseTreatment,
        item_type: "medicine",
        medicine_id: "5",
        quantity: 1,
      },
      {
        onUpdate,
        doseContext: { ...safeDoseContext, petSpecies: null },
      }
    );

    await new Promise((r) => setTimeout(r, 50));
    expect(screen.queryByText(/投与量パラメータを取得できなかった/)).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "1" }));
    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    await user.clear(quantityInput);
    await user.type(quantityInput, "3");
    await user.keyboard("{Enter}");

    expect(onUpdate).toHaveBeenCalledWith("1", { quantity: 3 });
  });

  it("dose-params 空配列では technical failure にせず従来どおり onUpdate 可能（missing data）", async () => {
    server.use(
      http.get("*/v1/masters/medicines/5/dose-params", () => HttpResponse.json([]))
    );
    const onUpdate = vi.fn();
    const user = userEvent.setup();
    renderRow(
      {
        ...baseTreatment,
        item_type: "medicine",
        medicine_id: "5",
        quantity: 1,
      },
      { onUpdate, doseContext: safeDoseContext }
    );

    await new Promise((r) => setTimeout(r, 50));
    expect(screen.queryByText(/投与量パラメータを取得できなかった/)).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "1" }));
    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    await user.clear(quantityInput);
    await user.type(quantityInput, "3");
    await user.keyboard("{Enter}");

    expect(onUpdate).toHaveBeenCalledWith("1", { quantity: 3 });
  });
});

// TASK-377: 上限内の著しい乖離 / 下限割れで inline 理由を要求し、空理由では mutation 0 回。
describe("TreatmentRow — dose deviation reason (TASK-377)", () => {
  const deviationParam = {
    id: 1,
    clinic_id: 1,
    medicine_id: 5,
    species: "dog",
    dose_basis: "per_administration",
    dose_per_kg: 5,
    max_mg_per_kg: 100,
    notes: "",
    created_at: "2026-07-15T00:00:00Z",
    updated_at: "2026-07-15T00:00:00Z",
  };

  const belowMinParam = {
    ...deviationParam,
    min_mg_per_kg: 4,
    max_mg_per_kg: 10,
    dose_per_kg: 5,
  };

  beforeEach(() => {
    server.use(
      http.get("*/v1/masters/medicines/5/dose-params", () =>
        HttpResponse.json([deviationParam])
      )
    );
  });

  it("著しい乖離で inline 理由を表示し、空理由では onUpdate を呼ばない", async () => {
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
    expect(screen.getByRole("alert")).toHaveTextContent(/推奨値/);
    expect(screen.getByLabelText("用量逸脱の理由")).toBeInTheDocument();
    // modal / confirm dialog を使わない
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.queryByText("投与量を確認してください")).not.toBeInTheDocument();
  });

  it("理由付きで quantity mutation を1回だけ送る", async () => {
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

    const reasonInput = screen.getByLabelText("用量逸脱の理由");
    await user.type(reasonInput, "体重再計測のため");
    await user.keyboard("{Enter}");
    // Enter 後の blur でも idempotent に 1 回だけ
    await user.tab();

    expect(onUpdate).toHaveBeenCalledTimes(1);
    expect(onUpdate).toHaveBeenCalledWith("1", {
      quantity: 5,
      dose_deviation_reason: "体重再計測のため",
    });
  });

  it("下限割れでも inline 理由を要求し空理由では送らない", async () => {
    server.use(
      http.get("*/v1/masters/medicines/5/dose-params", () =>
        HttpResponse.json([belowMinParam])
      )
    );
    const onUpdate = vi.fn();
    const user = userEvent.setup();
    // weightKg 1.5 → min 4*1.5=6mg → qty 0.1 = 1mg below min; recommended ~0.75
    const belowContext: MedicineDoseContext = {
      ...blockingDoseContext,
      weightKg: 1.5,
    };
    renderRow(
      {
        ...baseTreatment,
        item_type: "medicine",
        medicine_id: "5",
        quantity: 1,
      },
      { onUpdate, doseContext: belowContext }
    );

    await screen.findByText(/推奨0\.75/);
    await user.click(screen.getByRole("button", { name: "1" }));
    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    await user.clear(quantityInput);
    await user.type(quantityInput, "0.1");
    await user.keyboard("{Enter}");

    expect(onUpdate).not.toHaveBeenCalled();
    expect(screen.getByRole("alert")).toHaveTextContent(/下限|推奨値/);
    expect(screen.getByLabelText("用量逸脱の理由")).toBeInTheDocument();
  });

  it("safe dose へ戻した quantity は理由なしで1回送信する", async () => {
    const onUpdate = vi.fn();
    const user = userEvent.setup();
    renderRow(
      {
        ...baseTreatment,
        item_type: "medicine",
        medicine_id: "5",
        quantity: 5,
      },
      { onUpdate, doseContext: blockingDoseContext }
    );

    await screen.findByText(/推奨2/);
    await user.click(screen.getByRole("button", { name: "5" }));
    const quantityInput = screen.getByRole("spinbutton", { name: "数量" });
    await user.clear(quantityInput);
    await user.type(quantityInput, "2");
    await user.keyboard("{Enter}");

    expect(onUpdate).toHaveBeenCalledTimes(1);
    expect(onUpdate).toHaveBeenCalledWith("1", { quantity: 2 });
  });
});
