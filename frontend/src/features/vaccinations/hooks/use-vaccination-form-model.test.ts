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

import {
  mergeVaccinationFormData,
  buildCreateVaccinationRequest,
  buildUpdateVaccinationRequest,
} from "./use-vaccination-form-model";
import { jstDateStartISOString } from "@/lib/jst-date";
import type { VaccinationRecord } from "@/types";

function existingVaccination(overrides: Partial<VaccinationRecord> = {}): VaccinationRecord {
  return {
    id: "10",
    petId: "5",
    medicalRecordId: undefined,
    ownerName: "",
    petName: "",
    vaccineId: "1",
    vaccineName: "",
    doctor: "",
    date: "2026-08-08",
    nextDate: "2026-08-29",
    nextScheduleType: "1year",
    lot1: undefined,
    lot2: undefined,
    lot3: undefined,
    lot4: undefined,
    supplemental: undefined,
    remarks: undefined,
    version: 1,
    ...overrides,
  };
}

describe("mergeVaccinationFormData — 編集時の nextScheduleType 再判定（EMR-417）", () => {
  it("保存済み 1year に対し nextDate が不一致なら other に再判定し date/nextDate は保持する", () => {
    const merged = mergeVaccinationFormData(true, existingVaccination(), {});
    expect(merged.nextScheduleType).toBe("other");
    expect(merged.date).toBe("2026-08-08");
    expect(merged.nextDate).toBe("2026-08-29");
  });

  it.each([
    { type: "1year", nextDate: "2027-08-08", expected: "1year" },
    { type: "3weeks", nextDate: "2026-08-29", expected: "3weeks" },
    { type: "4weeks", nextDate: "2026-08-29", expected: "other" },
  ])("保存 type=$type + nextDate=$nextDate → $expected", ({ type, nextDate, expected }) => {
    const merged = mergeVaccinationFormData(
      true,
      existingVaccination({ nextScheduleType: type, nextDate }),
      {},
    );
    expect(merged.nextScheduleType).toBe(expected);
    expect(merged.nextDate).toBe(nextDate);
  });

  it.each([{ nextDate: "" }, { nextDate: undefined }])(
    "nextDate 未設定なら保存済み type を維持する（%j）",
    ({ nextDate }) => {
      const merged = mergeVaccinationFormData(
        true,
        existingVaccination({ nextScheduleType: "1year", nextDate }),
        {},
      );
      expect(merged.nextScheduleType).toBe("1year");
      expect(merged.nextDate).toBe("");
    },
  );

  it("nextScheduleType 未設定なら DEFAULT(1year) にフォールバックしてから再判定する", () => {
    const merged = mergeVaccinationFormData(
      true,
      existingVaccination({ nextScheduleType: undefined, nextDate: "2027-08-08" }),
      {},
    );
    expect(merged.nextScheduleType).toBe("1year");
  });

  it("タイムスタンプ形式の date/nextDate を YYYY-MM-DD に正規化してから再判定する", () => {
    const merged = mergeVaccinationFormData(
      true,
      existingVaccination({
        date: "2026-08-08T00:00:00+09:00",
        nextDate: "2026-08-29T00:00:00+09:00",
        nextScheduleType: "1year",
      }),
      {},
    );
    expect(merged.date).toBe("2026-08-08");
    expect(merged.nextDate).toBe("2026-08-29");
    expect(merged.nextScheduleType).toBe("other");
  });

  it("localOverrides は再判定結果より優先される", () => {
    const merged = mergeVaccinationFormData(true, existingVaccination(), {
      nextScheduleType: "4weeks",
    });
    expect(merged.nextScheduleType).toBe("4weeks");
  });

  it("新規作成はこれまで通り DEFAULT(1year) のまま", () => {
    const merged = mergeVaccinationFormData(false, undefined, {});
    expect(merged.nextScheduleType).toBe("1year");
  });

  it("修正後フォームの update request は next_schedule_type=other と保持された next_date を送る", () => {
    const merged = mergeVaccinationFormData(true, existingVaccination(), {});
    const req = buildUpdateVaccinationRequest(merged, 1);
    expect(req.next_schedule_type).toBe("other");
    expect(req.next_date).toBe(jstDateStartISOString("2026-08-29"));
  });
});

// EMR-431: 新規フォームは「当日＋既定間隔」の次回予定日を開いた時点で持つ。
// 日付に触れずに保存しても next_date が送られる（EMR-427 の LINE 期限配信は
// next_date で判定するため、空のまま保存されると配信対象から漏れる）。
describe("mergeVaccinationFormData — 新規フォームの次回予定日初期化（EMR-431）", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // UTC 16:00 = JST 翌日 01:00。JST 当日が 2026-10-15 になる瞬間を固定する。
    vi.setSystemTime(new Date("2026-10-14T15:01:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("既定間隔 1年後 で 当日+1年 の次回予定日を初期化する", () => {
    const merged = mergeVaccinationFormData(false, undefined, {});
    expect(merged.nextScheduleType).toBe("1year");
    expect(merged.date).toBe("2026-10-15");
    expect(merged.nextDate).toBe("2027-10-15");
  });

  it("日付に触れずに保存しても create request が next_date を送る", () => {
    const merged = mergeVaccinationFormData(false, undefined, {});
    const req = buildCreateVaccinationRequest(merged, "5");
    expect(req.next_date).toBe(jstDateStartISOString("2027-10-15"));
    expect(req.next_schedule_type).toBe("1year");
  });

  it("localOverrides の nextDate は初期値より優先される", () => {
    const merged = mergeVaccinationFormData(false, undefined, { nextDate: "2026-11-01" });
    expect(merged.nextDate).toBe("2026-11-01");
  });
});
