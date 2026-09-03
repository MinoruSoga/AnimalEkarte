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

// status は backend の ReservationStatus enum (backend/internal/model/reservation.go) と1:1で一致させる。
// staff_name は指名なし予約で Doctor が nil のため backend が省略する（omitempty 相当）。
// notes は空文字時に backend が omitempty で省略するため未着信を許容する。
export const reservationSchema = z.object({
  id: z.number(),
  course_name: z.string(),
  pet_name: z.string().optional().default(""),
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
});

export const createReservationResponseSchema = z.object({
  id: z.number(),
  notes: z.string(),
});
