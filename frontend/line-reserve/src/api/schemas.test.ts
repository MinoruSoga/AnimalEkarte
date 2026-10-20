import { describe, it, expect } from "vitest";
import {
  liffSettingsSchema,
  liffProfileSchema,
  courseSchema,
  trimmingCourseSchema,
  trimmingOptionSchema,
  staffSchema,
  availableDateSchema,
  availableDatesResponseSchema,
  availableTimeSchema,
  reservationSchema,
  createReservationResponseSchema,
  revisitIntakeLIFFReadSchema,
  revisitIntakeResultResponseSchema,
  revisitIntakeEnergyAppetiteSchema,
  revisitIntakeImprovementSchema,
  revisitIntakeExaminationPreferenceSchema,
} from "./schemas";

describe("line-reserve schemas（FE5-18: 実行時検証）", () => {
  it("liffSettingsSchema: models.ts どおりの fixture が通る", () => {
    const fixture = {
      liff_id: "123-abc",
      header_text: "ノア動物病院",
      phone_number: "03-1234-5678",
      status: "running",
      request_example: "",
      reservation_notice: "",
      cancel_notice: "",
      privacy_policy: "",
      show_no_staff_option: true,
      booking_window: 30,
    };
    expect(liffSettingsSchema.safeParse(fixture).success).toBe(true);
  });

  it("liffSettingsSchema: 必須フィールド欠落が reject される", () => {
    const { liff_id: _omit, ...rest } = {
      liff_id: "123-abc",
      header_text: "",
      phone_number: "",
      status: "running",
      request_example: "",
      reservation_notice: "",
      cancel_notice: "",
      privacy_policy: "",
      show_no_staff_option: true,
      booking_window: 30,
    };
    expect(liffSettingsSchema.safeParse(rest).success).toBe(false);
  });

  it("liffProfileSchema: models.ts どおりの fixture が通る（owner なし）", () => {
    const fixture = {
      line_user_id: "U123",
      display_name: "山田太郎",
      additional_fields: { name: "山田太郎", pets: [{ name: "ポチ", type: "柴犬" }] },
    };
    expect(liffProfileSchema.safeParse(fixture).success).toBe(true);
  });

  it("liffProfileSchema: 必須フィールド欠落が reject される", () => {
    const fixture = { line_user_id: "U123" };
    expect(liffProfileSchema.safeParse(fixture).success).toBe(false);
  });

  it("courseSchema: models.ts どおりの fixture が通る（category なし）", () => {
    const fixture = {
      id: 1,
      name: "一般診察",
      short_name: "一般",
      show_short_name: false,
      duration_minutes: 30,
      reservation_comment: "",
      reservation_image_url: "",
      sort_order: 1,
    };
    expect(courseSchema.safeParse(fixture).success).toBe(true);
  });

  it("courseSchema: 必須フィールド欠落が reject される", () => {
    const fixture = { id: 1, name: "一般診察" };
    expect(courseSchema.safeParse(fixture).success).toBe(false);
  });

  it("trimmingCourseSchema: models.ts どおりの fixture が通る（price null）", () => {
    const fixture = { id: 1, name: "フルコース", description: "", price: null, sort_order: 1 };
    expect(trimmingCourseSchema.safeParse(fixture).success).toBe(true);
  });

  it("trimmingCourseSchema: 必須フィールド欠落が reject される", () => {
    const fixture = { id: 1, name: "フルコース" };
    expect(trimmingCourseSchema.safeParse(fixture).success).toBe(false);
  });

  it("trimmingOptionSchema: models.ts どおりの fixture が通る", () => {
    const fixture = { id: 1, name: "爪切り", description: "", price: 500, sort_order: 1 };
    expect(trimmingOptionSchema.safeParse(fixture).success).toBe(true);
  });

  it("trimmingOptionSchema: 必須フィールド欠落が reject される", () => {
    const fixture = { id: 1, name: "爪切り" };
    expect(trimmingOptionSchema.safeParse(fixture).success).toBe(false);
  });

  it("staffSchema: models.ts どおりの fixture が通る", () => {
    const fixture = {
      id: 1,
      name: "佐藤",
      reservation_comment: "",
      reservation_image_url: "",
      sort_order: 1,
    };
    expect(staffSchema.safeParse(fixture).success).toBe(true);
  });

  it("staffSchema: 必須フィールド欠落が reject される", () => {
    const fixture = { id: 1 };
    expect(staffSchema.safeParse(fixture).success).toBe(false);
  });

  it("availableDatesResponseSchema: models.ts どおりの fixture が通る（reason なし）", () => {
    const fixture = { dates: [{ date: "2026-08-01", available: true }], window: { days: 30 } };
    expect(availableDatesResponseSchema.safeParse(fixture).success).toBe(true);
  });

  it("availableDateSchema: 必須フィールド欠落が reject される", () => {
    const fixture = { date: "2026-08-01" };
    expect(availableDateSchema.safeParse(fixture).success).toBe(false);
  });

  it("availableTimeSchema: models.ts どおりの fixture が通る（display_time なし）", () => {
    const fixture = { start_time: "1000", end_time: "1030" };
    expect(availableTimeSchema.safeParse(fixture).success).toBe(true);
  });

  it("availableTimeSchema: 必須フィールド欠落が reject される", () => {
    const fixture = { start_time: "1000" };
    expect(availableTimeSchema.safeParse(fixture).success).toBe(false);
  });

  it("reservationSchema: models.ts どおりの fixture が通る", () => {
    const fixture = {
      id: 1,
      course_name: "一般診察",
      staff_name: "佐藤",
      date: "2026-08-01",
      start_time: "10:00",
      end_time: "10:30",
      status: "confirmed",
      notes: "",
      created_at: "2026-07-01T00:00:00Z",
    };
    expect(reservationSchema.safeParse(fixture).success).toBe(true);
  });

  it("reservationSchema: 必須フィールド欠落が reject される", () => {
    const fixture = { id: 1, course_name: "一般診察" };
    expect(reservationSchema.safeParse(fixture).success).toBe(false);
  });

  it("reservationSchema: 指名なし予約（staff_name 省略）が通り '' デフォルトになる", () => {
    const { staff_name: _omit, ...fixture } = {
      id: 1,
      staff_name: "佐藤",
      course_name: "一般診察",
      date: "2026-08-01",
      start_time: "10:00",
      end_time: "10:30",
      status: "confirmed",
      notes: "",
      created_at: "2026-07-01T00:00:00Z",
    };
    const result = reservationSchema.safeParse(fixture);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.staff_name).toBe("");
    }
  });

  it("reservationSchema: notes 省略が通り '' デフォルトになる", () => {
    const { notes: _omit, ...fixture } = {
      id: 1,
      course_name: "一般診察",
      staff_name: "佐藤",
      date: "2026-08-01",
      start_time: "10:00",
      end_time: "10:30",
      status: "confirmed",
      notes: "",
      created_at: "2026-07-01T00:00:00Z",
    };
    const result = reservationSchema.safeParse(fixture);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.notes).toBe("");
    }
  });

  it.each([
    "confirmed",
    "pending",
    "cancelled",
    "checked_in",
    "in_consultation",
    "accounting",
    "completed",
    "no_show",
  ])("reservationSchema: backend ReservationStatus 値 %s を許容する", (status) => {
    const fixture = {
      id: 1,
      course_name: "一般診察",
      staff_name: "佐藤",
      date: "2026-08-01",
      start_time: "10:00",
      end_time: "10:30",
      status,
      notes: "",
      created_at: "2026-07-01T00:00:00Z",
    };
    expect(reservationSchema.safeParse(fixture).success).toBe(true);
  });

  it("reservationSchema: backend ReservationStatus に存在しない値は reject される", () => {
    const fixture = {
      id: 1,
      course_name: "一般診察",
      staff_name: "佐藤",
      date: "2026-08-01",
      start_time: "10:00",
      end_time: "10:30",
      status: "unknown_status",
      notes: "",
      created_at: "2026-07-01T00:00:00Z",
    };
    expect(reservationSchema.safeParse(fixture).success).toBe(false);
  });

  it("createReservationResponseSchema: models.ts どおりの fixture が通る", () => {
    const fixture = { id: 1, notes: "R-20260801-0001" };
    expect(createReservationResponseSchema.safeParse(fixture).success).toBe(true);
  });

  it("createReservationResponseSchema: 必須フィールド欠落が reject される", () => {
    const fixture = { id: 1 };
    expect(createReservationResponseSchema.safeParse(fixture).success).toBe(false);
  });
});

describe("revisit intake schemas（EMR-474）", () => {
  const unanswered = {
    availability: "available",
    form_definition_version: 1,
    current_target: {
      ref: { target_id: null, appointment_id: "12001", pet_id: "801", target_generation: "3" },
      head_revision: null,
      answer_state: "unanswered",
      latest_answer: null,
      content_available: false,
      editable: true,
      edit_block_reason: null,
    },
    history: [],
    next_cursor: null,
  };

  it("api.yaml の unpersisted 例が通る", () => {
    expect(revisitIntakeLIFFReadSchema.safeParse(unanswered).success).toBe(true);
  });

  it("ID を Number で返す応答は reject する（10進文字列のみ）", () => {
    const bad = {
      ...unanswered,
      current_target: {
        ...unanswered.current_target,
        ref: { ...unanswered.current_target.ref, appointment_id: 12001 },
      },
    };
    expect(revisitIntakeLIFFReadSchema.safeParse(bad).success).toBe(false);
  });

  it("未知の edit_block_reason は補完せず reject する", () => {
    const bad = {
      ...unanswered,
      current_target: { ...unanswered.current_target, edit_block_reason: "staff_locked" },
    };
    expect(revisitIntakeLIFFReadSchema.safeParse(bad).success).toBe(false);
  });

  it("未回答なのに head_revision がある不整合は reject する", () => {
    const bad = {
      ...unanswered,
      current_target: { ...unanswered.current_target, head_revision: "1" },
    };
    expect(revisitIntakeLIFFReadSchema.safeParse(bad).success).toBe(false);
  });

  it("可視 head の latest_answer.revision が head_revision と一致しない応答は reject する", () => {
    const ans = {
      answer_id: "9101",
      revision: "1",
      form_definition_version: 1,
      received_at: "2026-10-20T13:00:00Z",
      provenance: "owner_submission",
      answers: {
        visit_purpose: "合成目的",
        onset: null,
        energy: null,
        appetite: null,
        current_symptoms: null,
        improvement: null,
        examination_preference: null,
        main_concern: null,
      },
    };
    const target = {
      ...unanswered.current_target,
      ref: { ...unanswered.current_target.ref, target_id: "5401" },
      answer_state: "answered",
      latest_answer: ans,
      content_available: true,
    };
    const matched = { ...unanswered, current_target: { ...target, head_revision: "1" } };
    const mismatched = { ...unanswered, current_target: { ...target, head_revision: "2" } };
    expect(revisitIntakeLIFFReadSchema.safeParse(matched).success).toBe(true);
    expect(revisitIntakeLIFFReadSchema.safeParse(mismatched).success).toBe(false);
  });

  it("未知の選択肢 code は reject する", () => {
    expect(revisitIntakeEnergyAppetiteSchema.safeParse("fine").success).toBe(false);
    expect(revisitIntakeImprovementSchema.safeParse("new_symptoms").success).toBe(true);
    expect(revisitIntakeExaminationPreferenceSchema.safeParse("examination").success).toBe(true);
  });

  it("結果 union: redacted 成功と claimed / result_unavailable が通り、未知 state は reject", () => {
    const operationId = "00000000-0000-4000-8000-000000000001";
    const redacted = {
      operation_result: {
        operation_id: operationId,
        intent: "revisit-intake.submit",
        state: "succeeded",
        committed_at: "2026-10-20T13:00:00Z",
        result_refs: null,
        committed_versions: null,
      },
      current_state: null,
      content_available: false,
    };
    const claimed = {
      operation_result: {
        operation_id: operationId,
        intent: "revisit-intake.submit",
        state: "claimed_uncommitted",
      },
      current_state: null,
      content_available: false,
    };
    expect(revisitIntakeResultResponseSchema.safeParse(redacted).success).toBe(true);
    expect(revisitIntakeResultResponseSchema.safeParse(claimed).success).toBe(true);
    expect(
      revisitIntakeResultResponseSchema.safeParse({
        ...claimed,
        operation_result: { ...claimed.operation_result, state: "failed" },
      }).success,
    ).toBe(false);
  });
});
