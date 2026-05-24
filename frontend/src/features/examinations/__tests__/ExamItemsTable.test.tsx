import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { ExamItemsTable, type ExamItemRow } from "../components/ExamItemsTable";

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

describe("ExamItemsTable", () => {
  describe("empty 状態", () => {
    it("items が空のときテンプレ待ちメッセージを表示する", () => {
      render(<ExamItemsTable items={[]} onChangeInspectionValue={vi.fn()} />);
      expect(
        screen.getByText("検査種別を選択すると検査項目が表示されます"),
      ).toBeInTheDocument();
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
      render(
        <ExamItemsTable
          items={[makeItem()]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByText("項目名")).toBeInTheDocument();
      expect(screen.getByText("結果値")).toBeInTheDocument();
      expect(screen.getByText("単位")).toBeInTheDocument();
      expect(screen.getByText("基準値")).toBeInTheDocument();
      expect(screen.getByText("判定")).toBeInTheDocument();
    });

    it("項目名・単位・基準値を表示する", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({ name: "GLU", unit: "mg/dL", referenceValue: "70-110" }),
          ]}
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
      expect(screen.getByText("HIGH")).toBeInTheDocument();
      expect(screen.queryByText("LOW")).not.toBeInTheDocument();
    });

    it("status=low で LOW バッジを表示する", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "low", isAbnormal: true })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByText("LOW")).toBeInTheDocument();
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
      expect(screen.getByText("-")).toBeInTheDocument();
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
      expect(screen.getByTestId("exam-item-row")).toHaveAttribute(
        "data-abnormal",
        "true",
      );
    });

    it("isAbnormal=true status=low の行に data-abnormal='true' が付く", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "low", isAbnormal: true })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByTestId("exam-item-row")).toHaveAttribute(
        "data-abnormal",
        "true",
      );
    });

    it("isAbnormal=false の行に data-abnormal='false' が付く", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "normal", isAbnormal: false })]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByTestId("exam-item-row")).toHaveAttribute(
        "data-abnormal",
        "false",
      );
    });

    it("isAbnormal 未設定（undefined）の行に data-abnormal='false' が付く", () => {
      render(
        <ExamItemsTable
          items={[makeItem()]}
          onChangeInspectionValue={vi.fn()}
        />,
      );
      expect(screen.getByTestId("exam-item-row")).toHaveAttribute(
        "data-abnormal",
        "false",
      );
    });

    it("disabled=true でも isAbnormal=true の行は data-abnormal='true' を維持する", () => {
      render(
        <ExamItemsTable
          items={[makeItem({ status: "high", isAbnormal: true })]}
          onChangeInspectionValue={vi.fn()}
          disabled
        />,
      );
      expect(screen.getByTestId("exam-item-row")).toHaveAttribute(
        "data-abnormal",
        "true",
      );
      expect(screen.getByLabelText("GLUの結果値")).toBeDisabled();
    });

    it("異常行と正常行が混在する場合、それぞれ data-abnormal が正しく設定される", () => {
      render(
        <ExamItemsTable
          items={[
            makeItem({ key: "1", name: "GLU", status: "high", isAbnormal: true }),
            makeItem({ key: "2", name: "BUN", status: "normal", isAbnormal: false }),
            makeItem({ key: "3", name: "ALT", status: "low", isAbnormal: true }),
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
          items={[
            makeItem({ key: "tmpl-42", name: "GLU", inspectionValue: "" }),
          ]}
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
          items={[
            makeItem({ key: "row-A", name: "GLU" }),
            makeItem({ key: "row-B", name: "BUN" }),
          ]}
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
});
