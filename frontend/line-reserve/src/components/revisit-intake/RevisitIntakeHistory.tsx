// EMR-474: 回答履歴（read-only）。current（この予約の対象）と previous（以前の対象）を分けて表示し、
// previous を入力欄へコピーする操作は設けない。next_cursor があれば「さらに表示」で継続取得する。
import type { RevisitIntakeHistoryEntry } from "../../types/models";
import { RevisitIntakeAnswerView } from "./RevisitIntakeAnswerView";
import { formatRevisitIntakeDateTime } from "../../lib/revisit-intake-options";

interface RevisitIntakeHistoryProps {
  history: RevisitIntakeHistoryEntry[];
  hasMore: boolean;
  loadingMore: boolean;
  loadMoreError: string | null;
  /** 429 / 503 の Retry-After 期限中の待ち秒。期限まで「さらに表示」を無効にする */
  loadMoreRetryAfterSeconds: number | null;
  onLoadMore: () => void;
}

function HistoryList({ entries }: { entries: RevisitIntakeHistoryEntry[] }) {
  return (
    <ul className="space-y-3">
      {entries.map((entry) => (
        <li
          key={`${entry.target.target_id}-${entry.answer.answer_id}`}
          className="rounded-xl border border-noah-border bg-white p-4"
        >
          <p className="text-xs text-noah-text-sub">
            送信日時 {formatRevisitIntakeDateTime(entry.answer.received_at)}
          </p>
          <RevisitIntakeAnswerView answers={entry.answer.answers} />
        </li>
      ))}
    </ul>
  );
}

export function RevisitIntakeHistory({
  history,
  hasMore,
  loadingMore,
  loadMoreError,
  loadMoreRetryAfterSeconds,
  onLoadMore,
}: RevisitIntakeHistoryProps) {
  if (history.length === 0 && !hasMore) return null;
  const current = history.filter((h) => h.target_relation === "current");
  const previous = history.filter((h) => h.target_relation === "previous");

  return (
    <section aria-labelledby="revisit-intake-history-heading" className="space-y-4">
      <h2 id="revisit-intake-history-heading" className="text-base font-bold text-noah-teal-dark">
        回答履歴
      </h2>
      {current.length > 0 ? (
        <div className="space-y-2">
          <h3 className="text-sm font-semibold text-noah-text">この予約への回答</h3>
          <HistoryList entries={current} />
        </div>
      ) : null}
      {previous.length > 0 ? (
        <div className="space-y-2">
          <h3 className="text-sm font-semibold text-noah-text">以前の対象への回答（参照のみ）</h3>
          <p className="text-xs text-noah-text-sub">
            予約の取り直しや対象ペットの変更前の回答です。今回の回答には引き継がれません。
          </p>
          <HistoryList entries={previous} />
        </div>
      ) : null}
      {loadMoreError !== null ? (
        <p role="alert" className="text-sm text-noah-danger-text">
          {loadMoreError}
        </p>
      ) : null}
      {hasMore && loadMoreRetryAfterSeconds !== null ? (
        <p className="text-sm text-noah-text-sub">
          {`混み合っています。${loadMoreRetryAfterSeconds}秒ほど待ってから、もう一度お試しください。`}
        </p>
      ) : null}
      {hasMore ? (
        <button
          type="button"
          onClick={onLoadMore}
          disabled={loadingMore || loadMoreRetryAfterSeconds !== null}
          aria-busy={loadingMore}
          className="min-h-11 w-full rounded-lg border border-noah-border-input bg-white px-4 text-sm font-semibold text-noah-teal-dark disabled:opacity-50"
        >
          {loadingMore ? "読み込み中..." : "さらに表示"}
        </button>
      ) : null}
    </section>
  );
}
