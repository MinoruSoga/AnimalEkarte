/**
 * EMR-487: PDF の取り扱い。FE はレイアウトを描画しない（server renderer の PDF を表示・印刷するだけ）。
 * - 確認 preview は PDF の表示のみ（アプリ側に印刷ボタンを置かない）。
 * - 印刷は GET /vaccination-certificate-artifacts/{artifact_id}/content の PDF だけ。
 * blob: iframe での表示・印刷は、frontend の CSP の `frame-src 'self' blob:`（EMR-525）で許可している。
 */

export function base64ToPdfBlob(base64: string): Blob {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i);
  return new Blob([bytes], { type: "application/pdf" });
}

/** 組み込みビューアのツールバー（印刷・保存）を出さないための fragment。best effort でありコピー防止の保証ではない。 */
export const PDF_VIEWER_FRAGMENT = "#toolbar=0&navpanes=0";

/** iframe の読込が終わらない（CSP で拒否された等）場合に待ち続けない上限。 */
export const PRINT_LOAD_TIMEOUT_MS = 15_000;

/** 隠し iframe へ PDF を読み込み、そのまま印刷ダイアログを開く。完了後に後始末する。 */
export function printPdfBlob(blob: Blob, timeoutMs: number = PRINT_LOAD_TIMEOUT_MS): Promise<void> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(blob);
    const frame = document.createElement("iframe");
    frame.setAttribute("aria-hidden", "true");
    frame.tabIndex = -1;
    frame.style.cssText = "position:fixed;width:0;height:0;border:0;visibility:hidden";
    const cleanup = () => {
      window.setTimeout(() => {
        frame.remove();
        URL.revokeObjectURL(url);
      }, 60_000);
    };
    const timer = window.setTimeout(() => {
      frame.onload = null;
      cleanup();
      reject(new Error("PDF の読み込みがタイムアウトしました"));
    }, timeoutMs);
    frame.onload = () => {
      window.clearTimeout(timer);
      try {
        frame.contentWindow?.focus();
        frame.contentWindow?.print();
        resolve();
      } catch (error) {
        reject(error instanceof Error ? error : new Error("print failed"));
      } finally {
        cleanup();
      }
    };
    frame.onerror = () => {
      window.clearTimeout(timer);
      cleanup();
      reject(new Error("PDF を読み込めませんでした"));
    };
    frame.src = url;
    document.body.appendChild(frame);
  });
}
