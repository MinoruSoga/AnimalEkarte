import { formatJSTDate, formatJSTTime } from "@/lib/jst-date";

export { formatJSTDate };
export const DEFAULT_CHIEF_COMPLAINT =
  "# どんな症状\n\n# どこが\n\n# いつから\n\n# その他・備考\n\n# フリースペース";
/** 問診 notes 比較用。clinical_plan.treatment_policy の初期値には使わない（BUG-010）。 */
export const DEFAULT_TREATMENT_POLICY = "# 治療方針";

export const DEFAULT_RECEPTION_APPOINTMENT_MINUTES = 15;

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

export function isSupportedVisitTypeLabel(visitTypeLabel: string): boolean {
  return (
    visitTypeLabel === "初診" ||
    visitTypeLabel === "first" ||
    visitTypeLabel === "再診" ||
    visitTypeLabel === "revisit"
  );
}

/** Maps UI/API labels to the BE visit_type enum. Unknown labels must not become revisit. */
export function toVisitTypeValue(visitTypeLabel: string): "first" | "revisit" | null {
  if (visitTypeLabel === "初診" || visitTypeLabel === "first") return "first";
  if (visitTypeLabel === "再診" || visitTypeLabel === "revisit") return "revisit";
  return null;
}

// FE-RC-027: JST の時刻計算は lib/jst-date.ts の共通ヘルパーに委譲する
// （旧実装は JST_OFFSET_MS + padDatePart をこのファイルで再実装していた重複）。
function formatJSTDateTime(date: Date): string {
  return `${formatJSTDate(date)}T${formatJSTTime(date)}:00+09:00`;
}

// buildJSTWallDateTime は「ローカル tz = JST」前提のローカル Date 構築契約
// （DatePicker 系）のため、実行環境の tz に依存せず絶対時刻を保証する必要がある
// ここでは使わない。formatJSTTime の +09:00 オフセット計算のみ再利用する。
function createDateAtCurrentJSTTime(visitDate: string, timeSource: Date): Date {
  return new Date(`${visitDate}T${formatJSTTime(timeSource)}:00+09:00`);
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
  visitType: string,
) {
  const generalTypes = groups
    ?.flatMap((group) => group.types)
    .filter((type) => type.category === "general" && !type.is_internal)
    .sort((a, b) => a.sort_order - b.sort_order);

  return (
    generalTypes?.find((type) =>
      toVisitTypeValue(visitType) !== "first"
        ? type.name.includes("再診")
        : !type.name.includes("再診"),
    ) ?? generalTypes?.[0]
  );
}
