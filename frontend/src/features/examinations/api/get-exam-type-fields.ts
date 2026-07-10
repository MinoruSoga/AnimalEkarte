import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ExaminationType } from "@/types/generated/models";

/**
 * 検査種別の項目テンプレ（exam_type_fields）を表す行。
 * normal_value から ref_min/ref_max を抽出する責務はここで完結する（FE 側のパースのみ。判定は backend）。
 */
export interface ExamTypeFieldRow {
  id: number;
  name: string;
  unit: string;
  normalValue: string;
  /** normal_value から抽出した下限。抽出できない or 上限のみの表記なら undefined。 */
  refMin?: number;
  /** normal_value から抽出した上限。抽出できない or 下限のみの表記なら undefined。 */
  refMax?: number;
  sortOrder: number;
}

// 符号付き小数 / 整数を捉えるサブパターン。各分岐で再利用する。
const NUM = String.raw`(-?\d+(?:\.\d+)?)`;

// 範囲表記 (両端): "3.5-7.0", "3.5 - 7.0", "3.5〜7.0", "3.5~7.0"
const RANGE_RE = new RegExp(`^${NUM}\\s*[-~〜]\\s*${NUM}$`);
// 日本語両端: "3.5以上7.0以下" (スペース許容)
const BOTH_JP_RE = new RegExp(`^${NUM}\\s*以上\\s*${NUM}\\s*以下$`);
// 下限のみ: "3.5以上"
const MIN_JP_RE = new RegExp(`^${NUM}\\s*以上$`);
// 上限のみ: "7.0以下"
const MAX_JP_RE = new RegExp(`^${NUM}\\s*以下$`);
// 上限のみ (不等号): "<7.0", "＜7.0"
const LT_RE = new RegExp(`^[<＜]\\s*${NUM}$`);
// 下限のみ (不等号): ">3.5", "＞3.5"
const GT_RE = new RegExp(`^[>＞]\\s*${NUM}$`);

function toFiniteNumber(raw: string | undefined): number | undefined {
  if (raw === undefined) return undefined;
  const n = Number(raw);
  return Number.isFinite(n) ? n : undefined;
}

/**
 * normal_value 文字列から ref_min / ref_max を抽出する。
 *
 * 対応表記:
 *  - 範囲: "3.5-7.0", "3.5 - 7.0", "3.5〜7.0", "3.5~7.0" → refMin / refMax 両方
 *  - 日本語両端: "3.5以上7.0以下" → refMin / refMax 両方
 *  - 下限のみ: "3.5以上", ">3.5", "＞3.5" → refMin のみ (refMax は undefined)
 *  - 上限のみ: "7.0以下", "<7.0", "＜7.0" → refMax のみ (refMin は undefined)
 *  - 抽出できない形式は `{}` を返す（判定は backend に委ねる）
 *
 * 注: これは「テンプレ → 判定基準値の抽出」であって、異常値の判定そのものではない。
 *     判定 (low/high/normal) は backend service 層 (computeExamResultStatus) で行われる。
 *     片側のみ抽出された場合でも backend は対応する辺だけで判定する。
 */
export function parseNormalRange(normalValue: string): { refMin?: number; refMax?: number } {
  if (!normalValue) return {};
  const s = normalValue.trim();
  if (!s) return {};

  const range = s.match(RANGE_RE);
  if (range) {
    const lo = toFiniteNumber(range[1]);
    const hi = toFiniteNumber(range[2]);
    if (lo !== undefined && hi !== undefined) return { refMin: lo, refMax: hi };
    return {};
  }

  const both = s.match(BOTH_JP_RE);
  if (both) {
    const lo = toFiniteNumber(both[1]);
    const hi = toFiniteNumber(both[2]);
    if (lo !== undefined && hi !== undefined) return { refMin: lo, refMax: hi };
    return {};
  }

  const minJp = s.match(MIN_JP_RE);
  if (minJp) {
    const lo = toFiniteNumber(minJp[1]);
    return lo !== undefined ? { refMin: lo } : {};
  }

  const maxJp = s.match(MAX_JP_RE);
  if (maxJp) {
    const hi = toFiniteNumber(maxJp[1]);
    return hi !== undefined ? { refMax: hi } : {};
  }

  const lt = s.match(LT_RE);
  if (lt) {
    const hi = toFiniteNumber(lt[1]);
    return hi !== undefined ? { refMax: hi } : {};
  }

  const gt = s.match(GT_RE);
  if (gt) {
    const lo = toFiniteNumber(gt[1]);
    return lo !== undefined ? { refMin: lo } : {};
  }

  return {};
}

/**
 * GET /v1/masters/examination-types/:id — 検査種別の詳細（items=exam_type_fields 含む）を取得する。
 * 検査フォームで「検査項目テーブルのテンプレ」を組み立てるために使う。
 */
const getExamTypeFields = async (id: string): Promise<ExamTypeFieldRow[]> => {
  const { data } = await axios.get<ExaminationType>(`/v1/masters/examination-types/${id}`);
  const items = data.items ?? [];
  return items
    .slice()
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
    .map((field): ExamTypeFieldRow => {
      const { refMin, refMax } = parseNormalRange(field.normal_value ?? "");
      return {
        id: Number(field.id),
        name: field.name ?? "",
        unit: field.unit ?? "",
        normalValue: field.normal_value ?? "",
        refMin,
        refMax,
        sortOrder: field.sort_order ?? 0,
      };
    });
};

export const useGetExamTypeFields = (examTypeId: string) => {
  return useQuery({
    queryKey: ["exam-type-fields", examTypeId],
    queryFn: () => getExamTypeFields(examTypeId),
    enabled: !!examTypeId,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
