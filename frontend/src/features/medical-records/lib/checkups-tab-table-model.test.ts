import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { makeDefaultCheckupAddForm } from "./checkups-tab-table-model";

// EMR-431: カルテの健診タブの追加フォームは、開いた時点で「当日＋既定間隔」の
// 次回予定日を持つ。独立画面 (features/checkups/hooks/use-checkup-form.ts、PR665)
// と同じ挙動に揃え、日付に触れずに保存しても next_date が送られるようにする。
describe("makeDefaultCheckupAddForm — 次回予定日の初期化（EMR-431）", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // UTC 16:00 = JST 翌日 01:00。JST 当日が 2026-10-15 になる瞬間を固定する。
    vi.setSystemTime(new Date("2026-10-14T15:01:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("実施日は JST 当日、次回予定日は既定間隔 1年後", () => {
    const form = makeDefaultCheckupAddForm();
    expect(form.date).toBe("2026-10-15");
    expect(form.next_schedule_type).toBe("1year");
    expect(form.next_date).toBe("2027-10-15");
  });
});
