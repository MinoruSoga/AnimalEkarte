import type { CreateMedicalRecordRequest } from "@/features/medical-records";
import type { CreateReservationRequest } from "@/hooks/use-create-reservation";

import { SYNTHETIC_CREATED_AT, SYNTHETIC_IDS } from "./ui-design-clinical-constants";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function assertExactKeys(body: Record<string, unknown>, expectedKeys: readonly string[]): void {
  const actual = Object.keys(body).sort();
  const expected = [...expectedKeys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) {
    throw new Error("request body fields did not match the synthetic contract");
  }
}

export function validateReservationCreate(body: unknown): void {
  if (!isRecord(body)) throw new Error("reservation body must be an object");
  assertExactKeys(body, [
    "pet_id",
    "owner_id",
    "start_time",
    "end_time",
    "visit_type",
    "reservation_type_id",
    "status",
    "source",
    "reservation_route",
  ]);
  const expected = {
    pet_id: SYNTHETIC_IDS.pet,
    owner_id: SYNTHETIC_IDS.owner,
    visit_type: "revisit",
    reservation_type_id: SYNTHETIC_IDS.reservationType,
    status: "in_consultation",
    source: "manual",
    reservation_route: "record_shortcut",
  } satisfies Pick<
    CreateReservationRequest,
    | "pet_id"
    | "owner_id"
    | "visit_type"
    | "reservation_type_id"
    | "status"
    | "source"
    | "reservation_route"
  >;
  for (const [key, value] of Object.entries(expected)) {
    if (body[key] !== value) throw new Error(`reservation body mismatch: ${key}`);
  }
  const startTime = typeof body.start_time === "string" ? Date.parse(body.start_time) : Number.NaN;
  const endTime = typeof body.end_time === "string" ? Date.parse(body.end_time) : Number.NaN;
  if (!Number.isFinite(startTime) || !Number.isFinite(endTime)) {
    throw new Error("reservation body mismatch: appointment time format");
  }
  if (endTime - startTime !== 30 * 60 * 1000) {
    throw new Error("reservation body mismatch: appointment duration");
  }
  const jstDate = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Tokyo",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
  if (jstDate.format(new Date(startTime)) !== SYNTHETIC_CREATED_AT.slice(0, 10)) {
    throw new Error("reservation body mismatch: appointment date");
  }
}

export function validateMedicalRecordCreate(body: unknown): void {
  if (!isRecord(body)) throw new Error("medical record body must be an object");
  assertExactKeys(body, [
    "pet_id",
    "owner_id",
    "visit_date",
    "visit_type",
    "appointment_id",
    "status",
    "recommendation_reason",
  ]);
  const expected = {
    pet_id: String(SYNTHETIC_IDS.pet),
    owner_id: String(SYNTHETIC_IDS.owner),
    visit_date: SYNTHETIC_CREATED_AT.slice(0, 10),
    visit_type: "revisit",
    appointment_id: String(SYNTHETIC_IDS.reservation),
    status: "draft",
    recommendation_reason: "",
  } satisfies Pick<
    CreateMedicalRecordRequest,
    | "pet_id"
    | "owner_id"
    | "visit_date"
    | "visit_type"
    | "appointment_id"
    | "status"
    | "recommendation_reason"
  >;
  for (const [key, value] of Object.entries(expected)) {
    if (body[key] !== value) throw new Error(`medical record body mismatch: ${key}`);
  }
}
