import { describe, it, expect, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";

import { ExamItemsTable, type ExamItemRow } from "./ExamItemsTable";
import { C } from "@/lib/design-tokens";

// テスト用 ExamItemRow ファクトリ。デフォルトは status 未設定（保存前の新規行）。
const makeItem = (overrides: Partial<ExamItemRow> = {}): ExamItemRow => ({
  key: "tmpl-1",
  examTypeFieldId: 1,
  name: "GLU",
  inspectionValue: "",
  unit: "mg/dL",
  normalValue: "70-110",
  referenceValue: "70-110",
  refMin: 70,
  refMax: 110,
  sortOrder: 0,
  ...overrides,
});

function EditableItemsHarness({ initialItems }: { initialItems: ExamItemRow[] }) {
  const [items, setItems] = useState(initialItems);

  return (
    <ExamItemsTable
      items={items}
      onChangeInspectionValue={vi.fn()}
      onChangeName={vi.fn()}
      onAddItem={() =>
        setItems((previous) => [
          ...previous,
          makeItem({
            key: `manual-${previous.length + 1}`,
            examTypeFieldId: undefined,
            name: "",
          }),
        ])
      }
      onRemoveItem={(key) => setItems((previous) => previous.filter((item) => item.key !== key))}
    />
  );
}

describe("ExamItemsTable", () => {
  describe("empty 状態", () => {
    it("items が空のときテンプレ待ちメッセージを表示する", () => {
      render(<ExamItemsTable items={[]} onChangeInspectionValue={vi.fn()} />);
      expect(screen.getByText("検査種別を選択すると検査項目が表示されます")).toBeInTheDocument();
    });

    it("empty 状態ではテーブルヘッダ・行・input を描画しない", () => {
      render(<ExamItemsTable items={[]} onChangeInspectionValue={vi.fn()} />);
      expect(screen.queryByText("項目名")).not.toBeInTheDocument();
      expect(screen.queryByText("結果値")).not.toBeInTheDocument();
      expect(screen.queryByText("判定")).not.toBeInTheDocument();
      expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    });
  });

  describe("通常表示", () => {
    it("ヘッダ列（項目名・結果値・単位・基準値・判定）を描画する", () => {
      render(<ExamItemsTable items={[makeItem()]} onChangeInspectionValue={vi.fn()} />);
      expect(screen.getByText("項目名")).toBeInTheDocument();
      expect(screen.getByText("結果値")).toBeInTheDocument();
      expect(screen.getByText("単位")).toBeInTheDocument();
      expect(screen.getByText("基準値")).toBeInTheDocument();
      expect(screen.getByText("判定")).toBeInTheDocument();
    });

    it("項目名・単位・基準値を表示する", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ name: "GLU", unit: "mg/dL", referenceValue: "70-110" })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByText("GLU")).toBeInTheDocument();
      expect(screen.getByText("mg/dL")).toBeInTheDocument();
      expect(screen.getByText("70-110")).toBeInTheDocument();
    });

    it("input に inspectionValue が初期表示される", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ name: "GLU", inspectionValue: "95" })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      const input = screen.getByLabelText("GLUの結果値") as HTMLInputElement;
      expect(input.value).toBe("95");
    });

    it("結果値inputは44px以上の操作領域を持つ", () => {
      render(
        <ExamItemsTable items={[makeItem({ name: "GLU" })]} onChangeInspectionValue={vi.fn()} />,
      );

      expect(screen.getByLabelText("GLUの結果値")).toHaveClass("h-11", "min-w-11");
    });

    it("項目名が空でも結果値inputに一意なaccessible nameとid/nameを付ける", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({ key: "empty-1", examTypeFieldId: 101, name: "" }),
            makeItem({ key: "empty-2", examTypeFieldId: 102, name: "   " }),
          ]}
          onChangeInspectionValue={vi.fn()}
        />,
      );

      const firstInput = screen.getByRole("textbox", {
        name: "検査項目1の結果値",
      });
      const secondInput = screen.getByRole("textbox", {
        name: "検査項目2の結果値",
      });

      expect(firstInput).toHaveAttribute("id");
      expect(firstInput).toHaveAttribute("name", "examItems.0.inspectionValue");
      expect(secondInput).toHaveAttribute("id");
      expect(secondInput).toHaveAttribute("name", "examItems.1.inspectionValue");
      expect(firstInput.id).not.toBe(secondInput.id);
    });

    it("referenceValue が空のとき normalValue にフォールバックする", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({
              name: "BUN",
              referenceValue: "",
              normalValue: "6-25",
            }),
          ]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByText("6-25")).toBeInTheDocument();
    });

    it("referenceValue / normalValue が両方空なら基準値列に '-' を表示する", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({
              name: "X",
              unit: "mg/dL",
              referenceValue: "",
              normalValue: "",
              status: "normal",
            }),
          ]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      // 基準値列のフォールバック '-' のみが残る
      // (status=normal のため判定列は CheckCircle アイコン、unit は "mg/dL")
      expect(screen.getByText("-")).toBeInTheDocument();
    });

    it("unit が空なら単位列に '-' を表示する", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({
              name: "X",
              unit: "",
              referenceValue: "70-110",
              status: "normal",
            }),
          ]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      // unit のフォールバック "-" のみ
      expect(screen.getByText("-")).toBeInTheDocument();
    });

    it("複数行を全て描画する", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({ key: "1", name: "GLU" }),
            makeItem({ key: "2", name: "BUN" }),
            makeItem({ key: "3", name: "ALT" }),
          ]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByText("GLU")).toBeInTheDocument();
      expect(screen.getByText("BUN")).toBeInTheDocument();
      expect(screen.getByText("ALT")).toBeInTheDocument();
      expect(screen.getAllByRole("textbox")).toHaveLength(3);
    });
  });

  describe("StatusBadge — backend 由来 status の表示固定", () => {
    it("status=high で HIGH バッジを表示する", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "high", isAbnormal: true })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByText("HIGH")).toHaveClass(C.bgDanger);
      expect(screen.getByTestId("exam-item-row")).toHaveClass(C.bgDanger8);
      expect(screen.queryByText("LOW")).not.toBeInTheDocument();
    });

    it("status=low で LOW バッジを表示する", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "low", isAbnormal: true })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByText("LOW")).toHaveClass(
        C.textStatusBlue,
        C.borderBlue400,
        C.bgStatusBlueLight,
      );
      expect(screen.getByTestId("exam-item-row")).toHaveClass(C.bgStatusBlueLight);
      expect(screen.queryByText("HIGH")).not.toBeInTheDocument();
    });

    it("status=normal では HIGH/LOW バッジを描画しない（CheckCircle のみ）", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "normal", isAbnormal: false })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.queryByText("HIGH")).not.toBeInTheDocument();
      expect(screen.queryByText("LOW")).not.toBeInTheDocument();
      // unit と reference が埋まっているので、判定列の "-" も出てはならない
      expect(screen.queryByText("-")).not.toBeInTheDocument();
    });

    it("未評価の normal は評価済み normal と異なる表示になる", () => {
      const { rerender } = render(
        <ExamItemsTable
          items={[makeItem({ status: "normal", isAssessed: false })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );

      expect(screen.getByText("未判定")).toBeInTheDocument();
      expect(screen.getByText("（基準値未設定のため判定していない）")).toHaveClass("sr-only");

      rerender(
        <ExamItemsTable
          items={[makeItem({ status: "normal", isAssessed: true })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.queryByText("未判定")).not.toBeInTheDocument();
      expect(screen.getByRole("img", { name: "基準値内" })).toBeInTheDocument();
    });

    it.each([
      { status: "high" as const, label: "HIGH" },
      { status: "low" as const, label: "LOW" },
    ])("未評価フラグがあっても異常 status=$status を隠さない", ({ status, label }) => {
      render(
        <ExamItemsTable
          items={[makeItem({ status, isAssessed: false, isAbnormal: true })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );

      expect(screen.getByText(label)).toBeInTheDocument();
      expect(screen.queryByText("未判定")).not.toBeInTheDocument();
    });

    it("status=undefined（未判定）は判定列に '-' を表示する", () => {
      render(
        <ExamItemsTable
          items={[
            // unit / reference は埋め、判定列のみ '-' になるよう絞る
            makeItem({ name: "GLU", unit: "mg/dL", referenceValue: "70-110" }),
          ]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByText("-")).toHaveClass(C.text45);
      expect(screen.queryByText("HIGH")).not.toBeInTheDocument();
      expect(screen.queryByText("LOW")).not.toBeInTheDocument();
    });

    it("複数行で status が異なれば、それぞれ独立に表示される", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({ key: "1", name: "GLU", status: "high" }),
            makeItem({ key: "2", name: "BUN", status: "low" }),
            makeItem({ key: "3", name: "ALT", status: "normal" }),
          ]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByText("HIGH")).toBeInTheDocument();
      expect(screen.getByText("LOW")).toBeInTheDocument();
    });
  });

  describe("disabled 条件 — 確定済み状態の固定", () => {
    it("disabled=true で全 input が disabled になる", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({ key: "1", name: "A", inspectionValue: "" }),
            makeItem({ key: "2", name: "B", inspectionValue: "100" }),
            makeItem({ key: "3", name: "C", inspectionValue: "" }),
          ]}
          onChangeInspectionValue={vi.fn()}
          disabled
        />,
      );
      const inputs = screen.getAllByRole("textbox");
      expect(inputs).toHaveLength(3);
      inputs.forEach((input) => expect(input).toBeDisabled());
    });

    it("disabled が未指定（既定 false）のとき input は有効", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ key: "1", name: "A" })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByLabelText("Aの結果値")).not.toBeDisabled();
    });

    it("disabled=false を明示しても input は有効", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ key: "1", name: "A" })]}
          onChangeInspectionValue={vi.fn()}
          disabled={false}
        />,
      );
      expect(screen.getByLabelText("Aの結果値")).not.toBeDisabled();
    });
  });

  describe("isAbnormal 行ハイライト", () => {
    it("isAbnormal=true status=high の行に data-abnormal='true' が付く", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "high", isAbnormal: true })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByTestId("exam-item-row")).toHaveAttribute("data-abnormal", "true");
    });

    it("isAbnormal=true status=low の行に data-abnormal='true' が付く", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "low", isAbnormal: true })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByTestId("exam-item-row")).toHaveAttribute("data-abnormal", "true");
    });

    it("isAbnormal=false の行に data-abnormal='false' が付く", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "normal", isAbnormal: false })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByTestId("exam-item-row")).toHaveAttribute("data-abnormal", "false");
    });

    it("isAbnormal 未設定（undefined）の行に data-abnormal='false' が付く", () => {
      render(<ExamItemsTable items={[makeItem()]} onChangeInspectionValue={vi.fn()} />);
      expect(screen.getByTestId("exam-item-row")).toHaveAttribute("data-abnormal", "false");
    });

    it("disabled=true でも isAbnormal=true の行は data-abnormal='true' を維持する", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "high", isAbnormal: true })]}
          onChangeInspectionValue={vi.fn()}
          disabled
        />,
      );
      expect(screen.getByTestId("exam-item-row")).toHaveAttribute("data-abnormal", "true");
      expect(screen.getByLabelText("GLUの結果値")).toBeDisabled();
    });

    it("異常行と正常行が混在する場合、それぞれ data-abnormal が正しく設定される", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({
              key: "1",
              name: "GLU",
              status: "high",
              isAbnormal: true,
            }),
            makeItem({
              key: "2",
              name: "BUN",
              status: "normal",
              isAbnormal: false,
            }),
            makeItem({
              key: "3",
              name: "ALT",
              status: "low",
              isAbnormal: true,
            }),
          ]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      const rows = screen.getAllByTestId("exam-item-row");
      expect(rows[0]).toHaveAttribute("data-abnormal", "true");
      expect(rows[1]).toHaveAttribute("data-abnormal", "false");
      expect(rows[2]).toHaveAttribute("data-abnormal", "true");
    });

    it("isAbnormal=true status=high の行に HIGH バッジが表示され入力が有効", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ name: "GLU", status: "high", isAbnormal: true })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByText("HIGH")).toBeInTheDocument();
      expect(screen.getByLabelText("GLUの結果値")).not.toBeDisabled();
    });
  });

  describe("onChangeInspectionValue", () => {
    it("input 編集で (key, value) を渡してコールバックが呼ばれる", async () => {
      const user = userEvent.setup();
      const handleChange = vi.fn();
      render(
        <ExamItemsTable
          items={[makeItem({ key: "tmpl-42", name: "GLU", inspectionValue: "" })]}
          onChangeInspectionValue={handleChange}
        />,
      );
      const input = screen.getByLabelText("GLUの結果値");
      await user.type(input, "9");
      expect(handleChange).toHaveBeenCalledWith("tmpl-42", "9");
    });

    it("複数行のうち編集された行の key だけがコールバックに渡る", async () => {
      const user = userEvent.setup();
      const handleChange = vi.fn();
      render(
        <ExamItemsTable
          items={[makeItem({ key: "row-A", name: "GLU" }), makeItem({ key: "row-B", name: "BUN" })]}
          onChangeInspectionValue={handleChange}
        />,
      );
      await user.type(screen.getByLabelText("BUNの結果値"), "5");
      expect(handleChange).toHaveBeenCalledWith("row-B", "5");
      expect(handleChange).not.toHaveBeenCalledWith("row-A", expect.any(String));
    });

    it("disabled=true では入力が抑制されコールバックは呼ばれない", async () => {
      const user = userEvent.setup();
      const handleChange = vi.fn();
      render(
        <ExamItemsTable
          items={[makeItem({ key: "tmpl-42", name: "GLU" })]}
          onChangeInspectionValue={handleChange}
          disabled
        />,
      );
      const input = screen.getByLabelText("GLUの結果値");
      await user.type(input, "9");
      expect(handleChange).not.toHaveBeenCalled();
    });
  });

  describe("手動行の追加・編集・削除", () => {
    it("空状態でも44px以上の追加buttonを表示してcallbackを呼ぶ", async () => {
      const user = userEvent.setup();
      const onAddItem = vi.fn();

      render(<ExamItemsTable items={[]} onChangeInspectionValue={vi.fn()} onAddItem={onAddItem} />);

      const addButton = screen.getByRole("button", { name: "検査項目を追加" });
      expect(addButton).toHaveClass("h-11", "min-w-11");
      await user.click(addButton);
      expect(onAddItem).toHaveBeenCalledOnce();
    });

    it("手動行の項目名を一意にラベル付けし、名前変更を対象keyへ渡す", () => {
      const onChangeName = vi.fn();

      render(
        <ExamItemsTable
          items={[makeItem({ key: "manual-1", examTypeFieldId: undefined, name: "" })]}
          onChangeInspectionValue={vi.fn()}
          onChangeName={onChangeName}
        />,
      );

      const nameInput = screen.getByRole("textbox", {
        name: "検査項目1の項目名",
      });
      expect(nameInput).toHaveAttribute("id");
      expect(nameInput).toHaveAttribute("name", "examItems.0.name");
      expect(nameInput).toHaveClass("h-11", "min-w-11");
      fireEvent.change(nameInput, { target: { value: "手動項目" } });
      expect(onChangeName).toHaveBeenLastCalledWith("manual-1", "手動項目");
    });

    it("行固有の削除buttonから対象keyだけを渡す", async () => {
      const user = userEvent.setup();
      const onRemoveItem = vi.fn();

      render(
        <ExamItemsTable
          items={[
            makeItem({
              key: "manual-1",
              examTypeFieldId: undefined,
              name: "手動項目",
            }),
          ]}
          onChangeInspectionValue={vi.fn()}
          onRemoveItem={onRemoveItem}
        />,
      );

      const deleteButton = screen.getByRole("button", {
        name: "手動項目を削除",
      });
      expect(deleteButton).toHaveClass("h-11", "min-w-11");
      await user.click(deleteButton);
      expect(onRemoveItem).toHaveBeenCalledOnce();
      expect(onRemoveItem).toHaveBeenCalledWith("manual-1");
    });

    it("disabled時は追加・削除・項目名・結果値をすべて無効化する", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({
              key: "manual-1",
              examTypeFieldId: undefined,
              name: "手動項目",
            }),
          ]}
          onChangeInspectionValue={vi.fn()}
          onAddItem={vi.fn()}
          onRemoveItem={vi.fn()}
          onChangeName={vi.fn()}
          disabled
        />,
      );

      expect(screen.getByRole("button", { name: "検査項目を追加" })).toBeDisabled();
      expect(screen.getByRole("button", { name: "手動項目を削除" })).toBeDisabled();
      expect(screen.getByRole("textbox", { name: "手動項目の項目名" })).toBeDisabled();
      expect(screen.getByRole("textbox", { name: "手動項目の結果値" })).toBeDisabled();
    });

    it("追加後は新しい手動項目名へfocusを移す", async () => {
      const user = userEvent.setup();
      render(<EditableItemsHarness initialItems={[]} />);

      await user.click(screen.getByRole("button", { name: "検査項目を追加" }));

      expect(screen.getByRole("textbox", { name: "検査項目1の項目名" })).toHaveFocus();
    });

    it("削除後は次行、残行なしなら追加buttonへfocusを戻す", async () => {
      const user = userEvent.setup();
      render(
        <EditableItemsHarness
          initialItems={[
            makeItem({
              key: "manual-1",
              examTypeFieldId: undefined,
              name: "項目A",
            }),
            makeItem({
              key: "manual-2",
              examTypeFieldId: undefined,
              name: "項目B",
            }),
          ]}
        />,
      );

      await user.click(screen.getByRole("button", { name: "項目Aを削除" }));
      expect(screen.getByRole("button", { name: "項目Bを削除" })).toHaveFocus();

      await user.click(screen.getByRole("button", { name: "項目Bを削除" }));
      expect(screen.getByRole("button", { name: "検査項目を追加" })).toHaveFocus();
    });
  });
});
