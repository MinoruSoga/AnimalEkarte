import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import {
  DynamicCheckupFields,
  buildCheckupResultsPayload,
  type CheckupFieldValue,
} from "./DynamicCheckupFields";
import type { CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";

const fields: CheckupTypeFieldRow[] = [
  { id: 1, checkupTypeId: 10, name: "歯石除去必要の有無", fieldType: "boolean", unit: "", options: [], isProvisional: true, sortOrder: 1 },
  { id: 2, checkupTypeId: 10, name: "歯石付着度スコア", fieldType: "number", unit: "点", minValue: 0, maxValue: 4, options: [], isProvisional: true, sortOrder: 2 },
  {
    id: 3, checkupTypeId: 10, name: "歯科ケアアドバイス", fieldType: "multi_select", unit: "",
    options: [
      { value: "brush", label: "歯磨き" },
      { value: "scaling", label: "スケーリング" },
    ],
    isProvisional: true, sortOrder: 3,
  },
];

describe("DynamicCheckupFields", () => {
  it("field_type に応じた入力コントロールを描画する", () => {
    render(<DynamicCheckupFields fields={fields} values={{}} onChange={vi.fn()} />);

    // boolean → checkbox（フィールド名ラベルに紐づく）
    expect(screen.getByRole("checkbox", { name: "歯石除去必要の有無" })).toBeInTheDocument();
    // number → spinbutton（ラベルにフィールド名）
    expect(screen.getByRole("spinbutton", { name: /歯石付着度スコア/ })).toBeInTheDocument();
    // multi_select → 各選択肢が checkbox として並ぶ
    expect(screen.getByRole("checkbox", { name: "歯磨き" })).toBeInTheDocument();
    expect(screen.getByRole("checkbox", { name: "スケーリング" })).toBeInTheDocument();
    // フィールド名ラベルが表示される
    expect(screen.getByText("歯石付着度スコア")).toBeInTheDocument();
  });

  it("フィールド定義が空なら何も描画しない", () => {
    const { container } = render(<DynamicCheckupFields fields={[]} values={{}} onChange={vi.fn()} />);
    expect(container).toBeEmptyDOMElement();
  });
});

describe("buildCheckupResultsPayload", () => {
  it("型別の値を BE ペイロードに変換する", () => {
    const values: Record<number, CheckupFieldValue> = {
      1: { bool: true },
      2: { number: "5" },
      3: { list: ["brush"] },
    };
    const payload = buildCheckupResultsPayload(fields, values);

    expect(payload).toEqual([
      { checkup_type_field_id: 1, value_bool: true },
      { checkup_type_field_id: 2, value_number: 5 },
      { checkup_type_field_id: 3, value_list: ["brush"] },
    ]);
  });

  it("未入力フィールド（空文字・未選択）は送信しない", () => {
    const values: Record<number, CheckupFieldValue> = {
      2: { number: "" }, // 空文字 → skip
      3: { list: [] }, // 未選択 → skip
    };
    expect(buildCheckupResultsPayload(fields, values)).toEqual([]);
  });

  it("パース不能な数値は送信しない", () => {
    const payload = buildCheckupResultsPayload(fields, { 2: { number: "abc" } });
    expect(payload).toEqual([]);
  });
});
