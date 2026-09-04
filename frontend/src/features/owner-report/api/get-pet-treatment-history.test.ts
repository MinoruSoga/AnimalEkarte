import { describe, it, expect } from "vitest";
import { transformHistoryItem, type BackendPetTreatmentHistory } from "./get-pet-treatment-history";

function row(overrides: Partial<BackendPetTreatmentHistory> = {}): BackendPetTreatmentHistory {
  return {
    id: "1",
    medical_record_id: "9",
    date: "2026-06-01T00:00:00+09:00",
    item_type: "medicine",
    content: "",
    memo: "",
    admin_route: "",
    quantity: 1,
    unit_price: 0,
    status: "completed",
    ...overrides,
  };
}

describe("transformHistoryItem", () => {
  it("medicine は medicine_name を表示名にする", () => {
    const out = transformHistoryItem(
      row({
        item_type: "medicine",
        medicine_name: "アモキシシリン",
        content: "投薬メモ",
        admin_route: "経口",
        quantity: 2,
      }),
    );
    expect(out.name).toBe("アモキシシリン");
    expect(out.itemType).toBe("medicine");
    expect(out.adminRoute).toBe("経口");
    expect(out.quantity).toBe(2);
  });

  it("procedure は procedure_name と麻酔ラベルを表示する", () => {
    const out = transformHistoryItem(
      row({
        item_type: "procedure",
        procedure_name: "避妊手術",
        anesthesia: "general",
        content: "",
      }),
    );
    expect(out.name).toBe("避妊手術");
    expect(out.anesthesia).toBe("全身麻酔");
  });

  it("medicine/procedure 名が無ければ content にフォールバックする", () => {
    const out = transformHistoryItem(row({ item_type: "consultation", content: "再診" }));
    expect(out.name).toBe("再診");
    expect(out.anesthesia).toBeUndefined();
  });

  it("名前も content も空なら '-' を表示する", () => {
    const out = transformHistoryItem(
      row({ content: "", medicine_name: undefined, procedure_name: undefined }),
    );
    expect(out.name).toBe("-");
  });

  it("未知の麻酔値は原文を保持する", () => {
    const out = transformHistoryItem(
      row({ item_type: "procedure", procedure_name: "X", anesthesia: "spinal" }),
    );
    expect(out.anesthesia).toBe("spinal");
  });

  it("is_surgery=true は isSurgery=true にマップされる", () => {
    const out = transformHistoryItem(
      row({
        item_type: "procedure",
        procedure_name: "開腹手術",
        anesthesia: "general",
        is_surgery: true,
      }),
    );
    expect(out.isSurgery).toBe(true);
  });

  it("is_surgery が undefined のとき isSurgery は undefined", () => {
    const out = transformHistoryItem(
      row({ item_type: "medicine", medicine_name: "アモキシシリン" }),
    );
    expect(out.isSurgery).toBeUndefined();
  });

  it("日付は YY/M/D 形式、null は '-'", () => {
    expect(transformHistoryItem(row({ date: "2026-06-01T00:00:00+09:00" })).date).toMatch(
      /^\d{2}\/\d{1,2}\/\d{1,2}$/,
    );
    expect(transformHistoryItem(row({ date: null })).date).toBe("-");
  });

  it("日付は JST 壁日付で算出する（UTC 深夜は翌日になる）", () => {
    // 2026-06-10T15:30:00Z = JST 2026-06-11 00:30 → "26/6/11"（ローカル UTC 基準だと "26/6/10"）。
    expect(transformHistoryItem(row({ date: "2026-06-10T15:30:00Z" })).date).toBe("26/6/11");
  });
});
