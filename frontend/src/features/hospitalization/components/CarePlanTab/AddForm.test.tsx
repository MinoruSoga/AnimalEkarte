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
  }: {
    type: string;
    value: string | null;
    onChange: (v: string | null) => void;
  }) => (
    <input
      aria-label="ref-select-stub"
      data-type={type}
      value={value ?? ""}
      onChange={(e) => onChange(e.target.value || null)}
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
});
