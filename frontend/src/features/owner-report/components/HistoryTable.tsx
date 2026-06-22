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
              className={`sticky top-0 z-10 border-b py-1.5 pr-3 font-medium whitespace-nowrap ${C.borderLight} ${C.bgWhite} ${C.text50}`}
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
