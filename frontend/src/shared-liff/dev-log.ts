/**
 * FE-RC-077: liff/line-reserve は顧客端末（LINEアプリ内ブラウザ）で動くため、BE の
 * レスポンス本文をそのまま console.error すると本番でも顧客のブラウザコンソールに
 * 内部情報が露出する。import.meta.env.DEV のときだけ出力する薄いラッパーに統一する。
 *
 * フルロガー（送信・レベル分け等）は不要（対象は開発時デバッグ用の console 出力のみ）。
 */
export function devError(...args: unknown[]): void {
  if (import.meta.env.DEV) {
    console.error(...args);
  }
}
