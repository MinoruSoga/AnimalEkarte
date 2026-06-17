import type { ReactNode } from "react";
import { C } from "@/lib/design-tokens";

interface ReportSectionProps {
  title: string;
  /** リソース閲覧権限。false の場合は本文を「閲覧権限がありません」に縮退する（ページ全体は落とさない）。 */
  canView: boolean;
  isLoading: boolean;
  isError: boolean;
  isEmpty: boolean;
  emptyMessage?: string;
  children: ReactNode;
}

/**
 * #158 飼主レポートの各セクション共通シェル。
 * 権限なし / ローディング / エラー / 空 の縮退表示を一元化し、いずれでもない場合のみ children を描画する。
 */
export function ReportSection({
  title,
  canView,
  isLoading,
  isError,
  isEmpty,
  emptyMessage = "履歴はありません",
  children,
}: ReportSectionProps) {
  const body = !canView ? (
    <p className={`text-sm ${C.text50}`} data-testid="section-no-permission">
      閲覧権限がありません
    </p>
  ) : isLoading ? (
    <p className={`text-sm ${C.text50}`}>読み込み中...</p>
  ) : isError ? (
    <p className={`text-sm ${C.danger}`}>読み込みに失敗しました</p>
  ) : isEmpty ? (
    <p className={`text-sm ${C.text50}`}>{emptyMessage}</p>
  ) : (
    children
  );

  return (
    <section className={`rounded-lg border ${C.borderLight} ${C.bgWhite} p-4`}>
      <h2 className={`text-sm font-semibold ${C.text} mb-3`}>{title}</h2>
      {body}
    </section>
  );
}
