import type { ReactNode } from "react";
import { C } from "@/lib/design-tokens";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { ReportPanel } from "./ReportPanel";

interface ReportSectionProps {
  title: string;
  /** リソース閲覧権限。false の場合は本文を「閲覧権限がありません」に縮退する（ページ全体は落とさない）。 */
  canView: boolean;
  isLoading: boolean;
  isError: boolean;
  isEmpty: boolean;
  emptyMessage?: string;
  /** 件数バッジ（閲覧可・取得済み・非エラーのときのみヘッダーに表示）。 */
  count?: number;
  /**
   * SD-18: 取得 API が HISTORY_FETCH_LIMIT で打ち切っており、実件数がここに表示された
   * 件数より多い可能性がある場合 true。true のとき本文の先頭に打ち切り注記を表示する。
   */
  isTruncated?: boolean;
  children: ReactNode;
}

/**
 * #158 飼主レポートの各履歴セクション共通シェル。
 * 権限なし / ローディング / エラー / 空 の縮退表示を一元化し、いずれでもない場合のみ children を描画する。
 * 表示は密集ワークスペース用の ReportPanel（境界内スクロール）に委譲する。
 */
export function ReportSection({
  title,
  canView,
  isLoading,
  isError,
  isEmpty,
  emptyMessage = "履歴はありません",
  count,
  isTruncated = false,
  children,
}: ReportSectionProps) {
  const body = !canView ? (
    <p className={`text-sm ${C.text50}`} data-testid="section-no-permission">
      閲覧権限がありません
    </p>
  ) : isLoading ? (
    <p className={`text-sm ${C.text50}`} role="status" aria-live="polite">
      読み込み中...
    </p>
  ) : isError ? (
    <p className={`text-sm ${C.danger}`} role="alert">
      読み込みに失敗しました
    </p>
  ) : isEmpty ? (
    <div
      className={`flex h-full min-h-14 items-center justify-center rounded-lg border border-dashed px-4 py-3 text-center ${C.borderLight} ${C.bgPage30}`}
    >
      <p className={`text-sm ${C.text50}`}>{emptyMessage}</p>
    </div>
  ) : (
    <>
      {isTruncated ? (
        <p className={`mb-2 text-xs ${C.text50}`} data-testid="section-truncation-notice">
          直近{HISTORY_FETCH_LIMIT}件を表示しています
        </p>
      ) : null}
      {children}
    </>
  );

  const showCount = canView && !isLoading && !isError && typeof count === "number";

  return (
    <ReportPanel title={title} count={showCount ? count : undefined}>
      {body}
    </ReportPanel>
  );
}
