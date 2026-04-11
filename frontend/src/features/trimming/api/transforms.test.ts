import { describe, it, expect } from "vitest";
import { transformTrimming } from "./transforms";
import type { BackendTrimming } from "@/types/trimming";

/** BackendTrimming の最小スタブ */
const minimalBackend: BackendTrimming = {
  id: 1,
  clinic_id: 1,
  date: "2026-03-25T00:00:00Z",
  status: "reserved",
  style_request: "",
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformTrimming", () => {
  it("id を string に変換する", () => {
    const result = transformTrimming({ ...minimalBackend, id: 42 });
    expect(result.id).toBe("42");
  });

  it("id が null/undefined のとき '0' を返す", () => {
    const result = transformTrimming({ ...minimalBackend, id: undefined as unknown as number });
    expect(result.id).toBe("0");
  });

  it("date を T以前の日付部分のみに整形する", () => {
    const result = transformTrimming({ ...minimalBackend, date: "2026-03-25T10:00:00Z" });
    expect(result.date).toBe("2026-03-25");
  });

  it("date が未設定のとき空文字を返す", () => {
    const result = transformTrimming({ ...minimalBackend, date: undefined as unknown as string });
    expect(result.date).toBe("");
  });

  it("status: completed → '完了'", () => {
    const result = transformTrimming({ ...minimalBackend, status: "completed" });
    expect(result.status).toBe("完了");
  });

  it("status: reserved → '予約'", () => {
    const result = transformTrimming({ ...minimalBackend, status: "reserved" });
    expect(result.status).toBe("予約");
  });

  it("status: in_progress → '進行中'", () => {
    const result = transformTrimming({ ...minimalBackend, status: "in_progress" });
    expect(result.status).toBe("進行中");
  });

  it("未知の status は '予約' にフォールバックする", () => {
    const result = transformTrimming({ ...minimalBackend, status: "unknown" as "reserved" });
    expect(result.status).toBe("予約");
  });

  it("pet 情報がある場合、petId / petName / petNumber / ownerName を展開する", () => {
    const result = transformTrimming({
      ...minimalBackend,
      pet: {
        id: 10,
        name: "ポチ",
        pet_number: "P-001",
        owner: { id: 20, name: "田中太郎" } as BackendTrimming["pet"]["owner"],
      } as BackendTrimming["pet"],
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

  it("staff 情報がある場合、staff / staffId を展開する", () => {
    const result = transformTrimming({
      ...minimalBackend,
      staff: { id: 5, name: "山田" } as BackendTrimming["staff"],
    });
    expect(result.staff).toBe("山田");
    expect(result.staffId).toBe("5");
  });

  it("body_weight / body_temperature / usedShampoo / remarks を正しく変換する", () => {
    const result = transformTrimming({
      ...minimalBackend,
      body_weight: 3.5,
      bw_unit: "Kg",
      body_temperature: 38.5,
      used_shampoo: "シャンプーA",
      remarks: "特記なし",
    });
    expect(result.bw).toBe("3.5");
    expect(result.bwUnit).toBe("Kg");
    expect(result.bt).toBe("38.5");
    expect(result.usedShampoo).toBe("シャンプーA");
    expect(result.remarks).toBe("特記なし");
  });

  it("optionIds は options 配列の id を string 変換したリスト", () => {
    const result = transformTrimming({
      ...minimalBackend,
      options: [{ id: 1 }, { id: 2 }] as BackendTrimming["options"],
    });
    expect(result.optionIds).toEqual(["1", "2"]);
  });

  it("options が未設定のとき optionIds は空配列", () => {
    const result = transformTrimming({ ...minimalBackend, options: undefined });
    expect(result.optionIds).toEqual([]);
  });
});
