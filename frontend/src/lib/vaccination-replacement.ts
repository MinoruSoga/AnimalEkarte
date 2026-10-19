/**
 * 予防接種履歴の「置き換わり」判定（EMR-428）。
 *
 * 同じペット・同じ vaccineId で接種日がより後の記録がある記録は、ワクチンを
 * 接種し直したあとの古い記録なので期限警告を出さない（カードでは日付だけ残す）。
 * 同日に2件あればどちらも置き換わらない（厳密に「より後」のみ）。
 * vaccineName が「他院」で始まる記録と vaccineId "0"（汎用名・他院記録）は
 * 同じワクチンとみなさず、今どおり記録ごとに判定する（EMR-418 Q5）。
 *
 * 汎用名の判定規則（前方一致「他院」）は EMR-427 と同じ。呼び出し側は
 * VaccinationForm.tsx の絞り込み前の配列（同じペットの記録・接種日の新しい順）を渡す。
 * 新しい順に取得するため、置き換わり判定はページ内だけで正しく成立する。
 */

/** 置き換わり判定に必要な記録の最小フィールド。feature 間 import を避けるため構造で受ける */
export interface VaccinationReplacementRecord {
  id: string;
  petId: string;
  /** 接種日（YYYY-MM-DD または ISO 日時） */
  date: string;
  vaccineId: string;
  vaccineName: string;
}

/** 汎用名（他院で接種したことを示す名称）かどうか。前方一致「他院」。 */
export function isExternalClinicVaccinationName(vaccineName: string): boolean {
  return vaccineName.startsWith("他院");
}

/** 日付文字列を日単位のタイムスタンプへ正規化する（ISO 日時の時刻部分を無視） */
function toDay(dateString: string): number | null {
  if (!dateString) return null;
  const time = Date.parse(`${dateString.slice(0, 10)}T00:00:00Z`);
  return Number.isNaN(time) ? null : time;
}

/**
 * 置き換わった記録の id 集合を返す。
 * 汎用名（「他院」前方一致）と vaccineId "0" の記録は判定対象から除外する
 * （同じワクチンとみなさないため、置き換わる側にも置き換える側にもならない）。
 */
export function collectReplacedVaccinationIds(
  records: readonly VaccinationReplacementRecord[],
): Set<string> {
  const candidates = records.filter(
    (record) => record.vaccineId !== "0" && !isExternalClinicVaccinationName(record.vaccineName),
  );

  const replaced = new Set<string>();
  for (const record of candidates) {
    const recordDay = toDay(record.date);
    if (recordDay === null) continue;

    const isSuperseded = candidates.some((other) => {
      if (other.id === record.id) return false;
      if (other.petId !== record.petId) return false;
      if (other.vaccineId !== record.vaccineId) return false;
      const otherDay = toDay(other.date);
      return otherDay !== null && otherDay - recordDay > 0;
    });
    if (isSuperseded) replaced.add(record.id);
  }
  return replaced;
}
