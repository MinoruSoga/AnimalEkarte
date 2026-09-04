import { describe, it, expect } from "vitest";
import { transformTrimming } from "./trimming";
import type { BackendTrimming } from "@/types/trimming";

/** BackendTrimming の最小スタブ（BE-119: appointments ベース） */
const minimalBackend: BackendTrimming = {
  id: 1,
  clinic_id: 1,
  reservation_type_id: 9,
  start_time: "2026-03-25T10:00:00+09:00",
  end_time: "2026-03-25T11:30:00+09:00",
  status: "confirmed",
  source: "staff",
  has_detail: false,
  style_request: "",
  bw_unit: "Kg",
  used_shampoo: "",
  used_ribbon: "",
  remarks: "",
  style_image: "",
  completed_image: "",
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
  options: [],
};

describe("transformTrimming", () => {
  it("id を string に変換する", () => {
    const result = transformTrimming({ ...minimalBackend, id: 42 });
    expect(result.id).toBe("42");
  });

  it("has_detail を hasDetail に変換する", () => {
    const result = transformTrimming({ ...minimalBackend, has_detail: true });
    expect(result.hasDetail).toBe(true);
  });

  it("id が null/undefined のとき '0' を返す", () => {
    const result = transformTrimming({ ...minimalBackend, id: undefined as unknown as number });
    expect(result.id).toBe("0");
  });

  it("start_time を T以前の日付部分のみに整形する", () => {
    const result = transformTrimming({
      ...minimalBackend,
      start_time: "2026-03-25T10:00:00+09:00",
    });
    expect(result.date).toBe("2026-03-25");
  });

  it("start_time が未設定のとき空文字を返す", () => {
    const result = transformTrimming({
      ...minimalBackend,
      start_time: undefined as unknown as string,
    });
    expect(result.date).toBe("");
  });

  it("status: completed → '完了'", () => {
    const result = transformTrimming({ ...minimalBackend, status: "completed" });
    expect(result.status).toBe("完了");
  });

  it("status: confirmed → '予約'", () => {
    const result = transformTrimming({ ...minimalBackend, status: "confirmed" });
    expect(result.status).toBe("予約");
  });

  it("status: in_consultation → '進行中'", () => {
    const result = transformTrimming({ ...minimalBackend, status: "in_consultation" });
    expect(result.status).toBe("進行中");
  });

  it("status: canceled → 'キャンセル'", () => {
    const result = transformTrimming({ ...minimalBackend, status: "canceled" });
    expect(result.status).toBe("キャンセル");
  });

  it("status: cancelled → 'キャンセル'", () => {
    const result = transformTrimming({ ...minimalBackend, status: "cancelled" });
    expect(result.status).toBe("キャンセル");
  });

  it("未知の status は '予約' にフォールバックする", () => {
    const result = transformTrimming({ ...minimalBackend, status: "unknown" as "confirmed" });
    expect(result.status).toBe("予約");
  });

  it("pet 情報がある場合、petId / petName / petNumber / ownerName を展開する", () => {
    const result = transformTrimming({
      ...minimalBackend,
      pet: {
        id: 10,
        name: "ポチ",
        pet_number: "P-001",
        owner: { id: 20, name: "田中太郎" },
      },
    });
    expect(result.petId).toBe("10");
    expect(result.petName).toBe("ポチ");
    expect(result.petNumber).toBe("P-001");
    expect(result.ownerName).toBe("田中太郎");
  });

  it("pet が未設定のとき petId は undefined、petName は空文字", () => {
    const result = transformTrimming({ ...minimalBackend, pet: undefined });
    expect(result.petId).toBeUndefined();
    expect(result.petName).toBe("");
  });

  it("staff_id がある場合、staffId を展開する", () => {
    const result = transformTrimming({
      ...minimalBackend,
      staff_id: 5,
      staff: { id: 5, name: "山田" },
    });
    expect(result.staff).toBe("山田");
    expect(result.staffId).toBe("5");
  });

  it("bw / bt / usedShampoo / remarks を正しく変換する（新フィールド名）", () => {
    const result = transformTrimming({
      ...minimalBackend,
      bw: 3.5,
      bw_unit: "Kg",
      bt: 38.5,
      used_shampoo: "シャンプーA",
      remarks: "特記なし",
    });
    expect(result.bw).toBe("3.5");
    expect(result.bwUnit).toBe("Kg");
    expect(result.bt).toBe("38.5");
    expect(result.usedShampoo).toBe("シャンプーA");
    expect(result.remarks).toBe("特記なし");
  });

  it("course がある場合、courseId / courseName を展開する", () => {
    const result = transformTrimming({
      ...minimalBackend,
      course: { id: 3, name: "シャンプー＆カット", price: 5000 },
    });
    expect(result.courseId).toBe("3");
    expect(result.courseName).toBe("シャンプー＆カット");
  });

  it("course が未設定のとき courseId / courseName は空文字", () => {
    const result = transformTrimming({ ...minimalBackend, course: undefined });
    expect(result.courseId).toBe("");
    expect(result.courseName).toBe("");
  });

  it("optionIds は options 配列の id を string 変換したリスト", () => {
    const result = transformTrimming({
      ...minimalBackend,
      options: [
        { id: 1, name: "爪切り" },
        { id: 2, name: "耳掃除" },
      ],
    });
    expect(result.optionIds).toEqual(["1", "2"]);
  });

  it("options が未設定のとき optionIds は空配列", () => {
    const result = transformTrimming({
      ...minimalBackend,
      options: undefined as unknown as BackendTrimming["options"],
    });
    expect(result.optionIds).toEqual([]);
  });
});
