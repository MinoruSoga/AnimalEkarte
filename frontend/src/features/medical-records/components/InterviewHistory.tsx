// React/Framework
import { memo, useDeferredValue, useState, useMemo } from "react";

// External
import { Search, History } from "lucide-react";
import { Link } from "react-router";

// Internal
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { C, LAYOUT, ICON } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import { normalizedIncludes } from "@/lib/normalize-kana";
import { paths } from "@/config/paths";
import { formatRecordedAt } from "../lib/vitals-tab-table-model";
import { isMedicalHistoryRow } from "../lib/interview-weight-model";
import type { InterviewHistoryItem } from "../types";

interface InterviewHistoryProps {
  className?: string;
  historyItems: InterviewHistoryItem[];
  /** BUG-035: 確定済み/送信権限なし。検索・閲覧は読み取り操作のため残し、コピーのみ無効化する。 */
  isLocked?: boolean;
  /** EMR-356: カルテ履歴クエリの取得中（「履歴なし」と区別する）。 */
  isLoading?: boolean;
  /** EMR-356: カルテ履歴クエリの失敗（trimming 等の残行はそのまま表示する）。 */
  isError?: boolean;
  /** PR659-RR-1: 履歴クエリが打ち切られ未完（部分履歴 — 「該当なし」を断定しない）。 */
  historyIncomplete?: boolean;
  /** EMR-182: 行の コピー ボタン押下時のコールバック（複写は親が実施）。 */
  onCopyItem?: (item: InterviewHistoryItem) => void;
}

export const InterviewHistory = memo(function InterviewHistory({
  className,
  historyItems,
  isLocked = false,
  isLoading = false,
  isError = false,
  historyIncomplete = false,
  onCopyItem,
}: InterviewHistoryProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);

  // js-cache-function-results: API 由来の filter 結果は useMemo でキャッシュ
  const filteredItems = useMemo(() => {
    const compactTerm = deferredSearch.replace(/[/\-.]/g, "");
    return historyItems.filter(
      (item) =>
        normalizedIncludes(item.title, deferredSearch) ||
        normalizedIncludes(item.content, deferredSearch) ||
        normalizedIncludes(item.type, deferredSearch) ||
        normalizedIncludes(item.author, deferredSearch) ||
        normalizedIncludes(item.date, deferredSearch) ||
        (compactTerm !== "" && item.date.replace(/[/\-.]/g, "").includes(compactTerm)),
    );
  }, [historyItems, deferredSearch]);

  return (
    <div
      className={`flex flex-col border ${C.borderMedium} ${C.bgWhite} rounded-lg min-h-0 max-h-[600px] ${className ?? ""}`}
    >
      <div
        className={`p-3 border-b ${C.borderLight} ${C.bgPage} flex items-center justify-between min-h-12 shrink-0 gap-2`}
      >
        <div className="flex items-center gap-2 min-w-0">
          <History className={`${ICON.action} ${C.text}`} />
          {/* NO32: カルテ + トリミング記録の統合タイムライン */}
          <h3 className={`text-sm font-bold ${C.text}`}>治療履歴</h3>
          <p className={`text-sm ${C.text60} truncate`}>全文は詳細で確認できます</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="relative">
            <Search
              className={`absolute left-2.5 top-1/2 -translate-y-1/2 ${ICON.action} ${C.text60}`}
            />
            <label htmlFor="medical-record-history-search" className="sr-only">
              過去の履歴を検索
            </label>
            <Input
              id="medical-record-history-search"
              name="medicalRecordHistorySearch"
              placeholder="検索..."
              className={`${LAYOUT.touch.md} w-full max-w-[12rem] pl-9 text-sm ${C.bgWhite} ${C.borderMedium}`}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") e.preventDefault();
              }}
            />
          </div>
        </div>
      </div>

      <ScrollArea className="flex-1 min-h-0">
        <div className={`divide-y ${C.borderDivider}`}>
          {/* EMR-356 C3: 取得中/失敗は「履歴なし」と区別して表示する */}
          {isLoading ? <p className={`p-3 text-sm ${C.text60}`}>履歴を読み込み中です</p> : null}
          {isError && !isLoading ? (
            <p className={`p-3 text-sm ${C.text60}`}>履歴を取得できませんでした</p>
          ) : null}
          {/* PR659-RR-1: 部分履歴（打ち切り）では読み込み済みの一致行があっても、
              検索対象が全履歴でないことを示す */}
          {historyIncomplete && !isLoading ? (
            <p className={`p-3 text-sm ${C.text60}`}>一部の履歴のみを表示しています</p>
          ) : null}
          {filteredItems.map((item) => (
            <div key={item.id} className="flex items-stretch">
              <Link
                to={item.href ?? paths.medicalRecords.detail.getHref(item.id)}
                className={`block min-w-0 flex-1 p-3 transition-colors ${C.hoverBgPageHalf}`}
              >
                <div className="flex items-start justify-between mb-1">
                  <div className="flex items-center gap-2">
                    <span className={`font-mono text-sm font-semibold ${C.text}`}>{item.date}</span>
                    <Badge variant="secondary" className="text-sm px-2">
                      {item.type}
                    </Badge>
                  </div>
                  <span className={`text-sm ${C.text60}`}>{item.author}</span>
                </div>
                <h4 className={`text-sm font-bold ${C.text} mb-1`}>{item.title}</h4>
                <p className={`text-sm ${C.text80} leading-snug whitespace-pre-wrap line-clamp-2`}>
                  {item.content}
                </p>
                {/* EMR-356: このカルテ行に相関する体重のみ行内に表示する。
                    医療行で投影フィールド自体がない場合は「確認不可」、
                    明示的な空（[]）は「測定なし」と区別する。
                    trimming 等の医療行でない行には通知を付けない。 */}
                {item.weightMeasurements !== undefined ? (
                  item.weightMeasurements.length > 0 ? (
                    <p className={`text-sm ${C.text60} mt-1 flex flex-wrap gap-x-2 gap-y-0.5`}>
                      {item.weightMeasurements.map((m) => (
                        <span key={m.id} className="whitespace-nowrap">
                          体重 {m.value}
                          {m.unit}（{formatRecordedAt(m.measuredAt)}）
                        </span>
                      ))}
                    </p>
                  ) : isMedicalHistoryRow(item) ? (
                    <p className={`text-sm ${C.text60} mt-1`}>体重測定の記録なし</p>
                  ) : null
                ) : isMedicalHistoryRow(item) ? (
                  <p className={`text-sm ${C.text60} mt-1`}>体重履歴を確認できません</p>
                ) : null}
              </Link>
              {item.copySource ? (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="m-3 ml-0 self-center"
                  disabled={isLocked}
                  onClick={() => onCopyItem?.(item)}
                >
                  コピー
                </Button>
              ) : null}
            </div>
          ))}
          {/* 取得中/失敗の空は「履歴なし」の確定とは区別する。
              部分履歴では打ち切られた後続に該当行があり得るため断定しない */}
          {filteredItems.length === 0 && !isLoading && !isError ? (
            historyIncomplete ? (
              <p className={`p-3 text-sm ${C.text60}`}>
                読み込まれた履歴に該当するものはありません
              </p>
            ) : (
              <EmptyState message="該当する履歴はありません" />
            )
          ) : null}
        </div>
      </ScrollArea>
    </div>
  );
});
