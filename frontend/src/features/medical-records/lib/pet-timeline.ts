import { paths } from "@/config/paths";
import { formatDate } from "@/lib/format/date";
import type { TrimmingUI } from "@/lib/transforms/trimming";
import type { InterviewHistoryItem } from "../types";

/**
 * NO32（スプシNo.32 案A・治療履歴1本化）:
 * 同一ペットの診療カルテ履歴とトリミング記録履歴を、表示層だけで
 * 1本の時系列タイムラインにマージする。データモデル・API は統合しない。
 * 行リンクは各記録の既存詳細画面（/medical-records/:id, /trimming/:id）を使う。
 */

/** カルテ id と衝突しないよう、トリミング行の React key / item.id に付ける prefix。 */
export const TRIMMING_TIMELINE_ID_PREFIX = "trimming-";

/**
 * TrimmingUI → InterviewHistoryItem 変換。
 * - type バッジは「トリミング（ステータス）」でカルテ行（確定済/作成中）と区別する
 * - copySource は設定しない（トリミング記録は問診複写の対象外）
 * - コース名は一覧 API で preload 済みの表示用名を使う
 */
export function toTrimmingTimelineItem(trimming: TrimmingUI): InterviewHistoryItem {
  return {
    id: `${TRIMMING_TIMELINE_ID_PREFIX}${trimming.id}`,
    date: formatDate(trimming.date),
    author: trimming.staff || "-",
    type: `トリミング（${trimming.status}）`,
    title: trimming.styleRequest || trimming.courseName || "トリミング",
    content: trimming.remarks || "（記録なし）",
    href: paths.trimming.detail.getHref(trimming.id),
    sortDate: trimming.date,
  };
}

/**
 * カルテ履歴とトリミング履歴を sortDate 降順（新しい順）で1本にマージする。
 * sortDate は ISO datetime / YYYY-MM-DD の文字列比較で時系列が保証される。
 * 同日は stable sort によりカルテ行が先・トリミング行が後になる。
 * sortDate 未設定（""）の行は末尾に回る。
 */
export function mergePetTimelineItems(
  medicalItems: InterviewHistoryItem[],
  trimmings: TrimmingUI[],
): InterviewHistoryItem[] {
  const merged = [...medicalItems, ...trimmings.map(toTrimmingTimelineItem)];
  return merged.toSorted((a, b) => (b.sortDate ?? "").localeCompare(a.sortDate ?? ""));
}
