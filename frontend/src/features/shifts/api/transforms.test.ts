import { describe, it, expect } from "vitest";
import { transformShift } from "./transforms";
import type { ShiftEntry } from "./contract-types";

const minimal: ShiftEntry = {
  id: "7",
  clinic_id: "1",
  staff_id: "5",
  staff_name: "山田",
  date: "2026-03-25",
  shift_type: "full",
  starts_at: "2026-03-25T09:00:00+09:00",
  ends_at: "2026-03-25T18:00:00+09:00",
  version: 3,
  interval_state: "known",
  notes: "",
  breaks: [],
  created_at: "2026-03-25T00:00:00+09:00",
  updated_at: "2026-03-25T00:00:00+09:00",
};

describe("transformShift（EMR-514 固定API契約）", () => {
  it("契約DTOの正十進文字列IDをNumber変換せず保持する", () => {
    const result = transformShift(minimal);
    expect(result.id).toBe("7");
    expect(result.staff_id).toBe("5");
    expect(result.clinic_id).toBe("1");
  });

  it("version と interval_state を保持する", () => {
    const result = transformShift(minimal);
    expect(result.version).toBe(3);
    expect(result.interval_state).toBe("known");
  });

  it("明示NULL日時（off/paid_leave）を null のまま保持する", () => {
    const result = transformShift({
      ...minimal,
      shift_type: "off",
      starts_at: null,
      ends_at: null,
      breaks: [],
    });
    expect(result.starts_at).toBeNull();
    expect(result.ends_at).toBeNull();
  });

  it("unknown休憩を配列から黙って除外しない", () => {
    const result = transformShift({
      ...minimal,
      interval_state: "unknown",
      starts_at: null,
      ends_at: null,
      breaks: [
        { interval_state: "unknown", starts_at: null, ends_at: null },
        { interval_state: "unknown", starts_at: "2026-03-25T12:00:00+09:00", ends_at: null },
        {
          interval_state: "known",
          starts_at: "2026-03-25T15:00:00+09:00",
          ends_at: "2026-03-25T15:30:00+09:00",
        },
      ],
    });
    expect(result.breaks).toHaveLength(3);
    expect(result.breaks[0]?.interval_state).toBe("unknown");
    expect(result.breaks[1]?.starts_at).toBe("2026-03-25T12:00:00+09:00");
    expect(result.breaks[2]?.interval_state).toBe("known");
  });

  it("staff_name をそのまま返す（旧 staff オブジェクト合成はしない）", () => {
    const result = transformShift(minimal);
    expect(result.staff_name).toBe("山田");
  });

  it("notes / date をそのまま返す（dateは開始側JST日付の契約）", () => {
    const result = transformShift({ ...minimal, notes: "メモ", date: "2026-03-25" });
    expect(result.notes).toBe("メモ");
    expect(result.date).toBe("2026-03-25");
  });
});
