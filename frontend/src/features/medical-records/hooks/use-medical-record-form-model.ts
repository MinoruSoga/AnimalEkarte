import { formatJSTDate } from "@/lib/jst-date";

export { formatJSTDate };

export const DEFAULT_CHIEF_COMPLAINT = "# どんな症状\n\n# どこが\n\n# いつから\n\n# その他・備考\n\n# フリースペース";
/** 問診 notes 比較用。clinical_plan.treatment_policy の初期値には使わない（BUG-010）。 */
export const DEFAULT_TREATMENT_POLICY = "# 治療方針";
/**
 * 歴史的テンプレート固定文字列。BUG-010 以前は plan/assessment 初期値に使われ
 * 保存時にユーザー入力を上書きした。clinical_plan 3欄の初期値には使わない。
 */
export const DEFAULT_PLAN = "# 治療方針";
export const DEFAULT_ASSESSMENT = "# 診断詳細";

export const DEFAULT_RECEPTION_APPOINTMENT_MINUTES = 15;

const JST_OFFSET_MS = 9 * 60 * 60 * 1000;

export interface MedicalRecordReservationType {
  id: number;
  name: string;
  category: string;
  is_internal: boolean;
  sort_order: number;
  duration_minutes: number;
}

export interface MedicalRecordReservationTypeGroup {
  types: MedicalRecordReservationType[];
}

export function toVisitTypeValue(visitTypeLabel: string): "first" | "revisit" {
  return visitTypeLabel === "初診" || visitTypeLabel === "first" ? "first" : "revisit";
}

function padDatePart(value: number): string {
  return String(value).padStart(2, "0");
}

function formatJSTDateTime(date: Date): string {
  const jstDate = new Date(date.getTime() + JST_OFFSET_MS);
  return `${formatJSTDate(date)}T${padDatePart(jstDate.getUTCHours())}:${padDatePart(jstDate.getUTCMinutes())}:00+09:00`;
}

function createDateAtCurrentJSTTime(visitDate: string, timeSource: Date): Date {
  const jstTime = new Date(timeSource.getTime() + JST_OFFSET_MS);
  return new Date(
    `${visitDate}T${padDatePart(jstTime.getUTCHours())}:${padDatePart(jstTime.getUTCMinutes())}:00+09:00`
  );
}

export function createReceptionAppointmentTimeRange(durationMinutes: number, visitDate?: string) {
  const now = new Date();
  const start = visitDate ? createDateAtCurrentJSTTime(visitDate, now) : now;
  start.setSeconds(0, 0);
  const end = new Date(start);
  end.setMinutes(end.getMinutes() + durationMinutes);
  return {
    startTime: formatJSTDateTime(start),
    endTime: formatJSTDateTime(end),
  };
}

export function normalizeAppointmentId(value: unknown): string | undefined {
  if (typeof value === "string" && value !== "") return value;
  if (typeof value === "number" && Number.isFinite(value)) return String(value);
  return undefined;
}

export function normalizeVisitDate(value: unknown): string | undefined {
  if (typeof value !== "string") return undefined;
  return /^\d{4}-\d{2}-\d{2}$/.test(value) ? value : undefined;
}

export function findGeneralReservationType(
  groups: readonly MedicalRecordReservationTypeGroup[] | undefined,
  visitType: string
) {
  const generalTypes = groups
    ?.flatMap((group) => group.types)
    .filter((type) => type.category === "general" && !type.is_internal)
    .sort((a, b) => a.sort_order - b.sort_order);

  return generalTypes?.find((type) =>
    toVisitTypeValue(visitType) === "revisit"
      ? type.name.includes("再診")
      : !type.name.includes("再診")
  ) ?? generalTypes?.[0];
}
