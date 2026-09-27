import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EditRow } from "./EditRow";
import type { CarePlanItem } from "../../api/care-plan-items";

// AddForm.test.tsx と同じ理由でテストダブルに置き換える(CarePlanRefSelect.test.tsx で別途検証済み)。
vi.mock("./CarePlanRefSelect", () => ({
  CarePlanRefSelect: ({
    type,
    value,
    onChange,
    onUnitPriceChange,
  }: {
    type: string;
    value: string | null;
    onChange: (v: string | null) => void;
    onUnitPriceChange?: (price: number | null) => void;
  }) => (
    <input
      aria-label="ref-select-stub"
      data-type={type}
      value={value ?? ""}
      onChange={(e) => {
        const next = e.target.value || null;
        onChange(next);
        // 実物は type に応じた選択マスタの price を伝播する
        // (本テストでは medicine=100 / treatment=4000 / item=1200 固定)
        const price =
          type === "medicine" ? 100 : type === "treatment" ? 4000 : type === "item" ? 1200 : null;
        onUnitPriceChange?.(next !== null ? price : null);
      }}
    />
  ),
}));

afterEach(() => {
  vi.clearAllMocks();
});

const baseItem: CarePlanItem = {
  id: "item-1",
  hospitalization_id: "hosp-1",
  type: "instruction",
  name: "既存指示",
  description: "",
  timing: ["morning"],
  status: "active",
  notes: "",
  medicine_id: null,
  procedure_id: null,
  hospitalization_plan_id: null,
  unit_price: 0,
  category: "",
  manual: false,
  other_reason: "",
  sort_order: 0,
  created_at: "2026-07-17T00:00:00Z",
  updated_at: "2026-07-17T00:00:00Z",
};

async function selectType(user: ReturnType<typeof userEvent.setup>, label: string) {
  await user.click(screen.getByRole("combobox"));
  const listbox = await screen.findByRole("listbox");
  await user.click(within(listbox).getByText(label));
}

describe("EditRow — type連動マスタ参照(BUG-403)", () => {
  it("既存 medicine 項目を編集し type=処置・検査 へ変更すると、保存には処置選択が必須になる", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    const item: CarePlanItem = {
      ...baseItem,
      type: "medicine",
      medicine_id: "med-9",
      name: "既存投薬",
    };
    render(<EditRow item={item} onSave={onSave} onCancel={vi.fn()} />);

    await selectType(user, "処置・検査");

    expect(screen.getByRole("button", { name: /保存/ })).toBeDisabled();
  });

  it("type=持ち物 で入院プランを選択して保存すると hospitalization_plan_id が payload に含まれる", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    render(<EditRow item={baseItem} onSave={onSave} onCancel={vi.fn()} />);

    await selectType(user, "持ち物");
    await user.type(screen.getByLabelText("ref-select-stub"), "plan-1");
    await user.click(screen.getByRole("button", { name: /保存/ }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "item",
        hospitalization_plan_id: "plan-1",
        medicine_id: null,
        procedure_id: null,
      }),
    );
  });

  // 回帰(EMR-196 price-loss): 投薬・処置も持ち物と同じく、マスタ再選択時の price が
  // update payload に積まれ、既存項目は再選択なしでも保存時に永続化済みの単価を維持する。
  it("type=投薬 で薬剤(price=100)を選び直して保存すると update payload に unit_price=100 を積む", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    const item: CarePlanItem = {
      ...baseItem,
      type: "medicine",
      medicine_id: "med-9",
      name: "既存投薬",
    };
    render(<EditRow item={item} onSave={onSave} onCancel={vi.fn()} />);

    await user.clear(screen.getByLabelText("ref-select-stub"));
    await user.type(screen.getByLabelText("ref-select-stub"), "1");
    await user.click(screen.getByRole("button", { name: /保存/ }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "medicine",
        medicine_id: "1",
        unit_price: 100,
      }),
    );
  });

  it("type=処置・検査 で処置マスタ(price=4000)を選び直して保存すると update payload に unit_price=4000 を積む", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    const item: CarePlanItem = {
      ...baseItem,
      type: "treatment",
      procedure_id: "proc-9",
      name: "既存処置",
    };
    render(<EditRow item={item} onSave={onSave} onCancel={vi.fn()} />);

    await user.clear(screen.getByLabelText("ref-select-stub"));
    await user.type(screen.getByLabelText("ref-select-stub"), "2");
    await user.click(screen.getByRole("button", { name: /保存/ }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "treatment",
        procedure_id: "2",
        unit_price: 4000,
      }),
    );
  });

  it("type=投薬の既存項目はマスタ再選択なしでも保存時に unit_price=300 を維持する", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    const item: CarePlanItem = {
      ...baseItem,
      type: "medicine",
      medicine_id: "med-9",
      unit_price: 300,
      name: "既存投薬",
    };
    render(<EditRow item={item} onSave={onSave} onCancel={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /保存/ }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "medicine",
        medicine_id: "med-9",
        unit_price: 300,
      }),
    );
  });

  // 回帰(price-loss 修正): 入院プランを選び直すとマスタ price が update payload に積まる。
  it("type=持ち物 で入院プラン(price=1200)を選択して保存すると update payload に unit_price=1200 を積む", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    render(<EditRow item={baseItem} onSave={onSave} onCancel={vi.fn()} />);

    await selectType(user, "持ち物");
    await user.type(screen.getByLabelText("ref-select-stub"), "plan-1");
    await user.click(screen.getByRole("button", { name: /保存/ }));

    expect(onSave).toHaveBeenCalledTimes(1);
    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "item",
        hospitalization_plan_id: "plan-1",
        unit_price: 1200,
      }),
    );
  });

  // 回帰(price-loss 修正): 既存の持ち物項目は再選択なしでも保存時に永続化済みの単価を維持する。
  it("type=持ち物の既存項目はプラン再選択なしでも保存時に unit_price=1200 を維持する", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    const item: CarePlanItem = {
      ...baseItem,
      type: "item",
      hospitalization_plan_id: "3",
      unit_price: 1200,
      name: "入院プラン項目",
    };
    render(<EditRow item={item} onSave={onSave} onCancel={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /保存/ }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "item",
        hospitalization_plan_id: "3",
        unit_price: 1200,
      }),
    );
  });

  it("type=指示・その他(参照不要)のままなら参照選択欄は表示されない", () => {
    render(<EditRow item={baseItem} onSave={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.queryByLabelText("ref-select-stub")).not.toBeInTheDocument();
  });
});

describe("EditRow — 手入力（その他）明細(EMR-179)", () => {
  const manualItem: CarePlanItem = {
    ...baseItem,
    type: "item",
    name: "持ち込み療養食",
    manual: true,
    other_reason: "持ち込み品のため",
    category: "other",
    unit_price: 800,
    hospitalization_plan_id: null,
  };

  it("手入力行は理由・単価が初期値で表示され、参照選択欄は出ない", () => {
    render(<EditRow item={manualItem} onSave={vi.fn()} onCancel={vi.fn()} />);

    expect(screen.getByRole("checkbox", { name: /手入力/ })).toBeChecked();
    expect(screen.getByLabelText("その他理由")).toHaveValue("持ち込み品のため");
    expect(screen.getByLabelText("手入力の単価")).toHaveValue(800);
    expect(screen.queryByLabelText("ref-select-stub")).not.toBeInTheDocument();
  });

  it("手入力行の理由を編集して保存すると manual=true と trim 済み理由を送る", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    render(<EditRow item={manualItem} onSave={onSave} onCancel={vi.fn()} />);

    const reasonInput = screen.getByLabelText("その他理由");
    await user.clear(reasonInput);
    await user.type(reasonInput, "  新しい理由  ");
    await user.click(screen.getByRole("button", { name: /保存/ }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "item",
        manual: true,
        other_reason: "新しい理由",
        unit_price: 800,
        hospitalization_plan_id: null,
      }),
    );
  });

  it("手入力行で理由を空にすると保存ボタンが無効", async () => {
    const user = userEvent.setup();
    render(<EditRow item={manualItem} onSave={vi.fn()} onCancel={vi.fn()} />);

    await user.clear(screen.getByLabelText("その他理由"));

    expect(screen.getByRole("button", { name: /保存/ })).toBeDisabled();
  });

  it("手入力→OFF で入院プランを選ぶと manual=false + hospitalization_plan_id で保存される", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    render(<EditRow item={manualItem} onSave={onSave} onCancel={vi.fn()} />);

    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));
    expect(screen.queryByLabelText("その他理由")).not.toBeInTheDocument();
    await user.type(screen.getByLabelText("ref-select-stub"), "plan-7");
    await user.click(screen.getByRole("button", { name: /保存/ }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "item",
        manual: false,
        hospitalization_plan_id: "plan-7",
        unit_price: 1200,
      }),
    );
    const payload = onSave.mock.calls[0]?.[0] as Record<string, unknown>;
    expect(payload.other_reason).toBeUndefined();
  });

  it("マスタ参照行で手入力をONにすると参照がクリアされ手入力モードになる", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    const masterItem: CarePlanItem = {
      ...baseItem,
      type: "item",
      name: "入院プラン項目",
      hospitalization_plan_id: "3",
      unit_price: 1200,
    };
    render(<EditRow item={masterItem} onSave={onSave} onCancel={vi.fn()} />);

    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));

    expect(screen.queryByLabelText("ref-select-stub")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /保存/ })).toBeDisabled();

    await user.type(screen.getByLabelText("手入力の単価"), "500");
    await user.type(screen.getByLabelText("その他理由"), "特別持ち込み");
    await user.click(screen.getByRole("button", { name: /保存/ }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "item",
        manual: true,
        other_reason: "特別持ち込み",
        unit_price: 500,
        hospitalization_plan_id: null,
      }),
    );
  });
});
