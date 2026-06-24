import type { ReactNode } from "react";
import { createPortal } from "react-dom";

interface PrintPortalProps {
  /**
   * 印刷面ルート要素の data-testid。`@media print` で他要素を隠す際の
   * 「残す対象」セレクタにも使うため、ページ内で一意な定数を渡すこと。
   */
  testId: string;
  /** A4 の向き（横長の明細表は landscape を推奨）。既定は portrait。 */
  orientation?: "portrait" | "landscape";
  children: ReactNode;
}

/**
 * ブラウザ印刷／PDF出力（「PDFとして保存」）用の共通印刷ビュー基盤。
 *
 * #117 日次集計（DailyAccountingTab）で確立した
 * `createPortal` + `@media print` で他要素を非表示 + `@page { size: A4 }` の方式を
 * 単一コンポーネントに切り出し、#184 月次集計レポート・#153 レジ締めサマリーで再利用する。
 *
 * 画面上は `hidden` で常時非表示、印刷時のみ `display:block` で印刷面のみを出力する。
 * `document.body` 直下へ portal するため、印刷時は body 直下の他要素（アプリ本体・
 * トースト等）がまとめて `display:none` になり、操作 UI が出力に混入しない。
 */
export function PrintPortal({ testId, orientation = "portrait", children }: PrintPortalProps) {
  // SSR / 非ブラウザ環境では body 直下への portal は不可。SPA では常に存在する。
  if (typeof document === "undefined") return null;
  return createPortal(
    <div
      hidden
      className="bg-white"
      data-testid={testId}
      style={{ position: "fixed", inset: 0, zIndex: 9999, overflow: "auto", padding: "8mm" }}
    >
      <style type="text/css">
        {`
          @media print {
            [data-testid="${testId}"] { display: block !important; }
            body > :not([data-testid="${testId}"]) { display: none !important; }
            body { margin: 0; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
          }
          @page { size: A4 ${orientation === "landscape" ? "landscape" : "portrait"}; margin: 8mm; }
        `}
      </style>
      {children}
    </div>,
    document.body,
  );
}
