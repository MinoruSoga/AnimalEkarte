import type { ReactNode } from "react";
import { C } from "@/lib/design-tokens";

interface HistoryTableProps {
  headers: string[];
  children: ReactNode;
}

/** #158 飼主レポート履歴セクション共通のテーブルシェル（軽量・印刷向け）。 */
export function HistoryTable({ headers, children }: HistoryTableProps) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className={`text-left ${C.text50} border-b ${C.borderLight}`}>
            {headers.map((h) => (
              <th key={h} className="py-1.5 pr-3 font-medium whitespace-nowrap">
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className={`divide-y ${C.divideDivider}`}>{children}</tbody>
      </table>
    </div>
  );
}
