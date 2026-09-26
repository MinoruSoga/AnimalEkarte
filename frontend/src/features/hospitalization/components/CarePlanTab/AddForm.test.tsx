import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AddForm } from "./AddForm";

// CarePlanRefSelect 自体の挙動(マスタ取得・SearchableSelect の描画)は
// CarePlanRefSelect.test.tsx で検証済み。ここでは AddForm が
// 「type に応じて参照選択を要求し、選択された ID を正しいフィールドへ積んで
// onSubmit する」というオーケストレーションだけを検証するため、テストダブルに置き換える。
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

async function selectType(user: ReturnType<typeof userEvent.setup>, label: string) {
  await user.click(screen.getByRole("combobox"));
  const listbox = await screen.findByRole("listbox");
  await user.click(within(listbox).getByText(label));
}

describe("AddForm — type連動マスタ参照(BUG-403)", () => {
  it("type=投薬 選択時、薬剤未選択のままでは追加ボタンが無効", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<AddForm onSubmit={onSubmit} />);

    await selectType(user, "投薬");
    await user.type(screen.getByPlaceholderText("名称を入力"), "抗生剤");

    expect(screen.getByRole("button", { name: /追加/ })).toBeDisabled();
  });

  it("type=投薬 選択時、薬剤選択後に追加すると medicine_id が payload に含まれる", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<AddForm onSubmit={onSubmit} />);

    await selectType(user, "投薬");
    await user.type(screen.getByPlaceholderText("名称を入力"), "抗生剤");
    await user.type(screen.getByLabelText("ref-select-stub"), "med-1");

    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "medicine",
        name: "抗生剤",
        medicine_id: "med-1",
        procedure_id: null,
        hospitalization_plan_id: null,
      }),
    );
  });

  it("type=食事 では参照選択欄を表示せず、FK は全て null で送信される", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<AddForm onSubmit={onSubmit} />);

    await selectType(user, "食事");
    expect(screen.queryByLabelText("ref-select-stub")).not.toBeInTheDocument();

    await user.type(screen.getByPlaceholderText("名称を入力"), "療法食");
    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "food",
        name: "療法食",
        medicine_id: null,
        procedure_id: null,
        hospitalization_plan_id: null,
      }),
    );
  });

  // 回帰(EMR-196 price-loss): 投薬・処置も持ち物と同じく、CarePlanRefSelect が
  // 選択マスタの price を伝播し AddForm が create payload の unit_price に積む。
  // BE は request unit_price を保存し、退院会計は care_plan_items.unit_price を写す。
  it("type=投薬 で薬剤(price=100)を選択すると create payload に unit_price=100 を積む", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<AddForm onSubmit={onSubmit} />);

    await selectType(user, "投薬");
    await user.type(screen.getByPlaceholderText("名称を入力"), "抗生剤");
    await user.type(screen.getByLabelText("ref-select-stub"), "1");
    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "medicine",
        name: "抗生剤",
        medicine_id: "1",
        procedure_id: null,
        hospitalization_plan_id: null,
        unit_price: 100,
      }),
    );
  });

  it("type=処置・検査 で処置マスタ(price=4000)を選択すると create payload に unit_price=4000 を積む", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<AddForm onSubmit={onSubmit} />);

    await selectType(user, "処置・検査");
    await user.type(screen.getByPlaceholderText("名称を入力"), "血液検査");
    await user.type(screen.getByLabelText("ref-select-stub"), "2");
    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "treatment",
        name: "血液検査",
        procedure_id: "2",
        medicine_id: null,
        hospitalization_plan_id: null,
        unit_price: 4000,
      }),
    );
  });

  // 回帰(price-loss 修正): CarePlanRefSelect が選択プランのマスタ price を伝播し、
  // AddForm が create payload の unit_price に積む。BE は request unit_price を保存し、
  // 退院会計は care_plan_items.unit_price を写す。
  it("type=持ち物 で入院プラン(price=1200)を選択すると create payload に unit_price=1200 を積む", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<AddForm onSubmit={onSubmit} />);

    await selectType(user, "持ち物");
    await user.type(screen.getByPlaceholderText("名称を入力"), "スタンダード入院プラン");
    await user.type(screen.getByLabelText("ref-select-stub"), "3");
    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "item",
        name: "スタンダード入院プラン",
        hospitalization_plan_id: "3",
        medicine_id: null,
        procedure_id: null,
        unit_price: 1200,
      }),
    );
  });
});

describe("AddForm — 手入力（その他）明細(EMR-179)", () => {
  it("type=持ち物 で手入力をONにすると参照選択の代わりに単価・理由欄を表示する", async () => {
    const user = userEvent.setup();
    render(<AddForm onSubmit={vi.fn()} />);

    await selectType(user, "持ち物");
    expect(screen.getByLabelText("ref-select-stub")).toBeInTheDocument();

    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));

    expect(screen.queryByLabelText("ref-select-stub")).not.toBeInTheDocument();
    expect(screen.getByLabelText("手入力の単価")).toBeInTheDocument();
    expect(screen.getByLabelText("その他理由")).toBeInTheDocument();
  });

  it("手入力で名称・単価・理由を入れて追加すると manual=true と trim 済み理由を送る", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<AddForm onSubmit={onSubmit} />);

    await selectType(user, "持ち物");
    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));
    await user.type(screen.getByPlaceholderText("名称を入力"), "持ち込み療養食");
    await user.type(screen.getByLabelText("手入力の単価"), "800");
    await user.type(screen.getByLabelText("その他理由"), "  持ち込み品のため  ");
    await user.click(screen.getByRole("button", { name: /追加/ }));

    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        type: "item",
        name: "持ち込み療養食",
        manual: true,
        other_reason: "持ち込み品のため",
        unit_price: 800,
        hospitalization_plan_id: null,
        medicine_id: null,
        procedure_id: null,
      }),
    );
  });

  it("手入力で理由または単価が空のままでは追加ボタンが無効", async () => {
    const user = userEvent.setup();
    render(<AddForm onSubmit={vi.fn()} />);

    await selectType(user, "持ち物");
    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));
    await user.type(screen.getByPlaceholderText("名称を入力"), "持ち込み品");

    expect(screen.getByRole("button", { name: /追加/ })).toBeDisabled();

    await user.type(screen.getByLabelText("手入力の単価"), "800");
    expect(screen.getByRole("button", { name: /追加/ })).toBeDisabled();

    await user.type(screen.getByLabelText("その他理由"), "   ");
    expect(screen.getByRole("button", { name: /追加/ })).toBeDisabled();
  });

  it("手入力→OFF で理由と単価の入力をクリアして参照選択へ戻る", async () => {
    const user = userEvent.setup();
    render(<AddForm onSubmit={vi.fn()} />);

    await selectType(user, "持ち物");
    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));
    await user.type(screen.getByLabelText("その他理由"), "分類保留");

    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));

    expect(screen.getByLabelText("ref-select-stub")).toBeInTheDocument();
    expect(screen.queryByLabelText("その他理由")).not.toBeInTheDocument();
    // 再度 ON にしても理由は残っていない（モード間の持ち越し防止）
    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));
    expect(screen.getByLabelText("その他理由")).toHaveValue("");
  });

  it("持ち物の手入力中に別 type へ切り替えると手入力モードが解除される", async () => {
    const user = userEvent.setup();
    render(<AddForm onSubmit={vi.fn()} />);

    await selectType(user, "持ち物");
    await user.click(screen.getByRole("checkbox", { name: /手入力/ }));
    await user.type(screen.getByLabelText("その他理由"), "分類保留");

    await selectType(user, "指示・その他");

    expect(screen.queryByRole("checkbox", { name: /手入力/ })).not.toBeInTheDocument();
    expect(screen.queryByLabelText("その他理由")).not.toBeInTheDocument();
  });
});
