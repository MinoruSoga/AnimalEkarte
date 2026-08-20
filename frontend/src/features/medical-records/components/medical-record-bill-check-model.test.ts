import { describe, expect, it } from "vitest";

import {
  billCheckExtraLines,
  billCheckPricedExtras,
  isUnbillableMasterPrice,
} from "./medical-record-bill-check-model";

describe("medical-record-bill-check-model", () => {
  it("空・負の価格を請求不能と判定する", () => {
    expect(isUnbillableMasterPrice(undefined)).toBe(true);
    expect(isUnbillableMasterPrice(null)).toBe(true);
    expect(isUnbillableMasterPrice(Number.NaN)).toBe(true);
    expect(isUnbillableMasterPrice(-1)).toBe(true);
    expect(isUnbillableMasterPrice(0)).toBe(false);
    expect(isUnbillableMasterPrice(4200)).toBe(false);
  });

  it("このカルテの検査・接種だけを明細化する", () => {
    const lines = billCheckExtraLines(
      "exam",
      [
        { id: 1, name: "血液検査", price: 4200, medicalRecordId: "10" },
        { id: 2, name: "他院カルテ", price: 1000, medicalRecordId: "99" },
        { id: 3, name: "未紐付け", price: 800 },
      ],
      "10",
    );
    expect(lines).toEqual([
      { id: "exam_1", kind: "exam", name: "血液検査", unitPrice: 4200 },
    ]);
  });

  it("価格未設定行は合計から除外する", () => {
    const priced = billCheckPricedExtras([
      { id: "exam_1", kind: "exam", name: "血液検査", unitPrice: 4200 },
      { id: "exam_2", kind: "exam", name: "尿検査", unitPrice: null },
    ]);
    expect(priced.map((line) => line.id)).toEqual(["exam_1"]);
  });
});
