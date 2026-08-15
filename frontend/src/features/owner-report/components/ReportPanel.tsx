import type { ReactNode } from "react";
import { useId } from "react";
import { C } from "@/lib/design-tokens";

interface ReportPanelProps {
  title: string;
  /** ヘッダー右に表示する件数（業務ツールの密度シグナル）。未指定なら非表示。 */
  count?: number;
  children: ReactNode;
}

/**
 * #158 飼主レポートの密集ワークスペース用パネルシェル。
 *
 * 高さは親グリッドのセル（minmax(0,1fr)）から受け取り、本文だけを内部スクロールさせる。
 * これによりデスクトップではページ全体をスクロールさせず、各履歴を独立した境界内でスクロールできる。
 * - スクロール本文を見出しで命名した region ランドマークにする（aria-labelledby）。
 *   region をフォーカス可能なスクロール本文側に置くことで、二重ランドマークを避けつつ
 *   キーボードフォーカス時に AT が「○○」領域として読み上げられるようにする（WCAG 4.1.2）。
 * - 本文は min-h-0 + overflow-auto + tabindex=0 でキーボードスクロール可能にする。
 * - モバイル（行が auto 高さ）では本文が内容に合わせて伸び、内部スクロールは発生せずページスクロールに委ねる。
 */
export function ReportPanel({ title, count, children }: ReportPanelProps) {
  const headingId = useId();
  return (
    <section
      className={`flex min-h-0 min-w-0 flex-col overflow-hidden rounded-lg border ${C.borderLight} ${C.bgWhite}`}
    >
      <div
        className={`flex shrink-0 items-baseline justify-between gap-2 border-b px-3 py-2 ${C.borderLight} ${C.bgPage30} [@media(max-height:600px)]:py-0`}
      >
        <h2
          className={`truncate text-sm font-medium ${C.text} [@media(max-height:600px)]:text-xs`}
          id={headingId}
        >
          {title}
        </h2>
        {typeof count === "number" ? (
          <span className={`shrink-0 text-xs tabular-nums ${C.text50}`}>{count}</span>
        ) : null}
      </div>
      <div
        role="region"
        aria-labelledby={headingId}
        className="min-h-0 flex-1 overflow-auto px-3 py-2"
        tabIndex={0}
      >
        {children}
      </div>
    </section>
  );
}
