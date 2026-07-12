import { useEffect, useRef, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { Z } from "@/lib/design-tokens";

interface PrintPortalProps {
  /**
   * 印刷面ルート要素の data-testid。テスト・E2E のアンカー用途。
   *
   * #187 以降、印刷除外ルールは testId ではなく全ポータル共通の固定属性
   * `data-print-portal` をキーにするため、ページ内で一意である必要はあるが
   * 印刷の表示制御には関与しない（per-instance セレクタ依存を排除済み）。
   */
  testId: string;
  /** A4 の向き（横長の明細表は landscape を推奨）。既定は portrait。 */
  orientation?: "portrait" | "landscape";
  /**
   * 印刷時にこの印刷面を表示するか。既定 `true`。
   *
   * 同一ページに複数の `PrintPortal` を同居させ、印刷対象を 1 面へ限定したい場合に
   * 対象だけ `active`、他を `active={false}` とする。未指定（既定）なら全ポータルが
   * 表示されるため白紙化しない（#187 の主対象ケース）。ただし同居する全ポータルを
   * `active={false}` にすると印刷面が 0 になり白紙化するため、印刷時は最低 1 面を
   * active に保つこと（呼び出し側責務）。
   */
  active?: boolean;
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
 *
 * #187: 印刷除外ルールの除外キーを per-instance な testId から全ポータル共通の固定属性
 * `data-print-portal` へ変更。これにより複数ポータル同居時に互いの `:not(...)` が相手を
 * 隠して specificity 競合で全面白紙化する不具合を構造的に解消した。除外キーが共通のため、
 * 各ポータルの注入する `@media print` ルールは同一テキストになり、同居しても打ち消し合わない。
 */
export function PrintPortal({
  testId,
  orientation = "portrait",
  active = true,
  children,
}: PrintPortalProps) {
  const elementRef = useRef<HTMLDivElement>(null);

  // FD6 (FE-refactor.md): 同居する全 PrintPortal が active=false だと印刷面が
  // 0 件になり白紙化する（呼び出し側責務、実行時のガードは無し）。dev ビルド限定で
  // 印刷実行時にこの状態を検出し console.warn する（本番挙動は不変・挙動保存）。
  useEffect(() => {
    if (!import.meta.env.DEV) return;

    function warnIfAllPortalsInactive() {
      const portals = document.querySelectorAll<HTMLElement>("[data-print-portal]");
      if (portals.length === 0) return;
      // 複数 PrintPortal が同時に mount されている場合、DOM 順で最初のものだけが
      // チェックを実行する（各インスタンスの effect が重複して警告しないため）。
      if (portals[0] !== elementRef.current) return;

      const hasActivePortal = Array.from(portals).some(
        (el) => el.getAttribute("data-print-active") === "true",
      );
      if (!hasActivePortal) {
        console.warn(
          "[PrintPortal] 同居する全ての印刷ポータルが active=false です。印刷面が白紙化します。" +
            "呼び出し側で最低1つを active=true にしてください（#187）。",
        );
      }
    }

    window.addEventListener("beforeprint", warnIfAllPortalsInactive);
    return () => window.removeEventListener("beforeprint", warnIfAllPortalsInactive);
  }, []);

  // SSR / 非ブラウザ環境では body 直下への portal は不可。SPA では常に存在する。
  if (typeof document === "undefined") return null;
  return createPortal(
    <div
      ref={elementRef}
      hidden
      className="bg-white"
      data-testid={testId}
      data-print-portal=""
      data-print-active={active ? "true" : "false"}
      style={{ position: "fixed", inset: 0, zIndex: Z.overlay, overflow: "auto", padding: "8mm" }}
    >
      <style type="text/css">
        {`
          @media print {
            /* アプリ本体・トースト等（印刷ポータル以外の body 直下要素）を隠す。
               除外キーは全ポータル共通の固定属性 data-print-portal。per-instance な
               testId に依存しないため、複数ポータル同居時も互いを隠さない（#187）。*/
            body > :not([data-print-portal]) { display: none !important; }
            /* 印刷ポータルは既定で表示。active 未指定（既定）の同居では全ポータルが残り白紙化しない。
               全ポータルを active=false にした場合のみ印刷面が 0 になり得る（呼び出し側で 1 面を active に保つ）。*/
            [data-print-portal] { display: block !important; }
            /* 明示的に非アクティブ指定されたポータルのみ隠す
               （同居時に印刷面を 1 件へ限定する手段。specificity を 1 段上げて表示ルールに勝たせる）。*/
            [data-print-portal][data-print-active="false"] { display: none !important; }
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
