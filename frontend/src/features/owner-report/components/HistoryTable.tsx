import type { ReactNode } from "react";
import { C } from "@/lib/design-tokens";

interface HistoryTableProps {
  headers: string[];
  children: ReactNode;
}

/**
 * #158 飼主レポート履歴セクション共通のテーブルシェル。
 * 親パネル本文がスクロールコンテナなので、thead を sticky にして内部スクロール中も列見出しを保持する。
 * 列見出しは不透明背景で行の上に重なる（背景が透けると可読性が落ちるため）。
 */
export function HistoryTable({ headers, children }: HistoryTableProps) {
  return (
    <table className="w-full text-sm">
      <thead>
        <tr className="text-left">
          {headers.map((h) => (
            <th
              key={h}
              // DESIGN.md ex-data-table-cell: header は canvas-soft 背景 + eyebrow 相当タイポグラフィ。
              // sticky ヘッダーのためスクロール中も下の行が透けないよう不透明背景を維持する（bgWhite → bgPage）。
              className={`sticky top-0 z-10 border-b py-1.5 pr-3 text-2xs font-semibold uppercase whitespace-nowrap ${C.borderLight} ${C.bgPage} ${C.text55}`}
            >
              {h}
            </th>
          ))}
        </tr>
      </thead>
      <tbody className={`divide-y ${C.divideDivider}`}>{children}</tbody>
    </table>
  );
}
