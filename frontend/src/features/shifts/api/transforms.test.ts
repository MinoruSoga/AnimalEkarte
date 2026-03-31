import { describe, it, expect } from "vitest";
import { transformShift } from "./transforms";
import type { BackendShift } from "./types";

const minimal: BackendShift = {
  id: 1,
  clinic_id: 1,
  staff_id: 2,
  date: "2026-03-25T00:00:00Z",
  shift_type: "day",
  start_time: "09:00",
  end_time: "18:00",
  note: "",
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformShift", () => {
  it("id を string に変換する", () => {
    expect(transformShift({ ...minimal, id: 7 }).id).toBe("7");
  });

  it("clinic_id を string に変換する", () => {
    expect(transformShift({ ...minimal, clinic_id: 3 }).clinic_id).toBe("3");
  });

  it("staff_id を string に変換する", () => {
    expect(transformShift({ ...minimal, staff_id: 5 }).staff_id).toBe("5");
  });

  it("date を T以前の日付部分のみに整形する", () => {
    expect(transformShift({ ...minimal, date: "2026-03-25T00:00:00Z" }).date).toBe("2026-03-25");
  });

  it("date が未設定のとき空文字を返す", () => {
    expect(transformShift({ ...minimal, date: undefined as unknown as string }).date).toBe("");
  });

  it("staff.name を staff_name に優先する", () => {
    const result = transformShift({
      ...minimal,
      staff: { id: 5, clinic_id: 1, name: "山田" } as BackendShift["staff"],
      staff_name: "旧名前",
    });
    expect(result.staff_name).toBe("山田");
  });

  it("staff が未設定のとき staff_name フィールドを使う", () => {
    const result = transformShift({ ...minimal, staff: undefined, staff_name: "鈴木" });
    expect(result.staff_name).toBe("鈴木");
  });

  it("shift_type をそのまま返す", () => {
    expect(transformShift({ ...minimal, shift_type: "night" }).shift_type).toBe("night");
  });

  it("note が未設定のとき空文字を返す", () => {
    expect(transformShift({ ...minimal, note: undefined }).note).toBe("");
  });
});
