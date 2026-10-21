// FE5-18: line-reserve API 層の実行時検証（Zod）。
// フィールドの optional/nullable は ../types/models.ts の手書き型に厳密整合させる。
// 未知フィールド追加で壊れないよう .passthrough() は付けず Zod デフォルト（strip）とする。
import { z } from "zod";

export const liffSettingsSchema = z.object({
  liff_id: z.string(),
  header_text: z.string(),
  phone_number: z.string(),
  status: z.enum(["running", "stopped"]),
  request_example: z.string(),
  reservation_notice: z.string(),
  cancel_notice: z.string(),
  privacy_policy: z.string(),
  show_no_staff_option: z.boolean(),
  booking_window: z.number(),
});

const ownerPetSchema = z.object({
  id: z.number(),
  name: z.string(),
  animal_species: z.object({ name: z.string() }).optional(),
});

export const liffProfileSchema = z.object({
  line_user_id: z.string(),
  display_name: z.string(),
  additional_fields: z.object({
    name: z.string().optional(),
    phone: z.string().optional(),
    owner_name: z.string().optional(),
    pets: z
      .array(z.object({ name: z.string(), type: z.string(), is_new: z.boolean().optional() }))
      .optional(),
  }),
  owner: z
    .object({
      owner_name: z.string(),
      phone: z.string(),
      pets: z.array(ownerPetSchema).optional(),
    })
    .optional(),
});

export const courseSchema = z.object({
  id: z.number(),
  name: z.string(),
  short_name: z.string(),
  show_short_name: z.boolean(),
  duration_minutes: z.number(),
  reservation_comment: z.string(),
  reservation_image_url: z.string(),
  sort_order: z.number(),
  category: z.enum(["general", "trimming"]).optional(),
});

export const trimmingCourseSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string(),
  price: z.number().nullable(),
  sort_order: z.number(),
});

export const trimmingOptionSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string(),
  price: z.number().nullable(),
  sort_order: z.number(),
});

export const staffSchema = z.object({
  id: z.number(),
  name: z.string(),
  reservation_comment: z.string(),
  reservation_image_url: z.string(),
  sort_order: z.number(),
});

export const availableDateSchema = z.object({
  date: z.string(),
  available: z.boolean(),
  reason: z.enum(["closed", "holiday", "staff_off", "no_slots"]).optional(),
});

// GET /available-dates は { dates, window } のラッパー形状で返る（window は未使用のため unknown）
export const availableDatesResponseSchema = z.object({
  dates: z.array(availableDateSchema),
  window: z.unknown(),
});

// display_time は Backend 未実装のためオプショナル（models.ts コメント準拠）
export const availableTimeSchema = z.object({
  start_time: z.string(),
  end_time: z.string(),
  display_time: z.string().optional(),
});

// EMR-519: 予約 ID・pet_id は正の int64 範囲の正規形十進文字列。Number へ変換すると
// 2^53 超で precision を失うため、桁数 + 辞書順比較で上限を検査する
// （revisitIntakeDecimalId は範囲検査を持たないため別定義）。
const RESERVATION_ID_INT64_MAX = "9223372036854775807";
export const reservationDecimalId = z
  .string()
  .regex(/^[1-9][0-9]*$/)
  .refine(
    (v) =>
      v.length < RESERVATION_ID_INT64_MAX.length ||
      (v.length === RESERVATION_ID_INT64_MAX.length && v <= RESERVATION_ID_INT64_MAX),
  );

// status は backend の ReservationStatus enum (backend/internal/model/reservation.go) と1:1で一致させる。
// staff_name は指名なし予約で Doctor が nil のため backend が省略する（omitempty 相当）。
// notes は空文字時に backend が omitempty で省略するため未着信を許容する。
// pet_id / pet_name は「両方 string・両方 null・両方省略（旧応答）」の3形だけを受理し、
// parse 後は pet_id: string | null / pet_name: string | null に正規化する（両方省略は両方 null）。
export const reservationSchema = z
  .object({
    id: reservationDecimalId,
    pet_id: reservationDecimalId.nullable().optional(),
    pet_name: z.string().nullable().optional(),
    course_name: z.string(),
    staff_name: z.string().optional().default(""),
    date: z.string(),
    start_time: z.string(),
    end_time: z.string(),
    status: z.enum([
      "confirmed",
      "pending",
      "cancelled",
      "checked_in",
      "in_consultation",
      "accounting",
      "completed",
      "no_show",
    ]),
    notes: z.string().optional().default(""),
    created_at: z.string(),
  })
  .refine(
    (r) =>
      (r.pet_id === undefined) === (r.pet_name === undefined) &&
      (r.pet_id === null) === (r.pet_name === null),
  )
  .transform((r) => ({ ...r, pet_id: r.pet_id ?? null, pet_name: r.pet_name ?? null }));

export const createReservationResponseSchema = z.object({
  id: reservationDecimalId,
  notes: z.string(),
});

// ---------------------------------------------------------------------------
// EMR-474: 再診事前問診（backend/docs/api.yaml の RevisitIntake* schema）
// enum は api.yaml と同じ値集合。未知 code は補完・既定値化せず検証失敗（fail-closed）とする。
// ID・版・世代は正規形10進文字列のまま扱い Number へ変換しない。
// ---------------------------------------------------------------------------

const revisitIntakeDecimalId = z.string().regex(/^[1-9][0-9]*$/);
const revisitIntakeOperationKey = z
  .string()
  .regex(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/)
  .refine((v) => v !== "00000000-0000-0000-0000-000000000000");
const revisitIntakeFormVersion = z.number().int().min(1);

const RFC3339_DATE_TIME =
  /^(\d{4})-(\d{2})-(\d{2})[Tt](\d{2}):(\d{2}):(\d{2})(\.\d+)?(?:[Zz]|([+-])(\d{2}):(\d{2}))$/;

/**
 * RFC3339 date-time（api.yaml RevisitIntakeReceivedAt format: date-time）の検査。
 * 月・日（閏年を含む）・時・分・秒（閏秒 :60 可）・offset の範囲を確かめる。
 */
export function isRfc3339DateTime(value: string): boolean {
  const m = RFC3339_DATE_TIME.exec(value);
  if (!m) return false;
  const [year, month, day, hour, minute, second] = m.slice(1, 7).map(Number);
  if (month < 1 || month > 12) return false;
  const leap = (year % 4 === 0 && year % 100 !== 0) || year % 400 === 0;
  const daysInMonth = [31, leap ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31][month - 1];
  if (day < 1 || day > daysInMonth) return false;
  if (hour > 23 || minute > 59 || second > 60) return false;
  if (m[8] !== undefined) {
    if (Number(m[9]) > 23 || Number(m[10]) > 59) return false;
  }
  return true;
}

const revisitIntakeDateTime = z.string().refine(isRfc3339DateTime);

export const revisitIntakeEnergyAppetiteSchema = z.enum([
  "usual",
  "slightly_reduced",
  "markedly_reduced",
]);
export const revisitIntakeImprovementSchema = z.enum([
  "improved",
  "slightly_improved",
  "unchanged",
  "worse",
  "new_symptoms",
]);
export const revisitIntakeExaminationPreferenceSchema = z.enum([
  "consultation_only",
  "medication_only",
  "examination",
  "decide_with_vet",
]);

export const revisitIntakeFullAnswersSchema = z.object({
  visit_purpose: z.string().min(1),
  onset: z.string().nullable(),
  energy: revisitIntakeEnergyAppetiteSchema.nullable(),
  appetite: revisitIntakeEnergyAppetiteSchema.nullable(),
  current_symptoms: z.string().nullable(),
  improvement: revisitIntakeImprovementSchema.nullable(),
  examination_preference: revisitIntakeExaminationPreferenceSchema.nullable(),
  main_concern: z.string().nullable(),
});

export const revisitIntakeTargetRefSchema = z.object({
  target_id: revisitIntakeDecimalId.nullable(),
  appointment_id: revisitIntakeDecimalId,
  pet_id: revisitIntakeDecimalId,
  target_generation: revisitIntakeDecimalId,
});

const revisitIntakePersistedTargetRefSchema = revisitIntakeTargetRefSchema.extend({
  target_id: revisitIntakeDecimalId,
});

export const revisitIntakeAnswerSchema = z.object({
  answer_id: revisitIntakeDecimalId,
  revision: revisitIntakeDecimalId,
  form_definition_version: revisitIntakeFormVersion,
  received_at: revisitIntakeDateTime,
  provenance: z.literal("owner_submission"),
  answers: revisitIntakeFullAnswersSchema,
});

export const revisitIntakeOwnerCurrentTargetSchema = z
  .object({
    ref: revisitIntakeTargetRefSchema,
    head_revision: revisitIntakeDecimalId.nullable(),
    answer_state: z.enum(["unanswered", "answered"]),
    latest_answer: revisitIntakeAnswerSchema.nullable(),
    content_available: z.boolean(),
    editable: z.boolean(),
    edit_block_reason: z
      .enum([
        "reception_started",
        "cancelled",
        "deleted",
        "target_unresolved",
        "authorization_denied",
        "unsupported_lifecycle",
      ])
      .nullable(),
  })
  // RevisitIntakeCurrentAnswerState の3形（未回答 / 可視head / 不可視head）の整合を検査する。
  .refine((t) => {
    if (t.answer_state === "unanswered") {
      return t.head_revision === null && t.latest_answer === null && !t.content_available;
    }
    if (t.head_revision === null || t.ref.target_id === null) return false;
    // 可視 head の本文は当該 head の Answer でなければならない（旧回答を最新として扱わない）。
    if (t.content_available) {
      return t.latest_answer !== null && t.latest_answer.revision === t.head_revision;
    }
    return t.latest_answer === null;
  });

export const revisitIntakeHistoryEntrySchema = z.object({
  target: revisitIntakePersistedTargetRefSchema,
  answer: revisitIntakeAnswerSchema,
  target_relation: z.enum(["current", "previous"]),
});

export const revisitIntakeLIFFReadSchema = z.object({
  availability: z.enum(["available", "target_unresolved"]),
  form_definition_version: revisitIntakeFormVersion,
  current_target: revisitIntakeOwnerCurrentTargetSchema.nullable(),
  history: z.array(revisitIntakeHistoryEntrySchema),
  next_cursor: z.string().nullable(),
});

const revisitIntakeIntent = z.literal("revisit-intake.submit");

const revisitIntakeVisibleSuccessSchema = z.object({
  operation_result: z.object({
    operation_id: revisitIntakeOperationKey,
    intent: revisitIntakeIntent,
    state: z.literal("succeeded"),
    committed_at: revisitIntakeDateTime,
    result_refs: z.object({
      appointment_id: revisitIntakeDecimalId,
      pet_id: revisitIntakeDecimalId,
      target_id: revisitIntakeDecimalId,
      answer_id: revisitIntakeDecimalId,
    }),
    committed_versions: z.object({
      target_generation: revisitIntakeDecimalId,
      answer_revision: revisitIntakeDecimalId,
    }),
  }),
  current_state: revisitIntakeOwnerCurrentTargetSchema.nullable(),
  content_available: z.literal(true),
});

const revisitIntakeRedactedSuccessSchema = z.object({
  operation_result: z.object({
    operation_id: revisitIntakeOperationKey,
    intent: revisitIntakeIntent,
    state: z.literal("succeeded"),
    committed_at: revisitIntakeDateTime,
    result_refs: z.null(),
    committed_versions: z.null(),
  }),
  current_state: z.null(),
  content_available: z.literal(false),
});

const revisitIntakeUnresolvedSchema = z.object({
  operation_result: z.object({
    operation_id: revisitIntakeOperationKey,
    intent: revisitIntakeIntent,
    state: z.enum(["claimed_uncommitted", "result_unavailable"]),
  }),
  current_state: z.null(),
  content_available: z.literal(false),
});

export const revisitIntakeResultResponseSchema = z.union([
  revisitIntakeVisibleSuccessSchema,
  revisitIntakeRedactedSuccessSchema,
  revisitIntakeUnresolvedSchema,
]);

export const revisitIntakeLookupNotFoundSchema = z.object({
  error: z.literal("not found"),
  code: z.literal("NOT_FOUND"),
  operation_result: z.object({
    operation_id: revisitIntakeOperationKey,
    intent: revisitIntakeIntent,
    state: z.literal("not_found"),
  }),
  current_state: z.null(),
  content_available: z.literal(false),
});

/**
 * RevisitIntakeError（DomainError | MiddlewareError）。transport では code/details が省略され得るため
 * どちらも optional。error 文言では分岐しない（HTTP status と code だけで分岐する）。
 */
export const revisitIntakeErrorBodySchema = z.object({
  error: z.string().optional(),
  code: z.string().optional(),
  details: z.unknown().optional(),
});

export const revisitIntakeInvalidInputDetailsSchema = z
  .object({
    fields: z.array(z.object({ path: z.string(), reason: z.string() })),
  })
  .nullable();

export const revisitIntakeRevisionConflictDetailsSchema = z.object({
  head_revision: revisitIntakeDecimalId.nullable(),
});

export const revisitIntakeClosedDetailsSchema = z.object({
  reason: z.enum(["reception_started", "cancelled", "deleted", "unsupported_lifecycle"]),
});
