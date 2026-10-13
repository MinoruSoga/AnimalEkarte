import { formatJSTDate, jstDateStartISOString } from "@/lib/jst-date";
import type { InterviewHistoryItem, InterviewWeightMeasurement } from "../types";

/**
 * EMR-356 repair1 C2: 履歴各行の weightMeasurements 投影の集約結果。
 * - measurements: 選択カルテ時点より厳密に前の行ローカル測定（id 重複は1件に畳む）
 * - unavailable: 投影フィールドを持たない医療行が1行でもあれば true。
 *   trimming 等の医療行でない行（外部 href 持ち）は中立で数えない。
 *   正常取得・有効な境界を持つ空履歴のみ unavailable:false（測定なし確定）。
 */
export interface PastWeightMeasurements {
  measurements: InterviewWeightMeasurement[];
  unavailable: boolean;
}

/**
 * 医療行かどうかの判定。InterviewHistoryItem は href 未指定が医療行
 * （/medical-records/:id へフォールバック）、trimming 行は /trimming/:id を
 * 持つ。将来明示 href を持つ医療行が増えても /medical-records/ 前方一致で拾う。
 */
export function isMedicalHistoryRow(item: InterviewHistoryItem): boolean {
  return item.href === undefined || item.href.startsWith("/medical-records/");
}

/**
 * 履歴各行の weightMeasurements 投影を集約する。
 * 「未提供(undefined)」と「明示的な空([])」を区別するため、逐次 `?? []` ではなく
 * この helper を通す（ClinicalTabs の C2 契約）。
 */
export function collectPastWeightMeasurements(
  items: InterviewHistoryItem[],
  boundary?: Date,
): PastWeightMeasurements {
  const measurements: InterviewWeightMeasurement[] = [];
  const seen = new Set<string>();
  const boundaryTime = boundary?.getTime();
  if (boundaryTime === undefined || !Number.isFinite(boundaryTime)) {
    return { measurements: [], unavailable: true };
  }
  let unavailable = false;
  for (const item of items) {
    if (item.weightMeasurements === undefined) {
      if (isMedicalHistoryRow(item)) unavailable = true;
      continue;
    }
    for (const m of item.weightMeasurements) {
      const measuredTime = Date.parse(m.measuredAt);
      if (!/T.*(?:Z|[+-]\d{2}:\d{2})$/.test(m.measuredAt) || !Number.isFinite(measuredTime)) {
        unavailable = true;
        continue;
      }
      if (measuredTime < boundaryTime && !seen.has(m.id)) {
        seen.add(m.id);
        measurements.push(m);
      }
    }
  }
  return { measurements, unavailable };
}

/** medical_records.date is a calendar date; RFC3339 serialization does not establish a visit time. */
export function resolvePastWeightReferenceDate(input: {
  recordDateIso?: string;
  recordDateJst?: string;
}): Date | undefined {
  for (const value of [input.recordDateIso, input.recordDateJst]) {
    if (!value) continue;
    const dateOnly = /^\d{4}-\d{2}-\d{2}$/.test(value);
    // Clinical instants must identify their timezone rather than use the host clock.
    if (!dateOnly && !/T.*(?:Z|[+-]\d{2}:\d{2})$/.test(value)) continue;
    const date = new Date(dateOnly ? jstDateStartISOString(value) : value);
    if (dateOnly && formatJSTDate(date) !== value) continue;
    if (Number.isFinite(date.getTime()))
      return new Date(jstDateStartISOString(formatJSTDate(date)));
  }
  return undefined;
}
