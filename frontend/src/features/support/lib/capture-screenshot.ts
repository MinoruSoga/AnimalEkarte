/**
 * capture-screenshot.ts — 現在画面のスクリーンショット取得
 *
 * html2canvas-pro を dynamic import して初期バンドルに含めない。
 * Tailwind v4 の oklch 色は原版 html2canvas では解析できないため pro 版を使う。
 * サポートウィジェット自身は `data-html2canvas-ignore` でキャプチャ対象外になる。
 *
 * 失敗しても報告送信はブロックしない（null を返す）。
 */
export async function capturePageScreenshot(): Promise<File | null> {
  try {
    const { default: html2canvas } = await import("html2canvas-pro");
    const canvas = await html2canvas(document.body, {
      logging: false,
      useCORS: true,
    });
    const blob = await new Promise<Blob | null>((resolve) => {
      canvas.toBlob(resolve, "image/png");
    });
    if (blob === null) {
      return null;
    }
    const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
    return new File([blob], `screenshot-${timestamp}.png`, { type: "image/png" });
  } catch {
    return null;
  }
}
