import type { InterviewHistoryItem, InterviewWeightMeasurement } from "../types";

/**
 * EMR-356 repair1 C2: 履歴各行の weightMeasurements 投影の集約結果。
 * - measurements: 行ローカル測定を行順に flatten（id 重複は1件に畳む）
 * - unavailable: 投影フィールドを持たない医療行が1行でもあれば true。
 *   trimming 等の医療行でない行（外部 href 持ち）は中立で数えない。
 *   正常取得した空履歴は measurements:[] / unavailable:false（測定なし確定）。
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
): PastWeightMeasurements {
  const measurements: InterviewWeightMeasurement[] = [];
  const seen = new Set<string>();
  let unavailable = false;
  for (const item of items) {
    if (item.weightMeasurements === undefined) {
      if (isMedicalHistoryRow(item)) unavailable = true;
      continue;
    }
    for (const m of item.weightMeasurements) {
      if (!seen.has(m.id)) {
        seen.add(m.id);
        measurements.push(m);
      }
    }
  }
  return { measurements, unavailable };
}
