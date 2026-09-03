import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";
import { validateVaccinationForm, type VaccinationFormState } from "./use-vaccination-form-model";

function baseForm(overrides: Partial<VaccinationFormState> = {}): VaccinationFormState {
  return {
    vaccineId: "1",
    date: "",
    supplemental: "",
    lot1: "",
    lot2: "",
    lot3: "",
    lot4: "",
    nextScheduleType: "1year",
    nextDate: "",
    remarks: "",
    ...overrides,
  };
}

// FE-RC-003: validateVaccinationForm は date-only 契約を JST の文字列比較
// (todayJSTISO()) で判定しなければならない。new Date() のローカル時刻比較は
// 実行環境の TZ が JST でない場合（例: TZ=UTC の CI/Docker コンテナ）に
// JST 当日を未来日と誤検出する。本テストは TZ=UTC を強制し、
// UTC の日付境界と JST の日付境界がズレる瞬間（UTC 16:00 = JST 翌日 01:00）を
// vi.setSystemTime で固定して、JST 当日の入力が有効と判定されることを証明する。
describe("validateVaccinationForm — JST date compare (FE-RC-003)", () => {
  beforeAll(() => {
    process.env.TZ = "UTC";
  });

  beforeEach(() => {
    vi.useFakeTimers();
    // UTC 2026-07-10T16:00:00Z = JST 2026-07-11 01:00。
    // ローカル(UTC)日付は "2026-07-10" だが JST 当日は "2026-07-11"。
    vi.setSystemTime(new Date("2026-07-10T16:00:00.000Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("JST当日の接種日はTZ=UTC環境でも「今日以前」として有効になる（新規登録）", () => {
    const errors = validateVaccinationForm(false, baseForm({ date: "2026-07-11" }));
    expect(errors.date).toBeUndefined();
  });

  it("JST当日の接種日はTZ=UTC環境でも「今日以前」として有効になる（編集）", () => {
    const errors = validateVaccinationForm(true, baseForm({ date: "2026-07-11" }));
    expect(errors.date).toBeUndefined();
  });

  it("JST翌日（真の未来日）は依然としてエラーになる", () => {
    const errors = validateVaccinationForm(false, baseForm({ date: "2026-07-12" }));
    expect(errors.date).toBe("接種日は今日以前の日付を入力してください");
  });

  it("次回予定日がJST当日ならエラーにならない（本日以降チェック）", () => {
    const errors = validateVaccinationForm(
      false,
      baseForm({ date: "2026-07-01", nextDate: "2026-07-11" }),
    );
    expect(errors.nextDate).toBeUndefined();
  });

  it("次回予定日がJST当日より前（真の過去日）ならエラーになる", () => {
    const errors = validateVaccinationForm(
      false,
      baseForm({ date: "2026-07-01", nextDate: "2026-07-10" }),
    );
    expect(errors.nextDate).toBe("次回予定日は本日以降の日付を入力してください");
  });
});
