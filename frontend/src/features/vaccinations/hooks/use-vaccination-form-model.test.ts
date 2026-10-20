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
  vaccinationOverridesOnScheduleType,
  vaccinationOverridesOnVaccineId,
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

  // EMR-430（2026-10-20 本人回答）: 保存済み type が無いレコードは 1year に
  // フォールバックせず空欄扱いから再判定する。保存済み nextDate は解決可能な
  // 間隔が無いため other に落ちるが、日付の保存値は上書きしない。
  it("nextScheduleType 未設定なら 1year にフォールバックせず other 再判定・nextDate は保持する", () => {
    const merged = mergeVaccinationFormData(
      true,
      existingVaccination({ nextScheduleType: undefined, nextDate: "2027-08-08" }),
      {},
    );
    expect(merged.nextScheduleType).toBe("other");
    expect(merged.nextDate).toBe("2027-08-08");
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

  // EMR-430（2026-10-20 本人回答）: 新規フォームの初期 type は未選択（空欄）。
  // 一律 1year を補完しない。
  it("新規作成は未選択（空欄）で開く", () => {
    const merged = mergeVaccinationFormData(false, undefined, {});
    expect(merged.nextScheduleType).toBe("");
  });

  it("修正後フォームの update request は next_schedule_type=other と保持された next_date を送る", () => {
    const merged = mergeVaccinationFormData(true, existingVaccination(), {});
    const req = buildUpdateVaccinationRequest(merged, 1);
    expect(req.next_schedule_type).toBe("other");
    expect(req.next_date).toBe(jstDateStartISOString("2026-08-29"));
  });
});

// EMR-430（2026-10-20 本人回答「ワクチン種類ごとの設定を使う。設定がなければ空欄。」）:
// 新規フォームは種別未選択のため次回予定を空欄のまま開く。旧 EMR-431 の
// 「当日＋既定間隔」自動入力は予防接種では撤回。種別選択時にマスタ interval に
// 従属して初めて計算し、設定なし種別では空欄のまま保存できる。
describe("mergeVaccinationFormData — 新規フォームの次回予定日初期化（EMR-430）", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // UTC 16:00 = JST 翌日 01:00。JST 当日が 2026-10-15 になる瞬間を固定する。
    vi.setSystemTime(new Date("2026-10-14T15:01:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("種別未選択のままでは type/date とも空欄で開く（一律間隔を補わない）", () => {
    const merged = mergeVaccinationFormData(false, undefined, {});
    expect(merged.nextScheduleType).toBe("");
    expect(merged.date).toBe("2026-10-15");
    expect(merged.nextDate).toBe("");
  });

  it("空欄のまま保存すると create request は next_date=null / next_schedule_type 未送信", () => {
    const merged = mergeVaccinationFormData(false, undefined, {});
    const req = buildCreateVaccinationRequest(merged, "5");
    expect(req.next_date).toBeNull();
    expect(req.next_schedule_type).toBeUndefined();
  });

  it("localOverrides の nextDate は初期値より優先される", () => {
    const merged = mergeVaccinationFormData(false, undefined, { nextDate: "2026-11-01" });
    expect(merged.nextDate).toBe("2026-11-01");
  });
});

// EMR-430（2026-10-20 本人回答）: 種別選択時の次回予定はマスタ interval に従属する。
describe("vaccinationOverridesOnVaccineId — 種別選択の interval 従属（EMR-430）", () => {
  it("interval 1年 の種別を選ぶと 1year で当日から再計算する", () => {
    const overrides = vaccinationOverridesOnVaccineId(
      { date: "2026-10-15", nextScheduleType: "", nextDate: "" },
      "3",
      "2026-10-15",
      "1年",
    );
    expect(overrides.nextScheduleType).toBe("1year");
    expect(overrides.nextDate).toBe("2027-10-15");
  });

  it("interval 1ヶ月 の種別を選ぶと 4weeks で当日から再計算する", () => {
    const overrides = vaccinationOverridesOnVaccineId(
      { date: "2026-10-15", nextScheduleType: "", nextDate: "" },
      "3",
      "2026-10-15",
      "1ヶ月",
    );
    expect(overrides.nextScheduleType).toBe("4weeks");
    expect(overrides.nextDate).toBe("2026-11-12");
  });

  it("マスタ取得済みで interval 未設定（空文字）の種別は空欄へ上書きする", () => {
    const overrides = vaccinationOverridesOnVaccineId(
      { date: "2026-10-15", nextScheduleType: "1year", nextDate: "2027-10-15" },
      "3",
      "2026-10-15",
      "",
    );
    expect(overrides.nextScheduleType).toBe("");
    expect(overrides.nextDate).toBe("");
  });

  it("interval 未解決（マスタ取得中・失敗・遅延の undefined）では既存の type/date を保持する", () => {
    const overrides = vaccinationOverridesOnVaccineId(
      { date: "2026-10-15", nextScheduleType: "1year", nextDate: "2027-10-15" },
      "3",
      "2026-10-15",
      undefined,
    );
    expect(overrides.nextScheduleType).toBe("1year");
    expect(overrides.nextDate).toBe("2027-10-15");
    expect(overrides.vaccineId).toBe("3");
  });
});

// レビュー指摘（EMR-431）: 「以外（手動）」へ切り替えたとき、EMR-431 が既定で
// 自動計算した次回予定日が残ったまま保存されていた。間隔の切替時は必ず再計算し、
// 計算できない場合（other・実施日空）はクリアして手入力の余地を残す。
describe("vaccinationOverridesOnScheduleType — other 切替で次回予定日をクリア（レビュー指摘）", () => {
  it("other に切り替えたら既定で入った次回予定日をクリアする", () => {
    const overrides = vaccinationOverridesOnScheduleType(
      { date: "2026-08-29", nextDate: "2027-08-29" },
      "other",
      "2026-08-29",
    );
    expect(overrides.nextScheduleType).toBe("other");
    expect(overrides.nextDate).toBe("");
  });

  it("プリセットへ切り替えたら実施日から再計算する（従来どおり）", () => {
    const overrides = vaccinationOverridesOnScheduleType(
      { date: "2026-08-29", nextDate: "" },
      "1year",
      "2026-08-29",
    );
    expect(overrides.nextScheduleType).toBe("1year");
    expect(overrides.nextDate).toBe("2027-08-29");
  });
});
